/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"layers_cli/knowledge"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "A short description about doctor command",
	Long:  `A longer description about doctor command...`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("doctor called")
		// dir := getLayersDir()

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

		directory.Doctor()

	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// doctorCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// doctorCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
