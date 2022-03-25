package main

import "strings"

var sosNameReplacer = strings.NewReplacer(
	"/", "／",
	"\\", "、",
	":", "：",
	"*", "·",
	"?", "？",
	`"`, "“",
	"<", "《",
	">", "》",
	"|", "¦",
	"\b", "",
)

func sosCleanName(name string) string {
	return strings.TrimSpace(sosNameReplacer.Replace(name))
}
