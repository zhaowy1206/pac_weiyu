package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var mu sync.Mutex
var logfile *os.File

func init() {
	var err error
	logfile, err = os.OpenFile("pac_weiyu.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: pac_weiyu.go <function> [arguments]")
		fmt.Println("\nFunctions:")
		fmt.Println("  executeAndTime <script> <times> <pacingTime>")
		fmt.Println("  retrieveStackAndPackLogFiles")
		fmt.Println("  getStack <coreFile>")
		fmt.Println("  writeStackToFile <coreFile>")
		os.Exit(1)
	}

	// Call the function based on the first argument
	switch os.Args[1] {
	case "executeAndTime":
		if len(os.Args) < 5 {
			fmt.Println("Usage: pac_weiyu.go executeAndTime <script> <times> <pacingTime>")
			os.Exit(1)
		}
		times, _ := strconv.Atoi(os.Args[3])      // Convert os.Args[3] to int
		pacingTime, _ := strconv.Atoi(os.Args[4]) // Convert os.Args[4] to int
		executeAndTime(os.Args[2], times, pacingTime)
	case "retrieveStackAndPackLogFiles":
		retrieveStackAndPackLogFiles()
	case "getStack":
		if len(os.Args) < 3 {
			fmt.Println("Usage: pac_weiyu.go getStack <coreFile>")
			os.Exit(1)
		}
		stack, err := getStack(os.Args[2])
		if err != nil {
			fmt.Printf("Failed to get stack for %s: %v\n", os.Args[2], err)
			os.Exit(1)
		}
		fmt.Printf("Stack for %s: %s\n", os.Args[2], stack)
	case "writeStackToFile":
		if len(os.Args) < 3 {
			fmt.Println("Usage: pac_weiyu.go writeStackToFile <coreFile>")
			os.Exit(1)
		}
		err := writeStackToFile(os.Args[2])
		if err != nil {
			fmt.Printf("Error writing stack to file: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Println("Unknown function:", os.Args[1])
	}

	defer logfile.Close()
}

func writeLog(logfile *os.File, message string) {
	_, err := fmt.Fprint(logfile, message)
	if err != nil {
		fmt.Printf("Error writing to log file: %v\n", err)
	}
}

//	List all the files that are 10 latest updated in a directory,
//
// let the user choose which ones to monitor,
// and then tail those files in real-time.
func monitorLogs() {
	dir := "logs"
	_, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	type FileInfo struct {
		Path    string
		ModTime time.Time
	}

	var logFiles []FileInfo

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Ignore directories
		if !info.IsDir() {
			logFiles = append(logFiles, FileInfo{Path: path, ModTime: info.ModTime()})
		}

		return nil
	})

	// Sort the files by modification time
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].ModTime.Before(logFiles[j].ModTime)
	})

	// If there are more than 10 files, take the last 10
	if len(logFiles) > 10 {
		logFiles = logFiles[len(logFiles)-10:]
	}

	// Automatically tail the 10 latest files
	for _, file := range logFiles {
		go tailFile(file.Path)
		// fmt.Printf("Path: %s, Modification time: %s\n", file.Path, file.ModTime)
	}

	// Only block if at least one tailFile goroutine was started
	if len(logFiles) > 0 {
		// Prevent the program from exiting
		select {}
	}
}

func mustParseInt(s string) int {
	i, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return -1
	}
	return i
}

func tailFile(filename string) {
	cmd := exec.Command("tail", "-f", filename)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Error creating stdout pipe:", err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("Error creating stderr pipe:", err)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting command:", err)
		return
	}

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Println("Error output from tail command:", scanner.Text())
		}
	}()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Printf("%s: %s\n", filename, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading command output:", err)
	}

	if err := cmd.Wait(); err != nil {
		fmt.Println("Error waiting for command:", err)
	}
}

func executeAndTime(script string, times int, pacingTime int) {
	for i := 1; i <= times; i++ {
		start := time.Now()
		message := fmt.Sprintf("Starting %s run %d at %s\n", script, i, start)
		writeLog(logfile, message)

		cmd := exec.Command("bash", script)
		err := cmd.Run()
		if err != nil {
			message := fmt.Sprintf("%s failed with exit status %v\n", script, err)
			writeLog(logfile, message)
			break
		}

		end := time.Now()
		duration := end.Sub(start)
		message = fmt.Sprintf("Execution time for %s run %d: %v seconds\n", script, i, duration.Seconds())
		writeLog(logfile, message)
		message = fmt.Sprintf("Ended %s run %d at %s\n", script, i, end)
		writeLog(logfile, message)

		time.Sleep(time.Duration(pacingTime) * time.Second)
	}
}

