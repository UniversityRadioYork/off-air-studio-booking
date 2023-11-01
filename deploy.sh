#!/usr/bin/env bash

IMAGE="evergiven.ury.york.ac.uk:5000/off-air-studio-booking"
CONTAINER="off-air-studio-booking"
PROJECTDIR="/opt/off-air-studio-booking"
PORT=3090

docker build -t $IMAGE .
docker push $IMAGE
docker stop $CONTAINER || echo 0
docker rm $CONTAINER || echo 0
docker run -d --env-file $PROJECTDIR/.env --name $CONTAINER -p $PORT:8080 -v $PROJECTDIR/.myradio.key:/usr/src/app/.myradio.key $IMAGE