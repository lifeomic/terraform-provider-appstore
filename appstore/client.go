package appstore

import (
	"errors"

	"github.com/lifeomic/phc-sdk-go/client"

	"github.com/mitchellh/mapstructure"
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

const GRAPHQL_URL = "app-store-service:deployed/graphql"

type AppStoreClient struct {
	phc *client.Client
}

type app struct {
	Name          string
	Description   string
	AuthorDisplay string
	Image         string
	Url           string
}

func (self *AppStoreClient) getAppStoreListing(id string) (*app, error) {
	type Body struct {
		Data struct {
			App app `json:"app"`
		} `json:"data"`
	}
	res, err := self.phc.Gql(GRAPHQL_URL, GET_APP_STORE_LISTING, map[string]interface{}{"id": id})
	if err != nil {
		return nil, err
	}

	var data struct {
		App app
	}
	err = mapstructure.Decode(res, &data)
	if err != nil {
		return nil, err
	}
	return &data.App, nil
}

type appStoreCreate struct {
	Name          string
	AuthorDisplay string
	Url           string
	Description   string
	Image         string
}

func (self *AppStoreClient) createAppStoreListing(params appStoreCreate) (*string, error) {
	res, err := self.phc.Gql(GRAPHQL_URL, CREATE_APP_STORE_LISTING, map[string]interface{}{"input": map[string]string{
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
	var data struct {
		CreateWebApp struct {
			Id string
		}
	}
	err = mapstructure.Decode(res, &data)
	if err != nil {
		return nil, err
	}
	return &data.CreateWebApp.Id, nil
}

func (self *AppStoreClient) editAppStoreListing(id string, params appStoreCreate) error {
	res, err := self.phc.Gql(GRAPHQL_URL, EDIT_APP_STORE_LISTING, map[string]interface{}{
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

	var data struct {
		EditWebApp bool
	}
	err = mapstructure.Decode(res, &data)
	if err != nil {
		return err
	}
	if !data.EditWebApp {
		return errors.New("The app you're trying to edit does not exist")
	}
	return nil
}

func (self *AppStoreClient) deleteAppStoreListing(id string) error {
	res, err := self.phc.Gql(GRAPHQL_URL, DELETE_APP_STORE_LISTING, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return err
	}

	var data struct {
		DeleteApp bool
	}
	err = mapstructure.Decode(res, &data)
	if err != nil {
		return err
	}
	if !data.DeleteApp {
		return errors.New("The app you're trying to delete does not exist")
	}
	return nil
}

func BuildAppStoreClient() (*AppStoreClient, error) {
	phcClient, err := client.BuildClient("lifeomiclife", "app-store-tf", map[string]bool{
		"createData": true,
		"updateData": true,
		"deleteData": true,
	})
	if err != nil {
		return nil, err
	}
	client := AppStoreClient{phc: phcClient}
	return &client, nil
}
