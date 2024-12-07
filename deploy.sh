#!/usr/bin/env bash

IMAGE="evergiven.ury.york.ac.uk:5000/off-air-studio-booking"
CONTAINER="off-air-studio-booking"
PROJECTDIR="/opt/off-air-studio-booking"
PORT=3090
DATE=$(date +%s)

docker build -t $IMAGE:$DATE .
docker push $IMAGE:$DATE
docker service update --image $IMAGE:$DATE off-air-bookings