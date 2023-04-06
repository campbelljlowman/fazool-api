GO_VERSION=1.19.3
# To make it easy to connect from your dev machine, make your postgres username the same as your linux username
POSTGRES_USERNAME=clowman
POSTGRES_PASSWORD=asdf
POSTGRES_DEV_URL=postgres://${POSTGRES_USERNAME}:${POSTGRES_PASSWORD}@localhost:5432/fazool # This requires databases clowman and fazool to exist already

include .env
export

init: go-init postgres-init

go-init:
	wget https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz
	sudo tar -xvf go${GO_VERSION}.linux-amd64.tar.gz
	sudo mv go /usr/local
# Then add these lines to .bashrc
# export GOROOT=/usr/local/go
# export GOPATH=$HOME/go
# export PATH=$GOPATH/bin:$GOROOT/bin:$PATH


postgres-init:
	sudo apt update && apt upgrade
	sudo apt-get -y install postgresql postgresql-contrib
	sudo service postgresql start
	sudo -u postgres psql
	CREATE USER ${POSTGRES_USERNAME} PASSWORD ${POSTGRES_PASSWORD} CREATEDB;

run:
	POSTGRES_URL=${POSTGRES_DEV_URL} \
	go run .


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

integration-test-setup:
	docker-compose up -d --remove-orphans postgres;

INTEGRATION_TEST_POSTGRES_URL=postgres://postgres:asdf@localhost:5432/fazool-integration-test
integration-test-server-up:
	POSTGRES_URL=${INTEGRATION_TEST_POSTGRES_URL} \
	go run .

integration-test-teardown:
	docker-compose down;

integration-test:
	npm --prefix ./integration_tests run build
	npm --prefix ./integration_tests test