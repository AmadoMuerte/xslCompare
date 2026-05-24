package main

import (
	"fmt"
	"os"

	"github.com/AmadoMuerte/xslCompare/internal/constants"
	"github.com/AmadoMuerte/xslCompare/internal/excelutil"
	"github.com/AmadoMuerte/xslCompare/internal/fileutil"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Подготовка директории
	if err := fileutil.PrepareDirectory(constants.FilesDir); err != nil {
		return fmt.Errorf("подготовка директории: %w", err)
	}

	// Копирование файлов
	if err := fileutil.CopyFilesToWorkDir(constants.FilesDir, []string{constants.KoreaFileName, constants.EuropeFileName}); err != nil {
		return fmt.Errorf("копирование файлов: %w", err)
	}

	// Открытие Excel файлов
	files, err := excelutil.OpenExcelFiles([]string{constants.FullpriceFileName, constants.KoreaFileName, constants.EuropeFileName})
	if err != nil {
		return fmt.Errorf("открытие файлов: %w", err)
	}
	defer excelutil.CloseExcelFiles(files)

	return nil
}
