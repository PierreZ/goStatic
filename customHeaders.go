package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// HeaderConfigArray is the array which contains all the custom header rules
type HeaderConfigArray struct {
	Configs []HeaderConfig `json:"configs"`
}

// HeaderConfig is a single header rule specification
type HeaderConfig struct {
	Regex         string            `json:"regex"`
	Path          string            `json:"path"`
	FileExtension string            `json:"fileExtension"`
	Headers       []HeaderDefiniton `json:"headers"`

	CompiledRegex *regexp.Regexp
}

func (config *HeaderConfig) Init() (ok bool) {
	if len(config.Regex) > 0 {
		regex, err := regexp.Compile(config.Regex)

		if err != nil {
			fmt.Println("WARNING: Ignoring rule with regex error:", err)
			fmt.Println("")
			return
		}
		config.CompiledRegex = regex
	}

	ok = true
	return
}

func (config *HeaderConfig) UsesRegex() bool {
	return config.CompiledRegex != nil
}

// HeaderDefiniton is a key value pair of a specified header rule
type HeaderDefiniton struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var headerConfigs HeaderConfigArray

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func logHeaderConfig(config HeaderConfig) {
	if config.UsesRegex() {
		fmt.Println("Regex: " + config.Regex)
	} else {
		fmt.Println("Path: " + config.Path)
		fmt.Println("FileExtension: " + config.FileExtension)
	}

	for j := 0; j < len(config.Headers); j++ {
		headerRule := config.Headers[j]
		fmt.Println(headerRule.Key, ":", headerRule.Value)
	}

	fmt.Println("------------------------------")
}

func initHeaderConfig(headerConfigPath string) bool {
	headerConfigValid := false

	if fileExists(headerConfigPath) {
		jsonFile, err := os.Open(headerConfigPath)
		if err != nil {
			fmt.Println("Can't read header config file. Error:")
			fmt.Println(err)
		} else {
			byteValue, _ := ioutil.ReadAll(jsonFile)

			json.Unmarshal(byteValue, &headerConfigs)

			if len(headerConfigs.Configs) > 0 {
				headerConfigValid = true

				// Only keep valid config entries.
				keepers := make([]HeaderConfig, 0)
				for _, configEntry := range headerConfigs.Configs {
					ok := configEntry.Init()
					if ok {
						keepers = append(keepers, configEntry)
					}
				}

				headerConfigs.Configs = keepers

				// Print the config entries that are kept.
				fmt.Println("Found header config file. Rules:")
				fmt.Println("------------------------------")
				for _, configEntry := range headerConfigs.Configs {
					logHeaderConfig(configEntry)
				}
			} else {
				fmt.Println("No rules found in header config file.")
			}

		}
		jsonFile.Close()
	}

	return headerConfigValid
}

func customHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqFileExtension := filepath.Ext(r.URL.Path)

		for i := 0; i < len(headerConfigs.Configs); i++ {
			configEntry := headerConfigs.Configs[i]
			var matches bool

			if configEntry.UsesRegex() {
				matches = configEntry.CompiledRegex.MatchString(r.URL.Path)
			} else {
				matches =
					// Check if the file extension matches.
					(configEntry.FileExtension == "*" || reqFileExtension == "."+configEntry.FileExtension) &&
						// Check if the path matches.
						(configEntry.Path == "*" || strings.HasPrefix(r.URL.Path, configEntry.Path))
			}

			if matches {
				for j := 0; j < len(configEntry.Headers); j++ {
					headerEntry := configEntry.Headers[j]
					w.Header().Set(headerEntry.Key, headerEntry.Value)
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}
