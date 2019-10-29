// Copyright (C) 2019 The aws-exec-cmd Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package idp

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/codeactual/aws-exec-cmd/internal/cage/cli/handler"
	handler_cobra "github.com/codeactual/aws-exec-cmd/internal/cage/cli/handler/cobra"
	"github.com/codeactual/aws-exec-cmd/internal/cage/cli/handler/mixin/aws/auth"
	auth_idp "github.com/codeactual/aws-exec-cmd/internal/cage/cli/handler/mixin/aws/auth/idp"
	cmd_mixin "github.com/codeactual/aws-exec-cmd/mixin"
)

// Handler defines the sub-command flags and logic.
type Handler struct {
	handler.IO

	Auth         auth.Mixin
	IdentityPool auth_idp.Mixin
	Exec         cmd_mixin.Exec
}

// Init defines the command, its environment variable prefix, etc.
//
// It implements cli/handler/cobra.Handler.
func (h *Handler) Init() handler_cobra.Init {
	return handler_cobra.Init{
		Cmd: &cobra.Command{
			Use:   "idp",
			Short: "Acquire credentials from a Cognito identity pool",
		},
		EnvPrefix: "AWS_EXEC_CMD",
		Mixins: []handler.Mixin{
			&h.Auth,
			&h.Exec,
			&h.IdentityPool,
		},
	}
}

// BindFlags binds the flags to Handler fields.
//
// It implements cli/handler/cobra.Handler.
func (h *Handler) BindFlags(cmd *cobra.Command) []string {
	return []string{} // auth_idp.Mixin provides all of them
}

// Run performs the sub-command logic.
//
// It implements cli/handler/cobra.Handler.
func (h *Handler) Run(ctx context.Context, args []string) {
	creds, credsErr := h.Auth.Credentials(&h.IdentityPool)
	h.ExitOnErr(credsErr, "failed to acquire credentials", 1)
	h.Exec.Do(ctx, creds, args)
}

// New returns a cobra command instance based on Handler.
func New() *cobra.Command {
	return handler_cobra.NewHandler(&Handler{})
}

var _ handler_cobra.Handler = (*Handler)(nil)
