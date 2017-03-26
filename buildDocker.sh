#!/bin/sh

docker build -t roffe/k8s-sslmate .
# && docker run --rm -it --name k8s-sslmate -e SSLMATE_API_KEY="abcde" -e SSLMATE_CHECKTIME=60 roffe/k8s-sslmate