// Copyright (C) 2019 The CodeActual Go Environment Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

// Key contents are used to build the filename where the value is read from or written to.
//
// All characters are allowed in the string values. The final key will be a filename-safe hash.
type Key struct {
	MfaSerial string

	// Role can hold any type of identifier as long as the caller can rely on its uniqueness.
	// For example, it can hold an ARN but also a comma-separated role chain composed of
	// reliable identifiers.
	Role string
}

func (k Key) String() string {
	hash := sha256.Sum256([]byte(k.MfaSerial + k.Role))
	// use [:] to convert [32]byte to []byte
	return hex.EncodeToString(hash[:])
}

type Value struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Expires         int64
}

type Store struct {
	Dir string
}

func NewStore(dir string) Store {
	return Store{Dir: dir}
}

// Write saves the credentials triple if found by the given key.
//
// If the cache dir does not exist, it will be created.
func (s Store) Write(k Key, out Value) error {
	if dirErr := s.mkdir(); dirErr != nil {
		return errors.WithStack(dirErr)
	}

	buf, jsonErr := json.Marshal(out)
	if jsonErr != nil {
		return errors.WithStack(jsonErr)
	}

	filename := s.filename(k)
	if writeErr := ioutil.WriteFile(filename, buf, 0600); writeErr != nil {
		return errors.Wrapf(writeErr, "failed to read cache file [%s]", filename)
	}

	return nil
}

// Read returns the credentials triple if found by the given key.
//
// Callers will receive three empty strings and a nil error on a cache miss.
// An error is returned only if the read operation failed.
// If the cache dir does not exist, it will be created.
func (s Store) Read(k Key) (v Value, err error) {
	if dirErr := s.mkdir(); dirErr != nil {
		return Value{}, errors.WithStack(dirErr)
	}

	filename := s.filename(k)
	buf, readErr := ioutil.ReadFile(filename) // #nosec G304
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return Value{}, nil // don't treat cache miss as an error
		}
		return Value{}, errors.Wrapf(readErr, "failed to read cache file [%s]", filename)
	}

	if jsonErr := json.Unmarshal(buf, &v); jsonErr != nil {
		return Value{}, errors.WithStack(jsonErr)
	}

	if v.Expires-time.Now().Unix() > 0 {
		return v, nil
	}

	// No need to remove the stale file. Let the next successful Write operation replace it.

	return Value{}, nil
}

func (s Store) filename(k Key) string {
	return filepath.Join(s.Dir, k.String())
}

func (s Store) mkdir() (err error) {
	return os.MkdirAll(s.Dir, 0700)
}
