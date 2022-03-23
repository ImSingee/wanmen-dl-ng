package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"strings"
)

var cmdSplit = &cobra.Command{
	Use:   "split <file> <size>",
	Short: "Split a file into multiple files",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("split requires two arguments")
		}

		file := args[0]
		sizeString := args[1]

		size, err := strconv.Atoi(sizeString)
		if err != nil {
			return fmt.Errorf("size must be an integer")
		}
		if size < 1 {
			return fmt.Errorf("size must be greater than zero")
		}

		fileData, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("could not read file: %v", err)
		}

		parts := strings.Split(strings.TrimSpace(string(fileData)), "\n")
		result := make([]string, 0, len(parts))

		for _, part := range parts {
			part = strings.TrimSpace(part)

			if part != "" && !strings.HasPrefix(part, "#") {
				result = append(result, part)
			}
		}

		i := 1
		for len(result) != 0 {
			name := file + "-" + strconv.Itoa(i)

			var thisRound []string
			thisRound, result = shift(result, size)

			err = os.WriteFile(name, []byte(strings.Join(thisRound, "\n")), 0644)
			if err != nil {
				return fmt.Errorf("cannot write %v", err)
			}

			fmt.Println(name)
			i++
		}

		return nil
	},
}

func shift(slice []string, size int) ([]string, []string) {
	if size >= len(slice) {
		return slice, nil
	} else {
		return slice[:size], slice[size:]
	}
}

func init() {
	app.AddCommand(cmdSplit)
}
