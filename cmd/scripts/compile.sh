#!/bin/sh
GOOS=windows GOARCH=amd64 go build -o bin/xltemplate.exe xltemplate.go
GOOS=darwin GOARCH=amd64 go build -o bin/xltempate-mac xltemplate.go
go build -o bin/xltemplate xltemplate.go