// Copyright (C) 2019 The CodeActual Go Environment Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package auth

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/codeactual/aws-exec-cmd/internal/cage/aws/credentials/cache"
	"github.com/codeactual/aws-exec-cmd/internal/cage/cli/handler"
	"github.com/codeactual/aws-exec-cmd/internal/cage/os/terminal"
	cage_reflect "github.com/codeactual/aws-exec-cmd/internal/cage/reflect"
)

const (
	DefaultSessionTtlSec = 900
	DefaultMfaSource     = "prompt"

	defaultHomeCacheDir = ".cage-aws-cache"

	// cacheEarlyTtlSec reduces the opportunity for a command to receive a cached session
	// that expires before use.
	cacheEarlyTtlSec = 10
)

type ProviderInput struct {
	Ctx           context.Context
	MfaSerial     string
	MfaCode       string
	RoleChain     string
	SessionTtlSec int
}

type Provider interface {
	Get(ProviderInput) (*credentials.Credentials, error)
}

type Mixin struct {
	Ctx context.Context

	CacheDir  string
	CacheSkip bool   `usage:"Skip reading from cache (but still write after success)"`
	MfaSerial string `usage:"MFA serial ARN"`
	MfaSource string `usage:"MFA source (to read from an environment variable, provide the variable's name)"`

	// Normally this would live in the cli/handler/mixin/aws/auth/role mixin, but it's
	// needed earlier than the Provider.Get call for the cache read (key).
	RoleChain string `usage:"Comma-separated aliases, e.g. \"instance\" or ARNs (role auth mode only)"`

	SessionTtlSec int `usage:"Session length in seconds"`

	// RoleChainFlag is the CLI flag for the RoleChain field.
	//
	// It defaults to "role".
	RoleChainFlag string
}

// Implements cage/cli/handler.Mixin
func (m *Mixin) BindCobraFlags(cmd *cobra.Command) []string {
	roleChainFlag := m.RoleChainFlag
	if roleChainFlag == "" {
		roleChainFlag = "role"
	}

	cmd.Flags().StringVarP(&m.CacheDir, "cache-dir", "", "", "Defaults to ~/"+defaultHomeCacheDir)
	cmd.Flags().BoolVarP(&m.CacheSkip, "cache-skip", "", false, cage_reflect.GetFieldTag(*m, "CacheSkip", "usage"))
	cmd.Flags().StringVarP(&m.MfaSerial, "mfa-serial", "", "", cage_reflect.GetFieldTag(*m, "MfaSerial", "usage"))
	cmd.Flags().StringVarP(&m.MfaSource, "mfa-source", "", DefaultMfaSource, cage_reflect.GetFieldTag(*m, "MfaSource", "usage"))
	cmd.Flags().StringVarP(&m.RoleChain, roleChainFlag, "", "", cage_reflect.GetFieldTag(*m, "RoleChain", "usage"))
	cmd.Flags().IntVarP(&m.SessionTtlSec, "session-ttl", "", DefaultSessionTtlSec, cage_reflect.GetFieldTag(*m, "SessionTtlSec", "usage"))
	return []string{roleChainFlag}
}

// Implements cage/cli/handler.Mixin
func (m *Mixin) Name() string {
	return "cage/cli/handler/mixin/aws/auth"
}

// Implements cage/cli/handler.PreRun
func (m *Mixin) PreRun(ctx context.Context, args []string) error {
	m.Ctx = ctx
	return nil
}

func (m *Mixin) Credentials(provider Provider) (*credentials.Credentials, error) {
	if m.CacheDir == "" {
		homeDir, homeErr := homedir.Dir()
		if homeErr != nil {
			return nil, errors.Wrapf(homeErr, "failed to detect home dir for use as default --cache-dir")
		}
		m.CacheDir = filepath.Join(homeDir, defaultHomeCacheDir)
	}

	var mfaCode string
	if m.MfaSerial != "" {
		if m.MfaSource == DefaultMfaSource {
			mfaCodeBytes, promptErr := terminal.DefaultProvider{}.PromptHiddenf("MFA token:")
			if promptErr != nil {
				log.Fatalf("failed to read MFA token: %+v", promptErr)
				return nil, errors.Wrap(promptErr, "failed to MFA token from prompt")
			}
			mfaCode = string(mfaCodeBytes)
		} else {
			mfaCode = os.Getenv(m.MfaSource)
		}
	}

	var cacheVal cache.Value

	cacheStore := cache.NewStore(m.CacheDir)
	cacheKey := cache.Key{
		MfaSerial: m.MfaSerial,
		Role:      m.RoleChain,
	}

	if !m.CacheSkip {
		var readErr error
		cacheVal, readErr = cacheStore.Read(cacheKey)
		if readErr != nil {
			return nil, errors.Wrapf(readErr, "failed to read cache key [%s]", cacheKey)
		}
	}

	var creds *credentials.Credentials

	if cacheVal.AccessKeyID == "" {
		var providerErr error
		creds, providerErr = provider.Get(ProviderInput{
			Ctx:           m.Ctx,
			MfaSerial:     m.MfaSerial,
			MfaCode:       mfaCode,
			RoleChain:     m.RoleChain,
			SessionTtlSec: m.SessionTtlSec,
		})
		if providerErr != nil {
			return nil, errors.Wrap(providerErr, "failed to get credentials provider")
		}

		credsVal, credsErr := creds.Get()
		if credsErr != nil {
			return nil, errors.Wrap(credsErr, "failed to get credentials value")
		}

		cacheVal = cache.Value{
			AccessKeyID:     credsVal.AccessKeyID,
			SecretAccessKey: credsVal.SecretAccessKey,
			SessionToken:    credsVal.SessionToken,
			Expires:         time.Now().Add(time.Duration(m.SessionTtlSec-cacheEarlyTtlSec) * time.Second).Unix(),
		}

		writeErr := cacheStore.Write(cacheKey, cacheVal)
		if writeErr != nil {
			return nil, errors.Wrapf(writeErr, "failed to write cache key [%s]", cacheKey)
		}
	} else {
		creds = credentials.NewStaticCredentials(cacheVal.AccessKeyID, cacheVal.SecretAccessKey, cacheVal.SessionToken)
	}

	return creds, nil
}

var _ handler.Mixin = (*Mixin)(nil)
var _ handler.PreRun = (*Mixin)(nil)
