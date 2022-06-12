package components

import (
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Debug               bool     `yaml:"debug"`
	LogLevel            string   `yaml:"log_level"`
	ListenAddr          string   `yaml:"listen_addr"`
	Timeout             int      `yaml:"timeout"`
	MongoURL            string   `yaml:"mongo_url"`
	IPAPIKey            string   `yaml:"ipapi_key"`
	ACLAllowedCountries []string `yaml:"acl_allowed_countries"`
}

func ParseYAMLConfig(path string) (*Config, error) {
	configBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	if err = yaml.Unmarshal(configBytes, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *Config) GetDebug() bool {
	return c.Debug
}

func (c *Config) GetListenAddr() string {
	return c.ListenAddr
}

func (c *Config) GetTimeoutDuration() time.Duration {
	return time.Duration(c.Timeout) * time.Second
}

func (c *Config) GetAllowedCountries() []string {
	return c.ACLAllowedCountries
}
