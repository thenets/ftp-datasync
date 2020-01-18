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
	// TODO load connection settings from config file

	context := ftp_op.ServerContext{
		ConfigFilePath: "./test/server.yml",
	}

	// context.Test()

	context.Connect()
	context.Sync("/", "./test/data")
	context.Disconnect()

}
