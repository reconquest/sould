package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ExecutionError struct {
	Err     error
	Output  []byte
	Command string
}

type Mirror struct {
	Name string
	Dir  string
}

var (
	ErrMirrorNotFound      = errors.New("mirror not found")
	ErrMirrorAlreadyExists = errors.New("mirror already exists")
)

func CreateMirror(
	storageDir string, name string, cloneURL string,
) (Mirror, error) {
	mirrorDir := filepath.Join(storageDir, name)

	_, err := os.Stat(mirrorDir)
	if err == nil {
		return Mirror{}, ErrMirrorAlreadyExists
	} else if !os.IsNotExist(err) {
		return Mirror{}, err
	}

	err = os.MkdirAll(mirrorDir, 0770)
	if err != nil {
		return Mirror{}, err
	}

	mirror := Mirror{
		Name: name,
		Dir:  mirrorDir,
	}

	err = mirror.Clone(cloneURL)
	if err != nil {
		return Mirror{}, err
	}

	return mirror, err
}

func GetMirror(
	storageDir string, name string, cloneURL string,
) (Mirror, error) {
	mirrorDir := filepath.Join(storageDir, name)

	_, err := os.Stat(mirrorDir)
	if err != nil {
		if os.IsNotExist(err) {
			return Mirror{}, ErrMirrorNotFound
		}

		return Mirror{}, err
	}

	mirror := Mirror{
		Name: name,
		Dir:  mirrorDir,
	}

	return mirror, nil
}
func (mirror Mirror) GetTarArchive() ([]byte, error) {
	return mirror.execute(exec.Command("git", "archive", "HEAD"))
}

func (mirror Mirror) Pull() error {
	_, err := mirror.execute(exec.Command("git", "remote", "update"))
	return err
}

func (mirror Mirror) execute(command *exec.Cmd) ([]byte, error) {
	command.Dir = mirror.Dir

	output, err := command.CombinedOutput()
	if err != nil {
		execErr := ExecutionError{
			Err:     err,
			Output:  output,
			Command: strings.Join(command.Args, " "),
		}

		return output, execErr
	}

	return output, nil
}

func (mirror Mirror) Clone(url string) error {
	_, err := mirror.execute(
		exec.Command("git", "clone", "--recursive", "--mirror", url, "."),
	)

	return err
}

func (execErr ExecutionError) Error() string {
	return fmt.Sprintf(
		"%s: %s\n%s",
		execErr.Command, execErr.Err.Error(), execErr.Output,
	)
}
