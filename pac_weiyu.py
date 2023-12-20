# Parse the gc logs in logs/gc folder and generate a plot for each process.
# The gc logs for each process is names as "<process name>.<YYYYMMDDHH24MISS>.gc.log".
# The plot is named as "<process name>.<YYYYMMDDHH24MISS>.gc.png".
# The gc logs are generated by the following JVM options:
# -XX:+PrintGCDetails -XX:+PrintGCDateStamps -Xloggc:logs/gc/<process name>.gc.log
import os
from datetime import datetime, timedelta

def get_startup_time_from_filename(log_file):
    # Extract the filename from the full path
    filename = os.path.basename(log_file)

    # Extract the timestamp from the filename
    timestamp_str = filename.split('.')[1]

    # Convert the timestamp string to a datetime object
    startup_time = datetime.strptime(timestamp_str, '%Y%m%d%H%M%S')

    return startup_time

import re
def parse_gc_logs(log_file):
    startup_time = get_startup_time_from_filename(log_file)

    with open(log_file, 'r') as file:
        log_data = file.readlines()

    # 4794.607: [GC (Allocation Failure) 4794.607: [ParNew: 415412K->3354K(460096K), 0.0101589 secs] 3673497K->3261440K(6240384K), 0.0102789 secs] [Times: user=0.05 sys=0.00, real=0.01 secs]
    # This pattern matches "3673497K->3261440K" and captures "3261440K"
    # The pattern now includes two "->" patterns
    pattern1 = re.compile(r'(\d+K)->\d+K.*(\d+K)->(\d+K)')
    pattern2 = re.compile(r'(\d+K)->(\d+K)')

    heap_sizes = []
    gc_times = []
    last_timestamp = None

    for line in log_data:
        match1 = pattern1.search(line)
        match2 = pattern2.search(line)
        if match1:
            # The second "->" match is now in groups 2 and 3
            heap_size = match1.group(3)
        elif match2:
            heap_size = match2.group(2)
        else:
            continue

        # Extract the time since startup for this GC event
        timestamp_match = re.search(r'^\d+.\d+', line)
        if timestamp_match is not None:
            gc_time = timestamp_match.group()
            gc_time = startup_time + timedelta(seconds=float(gc_time))

            # If this timestamp is within 5 minutes of the last timestamp, skip this iteration
            if last_timestamp is not None and gc_time - last_timestamp < timedelta(minutes=5):
                #logging.info(f"Timestamp {gc_time} is within 5 minutes of the last timestamp {last_timestamp}. Skipping this iteration.")
                continue

            gc_times.append(gc_time)
            last_timestamp = gc_time  # Update the last timestamp

            heap_sizes.append(heap_size)
        else:
            logging.warning(f"No match found in line: {line}")
            continue  # Skip the rest of this iteration

    return heap_sizes, gc_times

import matplotlib.pyplot as plt
import matplotlib.dates as mdates
import csv

def generate_plot(process_name, heap_sizes, gc_times, output_folder):
    # Convert heap sizes from strings to integers and round to MB
    heap_sizes = [int(size[:-1]) / 1024 for size in heap_sizes]

    plt.figure(figsize=(10, 6))
    plt.plot_date(gc_times, heap_sizes, linestyle='solid', label='Heap size after GC')
    plt.xlabel('Time')
    plt.ylabel('Heap size (MB)')  # Update the y-axis label to MB
    plt.title(f'Heap size after GC for {process_name}')
    plt.legend()
    plt.grid(True)
    plt.gca().xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m-%d %H:%M:%S'))
    plt.gca().xaxis.set_major_locator(mdates.HourLocator())  # one tick per hour
    plt.gcf().autofmt_xdate()  # rotate and align the x labels

    # Save the plot as a PNG file
    output_file = os.path.join(output_folder, f'{process_name}.gc.png')
    plt.savefig(output_file)
    plt.close()

    # Save the data to a CSV file
    csv_file = os.path.join(output_folder, f'{process_name}.gc.csv')
    with open(csv_file, 'w', newline='') as f:
        writer = csv.writer(f)
        writer.writerow(['GC Time', 'Heap Size'])
        writer.writerows(zip(gc_times, heap_sizes))

