package ddb

import (
	"context"
	"errors"
	"fmt"
	"github.com/akijowski/tweek-2021-sam/internal/schema"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"reflect"
	"testing"
	"time"
)

type mockDynamoUpdateItemAPI func(ctx context.Context, input *dynamodb.UpdateItemInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)

func (m mockDynamoUpdateItemAPI) UpdateItem(ctx context.Context, input *dynamodb.UpdateItemInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return m(ctx, input, optFns...)
}

type mockDynamoScanAPI func(ctx context.Context, input *dynamodb.ScanInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.ScanOutput, error)

func (m mockDynamoScanAPI) Scan(ctx context.Context, input *dynamodb.ScanInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	return m(ctx, input, optFns...)
}

type mockDynamoQueryAPI func(ctx context.Context, input *dynamodb.QueryInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.QueryOutput, error)

func (m mockDynamoQueryAPI) Query(ctx context.Context, input *dynamodb.QueryInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	return m(ctx, input, optFns...)
}

func TestAddNote(t *testing.T) {
	cases := map[string]struct {
		clientBuilder func(t *testing.T, expectedNote *schema.Note, expectedError error) DynamoUpdateItemAPI
		table         string
		note          *schema.Note
		expectedErr   error
	}{
		"withCorrectInputReturnsCorrectly": {
			clientBuilder: func(t *testing.T, expectedNote *schema.Note, expectedError error) DynamoUpdateItemAPI {
				return mockDynamoUpdateItemAPI(func(ctx context.Context, input *dynamodb.UpdateItemInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
					t.Helper()
					validateUpdateInputKey(t, expectedNote, input.Key)

					return &dynamodb.UpdateItemOutput{
						Attributes: map[string]types.AttributeValue{},
					}, nil
				})
			},
			table: "MY_TABLE",
			note: &schema.Note{
				Owner:   "foo",
				Title:   "titlefoo",
				Message: "messagefoo",
			},
		},
		"returnsDynamoError": {
			clientBuilder: func(t *testing.T, expectedNote *schema.Note, expectedError error) DynamoUpdateItemAPI {
				return mockDynamoUpdateItemAPI(func(ctx context.Context, input *dynamodb.UpdateItemInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
					t.Helper()
					return nil, expectedError
				})
			},
			table: "MY_TABLE",
			note: &schema.Note{
				Owner:   "foo",
				Title:   "titlefoo",
				Message: "messagefoo",
			},
			expectedErr: errors.New("a DynamoDB error occurred"),
		},
		"missingTableNameReturnsError": {
			clientBuilder: func(t *testing.T, expectedNote *schema.Note, expectedError error) DynamoUpdateItemAPI {
				return mockDynamoUpdateItemAPI(func(ctx context.Context, input *dynamodb.UpdateItemInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
					t.Helper()
					return nil, nil
				})
			},
			note: &schema.Note{
				Owner:   "foo",
				Title:   "titlefoo",
				Message: "messagefoo",
			},
			expectedErr: errors.New("tableName must be provided"),
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			api := tt.clientBuilder(t, tt.note, tt.expectedErr)
			actual, err := AddNote(ctx, api, tt.table, tt.note)
			if tt.expectedErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if actual != tt.note.Owner {
					t.Fatalf("unexpected result: wanted %q, got %q", tt.note.Owner, actual)
				}
			} else {
				if err.Error() != tt.expectedErr.Error() {
					t.Fatalf("Unexpected error: wanted \"%v\" got %q", tt.expectedErr, err)
				}
			}
		})
	}
}

func TestScan(t *testing.T) {
	validNotesResponses := []schema.Note{
		{
			Owner:     "owner",
			Title:     "title",
			Message:   "message",
			Timestamp: time.Now().UnixMilli(),
		},
		{
			Owner:     "owner2",
			Title:     "title2",
			Message:   "message2",
			Timestamp: time.Now().UnixMilli(),
		},
	}
	validScanOut, err := marshalListOfMaps(validNotesResponses)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	cases := map[string]struct {
		clientBuilder func(t *testing.T) DynamoScanAPI
		tableName     string
		expectedNotes []schema.Note
		expectedErr   error
	}{
		"valid request returns successfully": {
			clientBuilder: func(t *testing.T) DynamoScanAPI {
				return mockDynamoScanAPI(func(ctx context.Context, input *dynamodb.ScanInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.ScanOutput, error) {
					t.Helper()
					return &dynamodb.ScanOutput{
						ScannedCount: int32(len(validScanOut)),
						Count:        int32(len(validScanOut)),
						Items:        validScanOut,
					}, nil
				})
			},
			tableName:     "MY_TABLE",
			expectedNotes: validNotesResponses,
		},
		"missing table name returns error": {
			clientBuilder: func(t *testing.T) DynamoScanAPI {
				return mockDynamoScanAPI(func(ctx context.Context, input *dynamodb.ScanInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.ScanOutput, error) {
					t.Helper()
					return nil, nil
				})
			},
			expectedErr: errors.New("tableName must be provided"),
		},
		"returns dynamo error": {
			clientBuilder: func(t *testing.T) DynamoScanAPI {
				return mockDynamoScanAPI(func(ctx context.Context, input *dynamodb.ScanInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.ScanOutput, error) {
					t.Helper()
					return nil, &DynamoDBError{ClientMessage: "foo"}
				})
			},
			tableName:   "MY_TABLE",
			expectedErr: errors.New("a DynamoDB error occurred"),
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			api := tt.clientBuilder(t)

			notes, err := Scan(ctx, api, tt.tableName)
			if tt.expectedErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if len(notes) != len(tt.expectedNotes) {
					t.Fatalf("unexpected number of notes: wanted %d got %d", len(tt.expectedNotes), len(notes))
				}
				for i, n := range notes {
					if !reflect.DeepEqual(n, tt.expectedNotes[i]) {
						t.Fatalf("unexpected note: wanted %+v got %+v", tt.expectedNotes[i], n)
					}
				}
			} else {
				if err.Error() != tt.expectedErr.Error() {
					t.Fatalf("unexpected error: wanted %q got %s", tt.expectedErr, err)
				}
			}
		})
	}
}

