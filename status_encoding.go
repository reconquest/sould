package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/seletskiy/hierr"
)

func (status ServerStatus) JSON() ([]byte, error) {
	var (
		buffer bytes.Buffer
		data   []byte
		err    error
	)

	if status.Role == "slave" {
		data, err = ffjson.Marshal(status.BasicServerStatus)
	} else {
		data, err = ffjson.Marshal(status)
	}

	err = json.Indent(&buffer, data, "", ServerStatusResponseIndentation)
	if err != nil {
		logger.Errorf("can't indent json %s: %s", data, err)
		return data, nil
	}

	return buffer.Bytes(), nil
}

func (status ServerStatus) TOML() ([]byte, error) {
	var buffer bytes.Buffer
	var err error

	encoder := toml.NewEncoder(&buffer)
	encoder.Indent = ServerStatusResponseIndentation
	if status.Role == "master" {
		err = encoder.Encode(status)
	} else {
		err = encoder.Encode(status.BasicServerStatus)
	}

	return buffer.Bytes(), err
}

func (status ServerStatus) Hierarchical() string {
	var hierarchy error
	if status.Address != "" {
		hierarchy = hierr.Push(status.Address)
	} else {
		hierarchy = hierr.Push(
			"status",
			fmt.Sprintf("role: %s", status.Role),
		)
	}

	if status.Error != nil {
		hierarchy = hierr.Push(
			"error",
			status.Error,
		)
	}

	hierarchy = hierr.Push(
		hierarchy,
		fmt.Sprintf("total: %d", len(status.Mirrors)),
	)

	if len(status.Mirrors) > 0 {
		mirrors := errors.New("mirrors")
		for _, mirror := range status.Mirrors {
			mirrors = hierr.Push(mirrors, mirror.Hierarchical())
		}
		hierarchy = hierr.Push(hierarchy, mirrors)
	}

	if status.Role == "master" {
		hierarchy = hierr.Push(hierarchy, status.Upstream.Hierarchical())
	}

	return hierr.String(hierarchy)
}

func (status MirrorStatus) Hierarchical() string {
	return hierr.String(
		hierr.Push(
			status.Name,
			fmt.Sprintf("state: %s", status.State),
			fmt.Sprintf("modify date: %v", status.ModifyDate),
		),
	)
}

func (status UpstreamStatus) Hierarchical() string {
	hierarchy := hierr.Push(
		"upstream",
		"total: "+strconv.Itoa(status.Total),
		fmt.Sprintf(
			"success: %d (%.2f%%)",
			status.Success, status.SuccessPercent,
		),
		fmt.Sprintf(
			"error: %d (%.2f%%)",
			status.Error, status.ErrorPercent,
		),
	)

	if len(status.Slaves) > 0 {
		slaves := errors.New("slaves")
		for _, slave := range status.Slaves {
			slaves = hierr.Push(slaves, slave.Hierarchical())
		}

		hierarchy = hierr.Push(hierarchy, slaves)
	}

	return hierr.String(hierarchy)
}
