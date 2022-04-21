#!/bin/bash

ginkgo -cover pkg/workflow
go tool cover -html=./pkg/workflow/workflow.coverprofile -o coverage.html
rm ./pkg/workflow/workflow.coverprofile
