/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run ecosystem",
	Long: `Run the ecosystem for the current project.
	Be sure to run "layers ecosystem setup" before this`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("run called")
		rootDir := os.Getenv("LAYERS_PATH")
		if rootDir == "" {
			rootDir = "../"
		}

		dir := getActualDir()

		if _, err := os.Stat(fmt.Sprintf("%s/%s.config.js", rootDir, dir)); err != nil {
			log.Fatalf("Didn't find any config for `%s`. Be sure that you ran `layers ecosystem setup`\n", dir)
		}

		shell := exec.Command("pm2", "start", dir+".config.js")

		shell.Dir = rootDir
		shell.Stdout = os.Stdout
		shell.Stderr = os.Stderr
		shell.Stdin = os.Stdin
		err := shell.Run()
		if err != nil {
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}

	},
}

func init() {
	ecosystemCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
