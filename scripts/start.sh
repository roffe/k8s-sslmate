#!/bin/bash

function shut_down() {
    echo shutting down k8s-sslmate
    exit

}

if [ -z "${SSLMATE_API_KEY}" ];	then
    echo "Could not get env SSLMATE_API_KEY"
    exit 1
else
    echo "api_key ${SSLMATE_API_KEY}" > /root/.sslmate
    echo "key_directory /etc/sslmate/keys" >> /root/.sslmate
    echo "cert_directory /etc/sslmate" >> /root/.sslmate
    echo "wildcard_filename star" >> /root/.sslmate
fi

trap "shut_down" SIGKILL SIGTERM SIGHUP SIGINT EXIT

/opt/bin/k8s-sslmate

