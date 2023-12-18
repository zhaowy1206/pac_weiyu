#!/bin/bash

# Function to call a script and record the execution time
execute_and_time() {
  start=$(date +%s)
  bash "$1"
  end=$(date +%s)
  duration=$((end - start))
  echo "Execution time for $1: $duration seconds"
}

