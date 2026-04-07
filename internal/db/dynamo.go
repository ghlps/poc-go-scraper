package db

import (
	"context"
	"fmt"
	appconfig "go-scraper/internal/config"
	"go-scraper/internal/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const tableName = "scraper_menu_executions"

type Store struct {
	client *dynamodb.Client
}

func NewStore(ctx context.Context, cfgApp appconfig.Config) (*Store, error) {
	var client *dynamodb.Client

	if cfgApp.IsDev {
		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithRegion("us-east-1"),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dummy", "dummy", "")),
		)

		if err != nil {
			return nil, fmt.Errorf("load aws config: %w", err)
		}

		client = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(cfgApp.DynamoURL)
		})
	} else {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("load aws config: %w", err)
		}
		client = dynamodb.NewFromConfig(cfg)
	}

	return &Store{client: client}, nil
}

func (s *Store) Save(ctx context.Context, data models.ScraperExecution) error {
	item, err := attributevalue.MarshalMap(data)
	if err != nil {
		return fmt.Errorf("marshal response data: %w", err)
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("put item: %w", err)
	}

	return nil
}

func (s *Store) GetByDate(ctx context.Context, date string) (*models.ResponseData, error) {
	result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"date": &types.AttributeValueMemberS{Value: date},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get item: %w", err)
	}
	if result.Item == nil {
		return nil, nil
	}

	var out models.ResponseData
	if err := attributevalue.UnmarshalMap(result.Item, &out); err != nil {
		return nil, fmt.Errorf("unmarshal response data: %w", err)
	}

	return &out, nil
}

func (s *Store) HasFailedExecutionForDate(ctx context.Context, date string) (bool, error) {
	// We filter by status = FAILED and where created_at begins with the date prefix.
	// Using a Scan with FilterExpression — suitable for small tables.
	// For large tables, add a GSI on (date_partition, status) instead.
	out, err := s.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(tableName),
		FilterExpression: aws.String("begins_with(created_at, :date) AND #st = :status"),
		ExpressionAttributeNames: map[string]string{
			"#st": "status", // "status" is a reserved word in DynamoDB
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":date":   &types.AttributeValueMemberS{Value: date}, // e.g. "2025-04-07"
			":status": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", models.ExecutionStatusFailed)},
		},
		Limit: aws.Int32(1), // We only need to know if at least one exists
	})
	if err != nil {
		return false, fmt.Errorf("scan failed executions for date %s: %w", date, err)
	}

	return len(out.Items) > 0, nil
}

func (s *Store) GetLatestByDate(ctx context.Context, date string, ruCode string) (*models.ScraperExecution, error) {
	out, err := s.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(tableName),
		FilterExpression: aws.String("menu.#date = :date AND ru.#code = :ru"),
		ExpressionAttributeNames: map[string]string{
			"#code": "code",
			"#date": "date",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":date": &types.AttributeValueMemberS{Value: date},
			":ru":   &types.AttributeValueMemberS{Value: ruCode},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("scan executions for date %s: %w", date, err)
	}
	if len(out.Items) == 0 {
		return nil, nil
	}

	var executions []models.ScraperExecution
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &executions); err != nil {
		return nil, fmt.Errorf("unmarshal executions: %w", err)
	}

	latest := executions[0]
	for _, e := range executions[1:] {
		if e.CreatedAt.After(latest.CreatedAt) {
			latest = e
		}
	}

	return &latest, nil
}
