#!/bin/bash

# make pushd and popd silent
pushd () { command pushd "$@" > /dev/null ; }
popd () { command popd "$@" > /dev/null ; }

echo "Generating models..."
pushd internal/models
    # without removing the templates first, xo_db.go.go will never be regenerated
    rm -rf *.xo.go
    xo pgsql://colin@127.0.0.1/twitch?sslmode=disable -o . --template-path templates/

    pushd schema
        pg_dump -h localhost \
        -f dump.sql \
        twitch
    popd
popd

echo "Generating protobuf files..."
pushd pb
    protoc --gogofaster_out=plugins=grpc:. *.proto
popd
