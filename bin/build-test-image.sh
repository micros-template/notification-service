#!/bin/sh

mkdir -p ./bin/dist

service_name="notification_service"
prefix_image="docker-registry.anandadf.my.id/micros-template/"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "./bin/dist/$service_name" ./cmd
wait

echo "Building Docker image for $prefix_image$service_name:test" >/dev/stderr
docker build -t "$prefix_image$service_name:test" --build-arg BIN_NAME=$service_name -f Dockerfile .

rm -rf ./bin/dist