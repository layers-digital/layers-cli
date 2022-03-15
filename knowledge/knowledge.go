package knowledge

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

const name = "Layers"

type RunInstructionCommand struct {
	starter string   // commands to start. Ex: { "yarn build" "yarn start" }
	setup   []string // commands to setup. Ex: { "yarn install" "yarn populate" }
}

type RunInstruction struct {
	path        string   // "root" | "web" | "app" | "package/sponte" ...
	interpreter string   // interpreter:major ex: "node:14"
	needs       []string // { mongodb redis file:fileName }
	commands    RunInstructionCommand
}

type LayersDirectory struct {
	requires     []string // layers directories. Ex: { "tendaedu-backend" "layers-auth-vanilla"}
	instructions []RunInstruction
}

type DirectoryKnowledge struct {
	name string // some-layers-directory
	this LayersDirectory
}

func New(name string) (*DirectoryKnowledge, error) {
	dir, err := getDirectoryKnowledgeByName(name)
	if err != nil {
		return &DirectoryKnowledge{}, errors.New(fmt.Sprintf("Couldn't initialize knowledge for `%s`. Reason: %s", name, err.Error()))
	}

	return &DirectoryKnowledge{
		name: name,
		this: dir,
	}, nil
}

type CheckItem struct {
	name      string
	succeeded bool
	reason    string
	solution  string
}

func printCheckItems(checkItems []CheckItem) {
	for _, check := range checkItems {
		fmt.Printf("%s -> succeeded %v\n", check.name, check.succeeded)
		if !check.succeeded {
			fmt.Printf("reason: %s\nsolution: %s\n", check.reason, check.solution)
		}
		fmt.Println()
	}
}

func (directory *DirectoryKnowledge) Doctor() {
	checkItems := []CheckItem{}

	for _, dir := range directory.this.requires {
		args := []string{}
		args = append(args, dir)
		succeeded, reason, solution := directoriesNeedsDictionary["dir"].check(args)
		checkItem := CheckItem{
			name:      "dir:" + dir,
			succeeded: succeeded,
			reason:    reason,
			solution:  solution,
		}
		checkItems = append(checkItems, checkItem)
	}

	for _, instruction := range directory.this.instructions {
		for _, need := range instruction.needs {
			args := []string{}
			succeeded, reason, solution := directoriesNeedsDictionary[need].check(args)
			checkItem := CheckItem{
				name:      need + ":" + directory.name + "/" + instruction.path,
				succeeded: succeeded,
				reason:    reason,
				solution:  solution,
			}
			checkItems = append(checkItems, checkItem)
		}
	}

	printCheckItems(checkItems)
}

func (directory *DirectoryKnowledge) Setup(args []string) (bool, error) {
	isInitial := false
	if args[0] == "initial" {
		isInitial = true
	}

	directoriesToSetup := []DirectoryKnowledge{}
	for _, layersDirectoryName := range directory.this.requires {
		layersDirectoryKnowledge := DirectoryKnowledge{
			name: layersDirectoryName,
			this: layersDirectoriesDictionary[layersDirectoryName],
		}
		directoriesToSetup = append(directoriesToSetup, layersDirectoryKnowledge)
	}

	directoriesToSetup = append(directoriesToSetup, *directory)

	ecosystemNeeds := []string{}

	for _, directoryToSetup := range directoriesToSetup {
		for _, instructions := range directoryToSetup.this.instructions {
			for _, needs := range instructions.needs {
				alreadyExists := false
				for _, e := range ecosystemNeeds {
					if e == needs {
						alreadyExists = true
						break
					}
				}
				if !alreadyExists {
					ecosystemNeeds = append(ecosystemNeeds, needs)
				}
			}
		}
	}

	fmt.Println(ecosystemNeeds)

	for _, needs := range ecosystemNeeds {
		splitted := strings.Split(needs, ":")
		args := []string{}
		for index, value := range splitted {
			if index > 0 {
				args = append(args, value)
			}
		}
		command := splitted[0]
		succeeded, _, _ := directoriesNeedsDictionary[command].check(args)

		if !succeeded {
			directoriesNeedsDictionary[command].run(args)
		}
	}

	if isInitial {
		installNpmBasics()
		for _, dir := range directoriesToSetup {
			colorGreen := "\033[32m"
			colorReset := "\033[0m"
			colorCyan := "\033[36m"
			fmt.Println(string(colorGreen), fmt.Sprintf("Accessing `%s` \n", dir.name), string(colorReset))
			for _, instruction := range dir.this.instructions {
				pathVar := changeNodeVersion(strings.Split(instruction.interpreter, ":")[1])
				for _, command := range instruction.commands.setup {
					fmt.Println(string(colorCyan), fmt.Sprintf("Running `%s` at `%s/%s` \n", command, dir.name, instruction.path), string(colorReset))
					runSetupCommand(command, instruction, dir.name, pathVar)
				}
			}
		}
	}

	ecosystem, err := generateEcosystemConfig(directoriesToSetup)
	if err != nil {
		log.Fatal(err.Error())
	}

	writeFile(ecosystem, directory.name)

	return true, nil
}

