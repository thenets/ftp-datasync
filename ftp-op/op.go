package ftpop

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
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
	items, err := context.conn.List(remoteDir)
	if err != nil {
		check(err, fmt.Sprintf("[copyDirContent] Can't list remoteDir '%s'", remoteDir))
	}

	for _, item := range items {

		if item.Type == 1 {
			// Recursive call if is a directory
			context.copyDirContent(
				fmt.Sprintf("%s/%s", remoteDir, item.Name),
				fmt.Sprintf("%s/%s", localDir, item.Name),
			)

		} else {
			// Download file if the remote and local file aren't equal
			// or local file doesn't exist
			remoteFilePath := fmt.Sprintf("%s/%s", remoteDir, item.Name)
			destinationLocalFilePath := fmt.Sprintf("%s/%s", localDir, item.Name)
			if context.fileHasChange(item, destinationLocalFilePath) {
				fmt.Println("Downloading file to...", destinationLocalFilePath)
				// Create dir if not exist
				ensureDirExist(localDir)

				// Download file
				context.downloadFile(item, remoteFilePath, destinationLocalFilePath)
			} else {
				fmt.Println("File already exist. Skipping...", destinationLocalFilePath)
			}
			// debug(item)
			// fmt.Println(remoteDir, item.Name)

		}
	}
}

// fileHasChange returns 'true' if the has change between remote and local file
// and return false if files are equal.
func (context *ServerContext) fileHasChange(remoteEntry *ftp.Entry, destinationLocalFilePath string) bool {
	// Check if file already exist
	if checkLocalFileExists(destinationLocalFilePath) {
		// Check if file size is equal
		sizeIsEqual := bool(remoteEntry.Size == getLocalFileSize(destinationLocalFilePath))

		// Check if createAt datetime is equal
		fileStat, _ := os.Stat(destinationLocalFilePath)
		modTimeIsEqual := remoteEntry.Time.Equal(fileStat.ModTime())

		if sizeIsEqual && modTimeIsEqual {
			return false
		}

	} else {
		// File not exist so change
		return true
	}

	return true
}

// Test nothing but testing things...
func (context *ServerContext) Test() {

	// debug(context)

	// Check file createAt time
	// fileStat, _ := os.Stat("test/sample-1.txt")
	// debug(fileStat.ModTime().Unix())

	// fmt.Println(context.conn.FileSize("/midgard/sample-1.txt"))
	// fmt.Println(getLocalFileSize("test/sample-1.txt"))

	// context.downloadFile("/midgard/sample-1.txt", "./test/data")
}

func (context *ServerContext) downloadFile(remoteEntry *ftp.Entry, remoteFilePath string, destinationLocalFilePath string) {
	// Download remote file
	res, err := context.conn.Retr(remoteFilePath)
	if err != nil {
		check(err, fmt.Sprintf("[downloadFile] Unable to download the file '%s'", remoteFilePath))
	}
	defer res.Close()

	// Write file on local storage
	buf, err := ioutil.ReadAll(res)
	err = ioutil.WriteFile(destinationLocalFilePath, buf, 0644)
	if err != nil {
		log.Fatal(err)
	}

	// Set 'access' and 'modification' time of downloaded file
	remoteFileModTime := remoteEntry.Time
	os.Chtimes(destinationLocalFilePath, remoteFileModTime, remoteFileModTime)
}

// checkLocalFileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func checkLocalFileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func getLocalFileSize(filePath string) uint64 {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	fi, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	return uint64(fi.Size())
}

func ensureDirExist(dirName string) error {

	err := os.MkdirAll(dirName, os.ModeDir)

	if err == nil || os.IsExist(err) {
		return nil
	} else {
		return err
	}
}

func getFileModTime() {
	fileStat, err := os.Stat("test.txt")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("File Name:", fileStat.Name())        // Base name of the file
	fmt.Println("Size:", fileStat.Size())             // Length in bytes for regular files
	fmt.Println("Permissions:", fileStat.Mode())      // File mode bits
	fmt.Println("Last Modified:", fileStat.ModTime()) // Last modification time
	fmt.Println("Is Directory: ", fileStat.IsDir())   // Abbreviation for Mode().IsDir()
}
