/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// monitorCmd represents the monitor command
var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Opens pm2 monitor",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		shell := exec.Command("pm2", "monit")

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
	ecosystemCmd.AddCommand(monitorCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// monitorCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// monitorCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
