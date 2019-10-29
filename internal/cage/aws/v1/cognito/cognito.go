// Copyright (C) 2019 The CodeActual Go Environment Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cognito

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/pkg/errors"

	core_cognito "github.com/codeactual/aws-exec-cmd/internal/cage/aws/cognito"
)

func SingleIdentityLogin(svc *cognitoidentity.CognitoIdentity, poolId, providerName, providerToken string) (core_cognito.IdentityLoginResult, error) {
	userId, err := svc.GetId(&cognitoidentity.GetIdInput{
		IdentityPoolId: aws.String(poolId),
		Logins: map[string]*string{
			providerName: aws.String(providerToken),
		},
	})
	if err != nil {
		return core_cognito.IdentityLoginResult{}, errors.Wrapf(err, "failed to get user identity from pool [%s] using provider [%s] with [%d] length token", poolId, providerName, len(providerToken))
	}

	creds, err := svc.GetCredentialsForIdentity(&cognitoidentity.GetCredentialsForIdentityInput{
		IdentityId: userId.IdentityId,
		Logins: map[string]*string{
			providerName: aws.String(providerToken),
		},
	})
	if err != nil {
		return core_cognito.IdentityLoginResult{}, errors.Wrapf(err, "failed to get credentials from pool [%s] using provider [%s] with [%d] length token", poolId, providerName, len(providerToken))
	}

	return core_cognito.IdentityLoginResult{
		Creds: credentials.NewStaticCredentials(
			*creds.Credentials.AccessKeyId,
			*creds.Credentials.SecretKey,
			*creds.Credentials.SessionToken,
		),
		Expiration: *creds.Credentials.Expiration,
		IdentityId: *creds.IdentityId,
	}, nil
}
