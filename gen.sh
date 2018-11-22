#!/bin/bash

# make pushd and popd silent
pushd () { command pushd "$@" > /dev/null ; }
popd () { command popd "$@" > /dev/null ; }

echo "Generating models..."
pushd internal/models
    gnorm gen

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
