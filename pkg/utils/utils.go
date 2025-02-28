package utils

import (
	"fmt"
	"strings"
)

func FormatAsMarkdown(header []string, rows [][]any) string {
	var sb strings.Builder
	sb.WriteString("| " + strings.Join(header, " | ") + " |\n")
	separator := make([]string, len(header))
	for i := range separator {
		separator[i] = "---"
	}
	sb.WriteString("| " + strings.Join(separator, " | ") + " |\n")

	for _, row := range rows {
		var rowValues []string
		for _, col := range row {
			var s string
			b, ok := col.([]byte)
			if ok {
				s = string(b)
			} else {
				s = fmt.Sprintf("%v", col)
			}
			s = strings.ReplaceAll(s, "\n", "\\n")
			rowValues = append(rowValues, s)
		}
		sb.WriteString("| " + strings.Join(rowValues, " | ") + " |\n")
	}

	return sb.String()
}
