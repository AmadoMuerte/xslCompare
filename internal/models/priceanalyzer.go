package models

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/AmadoMuerte/xslCompare/internal/constants"
	"github.com/AmadoMuerte/xslCompare/internal/log"
	"github.com/xuri/excelize/v2"
)

type PriceAnalyzer struct {
	Fullprice   *excelize.File
	KoreaFile   *excelize.File
	EuropeFile  *excelize.File
	Writer      *bufio.Writer
	KoreaCodes  map[string]CodeInfo
	EuropeCodes map[string]CodeInfo
}

func (p *PriceAnalyzer) AnalyzeAndLog() error {
	var err error
	p.Writer, err = log.InitLogFile(constants.CounterLogFile)
	if err != nil {
		return err
	}
	defer log.FinalizeLog(p.Writer)

	// Загружаем индексы кодов из файлов Кореи и Европы
	if err := p.loadCodeIndexes(); err != nil {
		return err
	}

	// Обрабатываем строки Fullprice
	return p.processFullpriceRows()
}

// Загружает коды из файла в map
func (p *PriceAnalyzer) LoadCodesFromFile(file *excelize.File, codesMap map[string]CodeInfo) error {
	sheetName := file.GetSheetName(0)
	rows, err := file.GetRows(sheetName)
	if err != nil {
		return err
	}

	for i, row := range rows {
		if i == 0 || len(row) == 0 {
			continue
		}

		code := strings.TrimSpace(row[0]) // Колонка A - код
		if code != "" {
			codesMap[code] = CodeInfo{
				RowIndex: i,
				Sheet:    sheetName,
			}
		}
	}

	return nil
}

// Загружает индексы для обоих файлов
func (p *PriceAnalyzer) loadCodeIndexes() error {
	if err := p.LoadCodesFromFile(p.KoreaFile, p.KoreaCodes); err != nil {
		return fmt.Errorf("загрузка кодов из Кореи: %w", err)
	}

	if err := p.LoadCodesFromFile(p.EuropeFile, p.EuropeCodes); err != nil {
		return fmt.Errorf("загрузка кодов из Европы: %w", err)
	}

	return nil
}

// Обрабатывает все строки Fullprice
func (p *PriceAnalyzer) processFullpriceRows() error {
	sheetName := p.Fullprice.GetSheetName(0)
	rows, err := p.Fullprice.GetRows(sheetName)
	if err != nil {
		return err
	}

	// Заголовок лога
	p.writeLog("%-15s | %-20s | %-10s | %-15s | %-15s | %-10s\n",
		"Код", "Марка", "количество", "Корея цена", "Европа цена", "Файлы")
	p.writeLog("%s\n", strings.Repeat("-", 100))

	for i, row := range rows {
		if i == 0 || len(row) < 5 {
			continue
		}

		code := strings.TrimSpace(row[1])        // Колонка B - код
		brand := strings.TrimSpace(row[4])       // Колонка E - марка
		quantityStr := strings.TrimSpace(row[3]) // Колонка D - количество

		if code == "" {
			continue
		}

		// Парсим количество в Fullprice
		quantity, err := strconv.ParseFloat(quantityStr, 64)
		if err != nil {
			continue
		}

		// Если в Fullprice количество > 0, проверяем другие файлы
		if quantity > 0 {
			p.checkAndLogZeroQuantity(code, brand, quantity)
		}
	}

	return nil
}

// Проверяет оба файла и логирует результат
func (p *PriceAnalyzer) checkAndLogZeroQuantity(code, brand string, fullpriceQuantity float64) {
	// Проверяем Korea файл
	koreaPrice := p.getPriceIfZeroQuantity(p.KoreaFile, p.KoreaCodes, code)

	// Проверяем Europe файл
	europePrice := p.getPriceIfZeroQuantity(p.EuropeFile, p.EuropeCodes, code)

	// Если хотя бы в одном файле количество = 0, логируем
	if koreaPrice != nil || europePrice != nil {
		koreaStr := "—"
		europeStr := "—"
		files := ""

		if koreaPrice != nil {
			koreaStr = fmt.Sprintf("%.2f", *koreaPrice)
			files += "Корея "
		}
		if europePrice != nil {
			europeStr = fmt.Sprintf("%.2f", *europePrice)
			files += "Европа"
		}

		p.writeLog("%-15s | %-20s | %-10.0f | %-15s | %-15s | %s\n",
			code, brand, fullpriceQuantity, koreaStr, europeStr, files)
	}
}

// Возвращает цену если количество = 0, иначе nil
func (p *PriceAnalyzer) getPriceIfZeroQuantity(file *excelize.File, codesMap map[string]CodeInfo, code string) *float64 {
	codeInfo, exists := codesMap[code]
	if !exists {
		return nil
	}

	excelRowNum := codeInfo.RowIndex + 1

	// Получаем количество из колонки D
	quantityCell := fmt.Sprintf("D%d", excelRowNum)
	quantityStr, err := file.GetCellValue(codeInfo.Sheet, quantityCell)
	if err != nil {
		return nil
	}

	quantity, err := strconv.ParseFloat(quantityStr, 64)
	if err != nil {
		return nil
	}

	// Если количество не 0, возвращаем nil
	if quantity != 0 {
		return nil
	}

	// Получаем цену из колонки E
	priceCell := fmt.Sprintf("E%d", excelRowNum)
	priceStr, err := file.GetCellValue(codeInfo.Sheet, priceCell)
	if err != nil {
		return nil
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return nil
	}

	return &price
}

func (p *PriceAnalyzer) writeLog(format string, args ...any) {
	if p.Writer != nil {
		fmt.Fprintf(p.Writer, format, args...)
	}
}
