package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/docopt/docopt-go"
	"github.com/zazab/zhash"
)

const usage = `Sould 1.0

Usage:
	sould [-c <config>] [--unsecure]

Options:
    -c <config>  Use specified file as config file.
                 [default: /etc/sould.conf]
	--unsecure   Allow create mirrors of local repositories.
`

func main() {
	args, err := docopt.Parse(usage, nil, true, "1.0", false)
	if err != nil {
		log.Fatal(err)
	}

	var (
		configPath   = args["-c"].(string)
		unsecureMode = args["--unsecure"].(bool)
	)

	config, err := getConfig(configPath)
	if err != nil {
		log.Fatalf("can't load config: %s", err.Error())
	}

	if unsecureMode {
		log.Printf(
			"WARNING! Sould server running in unsecure mode. " +
				"In this mode sould can create mirror to local repositories.",
		)
	}

	server, err := NewMirrorServer(config, MirrorStateTable{}, unsecureMode)
	if err != nil {
		log.Fatal(err)
	}

	go waitHangupSignals(server, configPath)

	err = server.ListenHTTP()
	if err != nil {
		log.Fatal(err)
	}
}

func getConfig(path string) (zhash.Hash, error) {
	var configData map[string]interface{}

	_, err := toml.DecodeFile(path, &configData)
	if err != nil {
		return zhash.Hash{}, err
	}

	return zhash.HashFromMap(configData), nil
}

func reloadConfig(
	server *MirrorServer, configPath string,
) (becomeMaster bool, becomeSlave bool, err error) {
	isMasterBeforeReload := server.IsMaster()

	newConfig, err := getConfig(configPath)
	if err != nil {
		return false, false, err
	}

	err = server.SetConfig(newConfig)
	if err != nil {
		return false, false, err
	}

	switch {
	case server.IsMaster() && !isMasterBeforeReload:
		return true, false, nil

	case !server.IsMaster() && isMasterBeforeReload:
		return false, true, nil
	}

	return false, false, nil
}

// waits for SIGHUP  and try to reload config
func waitHangupSignals(server *MirrorServer, configPath string) {
	hangup := make(chan os.Signal, 1)
	signal.Notify(hangup, syscall.SIGHUP)

	for _ = range hangup {
		becomeMaster, becomeSlave, err := reloadConfig(server, configPath)
		switch {
		case err != nil:
			log.Printf("can't reload config: %s", err.Error())

		case becomeMaster:
			log.Println("current sould server is now master")

		case becomeSlave:
			log.Println("current sould server is now slave")

		default:
			log.Println("config successfully reloaded")
		}
	}
}
