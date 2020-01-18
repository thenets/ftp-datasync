package ftpop

import (
	"fmt"
	"log"
	"time"

	"github.com/jlaffaye/ftp"
)

// ServerContext is an abstraction of a remote FTP directory
// and all information relative to it and the FTP server.
type ServerContext struct {
	ConfigFilePath string

	hostAddress  string
	hostPort     int
	hostUser     string
	hostPassword string

	conn *ftp.ServerConn
}

// Connect starts the connection between the client and the remote server.
// It's important to always disconnect in the end.
// Example:
//   context.Disconnect()
func (context *ServerContext) Connect() {
	var err error
	
	// Load config file
	context.readConfig()

	hostFullAddress := fmt.Sprintf("%s:%d", context.hostAddress, context.hostPort)

	// TODO add timeout to connection params
	context.conn, err = ftp.Dial(hostFullAddress, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Fatal(err)
	}

	err = context.conn.Login(context.hostUser, context.hostPassword)
	if err != nil {
		log.Fatal(err)
	}
}

// Disconnect close the connection between the client and the remote server
func (context *ServerContext) Disconnect() {
	if err := context.conn.Quit(); err != nil {
		log.Fatal(err)
	}
}

// Sync sincronizes files from remote directory to the the local directory
func (context *ServerContext) Sync(remoteDir string, localDir string) {
	items, _ := context.conn.List("/")
	for _, item := range items {
		debug(item)
	}
}

// Test nothing but testing things...
func (context *ServerContext) Test() {
	debug(context)
}
