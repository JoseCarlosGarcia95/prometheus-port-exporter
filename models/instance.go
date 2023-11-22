package models

import (
	"encoding/json"
	"io"
	"os"
	"strings"
)

// Instance is a struct that represents a single instance
type Instance struct {
	Labels map[string]string `json:"labels"`
	IP     string            `json:"ip"`
}

// ReadInstances is a function that reads instances from a file
func ReadInstances(path string) ([]*Instance, error) {
	var instances []*Instance

	// Read file
	jsonFile, err := os.Open(path)

	if err != nil {
		return instances, err
	}

	byteValue, err := io.ReadAll(jsonFile)

	if err != nil {
		return instances, err
	}

	err = json.Unmarshal(byteValue, &instances)

	// Check if IP contains /24
	for _, instance := range instances {
		// Remove the range if the IP contains /x
		if strings.Contains(instance.IP, "/") {
			instance.IP = strings.Split(instance.IP, "/")[0]
		}
	}

	return instances, nil
}
