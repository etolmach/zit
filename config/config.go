package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/google/go-jsonnet"
)

// EnvVarName TODO
const EnvVarName = "ZIT_CONFIG"

// ErrConfigNotFound TODO
type ErrConfigNotFound struct {
	EnvVar bool
	Path   string
}

func (err *ErrConfigNotFound) Error() string {
	envVar := ""
	if err.EnvVar {
		envVar = " defined in " + EnvVarName + " variable"
	}
	return fmt.Sprintf("config file%s is not found at %q", envVar, err.Path)
}

// HostMap TODO
type HostMap map[string]Config

// Get TODO
func (hm *HostMap) Get(host string) (*Config, error) {
	conf, ok := (*hm)[host]
	if !ok {
		return nil, fmt.Errorf("cannot find config for host %q", host)
	}

	return &conf, nil
}

// Config TODO
type Config struct {
	Default   *User      `json:"default"`
	Overrides []Override `json:"overrides"`
}

// User TODO
type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Override TODO
type Override struct {
	Owner string `json:"owner"`
	Repo  string `json:"repo,omitempty"`
	User  User   `json:"user"`
}

// ReadHostMap TODO
func ReadHostMap(filename string, r io.Reader) (*HostMap, error) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	confStr := buf.String()

	vm := jsonnet.MakeVM()
	confJSON, err := vm.EvaluateSnippet(filename, confStr)
	if err != nil {
		return nil, err
	}

	var hostMap HostMap
	if err := json.Unmarshal([]byte(confJSON), &hostMap); err != nil {
		return nil, err
	}

	return &hostMap, nil
}

// LocateConfFile locates the path of the configuration file.
func LocateConfFile() (string, error) {
	fileExists := func(filename string) bool {
		info, err := os.Stat(filename)
		if os.IsNotExist(err) {
			return false
		}
		return !info.IsDir()
	}

	var confPath string

	// check ZIT_CONFIG env variable
	confPath, defined := os.LookupEnv(EnvVarName)

	// if ZIT_CONFIG is not set, try default location
	if !defined {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		confPath = path.Join(home, ".zit", "config.jsonnet")
	}

	if !fileExists(confPath) {
		return "", &ErrConfigNotFound{
			EnvVar: defined,
			Path:   confPath,
		}
	}

	return confPath, nil
}
