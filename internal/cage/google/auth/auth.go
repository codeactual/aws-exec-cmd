// Copyright (C) 2019 The CodeActual Go Environment Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	cage_io "github.com/codeactual/aws-exec-cmd/internal/cage/io"
	"github.com/codeactual/aws-exec-cmd/internal/cage/trace"
)

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	IDToken     string `json:"id_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

func (r RefreshResponse) String() string {
	b, err := json.Marshal(r)
	if err != nil {
		return fmt.Sprintf("%+v", err)
	}
	return string(b)
}

const (
	// https://developers.google.com/identity/protocols/OAuth2WebServer#offline (HTTP/REST tab)
	RefreshTokenUrlVer4 = "https://www.googleapis.com/oauth2/v4/token" // #nosec
)

func ClientFromRefreshToken(ctx context.Context, clientId, clientSecret, refreshToken string) *http.Client {
	c := oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
	}
	return c.Client(ctx, &oauth2.Token{RefreshToken: refreshToken})
}

func RequestRefresh(ctx context.Context, clientId, clientSecret, refreshToken string) (RefreshResponse, error) {
	client := ClientFromRefreshToken(ctx, clientId, clientSecret, refreshToken)
	body := url.Values{}
	body.Add("client_id", clientId)
	body.Add("client_secret", clientSecret)
	body.Add("refresh_token", refreshToken)
	body.Add("grant_type", "refresh_token")

	res, err := client.PostForm(RefreshTokenUrlVer4, body)
	if err != nil {
		return RefreshResponse{}, errors.WithStack(err)
	}
	defer cage_io.CloseOrStderr(res.Body, trace.CallerID())

	// To debug the response: dump, err := httputil.DumpResponse(res, true)

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return RefreshResponse{}, errors.WithStack(err)
	}

	rr := RefreshResponse{}
	err = json.Unmarshal(bodyBytes, &rr)
	if err != nil {
		return RefreshResponse{}, errors.WithStack(err)
	}

	return rr, nil
}
