GO_VERSION=1.19.3
# To make it easy to connect from your dev machine, make your postgres username the same as your linux username
POSTGRES_USERNAME=clowman
POSTGRES_PASSWORD=asdf
# Application environment variables
POSTRGRES_URL=postgres://${POSTGRES_USERNAME}:${POSTGRES_PASSWORD}@localhost:5432/fazool # This requires databases clowman and fazool to exist already

include .env
export

init: go-init postgres-init

# Install go language
go-init:
	wget https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz
	sudo tar -xvf go${GO_VERSION}.linux-amd64.tar.gz
	sudo mv go /usr/local
# Then add these lines to .bashrc
# export GOROOT=/usr/local/go
# export GOPATH=$HOME/go
# export PATH=$GOPATH/bin:$GOROOT/bin:$PATH


# Setup local postgres database. Set postgres variables before running
postgres-init:
	sudo apt update && apt upgrade
	sudo apt-get -y install postgresql postgresql-contrib
	sudo service postgresql start
	sudo -u postgres psql
	CREATE USER ${POSTGRES_USERNAME} PASSWORD ${POSTGRES_PASSWORD} CREATEDB;

# Run project locally
run:
	POSTRGRES_URL=${POSTRGRES_URL} \
	go run .


# Start local postgres database
pg-up:
	sudo service postgresql start

pg-down:
	sudo service postgresql stop

build:
	docker build .

regen-graphql:
	go get github.com/99designs/gqlgen@v0.17.15 && go run github.com/99designs/gqlgen generate

unit-test:
	go test -v ./... -coverprofile=coverage.out

INTEGRATION_TEST_POSTGRES_URL=postgres://postgres:asdf@localhost:5432/fazool-integration-test
integration-test:
# docker-compose up -d --remove-orphans postgres;
	POSTGRES_URL=${INTEGRATION_TEST_POSTGRES_URL}
	go test -v -tags=integration
# docker-compose down;