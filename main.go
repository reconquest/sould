package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/docopt/docopt-go"
	"github.com/zazab/zhash"
)

const (
	usage = `Sould 1.0

Usage:
	sould [-c <config>] [--unsecure]

Options:
    -c <config>  Use specified file as config file.
                 [default: /etc/sould.conf]
	--unsecure   Allow create mirrors of local repositories.
`
)

func main() {
	args, err := docopt.Parse(usage, nil, true, "1.0", false)
	if err != nil {
		log.Fatal(err)
	}

	var (
		configPath = args["-c"].(string)
		unsecure   = args["--unsecure"].(bool)
	)

	config, err := getConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	if unsecure {
		log.Printf(
			"WARNING! Sould server running in unsecure mode. " +
				"In this mode sould can clone local repositories.",
		)
	}

	server, err := NewMirrorServer(config, MirrorStateTable{}, unsecure)
	if err != nil {
		log.Fatal(err)
	}

	//go waitHangupSignals(server, configPath)

	log.Printf("server.GetListenAddress(): %#v", server.GetListenAddress())
	err = server.ListenHTTP()
	if err != nil {
		log.Fatal(err)
	}
}

func waitHangupSignals(server *MirrorServer, configPath string) {
	hangup := make(chan os.Signal, 1)
	signal.Notify(hangup, syscall.SIGHUP)

	for _ = range hangup {
		isMasterBeforeReload := server.IsMaster()

		newConfig, err := getConfig(configPath)
		if err != nil {
			log.Println(err)
		}

		err = server.SetConfig(newConfig)
		if err != nil {
			log.Printf(
				"can't reload config: %s", err.Error(),
			)
		}

		log.Println("config successfully reloaded")

		if server.IsMaster() && !isMasterBeforeReload {
			log.Println("current sould server is now master")
		}
		if !server.IsMaster() && isMasterBeforeReload {
			log.Println("current sould server is now slave")
		}
	}
}

func getConfig(path string) (zhash.Hash, error) {
	var configData map[string]interface{}

	_, err := toml.DecodeFile(path, &configData)
	if err != nil {
		return zhash.Hash{}, fmt.Errorf(
			"could not load config: %s", err.Error(),
		)
	}

	return zhash.HashFromMap(configData), nil
}
