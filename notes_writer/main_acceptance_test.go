//go:build acceptance
// +build acceptance

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/akijowski/tweek-2021-sam/internal/schema"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/testcontainers/testcontainers-go"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const testTableName = "notes"

var (
	dockerComposeIdentifier string
	projectRootPath         string
	dockerComposeFilePath   string
	dynamoClient            *dynamodb.Client
	lambdaClient            *lambda.Client
)

var dockerComposeFile = flag.String("composeFile", "docker-compose-acceptance.yml", "the location of the docker-compose.yml file from the root project directory")
var defaultRegion = flag.String("region", "us-east-1", "the AWS region to pretend we are using for the tests")
var dynamoDBAPIURL = flag.String("dynamoURL", "http://localhost:4566", "the URL to use for the DynamoDB API")
var lambdaAPIURL = flag.String("lambdaURL", "http://localhost:3001", "the URL to use for the Lambda API")

var localstackCredentials = aws.Credentials{AccessKeyID: "test", SecretAccessKey: "test"}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func TestHandler_AcceptanceTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test")
	}
	cases := map[string]struct {
		note                *schema.Note
		expectedAPIResponse events.APIGatewayProxyResponse
	}{
		"valid request saves note": {
			note: &schema.Note{
				Owner:   "test-owner",
				Title:   "test-title",
				Message: "test-message",
			},
			expectedAPIResponse: events.APIGatewayProxyResponse{
				StatusCode: http.StatusCreated,
				Headers: map[string]string{
					"Location": "/test-owner",
				},
				Body: "",
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			if err := createTable(ctx, dynamoClient, testTableName); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			defer func() {
				if err := deleteTable(ctx, dynamoClient, testTableName); err != nil {
					t.Logf("unexpected error deleting table: %s", err)
				}
			}()
			payload, err := wrapNoteRequestWithAPIGateway(tt.note)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			out, err := lambdaClient.Invoke(ctx, &lambda.InvokeInput{
				FunctionName: aws.String("NotesWriterFunction"),
				Payload:      payload,
			})
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if out.FunctionError != nil {
				printContainerLogs("aws_sam")
				t.Fatalf("unexpected function error: %s", aws.ToString(out.FunctionError))
			}
			var actual events.APIGatewayProxyResponse
			err = json.Unmarshal(out.Payload, &actual)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !reflect.DeepEqual(actual, tt.expectedAPIResponse) {
				t.Fatalf("incorrect API response: wanted %+v got %+v", tt.expectedAPIResponse, actual)
			}
		})
	}
}

func setup() {
	if err := buildFilePaths(); err != nil {
		panic(err)
	}
	if err := startDockerCompose(); err != nil {
		panic(err)
	}
	if err := initAWSClients(); err != nil {
		panic(err)
	}
}

func teardown() {
	if err := stopDockerCompose(); err != nil {
		panic(err)
	}
}

func buildFilePaths() error {
	root, err := filepath.Abs("../")
	if err != nil {
		return err
	}
	projectRootPath = root
	dockerComposeFilePath = filepath.Join(projectRootPath, *dockerComposeFile)
	return nil
}

func startDockerCompose() error {
	dockerComposeIdentifier = fmt.Sprintf("notes-writer-test-%d", rand.Int31())
	log.Printf("starting docker-compose at %s (ID: %s)\n", dockerComposeFilePath, dockerComposeIdentifier)
	compose := testcontainers.NewLocalDockerCompose([]string{dockerComposeFilePath}, dockerComposeIdentifier)
	execError := compose.
		WithCommand([]string{"up", "--build", "-d"}).
		WithEnv(map[string]string{"PWD": projectRootPath}).
		Invoke()
	return execError.Error
}

func stopDockerCompose() error {
	log.Printf("stopping docker-compose at %s (ID: %s)\n", dockerComposeFilePath, dockerComposeIdentifier)
	compose := testcontainers.NewLocalDockerCompose([]string{dockerComposeFilePath}, dockerComposeIdentifier)
	return compose.Down().Error
}

func printContainerLogs(service string) {
	log.Printf("collecting container logs for %q\n", service)
	compose := testcontainers.NewLocalDockerCompose([]string{dockerComposeFilePath}, dockerComposeIdentifier)
	compose.WithCommand([]string{"logs", "--tail", "15", service}).Invoke()
}

func initAWSClients() error {
	var err error
	ctx := context.Background()
	dynamoConfig, err := newLocalstackDynamoConfig(ctx, *defaultRegion, *dynamoDBAPIURL, localstackCredentials)
	if err != nil {
		return err
	}
	dynamoClient = dynamodb.NewFromConfig(dynamoConfig)
	lambdaConfig, err := newLambdaConfig(ctx, *defaultRegion, *lambdaAPIURL)
	if err != nil {
		return err
	}
	lambdaClient = lambda.NewFromConfig(lambdaConfig, func(opts *lambda.Options) {
		opts.Credentials = nil
	})
	return nil
}

func newLocalstackDynamoConfig(ctx context.Context, defaultRegion, endpointURL string, credentials aws.Credentials) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx,
		config.WithDefaultRegion(defaultRegion),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(
			func(ctx context.Context) (aws.Credentials, error) {
				return credentials, nil
			})),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{PartitionID: "aws", URL: endpointURL}, nil
			})),
	)
}

func newLambdaConfig(ctx context.Context, defaultRegion, endpointURL string) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx,
		config.WithDefaultRegion(defaultRegion),
		config.WithHTTPClient(awshttp.NewBuildableClient().
			WithTransportOptions(func(tr *http.Transport) {
				tr.TLSClientConfig.InsecureSkipVerify = true
			})),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{PartitionID: "aws", URL: endpointURL}, nil
			})),
	)
}

func createTable(ctx context.Context, dynamoClient *dynamodb.Client, tableName string) error {
	log.Printf("creating table: %q", tableName)
	_, err := dynamoClient.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName:             aws.String(tableName),
		KeySchema:             schema.NotesKeySchema,
		AttributeDefinitions:  schema.NotesAttributeDefinitions,
		ProvisionedThroughput: schema.NotesProvisionedThroughput,
	})
	if err != nil {
		return err
	}
	tableListOut, err := dynamoClient.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		return err
	}
	log.Printf("tables: %v\n", tableListOut.TableNames)
	return nil
}

func deleteTable(ctx context.Context, dynamoClient *dynamodb.Client, tableName string) error {
	log.Printf("deleting table %q", tableName)
	_, err := dynamoClient.DeleteTable(ctx, &dynamodb.DeleteTableInput{TableName: aws.String(tableName)})
	return err
}

func wrapNoteRequestWithAPIGateway(note *schema.Note) ([]byte, error) {
	body, err := json.Marshal(note)
	if err != nil {
		return nil, err
	}
	gatewayReq := &events.APIGatewayProxyRequest{Body: string(body)}
	return json.Marshal(gatewayReq)
}
