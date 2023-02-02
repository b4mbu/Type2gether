package filemanager

import (
    "os"
)

func SaveToFile(filename, content string) error {
    // 0777 think about permission
    err := os.WriteFile(filename, []byte(content), 0666);
    if err != nil {
        return err
    }
    return nil
}

