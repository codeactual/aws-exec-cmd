// Copyright (C) 2019 The CodeActual Go Environment Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sts

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/pkg/errors"

	cage_resource "github.com/codeactual/aws-exec-cmd/internal/cage/aws/v1/resource"
	cage_crypto "github.com/codeactual/aws-exec-cmd/internal/cage/crypto"
)

const (
	// InstanceRoleChainAlias can be used at the head of a role chain to use seed the walk
	// with credentials from the metadata service.
	InstanceRoleChainAlias = "instance"

	// EnvTempRoleChainAlias can be used at the head of a role chain to use seed the walk
	// with temporary credentials from the environment.
	//
	// The keys match those of v1's aws/credentials/env_provider.go:
	//
	// AWS_ACCESS_KEY_ID
	// AWS_SECRET_ACCESS_KEY
	// AWS_SESSION_TOKEN
	EnvTempRoleChainAlias = "env-triple"
)

// ResolveRoleChainInput describes the chain to traverse and initial credentials, if any.
type ResolveRoleChainInput struct {
	// Initial access key to seed the traversal's first STS-dependent step
	AccessKey string
	// Initial secret access key to seed the traversal's first STS-dependent step
	SecretAccessKey string
	// Initial session token to seed the traversal's first STS-dependent step
	SessionToken string
	// SessioName will be applied to AssumeRole
	SessionName string
	// Region uses the format "us-west-2"
	Region string
	// Chain contains the EC2 metadata URL subpaths and role ARNs
	//         meta-data/iam/security-credentials/someRole1
	//         arn:aws:iam::123456789:role/someRole2
	Chain []string
	// SerialNumber is an MFA device hardware serial number or virtual devce ARN.
	SerialNumber string
	// TokenCode is a code from an MFA device.
	TokenCode string
	// DurationSeconds is the session lifetime (min 900).
	DurationSeconds int64
}

func (i ResolveRoleChainInput) String() string {
	return fmt.Sprintf(
		"session [%s] region [%s] mfa serial [%d chars] mfa code [%d chars] ttl [%d sec] chain [%s]",
		i.SessionName, i.Region, len(i.SerialNumber), len(i.TokenCode), i.DurationSeconds, strings.Join(i.Chain, ","),
	)
}

// ResolveRoleChainLog wraps a standard error and also include more details about
// the traversal progress to help identify where it stopped.
type ResolveRoleChainLog []string

// String implements the Stringer interface with a comma-separated list of completed resolution steps.
func (r ResolveRoleChainLog) String() string {
	var roles string
	if len(r) > 0 {
		roles = strings.Join(r, ",")
	} else {
		roles = "no roles resolved"
	}
	return roles
}

// NewBasicEC2RoleProvider returns an EC2RoleProvider for a given role.
func NewBasicEC2RoleProvider() (*ec2rolecreds.EC2RoleProvider, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &ec2rolecreds.EC2RoleProvider{Client: ec2metadata.New(sess)}, nil
}

// GetEC2RoleCreds returns credentials using the given role.
func GetEC2RoleCreds() (credentials.Value, error) {
	p, err := NewBasicEC2RoleProvider()
	if err != nil {
		return credentials.Value{}, errors.WithStack(err)
	}
	return credentials.NewCredentials(p).Get()
}

// GetAssumeRoleCreds returns credentials using the given role.
func GetAssumeRoleCreds(arn string, input *ResolveRoleChainInput, config *aws.Config) (creds credentials.Value, err error) {
	sess, err := session.NewSession(config)
	if err != nil {
		return credentials.Value{}, errors.WithStack(err)
	}
	svc := sts.New(sess)

	params := sts.AssumeRoleInput{
		RoleArn: aws.String(arn),
	}

	// Unlike general AWS config session names, which default to an automatically generated name,
	// AssumeRole requires one to avoid this error:
	//
	//   InvalidParameter: 1 validation error(s) found.
	//   - missing required field, AssumeRoleInput.RoleSessionName.
	if input.SessionName == "" {
		randStr, randStrErr := cage_crypto.RandHexString(2)
		if randStrErr != nil {
			return credentials.Value{}, errors.Wrap(err, "failed to generate random AssumeRole session name")
		}
		params.RoleSessionName = aws.String("cage.aws.v1.sts.GetAssumeRoleCreds." + randStr)
	} else {
		params.RoleSessionName = aws.String(input.SessionName)
	}

	if input.SerialNumber != "" {
		params.SerialNumber = aws.String(input.SerialNumber)
	}
	if input.TokenCode != "" {
		params.TokenCode = aws.String(input.TokenCode)
	}
	if input.DurationSeconds > 0 {
		params.DurationSeconds = aws.Int64(input.DurationSeconds)
	}

	resp, err := svc.AssumeRole(&params)
	if err != nil {
		return credentials.Value{}, errors.WithStack(err)
	}

	return SvcToBasicCreds(resp.Credentials), nil
}