import os
import logging
from collections import defaultdict

def process_logs(log_folder, output_folder, log_file):
    # Set up logging
    logging.basicConfig(filename=log_file, level=logging.INFO)

    # Create the output folder if it doesn't exist
    os.makedirs(output_folder, exist_ok=True)

    # Group log files by process name
    log_files_by_process = defaultdict(list)
    for log_file in os.listdir(log_folder):
        process_name = log_file.split('.')[0]
        log_files_by_process[process_name].append(os.path.join(log_folder, log_file))

    # Process each group of log files
    for process_name, log_files in log_files_by_process.items():
        logging.info(f'Processing {len(log_files)} log files for process {process_name}')

        heap_sizes = []
        gc_times = []
        for log_file in log_files:
            sizes, times = parse_gc_logs(log_file)
            heap_sizes.extend(sizes)
            gc_times.extend(times)

        # Limit the size of the arrays after processing all log files for this process
        MAX_DATA_POINTS = 100
        if len(heap_sizes) > MAX_DATA_POINTS:
            heap_sizes = heap_sizes[-MAX_DATA_POINTS:]
            gc_times = gc_times[-MAX_DATA_POINTS:]

        generate_plot(process_name, heap_sizes, gc_times, output_folder)

    logging.info(f'Processed {len(log_files_by_process)} processes')

# Scrape a webpage and save the HTML to a file
# The file should be saved in the pac_weiyu folder.
# The file name should be the title of the webpage.
from bs4 import BeautifulSoup
import requests
import os
import certifi
import string

def scrape_and_save(url):
    # Make a request to the website
    cookies = {
        '_pk_id.30.f982': 'a4a4e4a268a50cc0.1698194867.3.1701159916.1701159873.',
        '_pk_id.36.f982': 'a4a4e4a268a50cc0.1694498672.20.1702951434.1702951434.',
        '_pk_id.51.f982': 'a4a4e4a268a50cc0.1695628831.0.1702290093..',
        '_mkto_trk': 'id:876-RTE-754&token:_mch-murex.com-1694393149066-26709',
        '_pk_id.43.f982': '0ee009cb96be1e2d.1700620184.2.1701153663.1701153663.',
        '_ga_CK8VHLWMNX': 'GS1.2.1697444447.2.1.1697444627.0.0.0',
        '_ga': 'GA1.2.882601883.1696919315',
        'JSESSIONID': 'E977BE97711F55EF0333384967B767E7'
        # Add more cookies if needed
    }

    r = requests.get(url, cookies=cookies, verify=False)
    
    # Parse the page content
    soup = BeautifulSoup(r.content, 'html.parser')

    # Extract the title of the webpage
    title_tag = soup.find('title')

    if title_tag is None:
        # Use the last section of the URL as the title if no title tag is found
        title = url.split('/')[-1]
    else:
        title = title_tag.text

    # Replace any characters in the title that are not valid in file names
    valid_chars = "-_.() %s%s" % (string.ascii_letters, string.digits)
    filename = ''.join(c for c in title if c in valid_chars)

    # Add the .html extension to the file name
    filename += '.html'

    # Create the pac_weiyu folder if it doesn't exist
    if not os.path.exists('pac_weiyu'):
        os.makedirs('pac_weiyu')

    # Save the page content to a file in the pac_weiyu folder
    with open(os.path.join('pac_weiyu', filename), 'w') as f:
        f.write(r.text)

def main():
    log_folder = "logs/gc"
    output_folder = "pac_weiyu"
    log_file = "pac_weiyu.log"

    # process_logs(log_folder, output_folder, log_file)

    scrape_and_save('https://mxwiki.murex.com/confluence/display/OPEV/%5BPT%5D+Oracle+Database+Servers+Support+Matrix')

if __name__ == "__main__":
    main()