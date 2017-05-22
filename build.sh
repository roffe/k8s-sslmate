#!/bin/sh
mkdir bin
go get
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/k8s-sslmate .
cd bin
zip k8s-sslmate.zip k8s-sslmate
scp -4 k8s-sslmate.zip roffe@roffe.nu:/webroot/roffe/public_html/k8s-sslmate/
