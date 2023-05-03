BIN := "./bin/calendar"
SCHEDULER_BIN := "./bin/calendar_scheduler"
SENDER_BIN := "./bin/calendar_sender"
DOCKER_IMG="calendar:develop"
DOCKER_IMG_SHEDULER="sheduler:develop"
DOCKER_IMG_SENDER="sender:develop"
DSN="imapp:LightInDark@/OTUSFinalLab?parseTime=true"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

ex-services-img-up:
	docker-compose -f ./deployments/docker-compose_only_ex_services.yaml up -d
	
ex-services-img-down:
	docker-compose -f ./deployments/docker-compose_only_ex_services.yaml down

migrate-goose:
	goose --dir=migrations mysql $(DSN) up

generate:
	rm -rf internal/server/grpc/pb
	mkdir -p internal/server/grpc/pb
	protoc api/EventService.proto  --go_out=internal/server/grpc/pb --go-grpc_out=internal/server/grpc/pb 

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar
	go build -v -o $(SCHEDULER_BIN) -ldflags "$(LDFLAGS)" ./cmd/sheduler
	go build -v -o $(SENDER_BIN) -ldflags "$(LDFLAGS)" ./cmd/sender

run: build
	$(BIN) -config ./configs/config.env > calendarCLog.txt &
	$(SCHEDULER_BIN) -config ./configs/config_sheduler.env > shedulerClog.txt &
	$(SENDER_BIN)  -config ./configs/config_sender.env > senderClog.txt

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG_SHEDULER) \
		-f build/sheduler/Dockerfile .
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG_SENDER) \
		-f build/sender/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)
	docker run $(DOCKER_IMG_SHEDULER)
	docker run $(DOCKER_IMG_SENDER)

stop-img: 
	docker stop $(DOCKER_IMG)
	docker stop $(DOCKER_IMG_SHEDULER)
	docker stop $(DOCKER_IMG_SENDER)

version: build
	$(BIN) version

test:
	go test -race ./internal/... 

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.41.1

lint: install-lint-deps
	golangci-lint run ./...

up:
	docker-compose -f ./deployments/docker-compose.yaml up --build > deployLog.txt

down:
	docker-compose -f ./deployments/docker-compose.yaml down

integration-tests:
	docker-compose -f ./deployments/docker-compose.yaml -f ./deployments/docker-compose.test.yaml up --build --exit-code-from integration_tests && \
	docker-compose -f ./deployments/docker-compose.yaml -f ./deployments/docker-compose.test.yaml down > deployIntegrationTestsLog.txt

.PHONY: generate build run build-img ex-services-img-up run-img stop-img version test lint up down integration-tests ex-services-img-up ex-services-img-down
