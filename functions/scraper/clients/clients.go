package clients

import (
	"context"
	"fmt"
	"os"
	"site-tracker/functions/scraper/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

var (
	DynamoDBClient    *dynamodb.Client
	DynamoDBTableName string
	SQSClient         *sqs.Client
	SQSQueueURL       string
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}
	DynamoDBClient = dynamodb.NewFromConfig(cfg)
	DynamoDBTableName = os.Getenv("DYNAMODB_TABLE")
	if DynamoDBTableName == "" {
		panic("DYNAMODB_TABLE environment variable is not set")
	}

	SQSClient = sqs.NewFromConfig(cfg)
	SQSQueueURL = os.Getenv("SQS_URL")
	if SQSQueueURL == "" {
		panic("SQS_URL environment variable is not set")
	}
}

func GetItem(ctx context.Context, id string) (*models.Site, error) {
	out, err := DynamoDBClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &DynamoDBTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
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

	_, err = DynamoDBClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &DynamoDBTableName,
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}
	return nil
}

func UpdateItemShouldCheck(ctx context.Context, id string, shouldCheck bool) error {
	_, err := DynamoDBClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &DynamoDBTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression:          aws.String("SET #sc = :sc"),
		ExpressionAttributeNames:  map[string]string{"#sc": "shouldCheck"},
		ExpressionAttributeValues: map[string]types.AttributeValue{":sc": &types.AttributeValueMemberBOOL{Value: shouldCheck}},
	})
	if err != nil {
		return fmt.Errorf("failed to update shouldCheck: %w", err)
	}
	return nil
}

func DeleteItem(ctx context.Context, id string) error {
	_, err := DynamoDBClient.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &DynamoDBTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	return nil
}

func SendMessage(ctx context.Context, message string) error {
	_, err := SQSClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &SQSQueueURL,
		MessageBody: &message,
	})
	if err != nil {
		return fmt.Errorf("failed to send message to SQS queue, %w", err)
	}
	return nil
}
