#!/bin/bash

platforms = ("linux/amd64" "linux/arm" "linux/arm64" "darwin/amd64" "darwin/arm" "darwin/arm64")

for platform in "${platforms[@]}"
do
    platform_split = (${platform//\// })
    GOOS = ${platform_split[0]}
    GOARCH = ${platform_split[1]}
    go build -ldflags "-X github.com/arkenproject/ait/apis/github.clientID=$1 -X github.com/arkenproject/ait/cli.appVersion=$2" -o ait-v$2-${GOOS}-${GOARCH} .

done