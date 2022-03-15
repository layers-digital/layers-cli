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
- [x] Commands: status of each command
  - [ ] Doctor
    - [x] check mongodb
    - [x] check redis
    - [x] check file
    - [x] check required directories
    - [ ] use knowledge package
  - [ ] Ecosystem
    - [x] Run
      - [x] Run ecosystem for current project
    - [x] Monitor
      - [x] Open pm2's terminal monitor
    - [ ] Setup
      - [x] get project's required directories
      - [x] get needs for the project ecosystem
      - [x] setup needs
      - [ ] run setup scripts for each directory
      - [ ] get interpreters
        - [x] node
          - [x] from nvm
          - [ ] from asdf
      - [x] generate .config.js file
    - [x] Stop
      - [ ] Stop current ecosystem
- [x] *configuration file (LAYERS.md)*: a file to specify the project's configuration;
- [ ] *automatic installation*;
  - [ ] publish binary
  - [x] install script
