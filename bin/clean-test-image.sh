#!/bin/sh

service_name="notification_service"
prefix_image="docker-registry.anandadf.my.id/micros-template/"

echo "Removing test in local" >/dev/stderr
docker rmi "$prefix_image$service_name:test"
