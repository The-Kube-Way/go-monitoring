#!/bin/bash

set -x
set -e

IMAGE_NAME="go-monitoring:ci"
CONTAINER_NAME="go-monitoring"

docker kill $CONTAINER_NAME || true

docker build -t $IMAGE_NAME .

docker run -d --name $CONTAINER_NAME --rm -p 8080:8080 -v $(pwd)/tests/e2e/config/:/config $IMAGE_NAME -debug

sleep 10

docker logs $CONTAINER_NAME

METRICS=$(curl -s http://localhost:8080/metrics | grep go_monitoring_up)

grep "go_monitoring_up{id=\"github.com:22\",name=\"\",probe=\"raw_tcp\"} 1" <<< "$METRICS"
grep "go_monitoring_up{id=\"127.0.0.1\",name=\"\",probe=\"ping\"} 1" <<< "$METRICS"
grep "go_monitoring_up{id=\"1.2.3.4\",name=\"\",probe=\"ping\"} 0" <<< "$METRICS"
grep "go_monitoring_up{id=\"https://google.com\",name=\"test_200\",probe=\"http\"} 1" <<< "$METRICS"
grep "go_monitoring_up{id=\"https://httpstat.us/403\",name=\"\",probe=\"http\"} 0" <<< "$METRICS"
grep "go_monitoring_up{id=\"https://httpstat.us/404\",name=\"\",probe=\"http\"} 1" <<< "$METRICS"

echo "Tests passed!"

docker kill $CONTAINER_NAME || true
