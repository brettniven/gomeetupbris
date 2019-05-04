#!/bin/sh
# local quality checks
echo "go fmt"
go fmt ./...

echo "golint"
golint -set_exit_status $(go list ./... | grep -v mocks)

echo "go vet"
go vet ./...

echo "errcheck"
errcheck -ignoretests ./...

echo "staticcheck"
staticcheck ./...

echo "gosec"
gosec -quiet $(go list ./... | grep -v mocks)

echo "gocyclo"
gocyclo -over 10 . | grep -v vendor | grep -v main.go
