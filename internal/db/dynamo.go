package db

import (
	"context"
	"fmt"
	"go-scraper/internal/models"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const tableName = "MenuResponses"

type Store struct {
	client *dynamodb.Client
}

func NewStore(ctx context.Context) (*Store, error) {
	var client *dynamodb.Client

	if os.Getenv("APP_ENV") == "dev" {
		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithRegion("us-east-1"),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dummy", "dummy", "")),
		)
		if err != nil {
			return nil, fmt.Errorf("load aws config: %w", err)
		}

		client = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String("http://localhost:8000")
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

func (s *Store) Save(ctx context.Context, data models.ResponseData) error {
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
