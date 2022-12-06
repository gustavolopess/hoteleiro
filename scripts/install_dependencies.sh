#!/bin/bash
# for AWS Linux
sudo yum install golang -y

# Get the PID of go running process
pid=$(pgrep go)

# Check if the process is running
if [ -n "$pid" ]; then
  # Kill the process
  kill $pid
else
  echo "Process not found"
fi

