#!/bin/bash
set -e

sudo docker run -d \
	-p $HOST_PORT:$CONTAINER_PORT \
	-v /tmp/$CONFIG:/etc/example.cfg \
	--restart=always \
	--name $NAME $IMAGE
