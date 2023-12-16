#!/bin/bash

# Define the directory to search
dir="logs"

# Function to find the 10 most recently modified log files
find_latest_files() {
  find "$dir" -type f -mtime -1 -name "*.log" -print0 | xargs -0 -r ls -lt | head -n 10
}

# Call the appropriate function based on the first command line argument
if [ "$1" = "latest" ]; then
  find_latest_files
else
  echo "Invalid argument."
fi