// ResolveRoleChain returns the final credentials triple after walking a list of roles.
// Each chain element is acquired using the results of the prior AssumeRole API call.
// initialCreds can be nil, ex. when the first element of the chain is an instance
// profile that will seed the traversal.
func ResolveRoleChain(input *ResolveRoleChainInput) (accessKey string, secretAccessKey string, sessionToken string, err error) {
	log := ResolveRoleChainLog{}
	resolveErr := func(err error) error {
		return errors.Wrapf(err, "failed to resolve role chain (%s) log [%s]", input, log)
	}

	chain := input.Chain
	prior := credentials.Value{}

	if len(chain) == 0 {
		return "", "", "", errors.New("no links in role chain")
	}

	if input.AccessKey != "" && input.SecretAccessKey != "" {
		// E.g. to support use of a keypair on a laptop.
		awsConfig := &aws.Config{
			Credentials: credentials.NewStaticCredentials(input.AccessKey, input.SecretAccessKey, input.SessionToken),
		}
		prior, err = GetAssumeRoleCreds(chain[0], input, awsConfig.WithRegion(input.Region))
		if err != nil {
			return "", "", "", errors.Wrapf(err, "failed to assume role [%s] using intiial static creds", chain[0])
		}
	} else {
		// Support aliases that select a credentials source that seeds the role chain,
		// e.g. EC2 instance role credentials that have permission to assume the next link
		// that is identified by ARN.
		if !cage_resource.IsARN(chain[0]) {
			switch chain[0] {
			case EnvTempRoleChainAlias:
				prior, err = credentials.NewEnvCredentials().Get()
				if err != nil {
					return "", "", "", resolveErr(err)
				}

				log = append(log, "seeded chain with env-triple creds")
			case InstanceRoleChainAlias:
				prior, err = GetEC2RoleCreds()
				if err != nil {
					return "", "", "", resolveErr(err)
				}

				log = append(log, "seeded chain with instance role creds")
			default:
				return "", "", "", errors.Errorf("role chain first-link alias [%s] is not recognized", chain[0])
			}

			chain = chain[1:] // Simplify the final for-loop.
		}
	}

	for _, link := range chain {
		if link == "" { // Ex. chain was Split(..., ",") and there's a trailing ","
			continue
		}

		if !cage_resource.IsARN(link) {
			err = errors.Errorf("non-ARN role is only allowed in first chain role, chain [%s]", strings.Join(input.Chain, ","))
			return "", "", "", resolveErr(err)
		}

		priorCredsExist := prior.AccessKeyID != ""

		log = append(log, fmt.Sprintf("about to assume role from link [%s] with prior creds [%t]", link, priorCredsExist))

		// By default the SDK will fall back to EC2RoleProvider when creating a new session with no provider specified.
		//
		// As of 8234000, the flow is: NewSession->NewSessionWithOptions->newSession->mergeConfigSrcs->
		// cfg.Credentials = credentials.NewCredentials(&credentials.ChainProvider{...} where the final/3rd provider
		// is defaults.RemoteCredProvider which by default returns defaults.ec2RoleProvider().
		//
		// Here we remove that "magic" and force the input role chain to explicitly choose it via InstanceRoleChainAlias.
		assumeConfig := &aws.Config{Region: aws.String(input.Region)}
		if priorCredsExist {
			assumeConfig.Credentials = credentials.NewStaticCredentials(prior.AccessKeyID, prior.SecretAccessKey, prior.SessionToken)
		} else {
			assumeConfig.Credentials = credentials.AnonymousCredentials
		}

		prior, err = GetAssumeRoleCreds(link, input, assumeConfig)
		if err != nil {
			return "", "", "", resolveErr(err)
		}

		// Only support MFA for the first role assumption in the chain.
		input.SerialNumber = ""
		input.TokenCode = ""

		log = append(log, "assumed role from link: "+link)
	}

	accessKey = prior.AccessKeyID
	secretAccessKey = prior.SecretAccessKey
	sessionToken = prior.SessionToken

	return accessKey, secretAccessKey, sessionToken, nil
}

// SvcToBasicCreds returns a basic credentials triple from the STS version.
func SvcToBasicCreds(c *sts.Credentials) credentials.Value {
	return credentials.Value{
		AccessKeyID:     *c.AccessKeyId,
		SecretAccessKey: *c.SecretAccessKey,
		SessionToken:    *c.SessionToken,
	}
}
