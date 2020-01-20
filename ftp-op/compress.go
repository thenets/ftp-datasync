package ftpop

import (
	"archive/zip"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

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
	deleteObsoleteCompressedFiles(originDir, targetDir)
	deleteEmptyDirs(targetDir)
}

func compressFilesRecursive(originDir string, targetDir string) {
	// Get list of 'originEntries' in 'originDir'
	originEntries, err := ioutil.ReadDir(originDir)
	check(err, "[compressFilesRecursive] Can't read 'originDir' path")

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
			targetFilePath, err := filepath.Abs(fmt.Sprintf("%s/%s.zip", targetDir, originEntry.Name()))
			check(err, "[compressFilesRecursive] can't resolve absolute path from 'targetFilePath'")
			hashFilePath, err := filepath.Abs(fmt.Sprintf("%s/%s.hash", targetDir, originEntry.Name()))
			check(err, "[compressFilesRecursive] can't resolve absolute path from 'hashFilePath'")

			ensureDirExist(targetDir)
			compressFile(originFilePath, targetFilePath, hashFilePath)
		}
	}
}

func compressFile(originFilePath string, compressedFilePath string, hashFilePath string) {
	needToCompress := false

	// Check if hash file exist
	fileInfo, err := os.Stat(hashFilePath)
	if os.IsNotExist(err) {
		needToCompress = true
	} else {
		if fileInfo.IsDir() {
			os.RemoveAll(hashFilePath)
			needToCompress = true
		}
	}

	// Check hash file
	currentOriginalFileHash := getHashFromFile(originFilePath, "sha1")
	currentCompressedFileHash := getHashFromFile(compressedFilePath, "sha1")
	lastOriginalFileHash, lastCompressedFileHash := openHashFile(hashFilePath)
	if currentOriginalFileHash != lastOriginalFileHash {
		// Need to recompress if both hashes are not equal
		needToCompress = true
	}
	if currentCompressedFileHash != lastCompressedFileHash {
		// Need to recompress if both hashes are not equal
		needToCompress = true
	}

	// Compress only if needed
	if needToCompress {
		fmt.Println("Compressing:", compressedFilePath)
		files := []string{originFilePath}
		if err := zipFiles(compressedFilePath, files); err != nil {
			panic(err)
		}

	} else {
		fmt.Println("Skipping compress:", hashFilePath)
	}

	// Create new hash file
	newCompressedFileHash := getHashFromFile(compressedFilePath, "sha1")
	writeHashFile(hashFilePath, currentOriginalFileHash, newCompressedFileHash)

}

// openHashFile returns the 'originalFileHash' and 'compressedFileHash'
// from a valid hash file.
// Valid hash file format example:
// 62cdd0166772aa8de3b0c0ec60331d5249525ffa;b066df618ba28c33df2bcebfa9c879ea6632cbc6
func openHashFile(hashFilePath string) (string, string) {
	// Check if is a valid hash file format
	_, err := os.Stat(hashFilePath)
	if os.IsNotExist(err) {
		return "", ""
	}

	dat, err := ioutil.ReadFile(hashFilePath)
	check(err, "[openHashFile] Can't open hash file")

	hashList := strings.Split(string(dat), ";")

	// Check if is a valid hash file format
	if len(hashList) != 2 {
		return "", ""
	}

	originalFileHash := hashList[0]
	compressedFileHash := hashList[1]

	return originalFileHash, compressedFileHash
}

