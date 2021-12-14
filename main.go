package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

const GET_APP_STORE_LISTING = `
  query GetAppStoreListing($id: ID!) {
    app(id: $id) {
      name
      description
      authorDisplay
      image
      ... on AppStoreWebApplication {
        url
      }
    }
  }
`

type Payload struct {
	Headers               map[string]string `json:"headers"`
	Path                  string            `json:"path"`
	HttpMethod            string            `json:"httpMethod"`
	QueryStringParameters map[string]string `json:"queryStringParameters"`
	Body                  string            `json:"body"`
}

type Body struct {
	Query     string            `json:"query"`
	Variables map[string]string `json:"variables"`
}

func gqlQuery(query string, variables map[string]string) []byte {
	body, _ := json.Marshal(&Body{Query: query, Variables: variables})
	payload := &Payload{
		Headers:               map[string]string{"LifeOmic-Account": "lifeomiclife", "LifeOmic-User": "app-store-tf", "content-type": "application/json"},
		HttpMethod:            "POST",
		QueryStringParameters: map[string]string{},
		Path:                  "/graphql",
		Body:                  string(body),
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Failed to marshall payload %v", err)
	}
	return bytes
}

type AppStoreClient struct {
	c *lambda.Client
}

func (client *AppStoreClient) gql(query string, variables map[string]string) (*lambda.InvokeOutput, error) {
	APP_STORE_ARN := "app-store-service:deployed"
	return client.c.Invoke(context.TODO(), &lambda.InvokeInput{
		FunctionName: &APP_STORE_ARN,
		Payload:      gqlQuery(query, variables),
	})
}

func BuildAppStoreClient() (*AppStoreClient, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	client := AppStoreClient{c: lambda.NewFromConfig(cfg)}
	return &client, nil
}

func main() {
	client, err := BuildAppStoreClient()
	if err != nil {
		log.Fatalf("failed to build app store client, %v", err)
	}
	res, err := client.gql(GET_APP_STORE_LISTING, map[string]string{"id": "58e9ede8-eb28-40b6-82a6-d8b670d9c651"})
	if err != nil {
		log.Fatalf("Failed to query app-store-service, %v", err)
	}
	log.Print(string(res.Payload))
}
