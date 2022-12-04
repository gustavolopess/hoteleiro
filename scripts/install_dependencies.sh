#!/bin/bash
# for AWS Linux
sudo yum install golang -y
cd /app/
go mod download
go build -o main cmd/main.go
