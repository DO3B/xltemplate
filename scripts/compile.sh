#!/bin/sh
GOOS=windows GOARCH=amd64 go build -o $OUTPUT_NAME/xltemplate.exe lma/xltools/xltemplate-cli
GOOS=darwin GOARCH=arm64 go build -o $OUTPUT_NAME/xltemplate-mac lma/xltools/xltemplate-cli
go build -o $OUTPUT_NAME/xltemplate lma/xltools/xltemplate-cli