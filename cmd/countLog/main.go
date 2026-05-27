package main

import (
	"fmt"
	"os"

	"github.com/AmadoMuerte/xslCompare/internal/constants"
	"github.com/AmadoMuerte/xslCompare/internal/excelutil"
	"github.com/AmadoMuerte/xslCompare/internal/fileutil"
	"github.com/AmadoMuerte/xslCompare/internal/models"
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

	// Открытие Excel файлов
	fullpriceF := constants.FullpriceFileName
	koreaF := fmt.Sprintf("%s/%s", constants.FilesDir, constants.KoreaFileName)
	europeF := fmt.Sprintf("%s/%s", constants.FilesDir, constants.EuropeFileName)

	files, err := excelutil.OpenExcelFiles([]string{fullpriceF, koreaF, europeF})
	if err != nil {
		return fmt.Errorf("открытие файлов: %w", err)
	}
	defer excelutil.CloseExcelFiles(files)

	// Создание анализатора
	analyzer := &models.PriceAnalyzer{
		Fullprice:   files.Fullprice,
		KoreaFile:   files.Korea,
		EuropeFile:  files.Europe,
		KoreaCodes:  make(map[string]models.CodeInfo),
		EuropeCodes: make(map[string]models.CodeInfo),
	}

	// Запуск анализа
	if err := analyzer.AnalyzeAndLog(); err != nil {
		return fmt.Errorf("анализ цен: %w", err)
	}

	fmt.Println("Анализ завершен. Результаты в файле лога.")
	return nil
}
