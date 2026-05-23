package log

import (
	"bufio"
	"fmt"
	"os"
)

func InitLogFile(logFile string) (*bufio.Writer, error) {
	file, err := os.Create(logFile)
	if err != nil {
		return nil, fmt.Errorf("создание лог-файла: %w", err)
	}
	writer := bufio.NewWriter(file)
	return writer, nil
}

func FinalizeLog(writer *bufio.Writer) {
	if writer != nil {
		writer.Flush()
	}
}

func WriteSeparator(writer *bufio.Writer) {
	if writer != nil {
		writer.WriteString("------------------------------------------------------------------------\n")
	}
}
