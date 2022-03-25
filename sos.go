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

func getSosName(id string) (string, bool) {
	x := map[string]string{
		"60404dcb0c8d7de1725212e2": "计算机二级 MS Office精品课",
		"61134068fc88509376ff37df": "一级注册消防工程师：技术实务",
		"611342bd50ebac2a2cb301d5": "一级注册消防工程师：案例分析",
		"61134301b55b27f56cf93b5e": "一级注册消防工程师：全科刷题班",
	}

	if name, ok := x[id]; ok {
		return name, true
	}

	return GetName(id)

}
