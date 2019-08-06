FROM golang:1.12.6 AS build
ARG BRANCH
ARG COMMIT
ARG TAG
COPY . /app
WORKDIR /app
RUN curl -o ./ca-certificates.crt https://raw.githubusercontent.com/bagder/ca-bundle/master/ca-bundle.crt
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o walletsvc -ldflags="-X main.branch=${BRANCH} -X main.tag=${TAG} -X main.commit=${COMMIT} -s -w" ./cmd/walletsvc/main.go
RUN chmod +x ./walletsvc

FROM scratch
WORKDIR /app/
COPY --from=build /app/config/walletsvc/config.yaml ./config/walletsvc/config.yaml
COPY --from=build /app/db/dbconf.yaml ./db/dbconf.yaml
COPY --from=build /app/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /app/walletsvc ./walletsvc

ENTRYPOINT ["/app/walletsvc"]