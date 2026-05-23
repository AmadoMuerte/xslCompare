package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/AmadoMuerte/xslCompare/internal/constants"
	"github.com/AmadoMuerte/xslCompare/internal/excelutil"
	"github.com/AmadoMuerte/xslCompare/internal/fileutil"
	"github.com/AmadoMuerte/xslCompare/internal/log"
	"github.com/xuri/excelize/v2"
)

type CodeInfo struct {
	RowIndex int
	Sheet    string
}

type PriceComparator struct {
	fullprice   *excelize.File
	koreaFile   *excelize.File
	europeFile  *excelize.File
	writer      *bufio.Writer
	koreaCodes  map[string]CodeInfo
	europeCodes map[string]CodeInfo
}

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

	// Сравнение цен и обновление
	comparator := NewPriceComparator(files)
	if err := comparator.CompareAndUpdate(); err != nil {
		return fmt.Errorf("сравнение цен: %w", err)
	}

	fmt.Println("Обработка завершена успешно. Результат в файле output.txt")
	return nil
}

func NewPriceComparator(files *excelutil.ExcelFiles) *PriceComparator {
	return &PriceComparator{
		fullprice:   files.Fullprice,
		koreaFile:   files.Korea,
		europeFile:  files.Europe,
		koreaCodes:  make(map[string]CodeInfo),
		europeCodes: make(map[string]CodeInfo),
	}
}

func (p *PriceComparator) CompareAndUpdate() error {
	var err error
	p.writer, err = log.InitLogFile(constants.XmlCompareLogFile)
	if err != nil {
		return err
	}
	defer log.FinalizeLog(p.writer)

	if err := p.loadCodeIndexes(); err != nil {
		return err
	}

	return p.processFullpriceRows()
}

func (p *PriceComparator) loadCodeIndexes() error {
	if err := p.loadCodesFromFile(p.koreaFile, p.koreaCodes); err != nil {
		return fmt.Errorf("загрузка кодов из Кореи: %w", err)
	}

	if err := p.loadCodesFromFile(p.europeFile, p.europeCodes); err != nil {
		return fmt.Errorf("загрузка кодов из Европы: %w", err)
	}

	return nil
}

func (p *PriceComparator) loadCodesFromFile(file *excelize.File, codesMap map[string]CodeInfo) error {
	sheetName := file.GetSheetName(0)
	rows, err := file.GetRows(sheetName)
	if err != nil {
		return err
	}

	for i, row := range rows {
		if i == 0 || len(row) == 0 {
			continue
		}

		code := strings.TrimSpace(row[0])
		if code != "" {
			codesMap[code] = CodeInfo{
				RowIndex: i,
				Sheet:    sheetName,
			}
		}
	}

	return nil
}

func (p *PriceComparator) processFullpriceRows() error {
	sheetName := p.fullprice.GetSheetName(0)
	rows, err := p.fullprice.GetRows(sheetName)
	if err != nil {
		return err
	}

	for i, row := range rows {
		if i == 0 || len(row) < 4 {
			continue
		}

		p.processRow(row)
	}

	return p.saveChanges()
}

func (p *PriceComparator) processRow(row []string) {
	code := strings.TrimSpace(row[1])
	quantity := strings.TrimSpace(row[3])

	if code == "" {
		return
	}

	if quantity == "0" {
		p.handleZeroQuantity(code)
	}
}

func (p *PriceComparator) handleZeroQuantity(code string) {
	if updated, quantity := p.updateQuantityInFile(p.koreaFile, p.koreaCodes, code); updated {
		p.writeLog("Код: %s, Старое количество: %f,  ✓ Обновлено в XZAD_Корея.xlsx\n", code, quantity)
	}

	if updated, quantity := p.updateQuantityInFile(p.europeFile, p.europeCodes, code); updated {
		p.writeLog("Код: %s, Старое количество: %f,  ✓ Обновлено в XZAP_ Э.xlsx\n", code, quantity)
	}
}

func (p *PriceComparator) updateQuantityInFile(file *excelize.File, codesMap map[string]CodeInfo, code string) (bool, float64) {
	codeInfo, exists := codesMap[code]
	if !exists {
		return false, 0
	}

	excelRowNum := codeInfo.RowIndex + 1
	cellName := fmt.Sprintf("D%d", excelRowNum)
	quantity, err := file.GetCellValue(codeInfo.Sheet, cellName)
	if err != nil {
		fmt.Println(err)
		return false, 0
	}
	floatQuantity, err := strconv.ParseFloat(quantity, 64)
	if err != nil {
		fmt.Println(err)
		return false, 0
	}

	if floatQuantity != 0 {
		file.SetCellFloat(codeInfo.Sheet, cellName, 0, -1, 64)
		return true, floatQuantity
	}
	return false, 0
}

func (p *PriceComparator) saveChanges() error {
	if err := p.koreaFile.Save(); err != nil {
		return fmt.Errorf("ошибка сохранения XZAD_Корея.xlsx: %w", err)
	}
	if err := p.europeFile.Save(); err != nil {
		return fmt.Errorf("ошибка сохранения XZAP_ Э.xlsx: %w", err)
	}
	return nil
}

func (p *PriceComparator) writeLog(format string, args ...any) {
	if p.writer != nil {
		fmt.Fprintf(p.writer, format, args...)
	}
}
