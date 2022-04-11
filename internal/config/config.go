package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	defaultPath         = "./configs"
	defaultFileBaseName = "settings"
	defaultFileType     = "yaml"
	defaultEnvPrefix    = "K8SYNC"
)

// Config struct contains app configuration
type Config struct {
	runMode  string
	filename string                 // config file path and name
	rt       map[string]interface{} // runtime config
}

var cfg *Config

func Curr() *Config {
	if cfg == nil {
		fmt.Printf("ERROR: configuration not been initialized\n")
	}
	return cfg
}

// New creates new config object
func New(runMode string) (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}
	cfg = &Config{rt: make(map[string]interface{})}
	cfg.runMode = runMode
	cfg.filename = cfg.GetFilename()
	if cfg.filename == "" {
		return cfg, fmt.Errorf("can not get config filename")
	}
	if err := cfg.Load(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func (c *Config) GetFileBasename() string {
	if c.filename != "" {
		return filepath.Base(c.filename)
	}
	return defaultFileBaseName + "." + c.runMode
}

func (c *Config) GetFilename() string {
	configFile := c.filename
	if configFile == "" {
		configFile = filepath.Join(c.GetPath(), c.GetFileBasename()+"."+defaultFileType)
	}
	if _, err := os.Stat(configFile); err == nil {
		return configFile
	}
	file, err := os.Create(configFile)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return ""
	}
	file.Close()
	return configFile
}

func (c *Config) GetRtFilename() string {
	if c.filename == "" {
		return ""
	}
	filename := filepath.Base(c.filename)
	filename = filepath.Base(filename) + filepath.Ext(c.filename)
	return filename
}

// Load loads configuration from config file
func (c *Config) Load() error {
	if c.filename == "" {
		return errors.New("config file is null")
	}
	if _, err := os.Stat(c.filename); err != nil {
		return err
	}
	file, err := os.Open(c.filename)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	if len(b) != 0 {
		return yaml.Unmarshal(b, c.rt)
	}

	return nil
}

// CheckMissingResourceEnvvars will read the environment for equivalent config variables to set
func (c *Config) CheckMissingResourceEnvvars() {
	//if !c.Resource.DaemonSet && os.Getenv("KW_DAEMONSET") == "true" {
	//	c.Resource.DaemonSet = true
	//}
}

func (c *Config) Write() error {
	//f, err := os.OpenFile(c.GetRtFilename(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	//if err != nil {
	//	return err
	//}
	//defer f.Close()

	//enc := yaml.NewEncoder(f)
	//enc.SetIndent(2)
	//return enc.Encode(c.rt)
	return viper.WriteConfigAs(c.GetRtFilename())
}

func (c *Config) GetPath() string {
	if c.filename != "" {
		return filepath.Dir(c.filename)
	}
	if path := os.Getenv("GWA_CONFIG_PATH"); path != "" {
		return path
	}
	return defaultPath
}

func (c *Config) GetFileType() string {
	if c.filename != "" {
		return filepath.Ext(c.filename)
	}
	return defaultEnvPrefix
}

func (c *Config) GetEnvPrefix() string {
	return defaultEnvPrefix
}

func (c *Config) GetGrpcLocalAddress() string {
	return fmt.Sprintf("127.0.0.1:%s", viper.GetString("application.grpc-port"))
}

func Get(item string) interface{} {
	return viper.Get(item)
}

func GetString(item string) string {
	return viper.GetString(item)
}

func GetStringSlice(item string) []string {
	return viper.GetStringSlice(item)
}

func GetInt(item string) int {
	return viper.GetInt(item)
}

func GetBool(item string) bool{
	return viper.GetBool(item)
}

func GetAppGrpcDomain() string {
	return fmt.Sprintf("%s:%d", viper.Get("application.host"), viper.Get("application.grpc-port"))
}

func GetAppHttpDomain() string {
	return fmt.Sprintf("%s:%d", viper.Get("application.host"), viper.Get("application.http-port"))
}

func GetGwAdminBaseUrl() string {
	schema := "http://"
	if viper.GetBool("gateway.admin-secure") {
		schema = "https://"
	}
	return fmt.Sprintf("%s%s:%d", schema, GetGwAdminHost(), viper.Get("gateway.admin-port"))
}

func GetGwAdminHost() string {
	return viper.GetString("gateway.admin-host")
}

func GetGwSvcName() string {
	return viper.GetString("gateway.svc-name")
}

func GetSkipApps() []string {
	skips := viper.GetStringSlice("application.skip-apps")
	skips = append(skips, GetGwSvcName())
	return skips
}

func IsAppSvc(name string) bool {
	for _, s := range GetSkipApps() {
		if strings.HasPrefix(name, s) {
			return false
		}
	}
	return true
}
