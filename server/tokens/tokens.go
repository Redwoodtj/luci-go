// Copyright 2015 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package tokens provides means to generate and validate base64 encoded tokens
// compatible with luci-py's components.auth implementation.
package tokens

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"strconv"
	"strings"
	"time"

	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/server/secrets"
)

// allowedClockDrift is clock drift between machines we can tolerate.
const allowedClockDrift = 30 * time.Second

// TokenAlgo identifies how token is authenticated.
type TokenAlgo string

const (
	// TokenAlgoHmacSHA256 algorithm stores public portion of the token as plain
	// text and uses HMAC SHA256 to authenticate its integrity.
	TokenAlgoHmacSHA256 = "HMAC-SHA256"
)

// hash returns hash.Hash that computes the digest or nil if algo is unknown.
func (a TokenAlgo) hash(secret []byte) hash.Hash {
	switch a {
	case TokenAlgoHmacSHA256:
		return hmac.New(sha256.New, secret)
	}
	return nil
}

// digestLen returns length of digest generated by an algo or 0 if unknown.
func (a TokenAlgo) digestLen() int {
	switch a {
	case TokenAlgoHmacSHA256:
		return sha256.Size
	}
	return 0
}

// TokenKind is a configuration of particular type of a token. It can be
// defined statically in a module and then its Generate() and Validate() methods
// can be used to produce and verify tokens.
type TokenKind struct {
	Algo       TokenAlgo
	Expiration time.Duration // how long generated token lives
	SecretKey  string        // name of the secret key in secrets.Store
	Version    byte          // tokens with another version will be rejected
}

// Generate produces an urlsafe base64 encoded string that contains 'embedded'
// and MAC tag for 'state' + 'embedded' (but not the 'state' itself). The exact
// same 'state' then must be used in Validate to successfully verify the token.
//
// 'embedded' is an optional map with additional data to add to the token. It is
// embedded directly into the token and can be easily extracted from it by
// anyone who has the token. Should be used only for publicly visible data. It
// is tagged by token's MAC, so 'Validate' function can detect any modifications
// (and reject tokens tampered with).
//
// The context is used to grab secrets.Store and the current time.
func (k *TokenKind) Generate(c context.Context, state []byte, embedded map[string]string, exp time.Duration) (string, error) {
	extended := make(map[string]string, len(embedded))
	for k, v := range embedded {
		if len(k) == 0 {
			return "", fmt.Errorf("tokens: empty key in embedded map")
		}
		if k[0] == '_' {
			return "", fmt.Errorf("token: bad key %q in embedded map", k)
		}
		extended[k] = v
	}

	// Append 'issued' timestamp (in milliseconds) and expiration time (if not
	// default).
	extended["_i"] = strconv.FormatInt(clock.Now(c).UnixNano()/1e6, 10)
	if exp != 0 {
		if exp < 0 {
			return "", fmt.Errorf("tokens: expiration can't be negative")
		}
		extended["_x"] = strconv.FormatInt(exp.Nanoseconds()/1e6, 10)
	}

	// 'public' will be added to the token as is.
	public, err := json.Marshal(extended)
	if err != nil {
		return "", err
	}

	// Build HMAC tag.
	secret, err := secrets.RandomSecret(c, k.SecretKey)
	if err != nil {
		return "", err
	}
	mac, err := computeMAC(k.Algo, secret.Current, dataToAuth(k.Version, public, state))
	if err != nil {
		return "", err
	}

	encoded := base64.RawURLEncoding.EncodeToString(bytes.Join([][]byte{
		{k.Version},
		public,
		mac,
	}, nil))
	return strings.TrimRight(encoded, "="), nil
}

