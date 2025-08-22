package dynamodb

import (
	"context"
	"fmt"
	"os"
	"site-tracker/internal/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	ddbClient *ddb.Client
	ddbTable  string
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}
	ddbClient = ddb.NewFromConfig(cfg)
	ddbTable = os.Getenv("DYNAMODB_TABLE")
	if ddbTable == "" {
		panic("DYNAMODB_TABLE environment variable is not set")
	}

}

func Scan(ctx context.Context) (*ddb.ScanOutput, error) {
	out, err := ddbClient.Scan(ctx, &ddb.ScanInput{
		TableName:        &ddbTable,
		FilterExpression: aws.String("shouldCheck = :val"),
		ExpressionAttributeValues: map[string]ddbTypes.AttributeValue{
			":val": &ddbTypes.AttributeValueMemberBOOL{Value: true},
		},
	})
	return out, err
}

func GetItem(ctx context.Context, id string) (*models.Site, error) {
	out, err := ddbClient.GetItem(ctx, &ddb.GetItemInput{
		TableName: &ddbTable,
		Key: map[string]ddbTypes.AttributeValue{
			"id": &ddbTypes.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	if out.Item == nil {
		return nil, fmt.Errorf("item not found")
	}

	var site models.Site
	if err := attributevalue.UnmarshalMap(out.Item, &site); err != nil {
		return nil, fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return &site, nil
}

func PutItem(ctx context.Context, site models.Site) error {
	av, err := attributevalue.MarshalMap(site)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	_, err = ddbClient.PutItem(ctx, &ddb.PutItemInput{
		TableName: &ddbTable,
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}
	return nil
}

func UpdateItemShouldCheck(ctx context.Context, id string, shouldCheck bool) error {
	_, err := ddbClient.UpdateItem(ctx, &ddb.UpdateItemInput{
		TableName: &ddbTable,
		Key: map[string]ddbTypes.AttributeValue{
			"id": &ddbTypes.AttributeValueMemberS{Value: id},
		},
		UpdateExpression:          aws.String("SET #sc = :sc"),
		ExpressionAttributeNames:  map[string]string{"#sc": "shouldCheck"},
		ExpressionAttributeValues: map[string]ddbTypes.AttributeValue{":sc": &ddbTypes.AttributeValueMemberBOOL{Value: shouldCheck}},
	})
	if err != nil {
		return fmt.Errorf("failed to update shouldCheck: %w", err)
	}
	return nil
}

func DeleteItem(ctx context.Context, id string) error {
	_, err := ddbClient.DeleteItem(ctx, &ddb.DeleteItemInput{
		TableName: &ddbTable,
		Key: map[string]ddbTypes.AttributeValue{
			"id": &ddbTypes.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	return nil
}
