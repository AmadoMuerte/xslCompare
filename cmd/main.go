package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/xuri/excelize/v2"
)

func main() {
	err := os.Mkdir("files", 0777)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	copyFile("XZAD_Корея.xlsx", "files")
	copyFile("XZAP_ Э.xlsx", "files")

	fPrice, err := excelize.OpenFile("fullprice.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}

	xzadKorea, err := excelize.OpenFile("files/XZAD_Корея.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}

	xzadE, err := excelize.OpenFile("files/XZAP_ Э.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		// Close the spreadsheet.
		if err := fPrice.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	defer func() {
		// Close the spreadsheet.
		if err := xzadKorea.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	defer func() {
		// Close the spreadsheet.
		if err := xzadE.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	comparePrice(fPrice, xzadKorea, xzadE)
}

func comparePrice(fullprice, xzadKorea, xzadE *excelize.File) {
	rows, err := fullprice.GetRows(fullprice.GetSheetName(0))
	if err != nil {
		fmt.Println(err)
		return
	}

	file, err := os.Create("output.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for i, row := range rows {
		code := row[1]
		quantity := row[3]

		text := fmt.Sprintf("Строка %d, B%d: %s, Остаток: %s\n", i+1, i+1, code, quantity)
		writer.WriteString(text)
	}

	writer.Flush()
}

/*
 *
 *
 *
 */

func copyFile(filename, dir string) error {
	// Открываем исходный файл
	sourceFile, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("ошибка открытия источника: %w", err)
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dir + "/" + filename)
	if err != nil {
		return fmt.Errorf("ошибка создания получателя: %w", err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("ошибка копирования данных: %w", err)
	}

	return nil
}
