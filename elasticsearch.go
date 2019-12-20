package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/olivere/elastic/v7"
	v4 "github.com/olivere/elastic/v7/aws/v4"
	"net/url"
)

func newElasticClient(rawElasticSearchURL string) (*elastic.Client, error) {
	elasticSearchURL, err := url.Parse(rawElasticSearchURL)
	if err != nil {
		return nil, err
	}

	newSession, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	return elastic.NewClient(
		elastic.SetURL(elasticSearchURL.String()),
		elastic.SetScheme(elasticSearchURL.Scheme),
		elastic.SetHttpClient(v4.NewV4SigningClient(newSession.Config.Credentials, *newSession.Config.Region)),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
	)
}
