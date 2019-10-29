// Copyright (C) 2019 The CodeActual Go Environment Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package file

import (
	"path/filepath"
	"testing"

	"github.com/codeactual/aws-exec-cmd/internal/cage/testkit"
)

// dataDir defines the base directory for fixtures and test data.
const dataDir = "testdata"

func FixtureDataDir() string {
	return filepath.Join(dataDir, "fixture")
}

// FixturePath returns a path under FixtureDataDir() at a relative path built by joining the path parts.
func FixturePath(t *testing.T, pathPart ...string) (relPath string, absPath string) {
	pathPart = append([]string{FixtureDataDir()}, pathPart...)
	relPath = filepath.Join(pathPart...)

	absPath, err := filepath.Abs(relPath)
	testkit.FatalErrf(t, err, "failed to get absolute path [%s]", relPath)

	return relPath, absPath
}
