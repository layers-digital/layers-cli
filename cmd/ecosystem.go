/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/spf13/cobra"
)

// ecosystemCmd represents the ecosystem command
var ecosystemCmd = &cobra.Command{
	Use:   "ecosystem",
	Short: "Manage Layers Development Ecosystem",
	Long: `Manage Layers Development Ecosystem
This manage were built with pm2 package.`,
}

func init() {
	rootCmd.AddCommand(ecosystemCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ecosystemCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ecosystemCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