func changeNodeVersion(major string) string {
	newVersionPath, err := getNodeInterpreter("nvm", major)
	if err != nil {
		log.Fatalln(err)
	}
	out, err := exec.Command("node", "--version").Output()
	if err != nil {
		log.Fatalln(err)
	}
	old := strings.TrimPrefix(strings.Split(string(out), ".")[0], "v")
	oldVersionPath, err := getNodeInterpreter("nvm", old)
	if err != nil {
		log.Fatalln(err)
	}

	pathVar := os.Getenv("PATH")

	pathVar = strings.ReplaceAll(pathVar, strings.TrimSuffix(oldVersionPath, "/node"), strings.TrimSuffix(newVersionPath, "/node"))
	fmt.Println(pathVar)

	return pathVar
}

func installNpmBasics() {
	fmt.Println("start installation")

	nodeVersions := []string{"8", "12", "14"}
	scripts := []string{"yarn", "pm2"}

	fmt.Println(scripts)
	fmt.Println(nodeVersions)

	for _, nodeVersion := range nodeVersions {
		changeNodeVersion(nodeVersion)
		for _, script := range scripts {
			command := exec.Command("npm", "install", "-g", script)

			command.Stderr = os.Stderr
			command.Stdin = os.Stdin
			command.Stdout = os.Stdout
			command.Run()
		}
	}
}

func runSetupCommand(command string, instruction RunInstruction, dirName string, pathVar string) {
	splitted := strings.Split(command, " ")
	args := []string{}
	for index, arg := range splitted {
		if index > 0 {
			args = append(args, arg)
		}
	}
	script := splitted[0]

	path := os.Getenv("LAYERS_PATH") + "/" + dirName

	if instruction.path != "root" {
		path = path + "/" + instruction.path
	}

	runScript(script, args, path, instruction.interpreter, pathVar)
}

func getScriptPath(script string, interpreter string) (path string) {
	fmt.Println(interpreter)
	major := strings.Split(interpreter, ":")[1]
	fmt.Println(major)

	pathToNode, err := getNodeInterpreter("nvm", major)
	if err != nil {
		log.Fatalf("Couldn't get node path for v%s", major)
	}

	fmt.Println(pathToNode)

	// getNodeInterpreter -> "/Users/user/.nvm/versions/node/v12.22.8/bin/node"
	// change only last file to get script path
	// "/Users/user/.nvm/versions/node/v12.22.8/bin/node" -> "/Users/user/.nvm/versions/node/v12.22.8/bin/script"
	splittedPath := strings.Split(pathToNode, "/")
	splittedPath[len(splittedPath)-1] = script

	pathToScript := strings.Join(splittedPath, "/")

	return pathToScript
}

func runScript(script string, args []string, path string, interpreter string, pathVar string) {
	scriptPath := getScriptPath(script, interpreter)

	fmt.Println(scriptPath, args)

	command := exec.Command(scriptPath, args...)

	command.Env = append(command.Env, "PATH="+pathVar)

	command.Dir = path
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout

	command.Run()
}

func installPm2() {
	command := exec.Command("npm", "install", "-g", "pm2@latest")

	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout

	command.Run()
}

func IsPm2Installed() bool {
	output, err := exec.Command("pm2", "--version").Output()
	if err != nil {
		return false
	}

	if len(string(output)) == 0 {
		return false
	}
	return true
}

