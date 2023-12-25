package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

var mu sync.Mutex

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./pac_weiyu <function_name> <arg>")
		os.Exit(1)
	}

	funcName := os.Args[1]

	// Map function names to function calls
	funcs := map[string]func(){
		"monitorLogs":            monitorLogs,
		"TestDatabaseConnection": TestDatabaseConnection,
		// Add more functions here
	}

	// Call the function if it exists in the map
	if funcCall, ok := funcs[funcName]; ok {
		funcCall()
	} else {
		fmt.Printf("Unknown function: %s\n", funcName)
		os.Exit(1)
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

func gatherConnectionInfo(r io.Reader) (string, string, string, string, string, error) {
	reader := bufio.NewReader(r)

	fmt.Print("Enter IP address: ")
	ip, err := reader.ReadString('\n')
	if err != nil {
		return "", "", "", "", "", err
	}

	fmt.Print("Enter port: ")
	port, err := reader.ReadString('\n')
	if err != nil {
		return "", "", "", "", "", err
	}

	fmt.Print("Enter service name: ")
	serviceName, err := reader.ReadString('\n')
	if err != nil {
		return "", "", "", "", "", err
	}

	fmt.Print("Enter username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", "", "", "", "", err
	}

	fmt.Print("Enter password: ")
	passwordBytes, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", "", "", "", "", err
	}
	password := string(passwordBytes)
	fmt.Println() // Print a newline after the password input

	return strings.TrimSpace(ip), strings.TrimSpace(port), strings.TrimSpace(username), strings.TrimSpace(password), strings.TrimSpace(serviceName), nil
}

func TestDatabaseConnection() {
	ip, port, username, password, servicename, err := gatherConnectionInfo(os.Stdin)
	if err != nil {
		log.Fatalf("Error gathering connection info: %v", err)
	}

	connStr := fmt.Sprintf("%s/%s@%s:%s/%s", username, password, ip, port, servicename)

	// Attempt to connect to the database
	db, err := sql.Open("godror", connStr)
	if err != nil {
		fmt.Println("Failed to connect to the database: %v", err)
	}

	var count float64
	err = db.QueryRow("SELECT COUNT(*) FROM v$session").Scan(&count)
	if err != nil {
		fmt.Printf("Failed to run the SQL: %v\n", err)
		fmt.Printf("The connStr was: %v\n", connStr)
	} else {
		fmt.Printf("The number of sessions is: %f\n", count)
	}

	// Close the connection
	err = db.Close()
	if err != nil {
		fmt.Println("Failed to close the database connection: %v", err)
	}
}
