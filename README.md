## Layers Command Line Interface

Parallel project to simplify Layers's repos first usage

Now using golang

### About Go project
- To run in your own machine you must have [golang](https://go.dev/doc/install) installed
- This project is using [Cobra](https://github.com/spf13/cobra) to build the CLI
- With golang installed, use `go run main.go` to run the cli
- If you want to build, use this command `go build` in the root. It will generate `layers_cli` file, which could be initialized via terminal
- If you want to create a new command, you must install [CobraGenerator](https://github.com/spf13/cobra/blob/master/cobra/README.md). With CobraGenerator installed, use the command `cobra add <commandName>` in the root.

### How to install?
- run `bash ./install.sh --build` in the root to build and install
- OBS: At this moment you need to install golang to build

### Tasks
- [x] *layers doctor*: should check if the directory is well configurated;
  - [x] *node version step*: verify node version;
  - [x] *mongo step*: verify if mongoDB is running;
  - [-] *redis step*: verify if Redis is running;
- [ ] *configuration file (config.layers)*: a file to specify the project's configuration;
  - [ ] *SPECS*: specifies the technologies used in the project (<techName>:<version>);
  - [ ] *REQUIRES*: pre conditions to run the project;
  - [ ] *STEPS*: commands to run step-by-step (*to be discussed*);
- [ ] *automatic installation*;
