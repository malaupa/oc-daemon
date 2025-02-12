package client

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/telekom-mms/oc-daemon/pkg/xmlprofile"
)

var (
	// ConfigName is the name of the configuration file
	ConfigName = "oc-client.json"

	// ConfigDirName is the name of the directory where the configuration
	// file is stored
	ConfigDirName = "oc-daemon"

	// SystemConfigDirPath is the path of the directory where the directory
	// of the system configuration is stored
	SystemConfigDirPath = "/var/lib"

	// SocketFile is the socket file of the OC-Daemon
	SocketFile = "/run/oc-daemon/daemon.sock"

	// ConnectionTimeout is the timeout for the client connection attempt
	ConnectionTimeout = 30 * time.Second

	// RequestTimeout is the timeout for the entire request/response
	// exchange initiated by the client after a successful connection
	RequestTimeout = 30 * time.Second
)

// Config is a configuration for the OC client
type Config struct {
	ClientCertificate string
	ClientKey         string
	CACertificate     string
	XMLProfile        string
	VPNServer         string
	User              string
	Password          string

	SocketFile        string
	ConnectionTimeout time.Duration
	RequestTimeout    time.Duration
}

// Copy returns a copy of Config
func (c *Config) Copy() *Config {
	if c == nil {
		return nil
	}
	cp := *c
	return &cp
}

// Empty returns if the config is empty
func (c *Config) Empty() bool {
	if c == nil {
		return true
	}

	empty := &Config{}
	return reflect.DeepEqual(c, empty)
}

// expandPath expands tilde and environment variables in path
func expandPath(path string) string {
	// note: handling of tilde is limited:
	// it only works with file paths beginning with ~/
	if strings.HasPrefix(path, "~") {
		path = filepath.Join("$HOME", path[1:])
	}
	return os.ExpandEnv(path)
}

// expandPaths expands the paths in config
func (c *Config) expandPaths() {
	c.ClientCertificate = expandPath(c.ClientCertificate)
	c.ClientKey = expandPath(c.ClientKey)
	c.CACertificate = expandPath(c.CACertificate)
}

// expandUser expands the username in config
func (c *Config) expandUser() {
	c.User = os.ExpandEnv(c.User)
}

// Expand expands variables in config
func (c *Config) Expand() {
	c.expandPaths()
	c.expandUser()
}

// Save saves the config to file
func (c *Config) Save(file string) error {
	b, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, b, 0600)
}

// NewConfig returns a new Config
func NewConfig() *Config {
	return &Config{
		XMLProfile:        xmlprofile.SystemProfile,
		SocketFile:        SocketFile,
		ConnectionTimeout: ConnectionTimeout,
		RequestTimeout:    RequestTimeout,
	}
}

// LoadConfig loads a Config from file
func LoadConfig(file string) (*Config, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	conf := NewConfig()
	if err := json.Unmarshal(b, conf); err != nil {
		return nil, err
	}

	return conf, nil
}

// SystemConfig returns the default file path of the system configuration
func SystemConfig() string {
	return filepath.Join(SystemConfigDirPath, ConfigDirName, ConfigName)
}

// UserConfig returns the default file path of the current user's configuration
func UserConfig() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, ConfigDirName, ConfigName)
}

// LoadUserSystemConfig loads the user or system configuration from its
// default location, expands variables in config
func LoadUserSystemConfig() *Config {
	// try user config
	if config, err := LoadConfig(UserConfig()); err == nil && config != nil {
		config.Expand()
		return config
	}

	// try system config
	if config, err := LoadConfig(SystemConfig()); err == nil && config != nil {
		config.Expand()
		return config
	}

	// could not load a config
	return nil
}
