# wallets-task

## Getting Started

This microservice describe and realize business logic for simple wallet thought rest api interface for e-wallet handling
based on [DDD](https://dddcommunity.org) approach.

These instructions will help you to up and running service on your local machine for development and testing purposes.

### Prerequisites

**You can skip prerequisites, local setup and skip it for [docker-compose](#compose) point**

Service required [goose](https://github.com/steinbacher/goose) utility for migrations, [vegeta](https://github.com/tsenart/vegeta) for load testing

To install it automatically use:

```
$ make init
```

### Project structure

- **cmd** - contains main files for walletsvc and docgen.
- **config** - hold main app configuration and configs that mapped for [docker-compose](#compose).
- **db** - migration's sql files, initialization, config, types and models
- **docs** - project documentation.
- **domain** - domain layer source code, domain model, repository interfaces, domain services interfaces.
- **infrastructure** - infrastructure layer source code, services and gateway implementation.
- **internal** - generated mocks, internal modules
- **transport** - contains endpoints, application layer source code for http interactions
- **vendor** - contains vendored dependencies 

# Local setup 

Local stack:

- **golang go1.12.7**
- **postgresql 11.4**


### Configuration

- [config/walletsvc/config.yaml](config/walletsvc/config.yaml) - main configuration file, describes listen port & address of application 
- [db/dbconf.yaml](db/dbconf.yaml) - used to configure goose utility & app's database same time. 

### Run wallet service

After successful configuration you can run **walletscv** service by: 

```shell
$ go run ./cmd/walletsvc/main.go
```

### Api documentation 

Auto-generated api documentation stored in [docs/api.md](docs/api.md)

Project uses [chi-docgen](http://github.com/go-chi/docgen) utility to auto-generate, api documentation so you can easily regenerate it by:

```shell
$ go run ./cmd/docgen/main.go
```

#<a name="compose"></a> Docker-compose

- docker-compose version 1.24.1, build 4667896b
- docker engine 19.03.1

Docker compose stack:

- **db** - main database service, has local config [config/postgresql.conf](config/postgresql.conf)
- **nginx** - service that provides proxy & load balancer, that provides main application saleability, has local config [config/nginx.conf](config/nginx.conf)
- **app** - main service contain walletsvc microservice
- **migration** - short-live container that will migrate up your data

Optionally you can to build all related local images by running:

```shell
$ make docker-compose-build
```

This will also pass $TAG, $BRANCH and $COMMIT variables to app image. That will "sign" service binary in main app container in build phase.

 
To start up project services, and apply all migrations:

```shell
$ make docker-compose-up
```  

Shut down all services:

```shell
$ make docker-compose-down
``` 


#### Docker build

Separately you can build walletscv service container. Use this command for the auto build:
```shell
$ make build
```  

### Running the tests

```shell
$ make test
``` 
To run unit tests for all subdirectories.

```shell
$ make test-integration
```
Provides integration tests, based on parallel http servers that runs from one point, uses docker-compose system parts.

```shell
$ make docker-scale-load-test
```
For 2 nodes scale test based on vegeta util, was made for fun ;)

#### Mocking
Project uses [gomock](https://github.com/golang/mock) and [go-sqlmock](https://github.com/DATA-DOG/go-sqlmock) 

You can regenerate mocking structures by using:
```shell
$ go generate ./...
```

## Built With
- [chi framework](https://github.com/go-chi/chi)
- [viper](https://github.com/spf13/viper)
- [go-kit toolkit](https://gokit.io)
- [make](https://www.gnu.org/s/make/manual/make.html)
- [docker](https://www.docker.com)
- [docker-compose](https://docs.docker.com/compose/)

## Contributions

Please refer to each project's style and contribution guidelines for submitting patches and additions. In general, we follow the "fork-and-pull" Git workflow.

 1. **Fork** the repo on GitHub
 2. **Clone** the project to your own machine
 3. **Commit** changes to your own branch
 4. **Push** your work back up to your fork
 5. Submit a **Pull request** so that we can review your changes