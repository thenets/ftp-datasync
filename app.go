package main

import (
	"fmt"
	"os"

	ftp_op "github.com/thenets/ftp-datasync/ftp-op"
)

func main() {

	if len(os.Args) != 3 {
		fmt.Println("[ERROR] arguments not supplied!")
		fmt.Println("How to use:")
		fmt.Println("./ftpdatasync <configFilePath> <reportDestinationFilePath>")
	
		return
	}

	// Load args
	configFilePath := os.Args[1]
	reportDestinationFilePath := os.Args[2]

	// Load config and create context
	fmt.Printf("# Load config file...\n")
	context := ftp_op.ServerContext{
		ConfigFilePath: configFilePath,
	}

	// Connect
	fmt.Printf("# Connect to remote server...\n")
	context.Connect()
	defer context.Disconnect()

	// Sync remote and local dir
	fmt.Printf("# Sync remote and local dir...\n")
	context.Sync()

	// Compress
	fmt.Printf("\n# Compress...\n")
	context.Compress()

	// Create report
	fmt.Printf("\n# Generate compress report...\n")
	context.CompressCreateReport(reportDestinationFilePath)
}
