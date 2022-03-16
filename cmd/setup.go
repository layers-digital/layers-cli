/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"layers_cli/knowledge"

	"github.com/spf13/cobra"
)

func getActualDir() (dir string) {
	dir = ""
	absolutePath, err := exec.Command("pwd").Output()
	if err != nil {
		log.Fatalln("Couldn't get actual path")
	}
	cleanAbsolutePath := strings.TrimSpace(string(absolutePath))

	splittedPaths := strings.Split(cleanAbsolutePath, "/")

	dir = splittedPaths[len(splittedPaths)-1]
	return dir
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup ecosystem for the current project",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("setup called")

		isInitial, _ := cmd.Flags().GetBool("initial")
		populate, _ := cmd.Flags().GetBool("populate")

		colorGreen := "\033[32m"
		colorReset := "\033[0m"

		actual := getActualDir()
		fmt.Println(actual)

		if !knowledge.IsLayersDirectory(actual) {
			log.Fatalln(actual + " isn't a known directory")
		}

		rootDir := os.Getenv("LAYERS_PATH")
		if rootDir == "" {
			rootDir = "../"
		}

		directory, err := knowledge.New(actual)
		if err != nil {
			log.Fatal(err.Error())
		}

		succeeded, err := directory.Setup(args, isInitial, populate)
		if err != nil {
			log.Fatalf("Couldn't set up `%s` :(. Error: %s\n", actual, err.Error())
		}

		if succeeded {
			message := fmt.Sprintf("Yay! `%s` is already set up. Use `layers-cli ecosystem run` to run layers's ecosystem for this repository.\n", actual)
			fmt.Println(string(colorGreen), message, string(colorReset))
		}
	},
}

func init() {
	ecosystemCmd.AddCommand(setupCmd)

	setupCmd.PersistentFlags().Bool("initial", true, "Run initial installation.")
	setupCmd.PersistentFlags().Bool("populate", true, "Run populate.")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setupCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
