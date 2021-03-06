package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fnproject/cli/utils"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

const (
	rootConfigPathName = ".fn"

	contextsPathName       = "contexts"
	configName             = "config"
	contextConfigFileName  = "config.yaml"
	defaultContextFileName = "default.yaml"
	defaultLocalAPIURL     = "http://localhost:8080/v1"
	DefaultProvider        = "default"

	ReadWritePerms = os.FileMode(0755)

	CurrentContext  = "current-context"
	ContextProvider = "provider"

	EnvFnRegistry = "registry"
	EnvFnToken    = "token"
	EnvFnAPIURL   = "api-url"
	EnvFnContext  = "context"

	OracleKeyID         = "key-id"
	OraclePrivateKey    = "private-key"
	OracleCompartmentID = "compartment-id"
	OracleDisableCerts  = "disable-certs"
)

var defaultRootConfigContents = &utils.ContextMap{CurrentContext: ""}
var defaultContextConfigContents = &utils.ContextMap{
	ContextProvider: DefaultProvider,
	EnvFnAPIURL:     defaultLocalAPIURL,
	EnvFnRegistry:   "",
}

// ContextFile defines the internal structure of a default context
type ContextFile struct {
	ContextProvider string `yaml:"provider"`
	EnvFnAPIURL     string `yaml:"api-url"`
	EnvFnRegistry   string `yaml:"registry"`
}

// Init : Initialise/load config direc
func Init() error {
	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvPrefix("fn")

	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)

	viper.SetDefault(EnvFnAPIURL, defaultLocalAPIURL)

	return ensureConfiguration()
}

// EnsureConfiguration ensures context configuration directory hierarchy is in place, if not
// creates it and the default context configuration files
func ensureConfiguration() error {
	home := utils.GetHomeDir()

	rootConfigPath := filepath.Join(home, rootConfigPathName)
	if _, err := os.Stat(rootConfigPath); os.IsNotExist(err) {
		if err = os.Mkdir(rootConfigPath, ReadWritePerms); err != nil {
			return fmt.Errorf("error creating .fn directory %v", err)
		}
	}

	contextConfigFilePath := filepath.Join(rootConfigPath, contextConfigFileName)
	if _, err := os.Stat(contextConfigFilePath); os.IsNotExist(err) {
		file, err := os.Create(contextConfigFilePath)
		if err != nil {
			return fmt.Errorf("error creating config.yaml file %v", err)
		}

		err = utils.WriteYamlFile(file, defaultRootConfigContents)
		if err != nil {
			return err
		}
	}

	contextsPath := filepath.Join(rootConfigPath, contextsPathName)
	if _, err := os.Stat(contextsPath); os.IsNotExist(err) {
		if err = os.Mkdir(contextsPath, ReadWritePerms); err != nil {
			return fmt.Errorf("error creating contexts directory %v", err)
		}
	}

	defaultContextPath := filepath.Join(contextsPath, defaultContextFileName)
	if _, err := os.Stat(defaultContextPath); os.IsNotExist(err) {
		file, err := os.Create(defaultContextPath)
		if err != nil {
			return fmt.Errorf("error creating default.yaml context file %v", err)
		}

		err = utils.WriteYamlFile(file, defaultContextConfigContents)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetContextsPath : Returns the path to the contexts directory.
func GetContextsPath() string {
	contextsPath := filepath.Join(rootConfigPathName, contextsPathName)
	return contextsPath
}

func LoadConfiguration(c *cli.Context) error {
	// Find home directory.
	home := utils.GetHomeDir()
	context := ""
	if context = c.String(EnvFnContext); context == "" {
		viper.AddConfigPath(filepath.Join(home, rootConfigPathName))
		viper.SetConfigName(configName)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
		context = viper.GetString(CurrentContext)
	}

	viper.AddConfigPath(filepath.Join(home, rootConfigPathName, contextsPathName))
	viper.SetConfigName(context)

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("%v \n", err)
		err := WriteCurrentContextToConfigFile("default")
		if err != nil {
			return err
		}
		fmt.Println("current context has been set to default")
		return nil
	}

	viper.Set(CurrentContext, context)
	return nil
}

func WriteCurrentContextToConfigFile(value string) error {
	home := utils.GetHomeDir()

	configFilePath := filepath.Join(home, rootConfigPathName, contextConfigFileName)
	f, err := os.OpenFile(configFilePath, os.O_RDWR, ReadWritePerms)
	if err != nil {
		return err
	}
	defer f.Close()

	file, err := utils.DecodeYAMLFile(f)
	if err != nil {
		return err
	}

	configValues := utils.ContextMap{}
	for k, v := range *file {
		if k == CurrentContext {
			configValues[k] = value
		} else {
			configValues[k] = v
		}
	}

	err = utils.WriteYamlFile(f, &configValues)
	if err != nil {
		return err
	}

	return nil
}
