package main

import (
	"fmt"

	ftp_op "github.com/thenets/ftp-datasync/ftp-op"
)

func main() {
	fmt.Println("Loading...")
	sample()
}

func sample() {
	// Load config and connect
	context := ftp_op.ServerContext{
		ConfigFilePath: "./test/server.yml",
	}
	context.Connect()
	defer context.Disconnect()

	context.Sync()
	context.Compress()
	context.CompressCreateReport("./test/report.csv")
}

