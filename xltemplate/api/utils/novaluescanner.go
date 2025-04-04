package utils

import (
	"bufio"
	"bytes"
	"log/slog"
	"strings"
)

func NoValueScan(result string) {
	scanner := bufio.NewScanner(bytes.NewBufferString(result))
	lineIndex := 1
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "<no value>") {
			slog.Warn("<no value> detected", "line", lineIndex, "text", scanner.Text())
		}
		lineIndex++
	}
}
