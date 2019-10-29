// Copyright (C) 2019 The CodeActual Go Environment Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package aws

import (
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/pkg/errors"

	cage_exec "github.com/codeactual/aws-exec-cmd/internal/cage/os/exec"
)

func GetenvRegion() string {
	r := os.Getenv("AWS_REGION")
	if r != "" {
		return r
	}
	return os.Getenv("AWS_DEFAULT_REGION")
}

// ExecAs executes a local command with AWS credentials defined in the environment.
//
// If a pseudo-terminal is used, a zero value PipelineResult will be returned.
func ExecAs(ctx context.Context, creds *credentials.Credentials, out io.Writer, err io.Writer, in io.Reader, cmd *exec.Cmd, pty bool) (cage_exec.PipelineResult, error) {
	credsVal, credsErr := creds.Get()
	if credsErr != nil {
		return cage_exec.PipelineResult{}, errors.WithStack(credsErr)
	}

	// Based on https://docs.aws.amazon.com/cli/latest/userguide/cli-environment.html
	cmd.Env = append(
		os.Environ(),
		"AWS_ACCESS_KEY_ID="+credsVal.AccessKeyID,
		"AWS_SECRET_ACCESS_KEY="+credsVal.SecretAccessKey,
		"AWS_SESSION_TOKEN="+credsVal.SessionToken,
	)

	if pty {
		return cage_exec.PipelineResult{}, cage_exec.CommonExecutor{}.Pty(cmd)
	}

	return cage_exec.CommonExecutor{}.Standard(ctx, out, err, in, cmd)
}
