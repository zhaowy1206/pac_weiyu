package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	_ "github.com/godror/godror"
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
		"monitorLogs":     monitorLogs,
		"monitorSessions": monitorSessions,
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

// Create a plot in the console view, and feed it with the data retrieved from a Oracle database to monitor how many sessions are there in realtime.
func monitorSessions() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	p := widgets.NewPlot()
	p.Title = "Oracle Sessions"
	p.Data = make([][]float64, 1)
	p.Data[0] = make([]float64, 0)
	p.SetRect(0, 0, 50, 25)
	p.AxesColor = ui.ColorWhite
	p.LineColors[0] = ui.ColorGreen

	// Initialize p.Data[0] with a dummy data point
	p.Data[0] = append(p.Data[0], 0)

	//ui.Render(p)

	fmt.Print("Enter IP address: ")
	var ip string
	fmt.Scanln(&ip)

	fmt.Print("Enter port: ")
	var port string
	fmt.Scanln(&port)

	fmt.Print("Enter username: ")
	var username string
	fmt.Scanln(&username)

	fmt.Print("Enter password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal(err)
	}
	password := string(bytePassword)

	connStr := fmt.Sprintf("%s/%s@%s:%s/sid", username, password, ip, port)

	db, err := sql.Open("godror", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			var count float64
			err := db.QueryRow("SELECT COUNT(*) FROM v$session").Scan(&count)
			if err != nil {
				log.Fatal(err)
			}

			mu.Lock() // Lock the mutex before modifying p.Data
			// Remove the dummy data point after you've added the first real data point
			if len(p.Data[0]) == 1 && p.Data[0][0] == 0 {
				p.Data[0] = p.Data[0][1:]
			}

			p.Data[0] = append(p.Data[0], count)
			if len(p.Data[0]) > 50 {
				p.Data[0] = p.Data[0][1:]
			}
			mu.Unlock() // Unlock the mutex after modifying p.Data

			mu.Lock() // Lock the mutex before rendering
			ui.Render(p)
			mu.Unlock() // Unlock the mutex after rendering
		}
	}
}
