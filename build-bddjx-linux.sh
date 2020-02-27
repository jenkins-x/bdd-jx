#!/bin/sh

export GOOS=linux
export GOARCH=arm

go test  -v -timeout 2h -c github.com/jenkins-x/bdd-jx/test/suite/quickstart -o build/bddjx-$GOOS  -ldflags """"
