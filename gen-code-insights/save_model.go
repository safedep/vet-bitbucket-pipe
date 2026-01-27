package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func SaveModel(model any, filePath string) error {
	fmt.Printf("Saving Artifact to file: %s\n", filePath)

	data, err := json.MarshalIndent(model, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to json marshal model: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write json file: %w", err)
	}

	fmt.Printf("Saved!\n")
	return nil
}
