package helper

import (
	"encoding/base64"
	"errors"
	"io"
	"log"
	"os"
)

func ReadHTMLFileAsString(filepath string) (string, error) {
	// Open the file
	file, err := os.Open(filepath)
	if err != nil {
		log.Printf("Error opening file %s: %v", filepath, err)
		return "", err
	}
	defer file.Close()

	// Read the file content
	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	// Convert the content to a string and return
	return string(content), nil
}

func CheckIfFileExists(filepath string) bool {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return false
	}

	return true
}

func CreateDirectory(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func ConvertBase64ToBytes(base64Str string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, errors.New("failed to decode base64 string: " + err.Error())
	}
	return data, nil
}

func SaveBytesToFile(data []byte, filepath string) error {
	// Create the file
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the byte data to the file
	err = os.WriteFile(filepath, data, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
