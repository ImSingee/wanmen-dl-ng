package main

import (
	"fmt"
	"os"
	"strings"
)

func getList(filename string) ([]string, error) {
	f, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(f)), "\n")

	list := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		list = append(list, line)
	}

	return list, nil
}