// Validate checks token MAC and expiration, decodes data embedded into it.
//
// 'state' must be exactly the same as passed to Generate when creating a token.
// If it's different, the token is considered invalid. It usually contains some
// implicitly passed state that should be the same when token is generated and
// validated. For example, it may be an account ID of a current caller. Then if
// such token is used by another account, it is considered invalid.
//
// The context is used to grab secrets.Store and the current time.
func (k *TokenKind) Validate(c context.Context, token string, state []byte) (map[string]string, error) {
	digestLen := k.Algo.digestLen()
	if digestLen == 0 {
		return nil, fmt.Errorf("tokens: unknown algo %q", k.Algo)
	}
	blob, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}

	// One byte for version, at least one byte for public embedded dict portion,
	// the rest is MAC digest.
	if len(blob) < digestLen+2 {
		return nil, fmt.Errorf("tokens: the token is too small")
	}

	// Data inside the token.
	version := blob[0]
	public := blob[1 : len(blob)-digestLen]
	tokenMac := blob[len(blob)-digestLen:]

	// Data that should have been used to generate HMAC.
	toAuth := dataToAuth(version, public, state)

	// Token could have been generated by previous value of the secret, so check
	// them too.
	secret, err := secrets.RandomSecret(c, k.SecretKey)
	if err != nil {
		return nil, err
	}
	goodToken := false
	for _, blob := range secret.Blobs() {
		goodMac, err := computeMAC(k.Algo, blob, toAuth)
		if err != nil {
			return nil, err
		}
		if hmac.Equal(tokenMac, goodMac) {
			goodToken = true
			break
		}
	}
	if !goodToken {
		return nil, fmt.Errorf("tokens: bad token MAC")
	}

	// Token is authenticated, now check the rest.
	if version != k.Version {
		return nil, fmt.Errorf("tokens: bad version %q, expecting %q", version, k.Version)
	}
	embedded := map[string]string{}
	if err := json.Unmarshal(public, &embedded); err != nil {
		return nil, err
	}

	// Grab issued time, reject token from the future.
	now := clock.Now(c)
	issuedMs, err := popInt(embedded, "_i")
	if err != nil {
		return nil, err
	}
	issuedTs := time.Unix(0, issuedMs*1e6)
	if issuedTs.After(now.Add(allowedClockDrift)) {
		return nil, fmt.Errorf("tokens: issued timestamp is in the future")
	}

	// Grab expiration time embedded into the token, if any.
	expiration := k.Expiration
	if _, ok := embedded["_x"]; ok {
		expirationMs, err := popInt(embedded, "_x")
		if err != nil {
			return nil, err
		}
		expiration = time.Duration(expirationMs) * time.Millisecond
	}
	if expiration < 0 {
		return nil, fmt.Errorf("tokens: bad token, expiration can't be negative")
	}

	// Check token expiration.
	expired := now.Sub(issuedTs.Add(expiration))
	if expired > 0 {
		return nil, fmt.Errorf("tokens: token expired %s ago", expired)
	}

	return embedded, nil
}

// extractInt pops integer value from the map.
func popInt(m map[string]string, key string) (int64, error) {
	str, ok := m[key]
	if !ok {
		return 0, fmt.Errorf("tokens: bad token, missing %q key", key)
	}
	asInt, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("tokens: bad token, %q is not a number", str)
	}
	delete(m, key)
	return asInt, nil
}

// dataToAuth generates list of byte blobs authenticated by MAC.
func dataToAuth(version byte, public []byte, state []byte) [][]byte {
	out := [][]byte{
		{version},
		public,
	}
	if len(state) != 0 {
		out = append(out, state)
	}
	return out
}

// computeMAC packs dataToAuth into single blob and computes its MAC.
func computeMAC(algo TokenAlgo, secret []byte, dataToAuth [][]byte) ([]byte, error) {
	hash := algo.hash(secret)
	if hash == nil {
		return nil, fmt.Errorf("tokens: unknown algo %q", algo)
	}
	for _, chunk := range dataToAuth {
		// Separator between length header and the body is needed because length
		// encoding is variable-length (decimal string).
		if _, err := fmt.Fprintf(hash, "%d\n", len(chunk)); err != nil {
			return nil, err
		}
		if _, err := hash.Write(chunk); err != nil {
			return nil, err
		}
	}
	return hash.Sum(nil), nil
}
