// Copyright (C) 2019 The CodeActual Go Environment Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package idp

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	cage_sts "github.com/codeactual/aws-exec-cmd/internal/cage/aws/v1/sts"
	"github.com/codeactual/aws-exec-cmd/internal/cage/cli/handler"
	"github.com/codeactual/aws-exec-cmd/internal/cage/cli/handler/mixin/aws/auth"
)

// Mixin defines the sub-command flags and logic.
type Mixin struct {
	// Normally the role chain string would be defined here (instead of in the
	// cli/handler/mixin/aws/auth mixin), but the latter needs it earlier than
	// the Provider.Get call for the cache read (key).
}

// Implements cage/cli/handler.Mixin
func (m *Mixin) BindCobraFlags(cmd *cobra.Command) []string {
	return []string{}
}

// Implements cage/cli/handler.Mixin
func (m *Mixin) Name() string {
	return "cage/cli/handler/mixin/aws/auth/role"
}

// Implements cage/cli/handler/mixin/aws/auth.Provider
func (m *Mixin) Get(input auth.ProviderInput) (*credentials.Credentials, error) {
	var parsedRoleChain []string
	for _, role := range strings.Split(input.RoleChain, ",") {
		role = strings.TrimSpace(role)
		if role != "" {
			parsedRoleChain = append(parsedRoleChain, role)
		}
	}

	if len(parsedRoleChain) == 0 {
		return nil, errors.New("role chain required")
	}

	resolveInput := cage_sts.ResolveRoleChainInput{
		Chain:           parsedRoleChain,
		DurationSeconds: int64(input.SessionTtlSec),
	}
	if input.MfaSerial != "" {
		resolveInput.SerialNumber = input.MfaSerial
		resolveInput.TokenCode = input.MfaCode
	}

	id, secret, token, resolveErr := cage_sts.ResolveRoleChain(&resolveInput)
	if resolveErr != nil {
		return nil, errors.Wrapf(resolveErr, "failed to resolve role chain [%s]", strings.Join(parsedRoleChain, ","))
	}

	return credentials.NewStaticCredentials(id, secret, token), nil
}

var _ handler.Mixin = (*Mixin)(nil)
var _ auth.Provider = (*Mixin)(nil)