func TestFindNotesByOwner(t *testing.T) {
	validNotesResponses := []schema.Note{
		{
			Owner:     "owner",
			Title:     "title",
			Message:   "message",
			Timestamp: time.Now().UnixMilli(),
		},
		{
			Owner:     "owner",
			Title:     "title2",
			Message:   "message2",
			Timestamp: time.Now().UnixMilli(),
		},
	}
	validQueryOut, err := marshalListOfMaps(validNotesResponses)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	cases := map[string]struct {
		clientBuilder func(t *testing.T) DynamoQueryAPI
		owner         string
		tableName     string
		expectedNotes []schema.Note
		expectedErr   error
	}{
		"valid request returns successfully": {
			clientBuilder: func(t *testing.T) DynamoQueryAPI {
				return mockDynamoQueryAPI(func(ctx context.Context, input *dynamodb.QueryInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.QueryOutput, error) {
					t.Helper()
					if !isOwnerInKeyExpression(input.ExpressionAttributeNames, input.ExpressionAttributeValues, "owner") {
						t.Fatal("incorrect key expression")
					}
					return &dynamodb.QueryOutput{
						ScannedCount: int32(len(validQueryOut)),
						Count:        int32(len(validQueryOut)),
						Items:        validQueryOut,
					}, nil
				})
			},
			owner:         "owner",
			tableName:     "MY_TABLE",
			expectedNotes: validNotesResponses,
		},
		"missing table name returns error": {
			clientBuilder: func(t *testing.T) DynamoQueryAPI {
				return mockDynamoQueryAPI(func(ctx context.Context, input *dynamodb.QueryInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.QueryOutput, error) {
					t.Helper()
					return &dynamodb.QueryOutput{}, nil
				})
			},
			owner:       "owner",
			expectedErr: errors.New("tableName must be provided"),
		},
		"missing owner returns error": {
			clientBuilder: func(t *testing.T) DynamoQueryAPI {
				return mockDynamoQueryAPI(func(ctx context.Context, input *dynamodb.QueryInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.QueryOutput, error) {
					t.Helper()
					return &dynamodb.QueryOutput{}, nil
				})
			},
			tableName:   "MY_TABLE",
			expectedErr: errors.New("owner must be provided"),
		},
		"returns dynamo error": {
			clientBuilder: func(t *testing.T) DynamoQueryAPI {
				return mockDynamoQueryAPI(func(ctx context.Context, input *dynamodb.QueryInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.QueryOutput, error) {
					t.Helper()
					return nil, &DynamoDBError{ClientMessage: "foo"}
				})
			},
			tableName:   "MY_TABLE",
			owner:       "owner123",
			expectedErr: errors.New("a DynamoDB error occurred"),
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			api := tt.clientBuilder(t)

			actual, err := FindNotesByOwner(ctx, api, tt.tableName, tt.owner)
			if tt.expectedErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if len(actual) != len(tt.expectedNotes) {
					t.Fatalf("unexpected number of notes: wanted %d got %d", len(tt.expectedNotes), len(actual))
				}
				for i, n := range actual {
					if !reflect.DeepEqual(n, tt.expectedNotes[i]) {
						t.Fatalf("unexpected note: wanted %+v got %+v", tt.expectedNotes[i], n)
					}
				}
			} else {
				if err.Error() != tt.expectedErr.Error() {
					t.Fatalf("unexpected error: wanted %q got %s", tt.expectedErr, err)
				}
			}
		})
	}
}

func validateUpdateInputKey(t *testing.T, expectedNote *schema.Note, actualKeys map[string]types.AttributeValue) {
	var actualKeyMap map[string]string
	err := attributevalue.UnmarshalMap(actualKeys, &actualKeyMap)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if actualOwner, ok := actualKeyMap["owner"]; ok != true {
		t.Fatal("missing primary key for Owner")
	} else if actualOwner != expectedNote.Owner {
		t.Fatalf("expected owner %v, go %v", expectedNote.Owner, actualOwner)
	}
	if actualTitle, ok := actualKeyMap["title"]; ok != true {
		t.Fatal("missing sort key for Title")
	} else if actualTitle != expectedNote.Title {
		t.Fatalf("expected title %v, got %v", expectedNote.Title, actualTitle)
	}
}

// isOwnerInKeyExpression verifies that the names map contains the value of 'owner' and the values map contains the given 'ownerName'
func isOwnerInKeyExpression(names map[string]string, values map[string]types.AttributeValue, ownerName string) bool {
	hasName, hasValue := false, false
	var valueMap map[string]string
	for _, v := range names {
		if v == "owner" {
			hasName = true
		}
	}
	if err := attributevalue.UnmarshalMap(values, &valueMap); err != nil {
		fmt.Printf("error checking key expression: %s", err)
		return false
	}
	for _, v := range valueMap {
		if v == ownerName {
			hasValue = true
		}
	}
	return hasName && hasValue
}

func marshalListOfMaps(in []schema.Note) ([]map[string]types.AttributeValue, error) {
	var out []map[string]types.AttributeValue
	for _, n := range in {
		av, err := attributevalue.MarshalMap(n)
		if err != nil {
			return nil, err
		}
		out = append(out, av)
	}
	return out, nil
}
