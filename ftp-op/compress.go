package ftpop

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/blake2b"
)

func (context *ServerContext) Compress() {
	// Get all files from 'originDir', compress it, and
	// save them in 'targetDir'
	originDir, err := filepath.Abs(context.syncLocalDir)
	check(err, "[Compress] can't resolve absolute path from 'originDir'")
	targetDir, err := filepath.Abs(context.compressDir)
	check(err, "[Compress] can't resolve absolute path from 'targetDir'")

	compressFilesRecursive(originDir, targetDir)
}

func compressFilesRecursive(originDir string, targetDir string) {
	// Get list of 'originEntries' in 'originDir'
	originEntries, err := ioutil.ReadDir(originDir)
	if err != nil {
		log.Fatal(err)
	}

	// For each entry in 'originDir'
	for _, originEntry := range originEntries {
		if originEntry.IsDir() {
			// Recursive call if is a sub-directory
			originEntrySubPath := fmt.Sprintf("%s/%s", originDir, originEntry.Name())
			targetEntrySubPath := fmt.Sprintf("%s/%s", targetDir, originEntry.Name())
			compressFilesRecursive(originEntrySubPath, targetEntrySubPath)
		} else {
			// Compress if is a file
			originFilePath, err := filepath.Abs(fmt.Sprintf("%s/%s", originDir, originEntry.Name()))
			check(err, "[compressFilesRecursive] can't resolve absolute path from 'originFilePath'")
			targetFilePath, err := filepath.Abs(fmt.Sprintf("%s/%s", targetDir, originEntry.Name()))
			check(err, "[compressFilesRecursive] can't resolve absolute path from 'targetFilePath'")
			hashFilePath, err := filepath.Abs(fmt.Sprintf("%s/%s.hash", targetDir, originEntry.Name()))
			check(err, "[compressFilesRecursive] can't resolve absolute path from 'hashFilePath'")

			ensureDirExist(targetDir)
			compressFile(originFilePath, targetFilePath, hashFilePath)
		}
	}
}

func compressFile(originFilePath string, targetFilePath string, hashFilePath string) {
	needToCompress := false

	// Check if has changes and need to compress
	fileInfo, err := os.Stat(hashFilePath)
	if os.IsNotExist(err) {
		needToCompress = true
	} else {
		if fileInfo.IsDir() {
			os.RemoveAll(hashFilePath)
			needToCompress = true
		}
	}

	// Check hash
	content, err := ioutil.ReadFile(originFilePath)
	if err != nil {
		log.Fatal(err)
	}
	h := blake2b.Sum256(content)
	log.Println(hex.EncodeToString(h[:]))

	// Compress only if needed
	if needToCompress {
		fmt.Println("Compressing:", hashFilePath)

	} else {
		fmt.Println("Skipping compress:", hashFilePath)
	}
}
