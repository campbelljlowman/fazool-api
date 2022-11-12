GO_VERSION=1.19.3
# To make it easy to connect from your dev machine, make your postgres username the same as your linux username
POSTGRES_USERNAME=clowman
POSTGRES_PASSWORD=asdf
# Application environment variables
DATABASE_URL=postgres://${POSTGRES_USERNAME}:${POSTGRES_PASSWORD}@localhost:5432/fazool # This requires databases clowman and fazool to exist already

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
	sudo apt-get -y install postgresql postgresql-contrib
	sudo service postgresql start
	sudo -u postgres psql
	CREATE USER ${POSTGRES_USERNAME} PASSWORD ${POSTGRES_PASSWORD} CREATEDB;

# Run project locally
run:
	go run . \
	DATABASE_URL=${DATABASE_URL}

# Start local postgres database
pg:
	sudo service postgresql start
