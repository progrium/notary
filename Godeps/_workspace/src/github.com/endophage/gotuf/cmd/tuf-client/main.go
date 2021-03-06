package main

import (
	"fmt"
	"log"

	"github.com/endophage/gotuf/client"
	"github.com/endophage/gotuf/store"
	"github.com/flynn/go-docopt"
)

func main() {
	log.SetFlags(0)

	usage := `usage: tuf-client [-h|--help] <command> [<args>...]

Options:
  -h, --help

Commands:
  help         Show usage for a specific command
  init         Initialize with root keys
  list         List available target files
  get          Get a target file

See "tuf-client help <command>" for more information on a specific command.
`

	args, _ := docopt.Parse(usage, nil, true, "", true)
	cmd := args.String["<command>"]
	cmdArgs := args.All["<args>"].([]string)

	if cmd == "help" {
		if len(cmdArgs) == 0 { // `tuf-client help`
			fmt.Println(usage)
			return
		} else { // `tuf-client help <command>`
			cmd = cmdArgs[0]
			cmdArgs = []string{"--help"}
		}
	}

	if err := runCommand(cmd, cmdArgs); err != nil {
		log.Fatalln("ERROR:", err)
	}
}

type cmdFunc func(*docopt.Args, *client.Client) error

type command struct {
	usage string
	f     cmdFunc
}

var commands = make(map[string]*command)

func register(name string, f cmdFunc, usage string) {
	commands[name] = &command{usage: usage, f: f}
}

func runCommand(name string, args []string) error {
	argv := make([]string, 1, 1+len(args))
	argv[0] = name
	argv = append(argv, args...)

	cmd, ok := commands[name]
	if !ok {
		return fmt.Errorf("%s is not a tuf-client command. See 'tuf-client help'", name)
	}

	parsedArgs, err := docopt.Parse(cmd.usage, argv, true, "", true)
	if err != nil {
		return err
	}

	client, err := tufClient(parsedArgs)
	if err != nil {
		return err
	}
	return cmd.f(parsedArgs, client)
}

func tufClient(args *docopt.Args) (*client.Client, error) {
	storePath, ok := args.String["--store"]
	if !ok {
		storePath = args.String["-s"]
	}
	local := store.FileSystemStore(storePath, nil)
	remote, err := client.HTTPRemoteStore(args.String["<url>"], nil)
	if err != nil {
		return nil, err
	}
	return client.NewClient(local, remote), nil
}
