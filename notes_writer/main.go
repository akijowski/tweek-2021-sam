package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/akijowski/tweek-2021-sam/internal/ddb"
	"github.com/akijowski/tweek-2021-sam/internal/schema"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"
	"log"
	"net/http"
	"os"
)

var api ddb.DynamoUpdateItemAPI

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	lc, _ := lambdacontext.FromContext(ctx)
	log.Printf("CONTEXT: %+v", lc)

	primaryKey, err := handleRequest(ctx, request)
	if err != nil {
		log.Printf("error adding note: %s", err)
		var derr *ddb.DynamoDBError
		if errors.As(err, &derr) {
			log.Printf("client error: %s", derr.ClientMessage)
			return errorResponse(http.StatusBadGateway, lc.AwsRequestID, derr.Error()), nil
		} else {
			return errorResponse(http.StatusInternalServerError, lc.AwsRequestID, err.Error()), nil
		}
	}

	return events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Location": fmt.Sprintf("/%s", primaryKey),
		},
		StatusCode: http.StatusCreated,
	}, nil
}

func main() {
	lambda.Start(handler)
}

func init() {
	api = initDynamoClient()
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (string, error) {
	var creationRequest *schema.Note
	if err := json.Unmarshal([]byte(request.Body), &creationRequest); err != nil {
		log.Printf("error unmarshalling request: %s\n", err)
		return "", err
	}
	tableName := os.Getenv("WRITER_TABLE_NAME")
	return ddb.AddNote(ctx, api, tableName, creationRequest)
}

func errorResponse(statusCode int, requestID, message string) events.APIGatewayProxyResponse {
	e := &schema.LambdaHandlerError{
		StatusCode: statusCode,
		RequestID:  requestID,
		Message:    message,
	}
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       e.String(),
	}
}

func initDynamoClient() *dynamodb.Client {
	var optionsFuncs []func(options *config.LoadOptions) error
	if dynamoUri := os.Getenv("DYNAMODB_API_URL_OVERRIDE"); dynamoUri != "" {
		log.Printf("Overriding default DynamoDB API URI: %s", dynamoUri)
		// We are going to assume that if you are override the URL, you are likely trying to connect to a local service
		optionsFuncs = localstackConfigurationOptions(dynamoUri)
	}
	cfg, err := config.LoadDefaultConfig(context.Background(), optionsFuncs...)
	if err != nil {
		panic(err)
	}
	// Instrumenting AWS SDK v2
	awsv2.AWSV2Instrumentor(&cfg.APIOptions)
	return dynamodb.NewFromConfig(cfg)
}

func localstackConfigurationOptions(url string) []func(options *config.LoadOptions) error {
	return []func(options *config.LoadOptions) error{
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					PartitionID: "aws",
					URL:         url,
				}, nil
			})),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(
			func(ctx context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     "test",
					SecretAccessKey: "test",
				}, nil
			})),
		config.WithDefaultRegion("us-east-1"),
	}
}
