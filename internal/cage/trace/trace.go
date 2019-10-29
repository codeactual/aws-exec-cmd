// Copyright (C) 2019 The CodeActual Go Environment Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package trace

import (
	"fmt"

	"github.com/go-stack/stack"
)

// CallerID returns an ID in the format "pkg/path/source.go:10:CallerFunc" of its
// caller's caller.
func CallerID() string {
	c := stack.Caller(2) // skip this level and the immediate caller's
	return fmt.Sprintf("%+v", c) + ":" + fmt.Sprintf("%n", c)
}
