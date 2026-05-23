package excelutil

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

type ExcelFiles struct {
	Fullprice, Korea, Europe *excelize.File
}

func CloseExcelFiles(files *ExcelFiles) {
	files.Fullprice.Close()
	files.Korea.Close()
	files.Europe.Close()
}

func OpenExcelFiles(filelist []string) (*ExcelFiles, error) {
	excelFiles := &ExcelFiles{}

	for i, file := range filelist {
		f, err := excelize.OpenFile(file)
		if err != nil {
			return nil, fmt.Errorf("открытие %s: %w", file, err)
		}
		switch i {
		case 0:
			excelFiles.Fullprice = f
		case 1:
			excelFiles.Korea = f
		case 2:
			excelFiles.Europe = f
		}
	}

	return excelFiles, nil
}
