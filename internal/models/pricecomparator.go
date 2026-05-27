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

type PriceComparator struct {
	Fullprice   *excelize.File
	KoreaFile   *excelize.File
	EuropeFile  *excelize.File
	Writer      *bufio.Writer
	KoreaCodes  map[string]CodeInfo
	EuropeCodes map[string]CodeInfo
}

func (p *PriceComparator) CompareAndUpdate() error {
	var err error
	p.Writer, err = log.InitLogFile(constants.XmlCompareLogFile)
	if err != nil {
		return err
	}
	defer log.FinalizeLog(p.Writer)

	// Заголовок таблицы
	p.writeLog("%-15s | %-20s | %-15s | %s\n",
		"Код", "Марка", "Старое кол-во", "Файл")
	p.writeLog("%s\n", strings.Repeat("-", 80))

	if err := p.loadCodeIndexes(); err != nil {
		return err
	}

	return p.processFullpriceRows()
}

func (p *PriceComparator) LoadCodesFromFile(file *excelize.File, codesMap map[string]CodeInfo) error {
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

func (p *PriceComparator) loadCodeIndexes() error {
	if err := p.LoadCodesFromFile(p.KoreaFile, p.KoreaCodes); err != nil {
		return fmt.Errorf("загрузка кодов из Кореи: %w", err)
	}

	if err := p.LoadCodesFromFile(p.EuropeFile, p.EuropeCodes); err != nil {
		return fmt.Errorf("загрузка кодов из Европы: %w", err)
	}

	return nil
}

func (p *PriceComparator) processFullpriceRows() error {
	sheetName := p.Fullprice.GetSheetName(0)
	rows, err := p.Fullprice.GetRows(sheetName)
	if err != nil {
		return err
	}

	for i, row := range rows {
		if i == 0 || len(row) < 5 {
			continue
		}

		p.processRow(row)
	}

	return p.saveChanges()
}

func (p *PriceComparator) processRow(row []string) {
	code := strings.TrimSpace(row[1])     // Колонка B - код
	brand := strings.TrimSpace(row[4])    // Колонка E - марка
	quantity := strings.TrimSpace(row[3]) // Колонка D - количество

	if code == "" {
		return
	}

	if quantity == "0" {
		p.handleZeroQuantity(code, brand)
	}
}

func (p *PriceComparator) handleZeroQuantity(code, brand string) {
	p.updateQuantityInFile(p.KoreaFile, p.KoreaCodes, code, brand, "XZAD_Корея.xlsx")
	p.updateQuantityInFile(p.EuropeFile, p.EuropeCodes, code, brand, "XZAP_Э.xlsx")
}

func (p *PriceComparator) updateQuantityInFile(file *excelize.File, codesMap map[string]CodeInfo, code, brand, fileName string) {
	codeInfo, exists := codesMap[code]
	if !exists {
		return
	}

	excelRowNum := codeInfo.RowIndex + 1
	cellName := fmt.Sprintf("D%d", excelRowNum)
	quantity, err := file.GetCellValue(codeInfo.Sheet, cellName)
	if err != nil {
		fmt.Println(err)
		return
	}

	floatQuantity, err := strconv.ParseFloat(quantity, 64)
	if err != nil {
		fmt.Println(err)
		return
	}

	if floatQuantity != 0 {
		file.SetCellFloat(codeInfo.Sheet, cellName, 0, -1, 64)
		quantityFormatted := formatQuantity(floatQuantity)
		p.writeLog("%-15s | %-20s | %-15s | %s\n",
			code, brand, quantityFormatted, fileName)
	}
}

func (p *PriceComparator) saveChanges() error {
	if err := p.KoreaFile.Save(); err != nil {
		return fmt.Errorf("ошибка сохранения XZAD_Корея.xlsx: %w", err)
	}
	if err := p.EuropeFile.Save(); err != nil {
		return fmt.Errorf("ошибка сохранения XZAP_Э.xlsx: %w", err)
	}
	return nil
}

func (p *PriceComparator) writeLog(format string, args ...any) {
	if p.Writer != nil {
		fmt.Fprintf(p.Writer, format, args...)
	}
}

// formatQuantity убирает .0 у целых чисел
func formatQuantity(q float64) string {
	if q == float64(int64(q)) {
		return fmt.Sprintf("%.0f", q)
	}
	return fmt.Sprintf("%.2f", q)
}
