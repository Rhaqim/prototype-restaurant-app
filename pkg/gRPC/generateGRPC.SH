#!/bin/bash

# This script generates the gRPC code for the Go client and server.

# The script assumes that the protoc compiler is installed and available on the PATH.

# The script also assumes that the protoc-gen-go and protoc-gen-go-grpc plugins are installed and available on the PATH.

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative services.proto