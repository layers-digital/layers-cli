/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func checkProcessAtPort(port string, name string, dir string) (succeeded bool, reason string, solution string) {
	succeeded = false
	reason = ""
	solution = ""

	process, err := exec.Command("lsof", "-i", ":8009").Output()
	if err != nil {
		reason = fmt.Sprintf("No process running at port: %s", port)
		solution = fmt.Sprintf("Start the server for `%s`", dir)
		return succeeded, reason, solution
	}
	lines := strings.Split(string(process), "\n")

	// lines[0] is the header
	/*
		Example:
		COMMAND   PID USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
		node    59526 user   66u  IPv6 0xc2851b8f1c265133      0t0  TCP *:8009 (LISTEN)
	*/

	processLine := lines[1]

	hasCommandName := strings.HasPrefix(processLine, name)

	if hasCommandName {
		succeeded = true
		return succeeded, reason, solution
	}

	reason = fmt.Sprintf("Process running at port `%s` isn't `%s`", port, name)
	solution = fmt.Sprintf("Start the server for `%s`", dir)
	return succeeded, reason, solution
}

func dockerInstallationCheckup() (succeeded bool, reason string, solution string) {
	succeeded = false
	reason = ""
	solution = ""

	defaultReturn := map[string]func() (bool, string, string){
		"docker": func() (succeeded bool, reason string, solution string) {
			succeeded = false
			reason = "Docker engine isn't installed yet."
			solution = "Checkout this link https://docs.docker.com/engine/install/"

			return succeeded, reason, solution
		},
		"docker-compose": func() (succeeded bool, reason string, solution string) {
			succeeded = false
			reason = "docker-compose isn't installed yet."
			solution = "Checkout this link https://docs.docker.com/compose/install/"

			return succeeded, reason, solution
		},
	}

	dockerInstallation, err := exec.Command("docker", "--version").Output()
	if err != nil {
		defaultReturn["docker"]()
	}

	dockerMessage := string(dockerInstallation)

	if !strings.Contains(dockerMessage, "Docker version") {
		defaultReturn["docker"]()
	}

	dockerComposeInstallation, err := exec.Command("docker", "--version").Output()
	if err != nil {
		defaultReturn["docker-compose"]()
	}

	dockerComposeMessage := string(dockerComposeInstallation)

	if !strings.Contains(dockerComposeMessage, "docker-compose version") {
		defaultReturn["docker-compose"]()
	}

	succeeded = true
	return succeeded, solution, reason
}

func getDockerLine(imageName string) (columns []string) { // TODO: accept image version
	dockerInstances, err := exec.Command("docker", "ps").Output()

	columns = []string{}

	if err != nil {
		return columns
	}

	dockerLines := strings.Split(string(dockerInstances), "\n")
	for i, line := range dockerLines {
		if i == 0 || line == "" {
			// ignore header and empty lines
			continue
		}

		dockerColumns := strings.Split(line, "   ")

		lineImageName := strings.Split(dockerColumns[1], ":")[0]

		containsImageName := strings.Contains(dockerColumns[1], imageName)

		if lineImageName == imageName || containsImageName {
			columns = dockerColumns
			break
		}
	}

	return columns
}

type Command struct {
	run func(args []string) (succeeded bool, reason string, solution string)
}

var staticErrors = map[string]map[string]string{
	"node_installation": {
		"reason":   "Couldn't get node version, maybe node isn't installed yet",
		"solution": "Install node using nvm(https://github.com/nvm-sh/nvm) or asdf(https://github.com/asdf-vm/asdf-nodejs)",
	},
	"docker_init": {
		"solution": "Be sure that you ran `sudo docker-compose up -d` at `tendaedu-backend` or `payments`",
	},
}

