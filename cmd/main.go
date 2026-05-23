package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

const (
	outputFile     = "output.txt"
	fullpriceFile  = "fullprice.xlsx"
	koreaFileName  = "XZAD_Корея.xlsx"
	europeFileName = "XZAP_ Э.xlsx"
	filesDir       = "files"
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
	if err := prepareDirectory(); err != nil {
		return fmt.Errorf("подготовка директории: %w", err)
	}

	// Копирование файлов
	if err := copyFilesToWorkDir(); err != nil {
		return fmt.Errorf("копирование файлов: %w", err)
	}

	// Открытие Excel файлов
	files, err := openExcelFiles()
	if err != nil {
		return fmt.Errorf("открытие файлов: %w", err)
	}
	defer closeExcelFiles(files)

	// Сравнение цен и обновление
	comparator := NewPriceComparator(files.fullprice, files.korea, files.europe)
	if err := comparator.CompareAndUpdate(); err != nil {
		return fmt.Errorf("сравнение цен: %w", err)
	}

	fmt.Println("Обработка завершена успешно. Результат в файле output.txt")
	return nil
}

func prepareDirectory() error {
	return os.MkdirAll(filesDir, 0777)
}

func copyFilesToWorkDir() error {
	files := []string{koreaFileName, europeFileName}
	for _, file := range files {
		if err := copyFile(file, filesDir); err != nil {
			return err
		}
	}
	return nil
}

type excelFiles struct {
	fullprice, korea, europe *excelize.File
}

func openExcelFiles() (*excelFiles, error) {
	fullprice, err := excelize.OpenFile(fullpriceFile)
	if err != nil {
		return nil, fmt.Errorf("открытие fullprice.xlsx: %w", err)
	}

	korea, err := excelize.OpenFile(filesDir + "/" + koreaFileName)
	if err != nil {
		fullprice.Close()
		return nil, fmt.Errorf("открытие %s: %w", koreaFileName, err)
	}

	europe, err := excelize.OpenFile(filesDir + "/" + europeFileName)
	if err != nil {
		fullprice.Close()
		korea.Close()
		return nil, fmt.Errorf("открытие %s: %w", europeFileName, err)
	}

	return &excelFiles{
		fullprice: fullprice,
		korea:     korea,
		europe:    europe,
	}, nil
}

func closeExcelFiles(files *excelFiles) {
	files.fullprice.Close()
	files.korea.Close()
	files.europe.Close()
}

func NewPriceComparator(fullprice, korea, europe *excelize.File) *PriceComparator {
	return &PriceComparator{
		fullprice:   fullprice,
		koreaFile:   korea,
		europeFile:  europe,
		koreaCodes:  make(map[string]CodeInfo),
		europeCodes: make(map[string]CodeInfo),
	}
}

func (p *PriceComparator) CompareAndUpdate() error {
	if err := p.initLogFile(); err != nil {
		return err
	}
	defer p.finalizeLog()

	if err := p.loadCodeIndexes(); err != nil {
		return err
	}

	return p.processFullpriceRows()
}

func (p *PriceComparator) initLogFile() error {
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("создание лог-файла: %w", err)
	}
	p.writer = bufio.NewWriter(file)
	return nil
}

func (p *PriceComparator) finalizeLog() {
	if p.writer != nil {
		p.writer.Flush()
	}
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

		p.processRow(i, row)
	}

	return p.saveChanges()
}

func (p *PriceComparator) processRow(rowIndex int, row []string) {
	code := strings.TrimSpace(row[1])
	quantity := strings.TrimSpace(row[3])

	if code == "" {
		return
	}

	p.writeSeparator()

	if quantity == "0" {
		p.handleZeroQuantity(rowIndex, code, quantity)
	} else {
		p.writeLog("Строка %d, Код: %s, Остаток: %s - без изменений\n", rowIndex+1, code, quantity)
	}
}

func (p *PriceComparator) handleZeroQuantity(rowIndex int, code, quantity string) {
	p.writeLog("Строка %d, Код: %s, Остаток: %s - Ищем в других файлах.\n", rowIndex+1, code, quantity)

	updated := false

	if p.updateQuantityInFile(p.koreaFile, p.koreaCodes, code, quantity) {
		p.writeLog("  ✓ Обновлено в XZAD_Корея.xlsx\n")
		updated = true
	}

	if p.updateQuantityInFile(p.europeFile, p.europeCodes, code, quantity) {
		p.writeLog("  ✓ Обновлено в XZAP_ Э.xlsx\n")
		updated = true
	}

	if !updated {
		p.writeLog("  ✗ Код %s не найден ни в одном из файлов\n", code)
	}
}

func (p *PriceComparator) updateQuantityInFile(file *excelize.File, codesMap map[string]CodeInfo, code, quantity string) bool {
	codeInfo, exists := codesMap[code]
	if !exists {
		return false
	}

	excelRowNum := codeInfo.RowIndex + 1
	cellName := fmt.Sprintf("D%d", excelRowNum)

	file.SetCellStr(codeInfo.Sheet, cellName, quantity)
	return true
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

func (p *PriceComparator) writeLog(format string, args ...interface{}) {
	if p.writer != nil {
		fmt.Fprintf(p.writer, format, args...)
	}
}

func (p *PriceComparator) writeSeparator() {
	p.writeLog("------------------------------------------------------------------------\n")
}

func copyFile(filename, dir string) error {
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
