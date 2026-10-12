// utils-------------------------------------
// @file      : filetool.go
// @time      : 2025/5/25 15:54
// -------------------------------------------

package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func EnsureDir(path string) error {

	info, err := os.Stat(path)

	if os.IsNotExist(err) {

		err := os.MkdirAll(path, 0755)
		if err != nil {
			return fmt.Errorf("Failed to create directory: %w", err)
		}
		return nil
	}

	if err == nil && !info.IsDir() {
		return fmt.Errorf("Path exists but is not a directory: %s", path)
	}

	return nil
}

func WriteFile(filePath string, content []byte) error {

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		return fmt.Errorf("Failed to write file: %w", err)
	}

	return nil
}

func DownloadFile(url string, savePath string, timeout time.Duration) error {

	if err := os.MkdirAll(filepath.Dir(savePath), os.ModePerm); err != nil {
		return err
	}

	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("Request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected HTTP status: %d", resp.StatusCode)
	}

	out, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func DeleteFile(filePath string) {

	if filePath == "" {
		return
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return
	} else if err != nil {
		return
	}

	err := os.Remove(filePath)
	if err != nil {
		return
	}
}
