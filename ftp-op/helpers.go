package ftpop

import (
	"fmt"
	"log"
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
}