func generateEcosystemConfig(directories []DirectoryKnowledge) (ecosystem string, err error) {
	pm2DirectoriesAsString := []string{}

	for _, dir := range directories {
		for _, subDir := range dir.this.instructions {
			pm2DirAsString := getPm2DirAsString(subDir, dir.name)
			pm2DirectoriesAsString = append(pm2DirectoriesAsString, pm2DirAsString)
		}
	}

	pm2EcosystemAsString := strings.Join(pm2DirectoriesAsString, "\n")

	return pm2EcosystemAsString, nil
}

func writeFile(ecosystem string, dirName string) {
	layersPath := os.Getenv("LAYERS_PATH")

	path := layersPath + "/ecosystems"
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}

	f, err := os.Create(fmt.Sprintf("%s/ecosystems/%s.config.js", layersPath, dirName))
	if err != nil {
		log.Fatalln("Couldn't write file")
	}

	defer f.Close()
	w := bufio.NewWriter(f)
	n4, err := w.WriteString(`
	module.exports = {
		apps: [` + ecosystem + `
		],
	}
	`)
	if err != nil {
		log.Fatalln("Couldn't write file")
	}
	fmt.Printf("wrote %d bytes\n", n4)

	w.Flush()
}

func getPm2DirAsString(dir RunInstruction, dirName string) string {
	nodeMajor := strings.Split(dir.interpreter, ":")[1]

	interpreter, err := getNodeInterpreter("nvm", nodeMajor)
	if err != nil {
		log.Fatalln(fmt.Sprintf("Couldn't get interpreter for `%s`. Error: %s\n", dirName, err.Error()))
	}

	relativePath := fmt.Sprintf("/%s", dirName)
	name := dirName
	if dir.path != "root" {
		name = fmt.Sprintf("%s/%s", dirName, dir.path)
		relativePath = fmt.Sprintf("/%s/%s", dirName, dir.path)
	}

	splittedCommand := strings.Split(dir.commands.starter, " ")
	script := splittedCommand[0]
	args := strings.TrimPrefix(dir.commands.starter, script)

	return fmt.Sprintf(`
	{
		name: '%s',
		cwd: '../%s',
		script: '%s',
		args: '%s',
		interpreter: '%s',
		watch: true,
	},
	`, name, relativePath, script, args, interpreter)
}

func getNodeInterpreter(mode string, requiredMajor string) (interpreterPath string, err error) { // TODO: accept asdf in mode
	nvmDir := os.Getenv("NVM_DIR")

	nvmNodeDirs := nvmDir + "/versions/node"

	nodeVersions, err := ioutil.ReadDir(nvmNodeDirs)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Couldn't get node versions at path: %s", nvmNodeDirs))
	}

	for _, dir := range nodeVersions {
		if dir.IsDir() {
			majorVersion := strings.Split(strings.TrimPrefix(dir.Name(), "v"), ".")[0]
			if majorVersion == requiredMajor {
				return fmt.Sprintf("%s/%s/bin/node", nvmNodeDirs, dir.Name()), nil
			}
		}
		fmt.Println(dir.Name(), dir.IsDir())
	}
	return "", errors.New(fmt.Sprintf("Couldn't find %s version at `%s`\n", requiredMajor, nvmNodeDirs))
}

func IsLayersDirectory(name string) bool {
	layersDirectoriesNames := GetLayersDirectoriesNames()

	for _, dirName := range layersDirectoriesNames {
		if dirName == name {
			return true
		}
	}
	return false
}

func getDirectoryKnowledgeByName(name string) (LayersDirectory, error) {
	for key, layersDirectory := range layersDirectoriesDictionary {
		if key == name {
			return layersDirectory, nil
		}
	}
	return LayersDirectory{}, errors.New(name + " isn't a known directory.")
}

func GetLayersDirectoriesNames() []string {
	directories := []string{}

	for key := range layersDirectoriesDictionary {
		directories = append(directories, key)
	}

	return directories
}

type DirectoryNeeds struct {
	check func(args []string) (succeeded bool, reason string, solution string)
	run   func(args []string) error
}

