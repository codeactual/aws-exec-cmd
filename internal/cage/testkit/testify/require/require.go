// Copyright (C) 2019 The CodeActual Go Environment Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package require

import (
	"fmt"
	"regexp"
	"testing"

	std_require "github.com/stretchr/testify/require"
)

func MatchRegexp(t *testing.T, subject string, expectedReStr ...string) {
	for _, reStr := range expectedReStr {
		std_require.True(
			t,
			regexp.MustCompile(reStr).MatchString(subject),
			fmt.Sprintf("subject [%s]\nregexp [%s]", subject, reStr),
		)
	}
}
