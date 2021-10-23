package main

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Env   string `default:"development"`
	MySQL struct {
		Host    string `required:"true"`
		Port    uint   `required:"true"`
		User    string `required:"true"`
		Pass    string `required:"true"`
		DB      string `required:"true"`
		Migrate bool
	}

	Redis struct {
		Host string `required:"true"`
		Port uint   `required:"true"`
	}

	Log struct {
		Level string `required:"true"`
	}

	API struct {
		Port uint `required:"true"`
	}

	Auth0 struct {
		ClientID     string `envconfig:"AUTH0_CLIENT_ID" required:"true"`
		ClientSecret string `envconfig:"AUTH0_CLIENT_SECRET" required:"true"`
		RedirectURI  string `required:"true"`
		Audience     string `required:"true"`
		Issuer       string `required:"true"`
		Tenant       string `required:"true"`
		JWKSURI      string `required:"true"`
	}

	Plaid struct {
		ClientID     string `envconfig:"PLAID_CLIENT_ID" required:"true"`
		ClientSecret string `envconfig:"PLAID_CLIENT_SECRET" required:"true"`
		Environment  string `default:"sandbox"`
		Webhook      string
	}

	S3 struct {
		// ClientID     string `envconfig:"SPACES_CLIENT_ID" required:"true"`
		// ClientSecret string `envconfig:"SPACES_CLIENT_SECRET" required:"true"`
		// Endpoint     string `envconfig:"SPACES_ENDPOINT" required:"true"`
		Bucket string `envconfig:"S3_BUCKET_NAME" required:"true"`
	}
}

// buildConfig load environment variables from a
// .env file in the .config directory into the env
// then it maps those env variables to the struct above
// If the file is not present, we ignore the error
// and continue with trying to pull from the env like the application
// is in a container and the variables were injected in at runtime.
func buildConfig() {
	_ = godotenv.Load(".config/.env")

	cfg = new(config)
	err := envconfig.Process("", cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to config env: %s", err))
	}
}
