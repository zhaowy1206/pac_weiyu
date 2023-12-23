package main

import (
    "bufio"
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "strconv"
    "strings"
	"sort"
)

func main() {
    monitorLogs("./logs")
}

//  List all the files that are 10 latest updated in a directory, 
// let the user choose which ones to monitor, 
// and then tail those files in real-time. 
func monitorLogs(dir string) {
    files, err := ioutil.ReadDir(dir)
    if err != nil {
        fmt.Println("Error reading directory:", err)
        return
    }

    // Sort the files by modification time
    sort.Slice(files, func(i, j int) bool {
        return files[i].ModTime().Before(files[j].ModTime())
    })

    // If there are more than 10 files, take the last 10
    if len(files) > 10 {
        files = files[len(files)-10:]
    }

    fmt.Println("Select files to monitor (separated by comma):")
    for i, file := range files {
        fmt.Printf("%d: %s\n", i+1, file.Name())
    }

    reader := bufio.NewReader(os.Stdin)
    input, _ := reader.ReadString('\n')
    input = strings.TrimSpace(input)
    selected := strings.Split(input, ",")

    for _, s := range selected {
        index := mustParseInt(s) - 1
        if index >= 0 && index < len(files) {
            go tailFile(dir + "/" + files[index].Name())
        }
    }

    // Prevent the program from exiting
    select {}
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
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    _ = cmd.Run()
}