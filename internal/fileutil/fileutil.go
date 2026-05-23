package fileutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func PrepareDirectory(dir string) error {
	return os.MkdirAll(dir, 0777)
}

func CopyFilesToWorkDir(filesDir string, files []string) error {
	for _, file := range files {
		if err := CopyFile(file, filesDir); err != nil {
			return err
		}
	}
	return nil
}

func CopyFile(filename, dir string) error {
	destPath := filepath.Join(dir, filename)

	if info, err := os.Stat(destPath); err == nil {
		if !info.IsDir() {
			if err := os.Remove(destPath); err != nil {
				return fmt.Errorf("не удалось удалить %s: %w", destPath, err)
			}
			fmt.Printf("Удален старый файл: %s\n", destPath)
		}
	}

	sourceFile, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("не удалось открыть %s: %w", filename, err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("не удалось создать %s: %w", destPath, err)
	}
	defer destFile.Close()

	if written, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("ошибка копирования %s: %w", filename, err)
	} else {
		fmt.Printf("Скопировано %d байт: %s -> %s\n", written, filename, destPath)
	}

	return nil
}
