//go:build integration
// +build integration

package ddb

import (
	"context"
	"flag"
	"fmt"
	"github.com/akijowski/tweek-2021-sam/internal/schema"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/testcontainers/testcontainers-go"
	"math/rand"
	"os"
	"testing"
	"time"
)

var (
	dynamoClient    *dynamodb.Client
	dynamoContainer *dynamoDBContainer
)

var tableName = flag.String("tableName", "notes", "the table name to be created and used for the tests")
var defaultRegion = flag.String("region", "us-east-1", "the AWS region to pretend we are using for the tests")

var dbCredentials = aws.Credentials{AccessKeyID: "test", SecretAccessKey: "test"}

type dynamoDBContainer struct {
	testcontainers.Container
	URL string
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func TestAddNote_IntegrationTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	cases := map[string]struct {
		note      *schema.Note
		tableName string
		expectErr bool
	}{
		"valid note returns successfully": {
			note: &schema.Note{
				Owner:   "owner",
				Title:   "title-here",
				Message: "message goes here",
			},
			tableName: fmt.Sprintf("add-%s-%d", *tableName, rand.Int31()),
		},
		"missing table name returns error": {
			note:      &schema.Note{},
			expectErr: true,
		},
		"invalid note returns error": {
			note:      &schema.Note{Owner: "invalid"},
			tableName: fmt.Sprintf("add-%s-%d", *tableName, rand.Int31()),
			expectErr: true,
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			if tt.tableName != "" {
				if err := createTable(ctx, tt.tableName); err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				defer func() {
					if err := deleteTable(ctx, tt.tableName); err != nil {
						t.Fatalf("unexpected error: %s", err)
					}
				}()
			}
			primaryKey, err := AddNote(context.Background(), dynamoClient, tt.tableName, tt.note)
			if err != nil {
				if !tt.expectErr {
					t.Fatalf("unexpected error: %s", err)
				}
			} else {
				if primaryKey != tt.note.Owner {
					t.Fatalf("invalid value: wanted %q got %q", tt.note.Owner, primaryKey)
				}
			}
		})
	}
}

func TestScan_IntegrationTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	cases := map[string]struct {
		expectedNotes []schema.Note
		tableName     string
		expectErr     bool
	}{
		"valid input returns correctly": {
			expectedNotes: []schema.Note{
				{
					Owner:   "owner",
					Title:   "title",
					Message: "message",
				},
				{
					Owner:   "owner1",
					Title:   "title1",
					Message: "message1",
				},
			},
			tableName: fmt.Sprintf("scan-%s-%d", *tableName, rand.Int31()),
		},
		"missing table name return error": {
			expectErr: true,
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			if tt.tableName != "" {
				// setup
				if err := createTable(ctx, tt.tableName); err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				defer func() {
					if err := deleteTable(ctx, tt.tableName); err != nil {
						t.Fatalf("unexpected error: %s", err)
					}
				}()
				if err := saveToTable(ctx, tt.tableName, tt.expectedNotes); err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
			}

			actualNotes, err := Scan(ctx, dynamoClient, tt.tableName)
			if err != nil {
				if !tt.expectErr {
					t.Fatalf("unexpected error: %s", err)
				}
			} else {
				if len(actualNotes) != len(tt.expectedNotes) {
					t.Fatalf("incorrect response length: wanted %d got %d", len(tt.expectedNotes), len(actualNotes))
				}
			}
		})
	}
}

func TestFindNotesByOwner_IntegrationTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	cases := map[string]struct {
		owner         string
		seed          []schema.Note
		expectedNotes []schema.Note
		tableName     string
		expectErr     bool
	}{
		"valid input returns correctly": {
			owner: "owner",
			seed: []schema.Note{
				{
					Owner:   "owner",
					Title:   "title",
					Message: "message",
				},
				{
					Owner:   "not",
					Title:   "present",
					Message: "in test",
				},
			},
			expectedNotes: []schema.Note{
				{
					Owner:   "owner",
					Title:   "title",
					Message: "message",
				},
			},
			tableName: fmt.Sprintf("query-%s-%d", *tableName, rand.Int31()),
		},
		"missing table name returns error": {
			owner:     "owner",
			expectErr: true,
		},
		"missing owner returns error": {
			seed:      []schema.Note{{Owner: "owner", Title: "title", Message: "msg"}},
			tableName: fmt.Sprintf("query-%s-%d", *tableName, rand.Int31()),
			expectErr: true,
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			if tt.tableName != "" {
				// setup
				if err := createTable(ctx, tt.tableName); err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				defer func() {
					if err := deleteTable(ctx, tt.tableName); err != nil {
						t.Fatalf("unexpected error: %s", err)
					}
				}()
				if err := saveToTable(ctx, tt.tableName, tt.seed); err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
			}

			actualNotes, err := FindNotesByOwner(ctx, dynamoClient, tt.tableName, tt.owner)
			if err != nil {
				if !tt.expectErr {
					t.Fatalf("unexpected error: %s", err)
				}
			} else {
				if len(actualNotes) != len(tt.expectedNotes) {
					t.Fatalf("incorrect response length: wanted %d got %d", len(tt.expectedNotes), len(actualNotes))
				}
				for _, an := range actualNotes {
					if an.Owner != tt.owner {
						t.Fatalf("incorrect owner: wanted %q got %q", tt.owner, an.Owner)
					}
				}
			}
		})
	}
}

