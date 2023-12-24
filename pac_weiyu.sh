#!/bin/bash

logfile="pac_weiyu.log"
# Function to call a script and record the execution time
execute_and_time() {
  script=$1
  times=$2
  pacing_time=$3
  for ((i=1; i<=times; i++))
  do
    start=$(date +%s)
    echo "Starting $script run $i at $(date -u -d @$start)" >> "$logfile"
    bash "$script" > /dev/null 2>&1
    script_status=$?
    if [ $script_status -ne 0 ]; then
      echo "$script failed with exit status $script_status" >> "$logfile"
      break
    fi
    end=$(date +%s)
    duration=$((end - start))
    echo "Execution time for $script run $i: $duration seconds" >> "$logfile"
    echo "Ended $script run $i at $(date -u -d @$end)" >> "$logfile"
    sleep $pacing_time
  done
}

# Python script for plotting execution times
# import pandas as pd
# import matplotlib.pyplot as plt
# import re
#
# execution_times = []
# with open('pac_weiyu.log', 'r') as f:
#     for line in f:
#         match = re.search(r'Execution time for test.sh run (\d+): (\d+) seconds', line)
#         if match:
#             run_number = int(match.group(1))
#             execution_time = int(match.group(2))
#             execution_times.append((run_number, execution_time))
#
# df = pd.DataFrame(execution_times, columns=['Run', 'Execution Time'])
#
# plt.figure(figsize=(10, 6))
# plt.plot(df['Run'], df['Execution Time'])
# plt.xlabel('Run')
# plt.ylabel('Execution Time (seconds)')
# plt.title('Execution Time for Each Run of test.sh')
# plt.grid(True)
# plt.show()

# Rtrieve the stack from the core dump files, determin the process id associated with the core files 
# and find related log files that the file names containing the process id under the logs folder.
# The core dump files are named with core.<pid>.
# Use the command "./pmx -e core.<pid>" to retrieve the parameters and the stack.
# Then pack the stack file and log files into a zip file for each core dump file.
# Finally, add the zip files into one zip file named with stack_and_log_<YYYYMMDD_HH24MISS>.zip 
# and remove the stack files and zip files of each core dump.
retrieve_stack_and_pack_log_files() {
  core_files=$(ls core.*)
  for core_file in $core_files
  do
    pid=$(echo $core_file | cut -d '.' -f 2)
    echo "Retrieving stack and packing log files for core file $core_file" >> "$logfile"
    ./pmx -e $core_file > stack.$pid
    log_files=$(find logs -name "*$pid*")
    # Redirect the output of the zip command to the script log file
    zip -r stack_and_log_$pid.zip stack.$pid $log_files >> "$logfile"
    rm stack.$pid
  done
  final_zip_file="final_stack_and_log_$(date +%Y%m%d_%H%M%S).zip"
  zip -r $final_zip_file stack_and_log_*.zip >> "$logfile"
  rm stack_and_log_*.zip
}

# Call the function based on the first argument
case $1 in
  execute_and_time)
    shift # Remove the first argument
    execute_and_time "$@"
    ;;
  retrieve_stack_and_pack_log_files)
    shift # Remove the first argument
    retrieve_stack_and_pack_log_files "$@"
    ;;
  *)
    echo "Unknown function: $1"
    ;;
esac