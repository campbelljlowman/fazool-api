POSTGRES_USERNAME=clowman
POSTGRES_PASSWORD=asdf
POSTGRES_URL=postgres://${POSTGRES_USERNAME}:${POSTGRES_PASSWORD}@localhost:5432/fazool
POSTGRES_DEV_URL_DOCKER=postgres://${POSTGRES_USERNAME}:${POSTGRES_PASSWORD}@host.docker.internal:5432/fazool # This requires databases clowman and fazool to exist already

REGISTRY=ghcr.io
IMAGE_NAME=campbelljlowman/fazool-api
STABLE_VERSION=0.1.0
UNIQUE_VERSION=${STABLE_VERSION}-${shell date "+%Y.%m.%d"}-${shell git rev-parse --short HEAD}
STABLE_IMAGE_TAG=${REGISTRY}/${IMAGE_NAME}:${STABLE_VERSION}
UNIQUE_IMAGE_TAG=${REGISTRY}/${IMAGE_NAME}:${UNIQUE_VERSION}

ENV_FILE=./.env

include .env
export

pg-up:
	docker-compose up -d --remove-orphans postgres;

pg-down:
	docker-compose down;

run:
	POSTGRES_URL=${POSTGRES_URL} \
	go run .

run-docker:
	docker run --rm \
	-p 8080:8080 \
	--env-file ${ENV_FILE} \
	--env POSTGRES_URL=${POSTGRES_DEV_URL_DOCKER} \
	${UNIQUE_IMAGE_TAG}

build:
	docker build \
	-t ${STABLE_IMAGE_TAG} \
	-t ${UNIQUE_IMAGE_TAG} \
	.

publish:
	docker login ghcr.io -u campbelljlowman -p ${GITHUB_ACCESS_TOKEN}
	docker push ${UNIQUE_IMAGE_TAG}
ifeq ($(shell git rev-parse --abbrev-ref HEAD), master)
	docker push ${STABLE_IMAGE_TAG}
else 
	@echo "Not pushing stable version because not on master branch"
endif
	docker logout

regen-graphql:
	go get github.com/99designs/gqlgen@v0.17.15 && go run github.com/99designs/gqlgen generate

unit-test:
	go test -v ./... -coverprofile=coverage.out

integration-test:
	npm --prefix ./integration_tests run build
	npm --prefix ./integration_tests test