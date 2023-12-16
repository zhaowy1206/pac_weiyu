#!/bin/bash

# Define the directory to search
dir="logs"

# Use the find command to recursively search the directory
# Then, use ls to list the files sorted by modification time
# Finally, use head to get the 10 most recently modified files
find "$dir" -type f -name "*.log" -print0 | xargs -0 ls -lt | head -n 10


