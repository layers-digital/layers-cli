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

	"github.com/spf13/cobra"
)

type Command struct {
	run func(args []string) bool
}

var commandsDictionary = map[string]map[string]Command{
	"specs": {
		"nodejs": {
			run: func(args []string) (succeeded bool) {
				succeeded = false

				requiredVersion := args[0]

				nodeVersionOutput, err := exec.Command("node", "--version").Output()
				if err != nil {
					log.Fatal("couldn't get node version, maybe node is not installed yet")
				}

				nodeVersion := string(nodeVersionOutput)

				// v12.22.8 -> 12.22.8 -> 12
				major := strings.Split(strings.ReplaceAll(nodeVersion, "v", ""), ".")[0]

				majorMatches := major == requiredVersion

				if majorMatches {
					succeeded = true
				}

				return succeeded
			},
		},
		"mongodb": {
			run: func(args []string) (succeeded bool) {
				succeeded = false
				dockerInstances, err := exec.Command("docker", "ps").Output()
				if err != nil {
					log.Fatalln(err)
				}

				dockerLines := strings.Split(string(dockerInstances), "\n")
				for i, dockerLine := range dockerLines {
					if i == 0 || dockerLine == "" {
						// ignore header and empty lines
						continue
					}

					dockerColumns := strings.Split(dockerLine, "   ")

					// mongo:4.4.1 -> mongo
					isMongoDB := strings.Split(dockerColumns[1], ":")[0] == "mongo"
					isCorrectPort := strings.Contains(dockerColumns[5], "27017")
					isCorrectInstance := dockerColumns[6] == "tendaedu-backend_mongo_1" || dockerColumns[6] == "payments_mongo_1"

					if isCorrectInstance && isCorrectPort && isMongoDB {
						succeeded = true
						break
					}
				}

				return succeeded
			},
		},
	},
}

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "A short description about doctor command",
	Long:  `A longer description about doctor command...`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("doctor called")
		// dir := getLayersDir()

		instructions := getLayersInstructions()

		fmt.Println("Layers's instructions slice")
		for _, instruction := range instructions {
			fmt.Println(instruction)
		}

		results := runInstructions(instructions)

		for _, result := range results {
			fmt.Printf("%s -> %s\n", result.command, result.status)
		}
	},
}

type Instruction struct {
	status        string   // "toRun" | "failed" | "success"
	kind          string   // "specs" | "requires" | "steps"
	command       string   // ex: "mongodb", "nodejs", "redis"
	args          []string // command arguments
	resultMessage string
}

func getLayersInstructions() (instructions []Instruction) {

	configFile, err := os.ReadFile("./layers.md")
	if err != nil {
		log.Fatal(err)
	}

	rawLines := strings.Split(string(configFile), "\n")
	var lines []string

	fmt.Println("getting Layers's instructions")

	for i, line := range rawLines {
		fmt.Printf("getting line %d of %d\n", i+1, len(rawLines))
		isValid := strings.HasPrefix(line, "###") || strings.HasPrefix(line, "-")

		if isValid {
			lines = append(lines, line)
		}
	}

	// for _, lines := range lines {
	// 	println(lines)
	// }

	kind := "none"

	for _, line := range lines {
		if strings.HasPrefix(line, "###") {
			// "### STEPS anything" -> "steps"
			words := strings.Split(line, " ")
			kind = strings.ToLower(words[1])
			fmt.Println("getting header " + kind)
		} else if strings.HasPrefix(line, "-") {
			// fmt.Println("getting instruction with kind " + kind)
			if kind == "none" {
				log.Fatalln("no header was setted before the line: " + line)
			}

			// maybe it's a golang "gambiarra" to remove substrings from a string
			input := strings.ReplaceAll(line, "- ", "")

			splittedInputs := strings.Split(input, " ")

			var command string
			var args []string

			for i, word := range splittedInputs {
				if i == 0 {
					command = word
				} else {
					args = append(args, word)
				}
			}

			// maybe there is a better way

			instruction := Instruction{
				status:        "toRun",
				kind:          kind,
				command:       command,
				args:          args,
				resultMessage: "",
			}

			instructions = append(instructions, instruction)
		} else {
			log.Fatalln("an error occured to get instruction of this line: " + line)
		}
	}

	return instructions
}

func runInstructions(instructions []Instruction) (results []Instruction) {
	results = []Instruction{} // initialize empty slice

	for _, instruction := range instructions {

		// TODO: accept more kinds
		if instruction.kind != "specs" {
			continue
		}
		succeeded := commandsDictionary[instruction.kind][instruction.command].run(instruction.args)

		result := instruction

		if succeeded {
			result.status = "succeeded"
		} else {
			result.status = "failed"
		}

		results = append(results, result)
	}

	return results
}

func getLayersDir() (dir string) {
	out, err := exec.Command("pwd").Output()
	if err != nil {
		log.Fatal(err)
	}

	dirNames := strings.Split(string(out), "/")
	dir = dirNames[len(dirNames)-1]

	return dir
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
