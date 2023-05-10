BIN := "./bin/antibruteforce"
BIN_CLI := "./bin/cli"
DOCKER_IMG="antibruteforce:develop"
DSN="imapp:LightInDark@/OTUSAntibruteforce?parseTime=true"
XTERM="/usr/bin/xterm -bg RED -e ./bin/cli"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

ex-services-img-up:
	docker-compose -f ./deployments/docker-compose_only_ex_services.yaml up -d
	
ex-services-img-down:
	docker-compose -f ./deployments/docker-compose_only_ex_services.yaml down

migrate-goose:
	goose --dir=migrations mysql $(DSN) up

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/antibruteforce
	go build -v -o $(BIN_CLI) -ldflags "$(LDFLAGS)" ./cmd/cli

run: build
	$(BIN) -config ./configs/config.env > antibruteforceLog.txt 
	$(BIN_CLI) -config ./configs/config_cli.env > antibruteforceCLILog.txt 

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

stop-img: 
	docker stop $(DOCKER_IMG)

version: build
	$(BIN) version

test:
	go test -race ./internal/... 

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.41.1

lint: install-lint-deps
	golangci-lint run ./...

up:
	go build -v -o $(BIN_CLI) -ldflags "$(LDFLAGS)" ./cmd/cli && \
	gnome-terminal -- bash -c 'exec ./bin/cli -config ./configs/config_cli.env; bash' &&\
	docker-compose -f ./deployments/docker-compose.yaml up --build > deployLog.txt 
	

x:
	gnome-terminal --chmod +x $(BIN_CLI) -config ./configs/config_cli.env &
	exec chmod 777 $(BIN_CLI)
	-- bash -c $(BIN_CLI) -config ./configs/config_cli.env; bash" 

down:
	docker-compose -f ./deployments/docker-compose.yaml down

integration-tests:
	docker-compose -f ./deployments/docker-compose.yaml -f ./deployments/docker-compose.test.yaml up --build --exit-code-from integration_tests && \
	docker-compose -f ./deployments/docker-compose.yaml -f ./deployments/docker-compose.test.yaml down > deployIntegrationTestsLog.txt

.PHONY:  build run build-img ex-services-img-up run-img stop-img version test lint up down integration-tests ex-services-img-up ex-services-img-down
