package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kovetskiy/executil"
)

// Mirror of remote git repository.
type Mirror struct {
	// Name is unique identifier.
	Name string
	// URL will be passed to git for cloning and fetching updates.
	URL string
	// Dir is relative path to storage dir where mirror will be stored
	Dir string
}

// String returns a string representation of mirror (name + url)
func (mirror *Mirror) String() string {
	return mirror.Name + " (" + mirror.URL + ")"
}

// CreateMirror creates new git mirror repository in specified storage, and
// clones remote data using given url
func CreateMirror(
	storageDir string, name string, url string,
) (Mirror, error) {
	mirrorDir := filepath.Join(storageDir, name)

	_, err := os.Stat(mirrorDir)
	if err == nil {
		return Mirror{}, fmt.Errorf(
			"directory %s already exists",
			mirrorDir,
		)
	} else if !os.IsNotExist(err) {
		return Mirror{}, err
	}

	err = os.MkdirAll(mirrorDir, 0770)
	if err != nil {
		return Mirror{}, err
	}

	mirror := Mirror{
		Name: name,
		URL:  url,
		Dir:  mirrorDir,
	}

	err = mirror.Clone(url)
	if err != nil {
		return Mirror{}, err
	}

	return mirror, err
}

// GetMirror looks for the mirror with specified name in specified mirror
// storage  directory.
func GetMirror(
	storageDir string, name string,
) (Mirror, error) {
	mirrorDir := filepath.Join(storageDir, name)

	_, err := os.Stat(mirrorDir)
	if err != nil {
		return Mirror{}, err
	}

	mirror := Mirror{
		Name: name,
		Dir:  mirrorDir,
	}

	url, err := mirror.GetURL()
	if err != nil {
		return mirror, err
	}

	mirror.URL = url

	return mirror, nil
}

// Archive of specified git treeish will be written to given stdout writer.
func (mirror *Mirror) Archive(
	stdout io.Writer, treeish string,
) error {
	cmd := exec.Command("git", "archive", treeish)
	cmd.Dir = mirror.Dir
	cmd.Stdout = stdout

	_, _, err := executil.Run(cmd, executil.IgnoreStdout)
	return err
}

// Fetch remote changeset using git remote update.
func (mirror *Mirror) Fetch() error {
	cmd := exec.Command("git", "remote", "update")
	cmd.Dir = mirror.Dir
	_, _, err := executil.Run(cmd)
	return err
}

// SpoofChangeset forcely sets branch label on specified "deattached" tag and
// removes that tag from mirror repository.
func (mirror *Mirror) SpoofChangeset(branch, tag string) error {
	if tag == "0000000000000000000000000000000000000000" {
		cmd := exec.Command("git", "branch", "--delete", branch)
		cmd.Dir = mirror.Dir
		_, _, err := executil.Run(cmd)
		if err != nil {
			return err
		}

		return nil
	}

	cmd := exec.Command("git", "branch", "--force", branch, tag)
	cmd.Dir = mirror.Dir
	_, _, err := executil.Run(cmd)
	if err != nil {
		return err
	}

	return mirror.removeTag(tag)
}

func (mirror *Mirror) removeTag(tag string) error {
	cmd := exec.Command("git", "tag", "-d", tag)
	cmd.Dir = mirror.Dir
	_, _, err := executil.Run(cmd)
	return err
}

// Clone remote repository to directory of given mirror.
func (mirror *Mirror) Clone(url string) error {
	cmd := exec.Command("git", "clone", "--recursive", "--mirror", url, ".")
	cmd.Dir = mirror.Dir
	_, _, err := executil.Run(cmd)
	return err
}

// GetURL from git configuration which mirror uses for fetching changesets.
func (mirror *Mirror) GetURL() (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = mirror.Dir
	stdout, _, err := executil.Run(cmd)
	return strings.TrimSpace(string(stdout)), err
}

// GetModifyDateUnix calls GetModifyDate and returns unix timestamp
// representation.
func (mirror *Mirror) GetModifyDateUnix() (int64, error) {
	modDate, err := mirror.GetModifyDate()
	if err != nil {
		return 0, err
	}

	return modDate.Unix(), nil
}

// GetModifyDate returns last modify date of given mirror repository.
func (mirror *Mirror) GetModifyDate() (time.Time, error) {
	dirs := []string{
		"refs/heads",
		"refs/tags",
	}

	var (
		modDate  time.Time
		err      error
		fileinfo os.FileInfo
	)

	for _, dir := range dirs {
		fileinfo, err = os.Stat(mirror.Dir + "/" + dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			break
		}

		newModDate := fileinfo.ModTime()
		if newModDate.Unix() > modDate.Unix() {
			modDate = newModDate
		}
	}

	return modDate, err
}
