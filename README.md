# Ledger Backend

## What

Ledger is an application I am building in an effort to help better understand where my money is spent. I've alway been aware of how much money I've had in my account and how much was available to spend, when bills were due and ensure that they were paid, but outside of that, i'm not sure where my money goes. How much money have I spent at a specific consumer? How much money have I spent on rent? These are the questions I'll be looking to answer with this application

Ledger leverages [Auth0](https://auth0.com) for authentication and [Plaid](https://plaid.com) for interfacing with my banking institution. Since I only back with a single institution, this will most likely be the crux of this application and the hardest to code for the masses. There a few gotchu's with the Plaid API that my institution does not fall into, so I will not suffer with them.

## Configuration

You will need a Redis process and MySQL 5.7 Database running to support this application. The configuration below should speak for itself, however I've commented where to get certain values

```
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_USER=
MYSQL_PASS=
MYSQL_DB=ledger

REDIS_HOST=localhost
REDIS_PORT=6379

LOG_LEVEL=debug

API_PORT=9000

AUTH0_CLIENTID=
AUTH0_CLIENTSECRET=
AUTH0_AUDIENCE=
AUTH0_ISSUER=""
AUTH0_JWKSURI=""
AUTH0_TENANT=
AUTH0_REDIRECTURI=

# Plaid Environment must be one of sandbox, development, or production
PLAID_ENVIRONMENT=
PLAID_CLIENT_ID=
PLAID_CLIENT_SECRET=
# Plaid Webhook is the base uri that webhook will be received at. The path is hard coded to api/external/plaid/v1/webhook
PLAID_WEBHOOK=https://ledger.onetwentyseven.dev

# This application has minor support for NewRelic. This will be expanded in the future.
# All Environment variables are documented by the newrelic go-agent. Please review that packages document for information on which envs can be provided. As of the development of this API, the following are used
NEW_RELIC_ENABLED=
NEW_RELIC_APP_NAME=
NEW_RELIC_LICENSE_KEY=
NEW_RELIC_DISTRIBUTED_TRACING_ENABLED=

# Receipts for transaction that are uploaded by the users are stored in Digital Ocean spaces. This will probably be converted to S3 later on down the road. The following should be provided to configure the S3 Package that is used to write to DO Spaces. DO Spaces hosts an S3 compatable API, which allows us to use the AWS SDK to interact with the API

SPACES_CLIENT_ID=
SPACES_CLIENT_SECRET=
SPACES_ENDPOINT=https://nyc3.digitaloceanspaces.com
SPACES_BUCKET=
```

## Running the Application

Whilst the above can be provided as a `.env` file to the application, for the sake of my curiousity, I leveraged Terraform to setup AWS IAM users for development and wrote all of the envs to SSM. The application does not natively pull from SSM, but you can use AWS Vault and Chamber to inject SSM secrets into the env so that no application secrets are stored on the dev machine. Please follow the documentation on those various applications documentation portal for instructions on how to set them up. The Terraform code has been included in the .terrform directory and the following command is now the default method of the launching the application using the Makefile. Please note, to AWS Vault prompts for a password to unlock the secrets file. During development, I store the password in a local env called `AWS_VAULT_FILE_PASSPHRASE` so that I don't constantly have to type this in. the env is not exported in any `*rc` files and it is recommended not to export this variable by default.

```
aws-vault exec ledger-api-admin -- chamber exec ledger-api/development -- go run cmd/ledger/*.go server
```
