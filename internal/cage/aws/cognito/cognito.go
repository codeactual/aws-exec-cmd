// Copyright (C) 2019 The CodeActual Go Environment Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cognito

import (
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

const (
	GoogleProviderName = "accounts.google.com"
)

type IdentityLoginResult struct {
	Creds      *credentials.Credentials
	Expiration time.Time
	IdentityId string
}