func setup() {
	ctx := context.Background()
	err := initDBContainer(ctx)
	if err != nil {
		panic(err)
	}
	cfg, err := createDynamoConfig(ctx)
	if err != nil {
		panic(err)
	}
	dynamoClient = dynamodb.NewFromConfig(cfg)
}

func teardown() {
	ctx := context.Background()
	state, err := dynamoContainer.State(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("dynamoContainer is %q\n", state.Status)
	if state.Running {
		fmt.Printf("removing container\n")
		if err = dynamoContainer.Terminate(ctx); err != nil {
			fmt.Printf("error: %s", err)
		}
	}
}

func createDynamoConfig(ctx context.Context) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx,
		config.WithDefaultRegion(*defaultRegion),
		config.WithCredentialsProvider(
			aws.CredentialsProviderFunc(
				func(ctx context.Context) (aws.Credentials, error) {
					return dbCredentials, nil
				})),
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{PartitionID: "aws", URL: dynamoContainer.URL}, nil
				})),
	)
}

func createTable(ctx context.Context, tableName string) error {
	fmt.Printf("creating table: %s\n", tableName)
	if _, err := dynamoClient.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName:             aws.String(tableName),
		KeySchema:             schema.NotesKeySchema,
		AttributeDefinitions:  schema.NotesAttributeDefinitions,
		ProvisionedThroughput: schema.NotesProvisionedThroughput,
	}); err != nil {
		return err
	}
	fmt.Printf("waiting for %q to become ACTIVE\n", tableName)
	waiter := dynamodb.NewTableExistsWaiter(dynamoClient, func(opts *dynamodb.TableExistsWaiterOptions) {
		opts.LogWaitAttempts = true
	})
	return waiter.Wait(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(tableName)}, 3*time.Minute)
}

func deleteTable(ctx context.Context, tableName string) error {
	fmt.Printf("deleting table: %s\n", tableName)
	if _, err := dynamoClient.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	}); err != nil {
		return err
	}
	fmt.Printf("waiting for table %q to be removed\n", tableName)
	waiter := dynamodb.NewTableNotExistsWaiter(dynamoClient, func(opts *dynamodb.TableNotExistsWaiterOptions) {
		opts.LogWaitAttempts = true
	})
	return waiter.Wait(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(tableName)}, 3*time.Minute)
}

func saveToTable(ctx context.Context, tableName string, notes []schema.Note) error {
	var writes []types.WriteRequest
	for _, n := range notes {
		av, err := attributevalue.MarshalMap(n)
		if err != nil {
			return err
		}
		writes = append(writes, types.WriteRequest{PutRequest: &types.PutRequest{Item: av}})
	}
	_, err := dynamoClient.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			tableName: writes,
		},
	})
	return err
}

func initDBContainer(ctx context.Context) error {
	req := testcontainers.ContainerRequest{
		Image:        "amazon/dynamodb-local:1.17.0",
		ExposedPorts: []string{"8000"},
		Cmd:          []string{"-jar", "DynamoDBLocal.jar", "-inMemory"},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return err
	}
	ip, err := container.Host(ctx)
	if err != nil {
		return err
	}
	mappedPort, err := container.MappedPort(ctx, "8000")
	if err != nil {
		return err
	}
	uri := fmt.Sprintf("http://%s:%s", ip, mappedPort)
	state, err := container.State(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Container state: %q at %s\n", state.Status, uri)
	dynamoContainer = &dynamoDBContainer{
		Container: container,
		URL:       uri,
	}
	return nil
}
