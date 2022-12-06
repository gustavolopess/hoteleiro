#!/bin/bash

# Get the PID of the process
pid=$(pgrep go)

# Check if the process is running
if [ -n "$pid" ]; then
  # Kill the process
  kill $pid
else
  echo "Process not found"
fi