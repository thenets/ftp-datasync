package ftpop

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kr/pretty"
	"github.com/spf13/viper"
)

func check(err error, errMsg string) {
	if err != nil {
		log.Println(errMsg)
		panic(err)
	}
}

func debug(e interface{}) {
	pretty.Println(e)
}

func ensureDirExist(dirName string) error {
	err := os.MkdirAll(dirName, os.ModeDir)
	if err == nil || os.IsExist(err) {
		return nil
	}
	return err
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (context *ServerContext) readConfig() {
	// Split dir path and config file name
	var configDirPath string
	var configFileName string
	configFilePath := strings.ReplaceAll(context.ConfigFilePath, "\\", "/")
	if !strings.Contains(configFilePath, "/") {
		configDirPath = "."
		configFileName = configFilePath
	} else {
		s := strings.Split(configFilePath, "/")
		configDirPath = strings.Join(s[:len(s)-1], "/")
		configFileName = s[len(s)-1]
	}

	// Remove config file extension
	s := strings.Split(configFileName, ".")
	configFileName = strings.Join(s[:len(s)-1], ".")

	// Resolve absolute path
	absoluteConfigDirPath, err := filepath.Abs(configDirPath)
	check(err, "[readConfig] can't resolve absolute path from 'configDirPath'")

	// Load config file
	viper.AddConfigPath(absoluteConfigDirPath) // path to look for the config file in
	viper.SetConfigName(configFileName)        // name of config file (without extension)
	err = viper.ReadInConfig()                 // Find and read the config file
	if err != nil {                            // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	// Populate struct with connection data
	if viper.Get("hostAddress") == nil {
		panic("[readConfig] Variable 'hostAddress' not found in config file")
	}
	context.hostAddress = viper.GetString("hostAddress")

	if viper.Get("hostPort") == nil {
		panic("[readConfig] Variable 'hostPort' not found in config file")
	}
	context.hostPort = viper.GetInt("hostPort")

	if viper.Get("hostUser") == nil {
		panic("[readConfig] Variable 'hostUser' not found in config file")
	}
	context.hostUser = viper.GetString("hostUser")

	if viper.Get("hostPassword") == nil {
		panic("[readConfig] Variable 'hostPassword' not found in config file")
	}
	context.hostPassword = viper.GetString("hostPassword")

	if viper.Get("syncRemoteDir") == nil {
		panic("[readConfig] Variable 'syncRemoteDir' not found in config file")
	}
	context.syncRemoteDir = viper.GetString("syncRemoteDir")

	if viper.Get("syncLocalDir") == nil {
		panic("[readConfig] Variable 'syncLocalDir' not found in config file")
	}
	context.syncLocalDir = viper.GetString("syncLocalDir")

	if viper.Get("compressDir") == nil {
		panic("[readConfig] Variable 'compressDir' not found in config file")
	}
	context.compressDir = viper.GetString("compressDir")
}
