#!/bin/bash

COMMAND=$1
PORT=$2

if [ -z "${PORT}" ]; then
    PORT=5432
fi

export GOOSE_DRIVER="postgres"
export GOOSE_DBSTRING="postgres://user:password@localhost:$PORT/meteogo-weather-collector-service-db?sslmode=disable"

goose -dir migrations $COMMAND
