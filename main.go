package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/docopt/docopt-go"
	"github.com/seletskiy/hierr"
	"github.com/zazab/zhash"
)

var (
	version = `3.1`

	usage = `SOULD ` + version + `

The scalable failover service for mirroring git repositories.
https://github.com/reconquest/sould

Usage:
	sould [-c <config>] [--insecure]

Options:
    -c <config>  Use specified file as config file.
                 [default: /etc/sould.conf]
    --insecure   Allow create mirrors of local repositories.
				 In this mode sould will be able to give access to ANY local
				 repository readable by sould process.
				 It's inteded for tests only, so use with care.

`
)

var (
	logger = NewLogger()
)

func main() {
	args, err := docopt.Parse(usage, nil, true, version, false)
	if err != nil {
		logger.Fatal(err)
	}

	var (
		configPath   = args["-c"].(string)
		insecureMode = args["--insecure"].(bool)
	)

	config, err := getConfig(configPath)
	if err != nil {
		logger.Fatalf("can't load config: %s", err.Error())
	}

	if insecureMode {
		logger.Warning("Server running in insecure mode.")
		logger.Warning(
			"In this mode sould will be able to give access to ANY local " +
				"repository readable by sould process.",
		)
		logger.Warning("It's inteded for tests only, so use with care.")
	}

	mirrorStates := NewMirrorStates()

	server, err := NewMirrorServer(config, mirrorStates, insecureMode)
	if err != nil {
		logger.Fatal(err)
	}

	proxy, err := NewGitProxy(config, mirrorStates)
	if err != nil {
		logger.Fatal(err)
	}

	go serveHangupSignals(server, proxy, configPath)

	err = proxy.Start()
	if err != nil {
		logger.Fatal(hierr.Errorf(err, "failed to start git daemon proxy"))
	}

	err = server.ListenHTTP()
	if err != nil {
		logger.Fatal(hierr.Errorf(err, "failed to start http server"))
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
	defer func() {
		err := recover()
		if err != nil {
			err = fmt.Errorf("PANIC: %s\n%s", err, stack())
		}
	}()

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

func serveHangupSignals(
	server *MirrorServer, proxy *GitProxy, configPath string,
) {
	hangup := make(chan os.Signal, 1)
	signal.Notify(hangup, syscall.SIGHUP)

	for range hangup {
		becameMaster, becameSlave, err := reloadConfig(
			server, proxy, configPath,
		)
		switch {
		case err != nil:
			logger.Errorf("can't reload config: %s", err.Error())

		case becameMaster:
			logger.Info("current sould server is now master")

		case becameSlave:
			logger.Info("current sould server is now slave")

		default:
			logger.Info("config successfully reloaded")
		}
	}
}
