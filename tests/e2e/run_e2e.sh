#!/bin/bash

set -x
set -e

IMAGE_NAME="go-monitoring:ci"
CONTAINER_NAME="go-monitoring"

docker kill $CONTAINER_NAME || true

docker build -t $IMAGE_NAME .

docker run -d --name $CONTAINER_NAME --rm -p 8080:8080 -v $(pwd)/tests/e2e/config/:/config $IMAGE_NAME

sleep 10

docker logs $CONTAINER_NAME

METRICS=$(curl -s http://localhost:8080/metrics | grep go_monitoring_up)

grep "go_monitoring_up{id=\"github.com:22\",probe=\"raw_tcp\"} 1" <<< "$METRICS"
grep "go_monitoring_up{id=\"google.com\",probe=\"ping\"} 1" <<< "$METRICS"
grep "go_monitoring_up{id=\"https://google.com\",probe=\"http\"} 1" <<< "$METRICS"
grep "go_monitoring_up{id=\"https://httpstat.us/403\",probe=\"http\"} 0" <<< "$METRICS"

echo "Tests passed!"

docker kill $CONTAINER_NAME || true
