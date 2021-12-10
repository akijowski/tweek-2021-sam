// Package ddb is a wrapper for the AWS DynamoDB Client.
//
// All functions expect to take the client as a parameter.
package ddb

import (
	"context"
	"errors"
	"github.com/akijowski/tweek-2021-sam/internal/schema"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
	"time"
)

const (
	TableScanLimit = int32(25)
	TableQueryLimit = int32(25)
)

// DynamoUpdateItemAPI is a stand-in for the UpdateItem function that exists on the AWS DynamoDB Client
type DynamoUpdateItemAPI interface {
	UpdateItem(ctx context.Context, input *dynamodb.UpdateItemInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

// DynamoScanAPI is a stand-in for the Scan function that exists on the AWS DynamoDB Client
type DynamoScanAPI interface {
	Scan(ctx context.Context, input *dynamodb.ScanInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

// DynamoQueryAPI is a stand-in for the Query function that exists on the AWS DynamoDB Client
type DynamoQueryAPI interface {
	Query(ctx context.Context, input *dynamodb.QueryInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

// DynamoDBError encapsulates client errors and returns a consistent error string
type DynamoDBError struct {
	ClientMessage string
}

func (e *DynamoDBError) Error() string { return "a DynamoDB error occurred" }

// AddNote receives the schema.Note and calls the DynamoUpdateItemAPI.UpdateItem function, transforming the Note to the correct
// dynamodb.UpdateItemInput struct.
//
// The return value is the Owner of the Note
func AddNote(ctx context.Context, api DynamoUpdateItemAPI, tableName string, note *schema.Note) (string, error) {

	if tableName == "" {
		return "", errors.New("tableName must be provided")
	}

	keys, err := attributevalue.MarshalMap(map[string]string{"owner": note.Owner, "title": note.Title})
	if err != nil {
		return "", err
	}

	log.Printf("writing %+v to %s", note, tableName)
	expr, err := expression.
		NewBuilder().
		WithUpdate(buildUpdateExpression(note)).
		Build()
	if err != nil {
		return "", err
	}

	input := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(tableName),
		Key:                       keys,
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ReturnValues:              types.ReturnValueUpdatedNew,
	}

	log.Printf("input, %+v", input)
	output, err := api.UpdateItem(ctx, input)
	if err != nil {
		return "", &DynamoDBError{ClientMessage: err.Error()}
	}
	log.Printf("output, %+v", output)
	return note.Owner, nil
}

// Scan calls the DynamoScanAPI.Scan function, returning a []schema.Note.
func Scan(ctx context.Context, api DynamoScanAPI, tableName string) ([]schema.Note, error) {
	if tableName == "" {
		return nil, errors.New("tableName must be provided")
	}
	log.Printf("scanning table (limit: %d)\n", TableScanLimit)
	output, err := api.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(tableName),
		Limit: aws.Int32(TableScanLimit),
	})
	if err != nil {
		return nil, &DynamoDBError{ClientMessage: err.Error()}
	}
	log.Printf("scanned %d items\n", output.ScannedCount)
	var notes []schema.Note
	if err = attributevalue.UnmarshalListOfMaps(output.Items, &notes); err != nil {
		return nil, err
	}
	return notes, nil
}

// FindNotesByOwner calls the DynamoQueryAPI.Query function, returning a []schema.Note for the given owner.
func FindNotesByOwner(ctx context.Context, api DynamoQueryAPI, tableName, owner string) ([]schema.Note, error) {
	if tableName == "" {
		return nil, errors.New("tableName must be provided")
	}
	if owner == "" {
		return nil, errors.New("owner must be provided")
	}
	expr, err := expression.NewBuilder().
		WithKeyCondition(expression.KeyEqual(expression.Key("owner"), expression.Value(owner))).
		Build()
	if err != nil {
		return nil, err
	}
	input := &dynamodb.QueryInput{
		TableName: aws.String(tableName),
		Limit: aws.Int32(TableQueryLimit),
		ExpressionAttributeNames: expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression: expr.KeyCondition(),
	}
	log.Printf("querying for owner %q (limit: %d)\n", owner, TableQueryLimit)
	output, err := api.Query(ctx, input)
	if err != nil {
		return nil, &DynamoDBError{ClientMessage: err.Error()}
	}
	log.Printf("scanned %d items\n", output.ScannedCount)
	var notes []schema.Note
	if err = attributevalue.UnmarshalListOfMaps(output.Items, &notes); err != nil {
		return nil, err
	}
	return notes, nil
}

func buildUpdateExpression(note *schema.Note) expression.UpdateBuilder {
	return expression.
		Set(expression.Name("message"), expression.Value(note.Message)).
		Set(expression.Name("timestamp"), expression.Value(time.Now().Unix()))
}
