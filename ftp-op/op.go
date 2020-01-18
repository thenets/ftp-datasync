package ftpop

import (
	"fmt"
	"io/ioutil"
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
	// Copy root dir
	context.copyDirContent(remoteDir, localDir)
}

// copyDirContent will check the destination path and only replace
// if the file size is different or doesn't exist
func (context *ServerContext) copyDirContent(remoteDir string, localDir string) {
	items, _ := context.conn.List(remoteDir)
	for _, item := range items {
		if item.Type == 1 {
			// Is a directory
			context.copyDirContent(
				fmt.Sprintf("%s%s/", remoteDir, item.Name),
				fmt.Sprintf("%s%s/", localDir, item.Name),
			)

		} else {
			// Is a file
			// ... so do nothing
			fmt.Println(remoteDir, item.Name)
		}

		// debug(item)
	}
}

// Test nothing but testing things...
func (context *ServerContext) Test() {
	// debug(context)
	context.downloadFile("/midgard/sample-1.txt", "./test/data")
}

func (context *ServerContext) downloadFile(remoteFilePath string, destinationLocalFilePath string) {
	// Download remote file
	res, err := context.conn.Retr(remoteFilePath)
	if err != nil {
		check(err, fmt.Sprintf("[downloadFile] Unable to download the file '%s'", remoteFilePath))
	}
	defer res.Close()

	// Write file on local storage
	buf, err := ioutil.ReadAll(res)

	err = ioutil.WriteFile("test/sample-1.txt", buf, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res)

}
