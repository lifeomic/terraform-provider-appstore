package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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

type Payload struct {
	Headers               map[string]string `json:"headers"`
	Path                  string            `json:"path"`
	HttpMethod            string            `json:"httpMethod"`
	QueryStringParameters map[string]string `json:"queryStringParameters"`
	Body                  string            `json:"body"`
}

type Policy struct {
	Rules map[string]bool `json:"rules"`
}

func gqlQuery(query string, variables map[string]interface{}) []byte {
	type Body struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}
	policy, _ := json.Marshal(&Policy{
		Rules: map[string]bool{
			"createData": true,
			"updateData": true,
		},
	})
	body, _ := json.Marshal(&Body{Query: query, Variables: variables})
	payload := &Payload{
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
	return client.c.Invoke(context.TODO(), &lambda.InvokeInput{
		FunctionName: &APP_STORE_ARN,
		Payload:      gqlQuery(query, variables),
	})
}

type App struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	AuthorDisplay string `json:"authorDisplay"`
	Image         string `json:"image"`
	Url           string `json:"url"`
}

func (client *AppStoreClient) getAppStoreListing(id string) (*App, error) {
	type Payload struct {
		Body string `json:"body"`
	}
	type Body struct {
		Data struct {
			App App `json:"app"`
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

type AppStoreCreate struct {
	Name          string
	AuthorDisplay string
	Url           string
	Description   string
	Image         string
}

func (client *AppStoreClient) createAppStoreListing(params AppStoreCreate) (*string, error) {
	type Payload struct {
		Body string `json:"body"`
	}
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

func (client *AppStoreClient) editAppStoreListing(id string, params AppStoreCreate) error {
	type Payload struct {
		Body string `json:"body"`
	}
	type Body struct {
		Data struct {
			EdtiWebApp bool `json:"editWebApp"`
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
	if !body.Data.EdtiWebApp {
		return errors.New("Failed to edit web app")
	}
	return nil
}

func BuildAppStoreClient() (*AppStoreClient, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	client := AppStoreClient{c: lambda.NewFromConfig(cfg)}
	return &client, nil
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	return BuildAppStoreClient()
}

func readApplet(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AppStoreClient)
	id := d.Id()
	app, err := client.getAppStoreListing(id)
	if err != nil {
		return err
	}
	d.Set("name", app.Name)
	d.Set("description", app.Description)
	d.Set("author_display", app.AuthorDisplay)
	d.Set("image", app.Image)
	d.Set("url", app.Url)
	return nil
}

func createApplet(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AppStoreClient)
	id, err := client.createAppStoreListing(AppStoreCreate{
		Name:          d.Get("name").(string),
		AuthorDisplay: d.Get("author_display").(string),
		Image:         d.Get("image").(string),
		Url:           d.Get("url").(string),
		Description:   d.Get("description").(string),
	})
	if err != nil {
		return err
	}
	d.SetId(*id)
	return readApplet(d, meta)
}

func updateApplet(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AppStoreClient)
	err := client.editAppStoreListing(d.Id(), AppStoreCreate{
		Name:          d.Get("name").(string),
		AuthorDisplay: d.Get("author_display").(string),
		Image:         d.Get("image").(string),
		Url:           d.Get("url").(string),
		Description:   d.Get("description").(string),
	})
	if err != nil {
		return err
	}
	return readApplet(d, meta)
}

func deleteApplet(d *schema.ResourceData, meta interface{}) error {
	return errors.New("Unimplemented")
}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		ConfigureFunc: providerConfigure,
		Schema:        map[string]*schema.Schema{},
		ResourcesMap: map[string]*schema.Resource{
			"applet": {
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Required: true,
					},
					"description": {
						Type:     schema.TypeString,
						Required: true,
					},
					"author_display": {
						Type:     schema.TypeString,
						Required: true,
					},
					"image": {
						Type:     schema.TypeString,
						Required: true,
					},
					"url": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
				Create: createApplet,
				Read:   readApplet,
				Update: updateApplet,
				Delete: deleteApplet,
			},
		},
	}
}

func main() {
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: Provider})
}
