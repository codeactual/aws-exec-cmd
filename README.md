# aws-exec-cmd [![GoDoc](https://godoc.org/github.com/codeactual/aws-exec-cmd?status.svg)](https://pkg.go.dev/mod/github.com/codeactual/aws-exec-cmd) [![Go Report Card](https://goreportcard.com/badge/github.com/codeactual/aws-exec-cmd)](https://goreportcard.com/report/github.com/codeactual/aws-exec-cmd) [![Build Status](https://travis-ci.org/codeactual/aws-exec-cmd.png)](https://travis-ci.org/codeactual/aws-exec-cmd)

aws-exec-cmd acquires AWS credentials and runs an arbitrary command, providing it credentials through environment variables. It acquires credentials from the environment, IAM roles (with AssumeRole chaining), or Cognito identity pools.

## Use Case

Use authenticated commands with credential providers they do not natively support, e.g. EC2 instance role.

# Usage

> To install: `go get -v github.com/codeactual/aws-exec-cmd/cmd/aws-exec-cmd`

Environment variables passed to commands:

- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `AWS_SESSION_TOKEN`

## Examples

> Usage:

```bash
aws-exec-cmd --help
aws-exec-cmd role --help
aws-exec-cmd idp --help
```

> Use the IAM role, attached to an EC2 instance, to run "env | grep AWS_":

```bash
aws-exec-cmd role --chain instance -- env | grep AWS_
```

> Perform the same command but with credentials from role "backup" assumed from an EC2 instance role:

```bash
aws-exec-cmd role --chain instance,arn:aws:iam::123456789012:role/backup -- env | grep AWS_
```

> Perform the same command but with credentials from role "backup" assumed from enviroment credentials:

```bash
aws-exec-cmd role --chain env-triple,arn:aws:iam::123456789012:role/backup -- env | grep AWS_
```

> Perform the same command with credentials from Cognito identity pool, using federated Google auth:

```bash
aws-exec-cmd idp \
  --name accounts.google.com \
  --pool-id <pool ID> \
  --refresh <Google OAuth refresh token> \
  --client-id <Google OAuth client ID> \
  --client-secret <Google OAuth client secret>
```

> Supported AssumeRole chaining:

- environment variable credentials -> `AssumeRole` [-> `AssumeRole` ...]
- role (temporary credentials from STS) -> `AssumeRole` [-> `AssumeRole` ...]

## Travis CI

### Config

- Generate AWS API credentials which will be added to the config file as encrypted environment variables.
  - [travis-iam-assume-role.json](testdata/iam/travis-iam-assume-role.json) grants permissions to assume the role. For example, attach this to your Travis-dedicated IAM user.
    - The [Travis CI IP addresses](https://docs.travis-ci.com/user/ip-addresses/) in the policy conditions may be out-of-date.
  - [travis-iam-resource.json](testdata/iam/travis-iam-resource.json) grants permissions to the role to access the resource.
- To configure the environment variables used by the functional test against the EC2 API, use the [Travis CLI](https://docs.travis-ci.com/user/encryption-keys/#usage) to generate the `secure` string value.
  - Each `env` item expects all key/value pairs as one string, and multiple items define multiple build permutations so that all pair sets are tested. Input an entire set, e.g. `AWS_ACCESS_KEY_ID=... AWS_SECRET_ACCESS_KEY=... ROLE_ARN=...`, in the `encrypt` command.
  - Launch `travis` in interactive mode `-i` and input the pair set without trailing newline.

# Development

## License

[Mozilla Public License Version 2.0](https://www.mozilla.org/en-US/MPL/2.0/) ([About](https://www.mozilla.org/en-US/MPL/), [FAQ](https://www.mozilla.org/en-US/MPL/2.0/FAQ/))

## Contributing

- Please feel free to submit issues, PRs, questions, and feedback.
- Although this repository consists of snapshots extracted from a private monorepo using [transplant](https://github.com/codeactual/transplant), PRs are welcome. Standard GitHub workflows are still used.
