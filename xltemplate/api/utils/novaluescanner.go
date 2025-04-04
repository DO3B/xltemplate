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
			slog.Error("Warning: <no value> detected", "line", lineIndex, "text", scanner.Text())
			//fmt.Fprintf(os.Stderr, "\033[1;33m%s%d%s\033[0m", "Warning: <no value> detected at line ", lineIndex, ",  you may try to access a non existant YAML variable\n")
		}
		lineIndex++
	}
}
