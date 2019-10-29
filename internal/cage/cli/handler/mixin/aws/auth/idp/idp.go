// Copyright (C) 2019 The CodeActual Go Environment Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package role

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	cage_aws "github.com/codeactual/aws-exec-cmd/internal/cage/aws"
	cage_cognito_core "github.com/codeactual/aws-exec-cmd/internal/cage/aws/cognito"
	cage_cognito "github.com/codeactual/aws-exec-cmd/internal/cage/aws/v1/cognito"
	"github.com/codeactual/aws-exec-cmd/internal/cage/cli/handler"
	handler_aws_auth "github.com/codeactual/aws-exec-cmd/internal/cage/cli/handler/mixin/aws/auth"
	google_auth "github.com/codeactual/aws-exec-cmd/internal/cage/google/auth"
)

type Mixin struct {
	ClientId             string
	ClientSecret         string
	ProviderRefreshToken string

	IdentityPoolId  string
	ProviderIdToken string
	ProviderName    string
}

// Implements cage/cli/handler.Mixin
func (m *Mixin) BindCobraFlags(cmd *cobra.Command) []string {
	cmd.Flags().StringVarP(&m.ClientId, "client-id", "", "", "(--refresh requirement) Client ID")
	cmd.Flags().StringVarP(&m.ClientSecret, "client-secret", "", "", "(--refresh requirement) Client secret")
	cmd.Flags().StringVarP(&m.ProviderRefreshToken, "refresh", "", "", "Provider refresh token (used to obtain an ID token)")

	cmd.Flags().StringVarP(&m.IdentityPoolId, "pool-id", "", "", "Identity pool ID")
	cmd.Flags().StringVarP(&m.ProviderName, "name", "", "", "Provider name, e.g. "+cage_cognito_core.GoogleProviderName)
	cmd.Flags().StringVarP(&m.ProviderIdToken, "token", "", "", "Provider ID token, e.g. Google id_token")

	return []string{"pool-id", "name"}
}

// Implements cage/cli/handler.Mixin
func (m *Mixin) Name() string {
	return "cage/cli/handler/mixin/aws/auth/idp"
}

// Implements cage/cli/handler.PreRun
func (m *Mixin) PreRun(ctx context.Context, args []string) error {
	supportedRefreshProviders := map[string]bool{cage_cognito_core.GoogleProviderName: true}

	if m.ProviderIdToken == "" && m.ProviderRefreshToken == "" {
		return errors.New("must input --token or --refresh")
	}

	if m.ProviderIdToken != "" && m.ProviderRefreshToken != "" {
		return errors.New("only input --token or --refresh")
	}

	if m.ProviderIdToken == "" {
		if !supportedRefreshProviders[m.ProviderName] {
			return errors.Errorf("--refresh is not supported for provider %s", m.ProviderName)
		}

		if m.ClientId == "" {
			return errors.New("--refresh requires --id")
		}
		if m.ClientSecret == "" {
			return errors.New("--refresh requires --secret")
		}
	}

	return nil
}

func (m *Mixin) Get(input handler_aws_auth.ProviderInput) (*credentials.Credentials, error) {
	region := cage_aws.GetenvRegion()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})

	if err != nil {
		return nil, errors.Wrapf(err, "failed to create a new session for region [%s]", region)
	}

	svc := cognitoidentity.New(sess)

	var idToken string
	if m.ProviderIdToken == "" {
		refresh, refreshErr := google_auth.RequestRefresh(input.Ctx, m.ClientId, m.ClientSecret, m.ProviderRefreshToken)
		if refreshErr != nil {
			return nil, errors.Wrap(refreshErr, "failed to request refresh")
		}

		idToken = refresh.IDToken
	} else {
		idToken = m.ProviderIdToken
	}

	res, err := cage_cognito.SingleIdentityLogin(svc, m.IdentityPoolId, m.ProviderName, idToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request credentials")
	}

	return res.Creds, nil
}

var _ handler.Mixin = (*Mixin)(nil)
var _ handler.PreRun = (*Mixin)(nil)
var _ handler_aws_auth.Provider = (*Mixin)(nil)
