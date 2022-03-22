package main

import "github.com/spf13/cobra"

var app = &cobra.Command{
	Use:           "wanmen-dl",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func main() {

	err := app.Execute()
	if err != nil {
		panic(err)
	}
}
