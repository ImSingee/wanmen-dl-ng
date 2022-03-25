package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"math/rand"
	"os"
	"strings"
	"time"
)

func getList(filename string) ([]string, error) {
	random := false
	if strings.HasSuffix(filename, "?") {
		random = true
		filename = strings.TrimSuffix(filename, "?")
	}

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

	if random {
		rand.Shuffle(len(list), func(i, j int) {
			list[i], list[j] = list[j], list[i]
		})
	}

	return list, nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var list []string

func cobraParseList(cmd *cobra.Command, args []string) (err error) {
	if len(args) == 0 {
		list, err = getList("to_download")
		return
	}

	for i, arg := range args {
		l, err := getList(arg)
		if err != nil {
			return fmt.Errorf("cannot read list %d %s: %w", i+1, arg, err)
		}
		list = append(list, l...)
	}

	return nil
}
