package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var NotesKeySchema = []types.KeySchemaElement{
	{KeyType: types.KeyTypeHash, AttributeName: aws.String("owner")},
	{KeyType: types.KeyTypeRange, AttributeName: aws.String("title")},
}

var NotesAttributeDefinitions = []types.AttributeDefinition{
	{AttributeName: aws.String("owner"), AttributeType: types.ScalarAttributeTypeS},
	{AttributeName: aws.String("title"), AttributeType: types.ScalarAttributeTypeS},
}

var NotesProvisionedThroughput = &types.ProvisionedThroughput{ReadCapacityUnits: aws.Int64(10), WriteCapacityUnits: aws.Int64(5)}
