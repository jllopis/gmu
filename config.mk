# Project options
BLDDIR = _build
BLDDATE=$(shell date -u +%Y%m%dT%H%M%S)
VERSION ?= $(shell git describe --tags --always --dirty)
GIT_COMMIT ?= $(shell git rev-parse --short HEAD)
API_VERSION="1.0"

DOCKER=$(shell which docker)
COMPOSE=$(shell which docker-compose)

UID=$(shell id -u)
GID=$(shell id -g)

# Build options
LDFLAGS=" -s -X version.Name=$(BINNAME) -X version.BuildDate=$(BLDDATE) -X version.Version=$(VERSION) -X version.APIVersion=$(API_VERSION) -X version.GitCommit=$(GIT_COMMIT)"
SRCS = $(wildcard *.go ./**/*.go)
OS=$(shell uname -s | tr "[:upper:]" "[:lower:]")

BINNAME="gmu"
GITPROJECT="gmu"
ORG_PATH=github.com/jllopis
REPO_PATH=$(ORG_PATH)/$(GITPROJECT)

# Run test server options
PT  ?= 3000                                ## SERVER PORT
SH  ?= localhost                           ## STORE_HOST
SP  ?= 5432                                ## STORE_PORT
SN  ?= statsdb                             ## STORE_NAME
SU  ?= statsadm                            ## STORE_ADMIN
SW  ?= 00000000                            ## STORE_PASS
AH  ?= annotation.dev.acbdata.net:59800    ## ANNOTATION_HOST
StH ?= stats.dev.acbdata.net:59800         ## STATS_HOST
CH  ?= competicion.dev.acbdata.net:59800   ## COMPAPI_HOST
AA ?= arbitraje.dev.acbdata.net:59800      ## ARBAPI_HOST
MD  ?= dev                                 ## Mode. If "dev", all db queries will be logged