#!/bin/bash

echo "export GOPATH..."
HOME=$(pwd)
cd ../../
nodepath=$(pwd)
export GOPATH=$nodepath:$GOPATH
echo "GOPATH:"$GOPATH

echo "get packages..."
#go get -v github.com/docker/engine-api/client
#go get -v github.com/docker/engine-api/types
#go get -v github.com/docker/engine-api/types/events
#go get -v github.com/docker/engine-api/types/filters
go get -v github.com/docker/docker/client
go get -v github.com/docker/docker/api/types
go get -v github.com/docker/docker/api/types/events
go get -v github.com/docker/docker/api/types/filters
go get -v github.com/influxdata/influxdb/client/v2
go get -v golang.org/x/net/context
echo "get packages finished"

echo "build..."
cd -
GIT_COMMIT=$(git rev-parse --short HEAD || echo "GitNotFound")
BUILD_TIME=$(date +%FT%T%z)
echo $GIT_COMMIT $BUILD_TIME
LDFLAGS="-X main.GitCommit=${GIT_COMMIT} -X main.BuildTime=${BUILD_TIME}"
CGO_ENABLED=0 go build -ldflags "${LDFLAGS}" -a -o ${HOME}/nctler ${HOME}/main.go

if [[ $? -ne 0 ]]; then
	#build error
	echo "build ERROR"
	exit 1
fi
