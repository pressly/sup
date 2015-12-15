#!/bin/bash
set -e

cd /tmp/$IMAGE

# Cleanup.
sudo rm -rf bin

# Bulder image. Build binaries (make dist) into bin/ dir.
sudo docker run --rm \
	-v $(pwd):/go/src/$REPO/$NAME \
	-w /go/src/$REPO/$NAME \
	golang:1.5.2 go build

# Bake bin/* into the resulting image.
sudo docker build --no-cache -t $IMAGE .

sudo docker push $IMAGE
