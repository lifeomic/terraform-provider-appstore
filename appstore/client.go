package appstore

import (
	"context"
	"encoding/json"
	"errors"
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

const DELETE_APP_STORE_LISTING = `
  mutation DeleteAppStoreListing($id: ID!) {
	deleteApp(id: $id)
  }
`

const CREATE_APP_STORE_LISTING = `
  mutation CreateAppStoreListing($input: CreateWebAppInput!) {
    createWebApp(input: $input) {
      id
    }
  }
`

const EDIT_APP_STORE_LISTING = `
  mutation EditAppStoreListing($id: ID!, $edits: EditWebAppInput!) {
    editWebApp(id: $id, edits: $edits) 
  }
`

type payload struct {
	Headers               map[string]string `json:"headers"`
	Path                  string            `json:"path"`
	HttpMethod            string            `json:"httpMethod"`
	QueryStringParameters map[string]string `json:"queryStringParameters"`
	Body                  string            `json:"body"`
}

type policy struct {
	Rules map[string]bool `json:"rules"`
}

func gqlQuery(query string, variables map[string]interface{}) []byte {
	type Body struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}
	policy, _ := json.Marshal(&policy{
		Rules: map[string]bool{
			"createData": true,
			"updateData": true,
			"deleteData": true,
		},
	})
	body, _ := json.Marshal(&Body{Query: query, Variables: variables})
	payload := &payload{
		Headers:               map[string]string{"LifeOmic-Account": "lifeomiclife", "LifeOmic-User": "app-store-tf", "content-type": "application/json", "LifeOmic-Policy": string(policy)},
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

func (client *AppStoreClient) gql(query string, variables map[string]interface{}) (*lambda.InvokeOutput, error) {
	APP_STORE_ARN := "app-store-service:deployed"
	return client.c.Invoke(context.Background(), &lambda.InvokeInput{
		FunctionName: &APP_STORE_ARN,
		Payload:      gqlQuery(query, variables),
	})
}

type app struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	AuthorDisplay string `json:"authorDisplay"`
	Image         string `json:"image"`
	Url           string `json:"url"`
}

type Payload struct {
	Body string `json:"body"`
}

func (client *AppStoreClient) getAppStoreListing(id string) (*app, error) {
	type Body struct {
		Data struct {
			App app `json:"app"`
		} `json:"data"`
	}
	res, err := client.gql(GET_APP_STORE_LISTING, map[string]interface{}{"id": id})
	if err != nil {
		return nil, err
	}
	var payload Payload
	err = json.Unmarshal(res.Payload, &payload)
	if err != nil {
		return nil, err
	}
	var body Body
	err = json.Unmarshal([]byte(payload.Body), &body)
	if err != nil {
		return nil, err
	}
	app := body.Data.App
	return &app, nil
}

type appStoreCreate struct {
	Name          string
	AuthorDisplay string
	Url           string
	Description   string
	Image         string
}

func (client *AppStoreClient) createAppStoreListing(params appStoreCreate) (*string, error) {
	type Body struct {
		Data struct {
			CreateWebApp struct {
				Id string `json:"id"`
			} `json:"createWebApp"`
		} `json:"data"`
	}

	res, err := client.gql(CREATE_APP_STORE_LISTING, map[string]interface{}{"input": map[string]string{
		"name":          params.Name,
		"authorDisplay": params.AuthorDisplay,
		"url":           params.Url,
		"description":   params.Description,
		"image":         params.Image,
		"product":       "LX",
	}})
	if err != nil {
		return nil, err
	}
	var payload Payload
	err = json.Unmarshal(res.Payload, &payload)
	if err != nil {
		return nil, err
	}
	var body Body
	err = json.Unmarshal([]byte(payload.Body), &body)
	if err != nil {
		return nil, err
	}
	id := body.Data.CreateWebApp.Id
	return &id, nil
}

func (client *AppStoreClient) editAppStoreListing(id string, params appStoreCreate) error {
	type Body struct {
		Data struct {
			EditWebApp bool `json:"editWebApp"`
		} `json:"data"`
	}

	res, err := client.gql(EDIT_APP_STORE_LISTING, map[string]interface{}{
		"id": id,
		"edits": map[string]string{
			"name":          params.Name,
			"authorDisplay": params.AuthorDisplay,
			"url":           params.Url,
			"description":   params.Description,
			"image":         params.Image,
		}})
	if err != nil {
		return err
	}
	var payload Payload
	err = json.Unmarshal(res.Payload, &payload)
	if err != nil {
		return err
	}
	var body Body
	err = json.Unmarshal([]byte(payload.Body), &body)
	if err != nil {
		return err
	}
	if !body.Data.EditWebApp {
		return errors.New("Failed to edit web app")
	}
	return nil
}

func (client *AppStoreClient) deleteAppStoreListing(id string) error {
	res, err := client.gql(DELETE_APP_STORE_LISTING, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return err
	}

	var payload Payload
	err = json.Unmarshal(res.Payload, &payload)
	if err != nil {
		return err
	}
	var body struct {
		Data struct {
			DeleteApp bool `json:"deleteApp"`
		} `json:"data"`
	}
	err = json.Unmarshal([]byte(payload.Body), &body)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(payload.Body), &body)
	if err != nil {
		return err
	}
	if !body.Data.DeleteApp {
		return errors.New("Failed to delete web app")
	}
	return nil
}

func BuildAppStoreClient() (*AppStoreClient, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithSharedConfigProfile("lifeomic-dev"))
	if err != nil {
		return nil, err
	}
	client := AppStoreClient{c: lambda.NewFromConfig(cfg)}
	return &client, nil
}
