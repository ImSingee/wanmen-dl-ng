package main

import "strings"

var nameReplacer = strings.NewReplacer(
	"\\", " ",
	"/", " ",
	":", " ",
	"*", " ",
	"?", " ",
	`'`, " ",
	`"`, " ",
	"<", " ",
	">", " ",
	"|", " ",
)

func cleanName(name string) string {
	return nameReplacer.Replace(name)
}
