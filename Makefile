BRANCH := $(shell git branch | grep \* | cut -d ' ' -f2)
COMMIT := $(shell git rev-parse HEAD)
TAG := $(shell git describe --abbrev=0 --tags)

init:
	GO111MODULE=off go get -u github.com/tsenart/vegeta
	GO111MODULE=off go get github.com/steinbacher/goose/cmd/goose

build:
	docker build --rm -t wallet-svc:$(BRANCH)-$(TAG)-$(COMMIT) --build-arg TAG=$(TAG) --build-arg BRANCH=$(BRANCH) --build-arg COMMIT=$(COMMIT) .

docker-compose-build:
	docker-compose build --force-rm --build-arg TAG=$(TAG) --build-arg BRANCH=$(BRANCH) --build-arg COMMIT=$(COMMIT)

docker-compose-up:
	docker-compose up -d

docker-compose-down:
	docker-compose down

docker-compose-migrate:
	docker-compose up migration

test:
	go test -v -race ./...

test-integration:
	docker-compose up -d db
	go test -v -race ./... -tags integration -count 1
	docker-compose down

docker-scale-load-test:
	docker-compose up -d --scale app=2
	jq -ncM '{method: "POST", url: "http://localhost:8080/wallet/pay/bob456/alice123", body: "{\"amount\": 1}" | @base64, header: {"Content-Type": ["text/plain"]}}' | vegeta attack -duration=20s -format=json -rate=50 | vegeta report
	jq -ncM '{method: "POST", url: "http://localhost:8080/wallet/pay/alice123/bob456", body: "{\"amount\": 1}" | @base64, header: {"Content-Type": ["text/plain"]}}' | vegeta attack -duration=20s -format=json -rate=50 | vegeta report
	docker-compose down


