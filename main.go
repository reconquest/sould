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

const usage = `Sould 1.0

Usage:
	sould [-c <config>] [--insecure]

Options:
    -c <config>  Use specified file as config file.
                 [default: /etc/sould.conf]
    --insecure   Allow create mirrors of local repositories.
`

func main() {
	args, err := docopt.Parse(usage, nil, true, "1.0", false)
	if err != nil {
		log.Fatal(err)
	}

	var (
		configPath   = args["-c"].(string)
		insecureMode = args["--insecure"].(bool)
	)

	config, err := getConfig(configPath)
	if err != nil {
		log.Fatalf("can't load config: %s", err.Error())
	}

	if insecureMode {
		log.Printf(
			"WARNING! Sould server running in insecure mode. " +
				"In this mode sould will be able to give access to ANY local " +
				"repository readable by sould process. " +
				"It's inteded for tests only, so use with care.",
		)
	}

	mirrorStateTable := NewMirrorStateTable()

	server, err := NewMirrorServer(config, mirrorStateTable, insecureMode)
	if err != nil {
		log.Fatal(err)
	}

	proxy, err := NewGitProxy(config, mirrorStateTable)
	if err != nil {
		log.Fatal(err)
	}

	go serveHangupSignals(server, proxy, configPath)

	proxy.Start()

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
	server *MirrorServer, proxy *GitProxy, configPath string,
) (becameMaster bool, becameSlave bool, err error) {
	wasMaster := server.IsMaster()

	newConfig, err := getConfig(configPath)
	if err != nil {
		return false, false, err
	}

	err = server.SetConfig(newConfig)
	if err != nil {
		return false, false, fmt.Errorf(
			"can't set config for http server: %s", err,
		)
	}

	if server.IsMaster() == wasMaster {
		return false, false, nil
	}

	err = proxy.SetConfig(newConfig)
	if err != nil {
		return false, false, fmt.Errorf(
			"can't set config for git daemon proxy: %s", err,
		)
	}

	return server.IsMaster(), wasMaster, nil
}

// waits for SIGHUP and try to reload config
func serveHangupSignals(
	server *MirrorServer, proxy *GitProxy, configPath string,
) {
	hangup := make(chan os.Signal, 1)
	signal.Notify(hangup, syscall.SIGHUP)

	for _ = range hangup {
		becameMaster, becameSlave, err := reloadConfig(
			server, proxy, configPath,
		)
		switch {
		case err != nil:
			log.Printf("can't reload config: %s", err.Error())

		case becameMaster:
			log.Println("current sould server is now master")

		case becameSlave:
			log.Println("current sould server is now slave")

		default:
			log.Println("config successfully reloaded")
		}
	}
}
