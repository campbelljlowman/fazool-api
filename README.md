# Fazool
API for Fazool music service

# Env setup

1. Install go 
2. Setup postgres server (see make target)
3. cd to project root
4. run `make run`

# Integration tests

1. Install node
2. Install `sudo npm install -g typescript`
3. Install dependencies `npm install`

# Environment Variables
The API expects the following environment variables to be set
## SPOTIFY_ID
- The spotify ID comes from the spotify project. The name of this variable should not be changed, as it's picked up
by and used by the underlying spotify package
## SPOTIFY_SECRET
- The spotify secret also comes from the spotify project. The name of this variable should not be changed, as it's picked up
by and used by the underlying spotify package
## JWT_SECRET_KEY
- This is a seed phrase that's used to create and verify JWTs for account authentication. This phrase can really be any value,
however if is changed, existing tokens will stop working
## POSTGRES_URL
- URL for connecting to postgres database. Format is {POSTGRES_HOST}://{POSTGRES_USERNAME}:{POSTGRES_PASSWORD}@{POSTGRES_ADDRESS}:{POSTGRES_PORT}/{POSTGRES_DATABASE}
## LOG_LEVEL
- Log level to run the API in. Options are DEBUG, INFO, ERROR. Default is INFO