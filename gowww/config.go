package gowww

import (
	"errors"
	"path/filepath"
	"strings"

	"git.gohegan.uk/kaigoh/gowww/v2/utilities"
)

type Config struct {
	Path                string   `yaml:"-"`
	Host                string   `yaml:"-"`
	IsWatched           bool     `yaml:"-"`
	Hosts               []string `yaml:"hosts"`
	DefaultDocuments    []string `yaml:"default_documents"`
	AllowDirectoryIndex bool     `yaml:"allow_directory_index"`
}

var DefaultConfig = Config{DefaultDocuments: []string{"index.htm", "index.html", "default.htm"}, AllowDirectoryIndex: false}

func (c Config) HasHost(host string) bool {
	host = strings.ToLower(host)
	if strings.ToLower(c.Host) == host {
		return true
	}
	for _, h := range c.Hosts {
		if strings.ToLower(h) == host {
			return true
		}
	}
	return false
}

func (c Config) GetDefaultDocument(path string) (file string, err error) {
	path = c.Path + path
	if len(c.DefaultDocuments) == 0 {
		c.DefaultDocuments = DefaultConfig.DefaultDocuments
	}
	for _, i := range c.DefaultDocuments {
		index := filepath.Base(i)
		if utilities.FileExists(path + index) {
			return index, nil
		}
	}

	// Return an error if no default document found...
	return "", errors.New("no default document found")

}

// Get the configuration for a web site...
func GetConfig(configs []Config, defaultConfig Config, host string) (config Config) {
	for _, c := range configs {
		if c.HasHost(host) {
			return c
		}
	}
	return defaultConfig
}
