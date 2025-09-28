package utils

import "strings"

func GetName(name string) string {
	finalName := name
	finalName = strings.ReplaceAll(finalName, " ", "_")
	finalName = strings.ReplaceAll(finalName, "-", "_")
	finalName = strings.ReplaceAll(finalName, ".", "_")
	finalName = strings.ReplaceAll(finalName, "'", "_")
	finalName = strings.ReplaceAll(finalName, "(", "_")
	finalName = strings.ReplaceAll(finalName, ")", "_")
	finalName = strings.ReplaceAll(finalName, "/", "_")
	finalName = strings.ReplaceAll(finalName, "&", "_")
	finalName = strings.ReplaceAll(finalName, "%", "_")
	finalName = strings.ToLower(finalName)

	if len(finalName) > 50 {
		finalName = finalName[:50]
	}

	return finalName
}
