package dynamo

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type DynamoClient struct {
	*dynamodb.DynamoDB
}

const tableName = "spreadsheet"

var tablesDefinitions = []*dynamodb.CreateTableInput{
	{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("ApartmentName"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("Type"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("ApartmentName"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("Type"),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String(tableName),
	},
}

func NewDynamoClient(endpoint, region string) *DynamoClient {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:   aws.String(region),
			Endpoint: aws.String(endpoint),
		},
	}))
	d := &DynamoClient{
		DynamoDB: dynamodb.New(sess),
	}
	d.CreateTablesIfNotExists()
	return d
}

func (d *DynamoClient) CreateTablesIfNotExists() {
	for _, td := range tablesDefinitions {
		_, err := d.CreateTable(td)
		if err != nil {
			log.Fatalf("Got error calling CreateTable: %s", err)
		}
	}
}

func (d *DynamoClient) AddModel(m interface{}) error {
	marsheledModel, err := dynamodbattribute.MarshalMap(m)
	if err != nil {
		return err
	}
	_, err = d.PutItem(&dynamodb.PutItemInput{
		Item:      marsheledModel,
		TableName: aws.String(tableName),
	})
	return err
}
