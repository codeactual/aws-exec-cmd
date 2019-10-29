// Copyright (C) 2019 The aws-exec-cmd Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package mixin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/spf13/cobra"

	"github.com/codeactual/aws-exec-cmd/internal/cage/aws"
	"github.com/codeactual/aws-exec-cmd/internal/cage/cli/handler"
	cage_reflect "github.com/codeactual/aws-exec-cmd/internal/cage/reflect"
)

const (
	defaultTimeout = 0
)

type Exec struct {
	handler.IO

	Pty     bool `usage:"Run in a pseudo-terminal"`
	Timeout int  `usage:"Number of seconds to wait for a non-[--pty] command to finish"`
}

// Implements cage/cli/handler.Mixin
func (m *Exec) BindCobraFlags(cmd *cobra.Command) []string {
	cmd.Flags().IntVarP(&m.Timeout, "timeout", "", defaultTimeout, cage_reflect.GetFieldTag(*m, "Timeout", "usage"))
	cmd.Flags().BoolVarP(&m.Pty, "pty", "", false, cage_reflect.GetFieldTag(*m, "Pty", "usage"))
	return []string{}
}

// Implements cage/cli/handler.Mixin
func (m *Exec) Name() string {
	return "cage/cmd/aws-exec-cmd/mixin"
}

// Implements cage/cli/handler.Mixin
func (m *Exec) PreRun(ctx context.Context) error {
	return nil
}

func (m *Exec) Do(ctx context.Context, creds *credentials.Credentials, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(m.Err(), "command not specified")
		os.Exit(1)
	}

	var cmd *exec.Cmd

	if m.Pty {
		if len(args) == 1 {
			cmd = exec.Command(args[0]) // #nosec
		} else {
			cmd = exec.Command(args[0], args[1:]...) // #nosec
		}
	} else {
		var cancel func()

		if m.Timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, time.Duration(m.Timeout)*time.Second)
			defer cancel()
		}

		cmd = exec.CommandContext(ctx, args[0], args[1:]...) // #nosec
	}

	if res, execErr := aws.ExecAs(ctx, creds, m.Out(), m.Err(), m.In(), cmd, m.Pty); execErr != nil {
		fmt.Fprintln(m.Err(), execErr)
		os.Exit(res.Cmd[cmd].Code)
	}
}

var _ handler.Mixin = (*Exec)(nil)
