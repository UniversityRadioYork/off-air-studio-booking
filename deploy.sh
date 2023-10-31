#!/usr/bin/env bash

IMAGE="evergiven.ury.york.ac.uk:5000/off-air-studio-booking"
CONTAINER="off-air-studio-booking"
ENV="/opt/off-air-studio-booking/.env"
PORT=3090

docker build -t $IMAGE .
docker push $IMAGE
docker stop $CONTAINER || echo 0
docker run -d --env-file $ENV --name $CONTAINER -p $PORT:8080 --rm $IMAGE