package main

import (
	"encoding/base64"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type Secret struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   map[string]interface{} `yaml:"metadata"`
	Data       map[string]interface{} `yaml:"data"`
}

func main() {
	// Directory containing YAML files
	dir := "."

	// List all .yaml files in the directory
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		log.Fatalf("Failed to list YAML files: %v", err)
	}

	for _, file := range files {
		fmt.Printf("Processing file: %s\n", file)

		// Read the YAML file
		data, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Failed to read file %s: %v\n", file, err)
			continue
		}

		// Parse the YAML data
		var secret Secret
		err = yaml.Unmarshal(data, &secret)
		if err != nil {
			fmt.Printf("Failed to parse YAML in file %s: %v\n", file, err)
			continue
		}

		// Validate the expected structure
		if secret.APIVersion == "" || secret.Kind == "" || secret.Data == nil {
			fmt.Printf("Invalid structure in file %s: skipping\n", file)
			continue
		}

		// Decode base64 values and detect their type
		valid := true
		for key, value := range secret.Data {
			strValue, ok := value.(string)
			if !ok {
				fmt.Printf("Value for key %s is not a string in file %s: skipping\n", key, file)
				valid = false
				break
			}

			decoded, err := base64.StdEncoding.DecodeString(strValue)
			if err != nil {
				fmt.Printf("Failed to decode base64 for key %s in file %s: %v\n", key, file, err)
				valid = false
				break
			}

			decodedStr := string(decoded)

			// Try to determine the type of the decoded value
			if intValue, err := strconv.Atoi(decodedStr); err == nil {
				secret.Data[key] = intValue
			} else if boolValue, err := strconv.ParseBool(decodedStr); err == nil {
				secret.Data[key] = boolValue
			} else {
				secret.Data[key] = decodedStr
			}
		}

		// Skip the file if any value was invalid
		if !valid {
			fmt.Printf("Skipping file %s due to invalid data\n", file)
			continue
		}

		// Marshal the new YAML data
		outputData, err := yaml.Marshal(&secret)
		if err != nil {
			fmt.Printf("Failed to marshal YAML for file %s: %v\n", file, err)
			continue
		}

		// Create output file name
		outputFile := file[:len(file)-len(filepath.Ext(file))] + "_decoded.yaml"

		// Write the output YAML file
		err = os.WriteFile(outputFile, outputData, 0644)
		if err != nil {
			fmt.Printf("Failed to write output file %s: %v\n", outputFile, err)
			continue
		}

		fmt.Printf("Decoded YAML saved to %s\n", outputFile)
	}
}
