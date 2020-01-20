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

	syncRemoteDir string
	syncLocalDir  string

	compressDir string

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
func (context *ServerContext) Sync() {
	remoteDir := context.syncRemoteDir
	localDir := context.syncLocalDir

	// Recursive localDir if not exist
	ensureDirExist(localDir)

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

	// Delete local files that doesn't exist in remote
	context.deleteLocalFiles(items, remoteDir, localDir)

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

// deleteLocalFiles deletes all local files that doesn't exist in remote directory
func (context *ServerContext) deleteLocalFiles(remoteEntries []*ftp.Entry, remoteDir string, localDir string) {
	// Just return if 'localDir' doesn't exist
	if _, err := os.Stat(localDir); os.IsNotExist(err) {
		return
	}

	// Get list of 'localEntries' in 'localDir'
	localEntries, err := ioutil.ReadDir(localDir)
	if err != nil {
		log.Fatal(err)
	}

	// Check if exist in remote
	for _, localEntry := range localEntries { // for each localEntry
		localEntryFoundInRemote := false

		// Search localFile in remoteEntries
		// TODO probably exist a better way to do it
		for _, remoteEntry := range remoteEntries {
			if remoteEntry.Type == 1 { // Is a directory
				if localEntry.IsDir() && localEntry.Name() == remoteEntry.Name {
					localEntryFoundInRemote = true
				}

			} else { // Is a file
				if !localEntry.IsDir() && localEntry.Name() == remoteEntry.Name {
					localEntryFoundInRemote = true
				}
			}
		}

		// Delete 'localEntry' if not found in remote
		if !localEntryFoundInRemote {
			localEntryPath := fmt.Sprintf("%s/%s", localDir, localEntry.Name())

			fmt.Printf("File '%s' not found on remote. Removing...\n", localEntryPath)

			if localEntry.IsDir() {
				os.RemoveAll(localEntryPath)
			} else {
				os.Remove(localEntryPath)
			}
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
