#!/bin/bash

go test ./pkg/workflow -cover -gcflags=-l -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
