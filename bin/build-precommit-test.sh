#!/bin/sh

mkdir -p ./bin/dist

service_name="notification_service"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o "./bin/dist/$service_name" ./cmd
wait

if ! command -v upx >/dev/null 2>&1; then
  echo "UPX not found. Installing..."
  apk update && apk add upx
fi

upx --best --lzma ./bin/dist/$service_name
wait


echo "Building Docker image for $service_name:test" >/dev/stderr
docker build -t "10.1.20.130:5001/dropping/notification-service:test" --build-arg BIN_NAME=$service_name -f Dockerfile .

rm -rf ./bin/dist