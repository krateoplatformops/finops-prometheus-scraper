package utils

import (
	"strings"
)

type ContextKey string

const CertFilePathKey ContextKey = "certFilePath"

func ConvertIntoRowWithMaximumLength10k(allRows []string, byteSizeLimit int) []string {
	result := []string{}
	temp := ""
	for _, row := range allRows {
		if len(temp)+len(row) > byteSizeLimit {
			result = append(result, strings.Trim(temp, "\n"))
			temp = ""
		}
		temp += row + "\n"
	}
	if temp != "" {
		return append(result, strings.Trim(temp, "\n"))
	}
	return result
}
