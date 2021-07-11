package main

import (
	"net/url"

	"golang.org/x/oauth2"
)

func oauth2Config() *oauth2.Config {

	tokenURL := &url.URL{
		Scheme: "https",
		Host:   cfg.Auth0.Tenant,
		Path:   "/oauth/token",
	}

	return &oauth2.Config{
		ClientID:     cfg.Auth0.ClientID,
		ClientSecret: cfg.Auth0.ClientSecret,
		RedirectURL:  cfg.Auth0.RedirectURI,
		Endpoint: oauth2.Endpoint{
			TokenURL: tokenURL.String(),
		},
	}
}
