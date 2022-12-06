#!/bin/bash
cd /app/
nohup go run cmd/main.go >/dev/null 2>&1 &
