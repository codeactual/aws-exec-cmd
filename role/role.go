// Copyright (C) 2019 The aws-exec-cmd Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package role

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/codeactual/aws-exec-cmd/internal/cage/cli/handler"
	handler_cobra "github.com/codeactual/aws-exec-cmd/internal/cage/cli/handler/cobra"
	"github.com/codeactual/aws-exec-cmd/internal/cage/cli/handler/mixin/aws/auth"
	auth_role "github.com/codeactual/aws-exec-cmd/internal/cage/cli/handler/mixin/aws/auth/role"
	cmd_mixin "github.com/codeactual/aws-exec-cmd/mixin"
)

// Handler defines the sub-command flags and logic.
type Handler struct {
	handler.Session

	// Auth acquires credentials from a cage/cli/handler/mixin/aws/auth.Provider implementation.
	Auth auth.Mixin

	// Exec runs a local command AWS credentials set in environment variables.
	Exec cmd_mixin.Exec

	// RoleChain defines/collects the CLI flags and provides the implementation for Handler.Auth.
	RoleChain auth_role.Mixin
}

// Init defines the command, its environment variable prefix, etc.
//
// It implements cli/handler/cobra.Handler.
func (h *Handler) Init() handler_cobra.Init {
	h.Auth.RoleChainFlag = "chain" // use "role --chain" to avoid "role --role" invocation.

	h.Exec = cmd_mixin.New()

	return handler_cobra.Init{
		Cmd: &cobra.Command{
			Use:   "role",
			Short: "Use a role chain to provide the credentials",
		},
		EnvPrefix: "AWS_EXEC_CMD",
		Mixins: []handler.Mixin{
			&h.Auth,
			&h.Exec,
			&h.RoleChain,
		},
	}
}

// BindFlags binds the flags to Handler fields.
//
// It implements cli/handler/cobra.Handler.
func (h *Handler) BindFlags(cmd *cobra.Command) []string {
	return []string{} // auth_role.Mixin provides all of them
}

// Run performs the sub-command logic.
//
// It implements cli/handler/cobra.Handler.
func (h *Handler) Run(ctx context.Context, input handler.Input) {
	creds, credsErr := h.Auth.Credentials(&h.RoleChain)
	h.ExitOnErr(credsErr, "failed to acquire credentials", 1)
	h.Exec.Do(ctx, creds, input.Args)
}

// New returns a cobra command instance based on Handler.
func NewCommand() *cobra.Command {
	return handler_cobra.NewHandler(&Handler{
		Session: &handler.DefaultSession{},
	})
}

var _ handler_cobra.Handler = (*Handler)(nil)
