// Copyright (C) 2019 The aws-exec-cmd Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Command aws-exec-cmd acquires AWS credentials and runs an arbitrary command, providing it credentials through environment variables. It acquires credentials from the environment, IAM roles (with AssumeRole chaining), or Cognito identity pools.
//
// Environment variables:
//
//   AWS_ACCESS_KEY_ID
//   AWS_SECRET_ACCESS_KEY
//   AWS_SESSION_TOKEN
//
// Usage:
//
//   aws-exec-cmd --help
//   aws-exec-cmd role --help
//   aws-exec-cmd idp --help
//
// Use the IAM role, attached to an EC2 instance, to run "env | grep AWS_":
//
//   aws-exec-cmd role --chain instance -- env | grep AWS_
//
// Perform the same command but with credentials from role "backup" assumed from an EC2 instance role:
//
//   aws-exec-cmd role --chain instance,arn:aws:iam::123456789012:role/backup -- env | grep AWS_
//
// Perform the same command but with credentials from role "backup" assumed from enviroment credentials:
//
//   aws-exec-cmd role --chain env-triple,arn:aws:iam::123456789012:role/backup -- env | grep AWS_
//
// Perform the same command with credentials from Cognito identity pool, using federated Google auth:
//
//   aws-exec-cmd idp \
//     --name accounts.google.com \
//     --pool-id <pool ID> \
//     --refresh <Google OAuth refresh token> \
//     --client-id <Google OAuth client ID> \
//     --client-secret <Google OAuth client secret>
//
// Supported AssumeRole chaining:
//
//   environment variable credentials -> AssumeRole [-> AssumeRole ...]
//   role (temporary credentials from STS) -> AssumeRole [-> AssumeRole ...]
package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/codeactual/aws-exec-cmd/idp"
	"github.com/codeactual/aws-exec-cmd/internal/ldflags"
	"github.com/codeactual/aws-exec-cmd/role"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "aws-exec-cmd",
		Short: "Run a local command with AWS credentials set in the environment",
	}

	rootCmd.Version = ldflags.Version
	rootCmd.AddCommand(role.New())
	rootCmd.AddCommand(idp.New())

	if err := rootCmd.Execute(); err != nil {
		panic(errors.WithStack(err))
	}
}