func writeHashFile(hashFilePath string, originalFileHash string, compressedFileHash string) {
	f, err := os.Create(hashFilePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f.WriteString(
		fmt.Sprintf("%s;%s", originalFileHash, compressedFileHash),
	)
}

// getHashFromFile returns the hash content of a file in string format
// hashAlgorithm options: [black2b, sha1]
func getHashFromFile(filePath string, hashAlgorithm string) string {
	var hashString string
	var err error

	if !fileExists(filePath) {
		return ""
	}

	if hashAlgorithm == "blake2b" {
		content, err := ioutil.ReadFile(filePath)
		check(err, "[getHashFromFile] Fail trying to hash using 'blake2b'")
		hash := blake2b.Sum256(content)
		hashString = hex.EncodeToString(hash[:])
	} else if hashAlgorithm == "sha1" {
		hashString, err = func(filePath string) (string, error) {
			var returnSHA1String string
			file, err := os.Open(filePath)
			if err != nil {
				return returnSHA1String, err
			}
			defer file.Close()
			hash := sha1.New()
			if _, err := io.Copy(hash, file); err != nil {
				return returnSHA1String, err
			}
			hashInBytes := hash.Sum(nil)[:20]
			returnSHA1String = hex.EncodeToString(hashInBytes)
			return returnSHA1String, nil
		}(filePath)
		check(err, "[getHashFromFile] Fail trying to hash using 'sha1'")
	} else {
		log.Fatalln("hashAlgorithm not supported!")
	}

	return hashString
}

// zipFiles compresses one or many files into a single zip archive file.
// Param 1: filename is the output zip file's name.
// Param 2: files is a list of files to add to the zip.
func zipFiles(filename string, files []string) error {

	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		if err = addFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string) error {

	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	s := strings.Split(strings.ReplaceAll(filename, "\\", "/"), "/")
	header.Name = s[len(s)-1]

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

// deleteObsoleteCompressedFiles deletes all compressed files that doesn't exist in 'originDir' directory
func deleteObsoleteCompressedFiles(originDir string, compressDir string) {
	// Get list of 'compressEntries' in 'compressDir'
	compressEntries, err := ioutil.ReadDir(compressDir)
	if err != nil {
		check(err, "[deleteObsoleteCompressedFiles] can't read 'compressDir' dir")
	}

	// Check if exist in origin
	for _, compressEntry := range compressEntries { // for each compressEntry
		if compressEntry.IsDir() {
			// Recursive call if is a dir
			deleteObsoleteCompressedFiles(
				fmt.Sprintf("%s/%s", originDir, compressEntry.Name()),
				fmt.Sprintf("%s/%s", compressDir, compressEntry.Name()),
			)
			continue
		}

		// Check if is a hash file
		if strings.HasSuffix(compressEntry.Name(), ".hash") {
			// Check if exist origin file relative to hash
			// delete it if do not
			fileNameWithoutHashExtension := compressEntry.Name()[:len(compressEntry.Name())-5]
			if !fileExists(compressDir + "/" + fileNameWithoutHashExtension + ".zip") {
				os.Remove(compressDir + "/" + compressEntry.Name())
			}

			// Else ignore hash file
			continue
		}

		// If is a file, search original file in original dir
		fileNameWithoutZipExtension := compressEntry.Name()[:len(compressEntry.Name())-4]
		originFilePath := fmt.Sprintf("%s/%s", originDir, fileNameWithoutZipExtension)
		fileFoundInOrigin := fileExists(originFilePath)

		// Delete 'compressEntry' and hash file if not found in origin
		if !fileFoundInOrigin {
			fmt.Printf("File '%s' not found on origin. Removing...\n", originFilePath)

			compressEntryPath := compressDir + "/" + fileNameWithoutZipExtension
			os.Remove(compressEntryPath + ".zip")
			os.Remove(compressEntryPath + ".hash")
		} else {
			// fmt.Println("File found!", originFilePath)
		}

	}
}

// deleteEmptyDirs delete all subdirs if is empty
func deleteEmptyDirs(targetDir string) {

	// Recursive call all other subdirs
	entries, err := ioutil.ReadDir(targetDir)
	if err != nil {
		check(err, "[deleteEmptyDirs] can't read 'targetDir' dir")
	}
	for _, entry := range entries {
		if entry.IsDir() {
			deleteEmptyDirs(targetDir + "/" + entry.Name())
		}
	}

	// Delete dir if empty
	entries, err = ioutil.ReadDir(targetDir)
	if err != nil {
		check(err, "[deleteEmptyDirs] can't read 'targetDir' dir")
	}
	if len(entries) == 0 {
		os.RemoveAll(targetDir)
	}
}

func (context *ServerContext) CompressCreateReport(reportFilePath string) {
	// Create report header
	header := []byte("originalFileHash,compressedFileHash,compressedFilePath\n")
	err := ioutil.WriteFile(reportFilePath, header, 0644)
	if err != nil {
		panic(err)
	}

	// Scan dir
	compressReportScanDir(
		context.compressDir,
		reportFilePath,
	)
}

func compressReportScanDir(targetDir string, reportFilePath string) {
	entries, err := ioutil.ReadDir(targetDir)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		// If is a dir
		if entry.IsDir() {
			compressReportScanDir(
				targetDir+"/"+entry.Name(),
				reportFilePath,
			)
			continue
		}

		// If is a .hash file
		if strings.HasSuffix(entry.Name(), ".hash") {
			// Create line to write
			fileNameWithoutHashExtension := entry.Name()[:len(entry.Name())-5]
			absoluteCompressedFilePath, _ := filepath.Abs(targetDir + "/" + fileNameWithoutHashExtension + ".zip")
			originalFileHash, compressedFileHash := openHashFile(targetDir + "/" + entry.Name())
			line := fmt.Sprintf("%s,%s,%s\n", originalFileHash, compressedFileHash, absoluteCompressedFilePath)

			// Write in report file
			f, err := os.OpenFile(reportFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Println(err)
			}
			defer f.Close()
			if _, err := f.WriteString(line); err != nil {
				log.Println(err)
			}
		}

	}
}
