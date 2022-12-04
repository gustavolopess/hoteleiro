#!/bin/bash
sudo mkdir -p /app/
cd /app/
go mod download
go build -o main cmd/main.go