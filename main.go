package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {

	// Listening for commands in a goroutine because the
	// http server blocks the thread after it starts.
	go listenForCommands()

	server := Server{url: "127.0.0.1:8080"}
	server.Start()
}

// Funcion scans command line input and executes a specific command.
func listenForCommands() {

	input := bufio.NewScanner(os.Stdin)
	input.Scan()

	result := handleCommand(input.Text())
	fmt.Println(result)

}

// Function takes in the input from command line and executes
// a function accordingly. Returns a string with an execution status.
func handleCommand(command string) string {
	parts := strings.Split(command, " ")

	// Continue to listen for commands after execution.
	defer func() {
		go listenForCommands()
	}()

	switch parts[0] {

	case "merge":
		// Command requires a path tp the database
		if len(parts) < 2 {
			return "No destination db provided"
		}
		// see retrieve.go
		return merge(parts)
	case "initialize":
		// Command requires a path tp the database
		if len(parts) < 2 {
			return "No destination db provided"
		}
		// see db.go
		return initializeTitles(parts[1])
	default:
		return "Command not recognized"
	}
}
