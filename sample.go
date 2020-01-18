package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jlaffaye/ftp"
)

// ServerContext is an abstraction of a remote FTP directory
// and all information relative to it and the FTP server.
type ServerContext struct {
	HostAddress  string
	HostPort     int
	HostUser     string
	HostPassword string

	Conn *ftp.ServerConn
}

// Connect starts the connection between the client and the remote server.
// It's important to always disconnect in the end.
// Example:
//   context.Disconnect()
func (context ServerContext) Connect() {
	// TODO validate connection

	hostFullAddress := fmt.Sprintf("%s:%d", context.HostAddress, context.HostPort)

	// TODO add timeout to connection params
	c, err := ftp.Dial(hostFullAddress, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Fatal(err)
	}

	err = c.Login(context.HostUser, context.HostPassword)
	if err != nil {
		log.Fatal(err)
	}

	context.Conn = c
}

// Disconnect close the connection between the client and the remote server
func (context ServerContext) Disconnect() {
	if err := context.Conn.Quit(); err != nil {
		log.Fatal(err)
	}
}

// func sync(string path) {

// }

func sample() {
	// TODO load connection settings from config file

	context := ServerContext{
	}

	context.Connect()

	// List dir
	dirList, err := context.Conn.List("/")
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range dirList {
		fmt.Printf("Type: %d, %s\n", item.Type, item.Name)
	}

	context.Disconnect()

}
