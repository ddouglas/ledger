# Ledger

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

# This is just a randomly generated string. It is used by as a secret on an Auth0 Action that I've copied into the .auth0 folder
AUTH0_SERVERTOKEN=

PLAID_CLIENT_ID=
PLAID_CLIENT_SECRET=

# This is user provided, it is the endpoint on this API that Plaid will hit when specific event such has a users registration or new transactions are detected
PLAID_WEBHOOK=
```