var directoriesNeedsDictionary = map[string]DirectoryNeeds{
	"mongodb": {
		check: func(args []string) (succeeded bool, reason string, solution string) {
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
		run: func(args []string) error {
			pathToBackendCore := os.Getenv("LAYERS_PATH") + "/tendaedu-backend"

			cmd := exec.Command("sudo", "docker-compose", "up", "-d")
			cmd.Dir = pathToBackendCore
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			cmd.Run()

			return nil
		},
	},
	"redis": {
		check: func(args []string) (succeeded bool, reason string, solution string) {
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
		run: func(args []string) error {
			pathToBackendCore := os.Getenv("LAYERS_PATH") + "/tendaedu-backend"

			cmd := exec.Command("sudo", "docker-compose", "up", "-d")
			cmd.Dir = pathToBackendCore
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			cmd.Run()

			return nil
		},
	},
	"file": {
		check: func(args []string) (succeeded bool, reason string, solution string) {
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
		run: func(args []string) error {
			// can't check this
			return nil
		},
	},
	"dir": {
		check: func(args []string) (succeeded bool, reason string, solution string) {
			dir := args[0]
			ports := map[string]string{
				"tendaedu-backend":    "8009",
				"layers-auth-vanilla": "8090",
				"layers-webapp":       "8080",
			}

			succeeded, reason, solution = checkProcessAtPort(ports[dir], "node", dir)
			return succeeded, reason, solution
		},
		run: func(args []string) error {
			// can't check this
			return nil
		},
	},
}

var layersDirectoriesDictionary = map[string]LayersDirectory{
	"tendaedu-backend": {
		requires: []string{},
		instructions: []RunInstruction{
			{
				path:        "root",
				interpreter: "node:14",
				needs:       []string{"mongodb", "redis"},
				commands: RunInstructionCommand{
					starter: "yarn start",
					setup:   []string{"yarn install", "yarn populate both"},
				},
			},
		},
	},
	"layers-auth-vanilla": {
		requires: []string{"tendaedu-backend"},
		instructions: []RunInstruction{
			{
				path:        "root",
				interpreter: "node:12",
				needs:       []string{},
				commands: RunInstructionCommand{
					starter: "yarn start",
					setup:   []string{"yarn install"},
				},
			},
		},
	},
	"tendaedu-web": {
		requires: []string{
			"tendaedu-backend",
			"layers-webapp",
			"layers-auth-vanilla",
		},
		instructions: []RunInstruction{
			{
				path:        "web",
				interpreter: "node:8",
				needs:       []string{},
				commands: RunInstructionCommand{
					starter: "yarn start",
					setup:   []string{"yarn install"},
				},
			},
		},
	},
	"layers-webapp": {
		requires: []string{
			"tendaedu-backend",
			"layers-auth-vanilla",
		},
		instructions: []RunInstruction{
			{
				path:        "root",
				interpreter: "node:12",
				needs:       []string{},
				commands: RunInstructionCommand{
					starter: "yarn start",
					setup:   []string{"yarn install"},
				},
			},
		},
	},
	"layers-comunicados": {
		requires: []string{
			"tendaedu-backend",
			"layers-auth-vanilla",
			"layers-webapp",
		},
		instructions: []RunInstruction{
			{
				path:        "web",
				interpreter: "node:12",
				needs:       []string{},
				commands: RunInstructionCommand{
					starter: "yarn start",
					setup:   []string{"yarn install"},
				},
			},
			{
				path:        "app",
				interpreter: "node:12",
				needs:       []string{"mongodb"},
				commands: RunInstructionCommand{
					starter: "yarn start",
					setup:   []string{"yarn install", "yarn populate both"},
				},
			},
		},
	},
	"layers-agenda": {
		requires: []string{
			"tendaedu-backend",
			"layers-auth-vanilla",
			"layers-webapp",
		},
		instructions: []RunInstruction{
			{
				path:        "web",
				interpreter: "node:12",
				needs:       []string{},
				commands: RunInstructionCommand{
					starter: "yarn start",
					setup:   []string{"yarn install"},
				},
			},
			{
				path:        "app",
				interpreter: "node:12",
				needs:       []string{"mongodb"},
				commands: RunInstructionCommand{
					starter: "yarn start",
					setup:   []string{"yarn install"},
				},
			},
		},
	},
}

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

var staticErrors = map[string]map[string]string{
	"node_installation": {
		"reason":   "Couldn't get node version, maybe node isn't installed yet",
		"solution": "Install node using nvm(https://github.com/nvm-sh/nvm) or asdf(https://github.com/asdf-vm/asdf-nodejs)",
	},
	"docker_init": {
		"solution": "Be sure that you ran `sudo docker-compose up -d` at `tendaedu-backend` or `payments`",
	},
}
