package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// var db *gorm.DB

type DB interface {
	Save(table string, item map[string]types.AttributeValue) error

	// TODO: Combine all the GET into 1 function
	Get(table string, key map[string]types.AttributeValue, output interface{}) error
	GetItemById(table string, id string, output interface{}) error
	GetItem(table string, key map[string]types.AttributeValue) (map[string]types.AttributeValue, error)
	QueryOne(queryInput *dynamodb.QueryInput, output interface{}) error
	QueryAll(queryInput *dynamodb.QueryInput, output interface{}) error
	Scan(queryParams *dynamodb.ScanInput, output interface{}) error
	Delete(table string, id string) error
}

type DBClient struct {
	Svc *dynamodb.Client
}

func Init() DB {
	dbClient := &DBClient{}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	if endpoint := os.Getenv("DYNAMODB_ENDPOINT"); endpoint != "" {
		log.Print("Starting Dynamo Locally at ", endpoint)
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if service == dynamodb.ServiceID {
				return aws.Endpoint{
					PartitionID:   "aws",
					URL:           endpoint,
					SigningRegion: "us-east-1",
				}, nil
			}
			// Fallback to default resolution
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		})

		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithEndpointResolverWithOptions(customResolver),
			config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
				Value: aws.Credentials{
					AccessKeyID: "dummy", SecretAccessKey: "dummy", SessionToken: "dummy",
					Source: "Hard-coded credentials; values are irrelevant for local DynamoDB",
				},
			}),
		)
		if err != nil {
			panic(fmt.Sprintf("unable to load SDK config, %v", err))
		}

	}

	dbClient.Svc = dynamodb.NewFromConfig(cfg)

	return dbClient
}

func (dbClient *DBClient) Save(table string, item map[string]types.AttributeValue) error {
	_, err := dbClient.Svc.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item:      item,
	})

	return err
}

func (dbClient *DBClient) Get(table string, key map[string]types.AttributeValue, output interface{}) error {
	result, err := dbClient.Svc.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key:       key,
	})

	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	err = attributevalue.UnmarshalMap(result.Item, &output)
	if err != nil {
		return fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return err
}

func (dbClient *DBClient) GetItem(table string, key map[string]types.AttributeValue) (map[string]types.AttributeValue, error) {
	out, err := dbClient.Svc.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key:       key,
	})

	return out.Item, err
}

func (dbClient *DBClient) Delete(table string, id string) error {
	_, err := dbClient.Svc.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})

	return err
}

func (dbClient *DBClient) GetItemById(table string, id string, output interface{}) error {
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(table),
		KeyConditionExpression: aws.String("id = :id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: id},
		},
		Limit: aws.Int32(1),
	}

	return dbClient.QueryOne(queryInput, output)
}

func (dbClient *DBClient) QueryOne(queryInput *dynamodb.QueryInput, output interface{}) error {
	result, err := dbClient.Svc.Query(context.TODO(), queryInput)
	if err != nil {
		return fmt.Errorf("error querying: %v", err)
	}

	if len(result.Items) == 0 {
		return fmt.Errorf("no record")
	}

	err = attributevalue.UnmarshalMap(result.Items[0], &output)
	if err != nil {
		return fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return err
}

func (dbClient *DBClient) QueryAll(queryInput *dynamodb.QueryInput, output interface{}) error {
	result, err := dbClient.Svc.Query(context.TODO(), queryInput)
	if err != nil {
		return fmt.Errorf("error querying: %v", err)
	}

	err = attributevalue.UnmarshalListOfMaps(result.Items, &output)
	return err
}

func (dbClient *DBClient) Scan(queryParams *dynamodb.ScanInput, output interface{}) error {
	result, err := dbClient.Svc.Scan(context.Background(), queryParams)
	if err != nil {
		return fmt.Errorf("error scanning table: %v", err)
	}

	err = attributevalue.UnmarshalListOfMaps(result.Items, &output)
	if err != nil {
		return fmt.Errorf("error unmarshalling result items: %v", err)
	}
	return nil
}
