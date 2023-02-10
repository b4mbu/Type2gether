package filemanager

import (
	"errors"
	"os"
)

func SaveToFile(filename, content string) error {
    // 0777 think about permission
    err := os.WriteFile(filename, []byte(content), 0666)
    if err != nil {
        return errors.New("Cannot save file")
    }
    return nil
}

func ReadFromFile(filename string) (string, error) {
    str, err := os.ReadFile(filename)
    if err != nil {
        return "", errors.New("Cannot open file")
    }
    return string(str), nil
}

