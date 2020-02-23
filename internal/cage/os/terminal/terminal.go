// Copyright (C) 2019 The CodeActual Go Environment Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package terminal

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"syscall"

	ssh_terminal "golang.org/x/crypto/ssh/terminal"

	"github.com/pkg/errors"
)

// interaction prevents multiple goroutines from requesting user interaction at the same time.
// The mutex is package-level based on the assumption that there will be at most one interactive
// user per process.
var interaction *sync.Mutex

func init() {
	interaction = &sync.Mutex{}
}

func Lock() {
	interaction.Lock()
}

func Unlock() {
	interaction.Unlock()
}

type DefaultProvider struct{}

// Promptf displays a formatted message and collects a string response.
//
// A trailing space will be appended to the message if not already present.
// A package-level mutex prevents overlapping prompts.
// Promptf implements Provider.
func (d DefaultProvider) Promptf(format string, a ...interface{}) (string, error) {
	Lock()
	defer Unlock()

	fmt.Printf(format+" ", a...)

	r := bufio.NewReader(os.Stdin)

	buf, _, err := r.ReadLine()
	if err != nil {
		return "", errors.WithStack(err)
	}

	return string(buf), nil
}

// PromptHiddenf behaves the same as Promptf except that the response is not echoed.
//
// A trailing space will be appended to the message if not already present.
// A package-level mutex prevents overlapping prompts.
// PromptHiddenf implements Provider.
func (d DefaultProvider) PromptHiddenf(format string, a ...interface{}) (string, error) {
	Lock()
	defer Unlock()

	fmt.Printf(format+" \n", a...)

	buf, err := ssh_terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", errors.WithStack(err)
	}

	return string(buf), nil
}