func getStack(coreFile string) ([]byte, error) {
	cmd := exec.Command("./pmx", "-e", coreFile)
	return cmd.Output()
}

func getPid(coreFile string) string {
	pid := strings.Split(coreFile, ".")[1]
	return pid
}

func writeStackToFile(coreFile string) error {
	pid := getPid(coreFile)
	stack, err := getStack(coreFile)
	if err != nil {
		writeLog(logfile, fmt.Sprintf("Failed to get stack for %s: %v\n", coreFile, err))
		return fmt.Errorf("Failed to get stack for %s: %v", coreFile, err)
	}

	stackFileName := fmt.Sprintf("stack.%s", pid)
	if len(stack) > 0 && pid != "" {
		err := ioutil.WriteFile(stackFileName, stack, 0644)
		if err != nil {
			writeLog(logfile, fmt.Sprintf("Failed to write to %s: %v\n", stackFileName, err))
			return fmt.Errorf("Failed to write to %s: %v", stackFileName, err)
		}
		writeLog(logfile, fmt.Sprintf("Successfully wrote to %s\n", stackFileName))
	} else {
		if len(stack) == 0 {
			writeLog(logfile, "No data to write.\n")
		}
		if pid == "" {
			writeLog(logfile, "Invalid PID.\n")
		}
	}
	return nil
}

func findLogFiles(pid string) ([]string, error) {
	var logFiles []string
	err := filepath.Walk("logs", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.Contains(path, pid) {
			logFiles = append(logFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return logFiles, nil
}

func retrieveStackAndPackLogFiles() {
	coreFiles, _ := filepath.Glob("core.*")
	for _, coreFile := range coreFiles {
		pid := getPid(coreFile)
		message := fmt.Sprintf("Retrieving stack and packing log files for core file %s\n", coreFile)
		writeLog(logfile, message)

		err := writeStackToFile(coreFile)
		if err != nil {
			fmt.Println(err)
		}

		logFiles, err := findLogFiles(pid)
		if err != nil {
			writeLog(logfile, fmt.Sprintf("Failed to find log files: %v\n", err))
			return
		}
		writeLog(logfile, fmt.Sprintf("Log files: %s\n", strings.Join(logFiles, ", ")))

		args := append([]string{"zip", "-r", fmt.Sprintf("stack_and_log_%s.zip", pid), fmt.Sprintf("stack.%s", pid)}, logFiles...)
		cmd := exec.Command(args[0], args[1:]...)

		writeLog(logfile, fmt.Sprintf("Running command: %s\n", strings.Join(cmd.Args, " ")))
		cmdOutput, err := cmd.CombinedOutput()
		if err != nil {
			writeLog(logfile, fmt.Sprintf("Failed to zip files: %v\n", err))
		} else {
			writeLog(logfile, fmt.Sprintf("Zip command output: %s\n", cmdOutput))
		}

		os.Remove(fmt.Sprintf("stack.%s", pid))
	}

	files, err := filepath.Glob("stack_and_log_*.zip")
	if err != nil {
		writeLog(logfile, fmt.Sprintf("Failed to find files: %v\n", err))
		return
	}

	if len(files) == 0 {
		writeLog(logfile, "No files match the pattern stack_and_log_*.zip\n")
		return
	}

	writeLog(logfile, fmt.Sprintf("%d files match the pattern stack_and_log_*.zip\n", len(files)))

	finalZipFile := fmt.Sprintf("final_stack_and_log_%s.zip", time.Now().Format("20060102_150405"))
	cmd := exec.Command("zip", "-r", finalZipFile)
	cmd.Args = append(cmd.Args, files...)

	writeLog(logfile, fmt.Sprintf("Running command: %s\n", strings.Join(cmd.Args, " ")))
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		writeLog(logfile, fmt.Sprintf("Failed to create final zip file: %v\n", err))
	} else {
		writeLog(logfile, fmt.Sprintf("Successfully created final zip file: %s\n", finalZipFile))
		writeLog(logfile, fmt.Sprintf("Zip command output: %s\n", cmdOutput))
	}

	files, _ = filepath.Glob("stack_and_log_*.zip")
	for _, file := range files {
		os.Remove(file)
	}
}