var commandsDictionary = map[string]map[string]Command{
	"check": {
		"nodejs": {
			run: func(args []string) (succeeded bool, reason string, solution string) {
				succeeded = false
				reason = ""
				solution = ""

				requiredVersion := args[0]

				nodeVersionOutput, err := exec.Command("node", "--version").Output()
				if err != nil {
					reason = staticErrors["node_installation"]["reason"]
					solution = staticErrors["node_installation"]["solution"]
					return succeeded, reason, solution
				}

				nodeVersion := string(nodeVersionOutput)

				// v12.22.8 -> 12.22.8 -> 12
				major := strings.Split(strings.ReplaceAll(nodeVersion, "v", ""), ".")[0]

				majorMatches := major == requiredVersion

				if majorMatches {
					succeeded = true
				} else {
					reason = fmt.Sprintf("wrong node version. Expected `%s` but got `%s`", requiredVersion, major)
					solution = fmt.Sprintf("run `nvm use %s` or `asdf local nodejs %s`", requiredVersion, requiredVersion)
				}

				return succeeded, reason, solution
			},
		},
		"mongodb": {
			run: func(args []string) (succeeded bool, reason string, solution string) {
				succeeded = false
				defaultSolution := staticErrors["docker_init"]["solution"]

				succeeded, reason, solution = dockerInstallationCheckup()
				if !succeeded {
					return succeeded, reason, solution
				}

				mongoInstanceColumns := getDockerLine("mongo")
				if len(mongoInstanceColumns) == 0 {
					succeeded = false
					reason = "There is no MongoDB instance running in Docker."
					solution = defaultSolution
					return succeeded, reason, solution
				}

				// mongo:4.4.1 -> mongo
				isMongoDB := strings.Split(mongoInstanceColumns[1], ":")[0] == "mongo"
				isCorrectPort := strings.Contains(mongoInstanceColumns[5], "27017")
				isCorrectInstance := mongoInstanceColumns[6] == "tendaedu-backend_mongo_1" || mongoInstanceColumns[6] == "payments_mongodb-primary_1"

				if isCorrectInstance && isCorrectPort && isMongoDB {
					succeeded = true
				} else {
					reason = "Didn't find any MongoDB at docker instances."
					solution = defaultSolution
				}

				return succeeded, reason, solution
			},
		},
		"redis": {
			run: func(args []string) (succeeded bool, reason string, solution string) {
				succeeded = false
				reason = ""
				solution = ""

				defaultSolution := staticErrors["docker_init"]["solution"]

				succeeded, reason, solution = dockerInstallationCheckup()
				if !succeeded {
					return succeeded, reason, solution
				}

				redisInstanceColumn := getDockerLine("redis")
				if len(redisInstanceColumn) == 0 {
					succeeded = false
					reason = "There is no Redis instance running in Docker."
					solution = defaultSolution
					return succeeded, reason, solution
				}

				isCorrectPort := strings.Contains(redisInstanceColumn[5], "27017")
				isCorrectInstance := redisInstanceColumn[6] == "tendaedu-backend_redis_1" || redisInstanceColumn[6] == "payments_redis_1"

				if isCorrectInstance && isCorrectPort {
					succeeded = true
				} else {
					reason = "Didn't find any MongoDB at docker instances."
					solution = defaultSolution
				}

				return succeeded, reason, solution
			},
		},
		"file": {
			run: func(args []string) (succeeded bool, reason string, solution string) {
				path := args[0]
				succeeded = false
				reason = ""
				solution = ""

				directories := strings.Split(path, "/")
				filename := directories[len(directories)-1]

				if _, err := os.Stat(path); err == nil {
					file, err := os.ReadFile(path)
					if err != nil {
						reason = fmt.Sprintf("Couldn't read the file content at path: %s", path)
						solution = fmt.Sprintf("Be sure that a file exists at this path: %s", path)
					}
					if len(string(file)) == 0 {
						reason = fmt.Sprintf("File at path '%s' is empty.", path)
						solution = fmt.Sprintf("If you don't have it, call some experient dev to get this, or maybe you can get this in the repository's environment variables section.")
					} else {
						succeeded = true
					}
				} else if errors.Is(err, os.ErrNotExist) {
					reason = fmt.Sprintf("File at path '%s' doesn't exist", path)
					solution = fmt.Sprintf("Be sure that file '%s' is at this path '%s'. If you don't have it, call some experient dev to get this, or maybe you can get this in the repository's environment variables section.", filename, path)
				} else {
					// fatal error
					reason = fmt.Sprintf("Unknown file error. Error: %s", err)
					solution = "Unfortunately I don't know the solution for this error :("
				}
				return succeeded, reason, solution
			},
		},
		"layersDir": {
			run: func(args []string) (succeeded bool, reason string, solution string) {
				dir := args[0]
				ports := map[string]string{
					"tendaedu-backend":    "8009",
					"layers-auth-vanilla": "8090",
					"layers-webapp":       "8090",
				}

				succeeded, reason, solution = checkProcessAtPort(ports[dir], "node", dir)
				return succeeded, reason, solution
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

		fmt.Println("\nLayers's instructions slices")
		for _, instruction := range instructions {
			fmt.Println(instruction)
		}

		results := runInstructions(instructions)

		fmt.Println("\nRESULTS")
		for _, result := range results {
			fmt.Printf("%s -> %s\n", result.command, result.status)
			if result.status == "failed" {
				fmt.Println(result.resultMessage)
			}
			fmt.Println("")
		}
	},
}

type Instruction struct {
	status        string   // "toRun" | "failed" | "success"
	kind          string   // "check" | "run"
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

		// TODO: doesn't accept run steps yet
		if instruction.kind != "check" {
			continue
		}
		succeeded, reason, solution := commandsDictionary[instruction.kind][instruction.command].run(instruction.args)

		result := instruction

		if succeeded {
			result.status = "succeeded"
		} else {
			result.status = "failed"
			result.resultMessage = fmt.Sprintf("Reason: %s\nSolution: %s", reason, solution)
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
