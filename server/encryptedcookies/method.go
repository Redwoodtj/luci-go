// Copyright 2021 The LUCI Authors.
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

package encryptedcookies

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/google/tink/go/tink"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.chromium.org/luci/auth/identity"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/retry/transient"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/openid"
	"go.chromium.org/luci/server/encryptedcookies/internal"
	"go.chromium.org/luci/server/encryptedcookies/internal/encryptedcookiespb"
	"go.chromium.org/luci/server/encryptedcookies/session"
	"go.chromium.org/luci/server/encryptedcookies/session/sessionpb"
	"go.chromium.org/luci/server/router"
)

// OpenIDConfig is a configuration related to OpenID Connect provider.
//
// All parameters are required.
type OpenIDConfig struct {
	// DiscoveryURL is where to grab discovery document with provider's config.
	DiscoveryURL string

	// ClientID identifies OAuth2 Web client representing the application.
	//
	// Can be obtained by registering the OAuth2 client with the identity
	// provider.
	ClientID string

	// ClientSecret is a secret associated with ClientID.
	//
	// Can be obtained by registering the OAuth2 client with the identity
	// provider.
	ClientSecret string

	// RedirectURI must be `https://<host>/auth/openid/callback`.
	//
	// The OAuth2 client should be configured to allow this redirect URL.
	RedirectURI string
}

// discoveryDoc returns the cached OpenID discovery document.
func (cfg *OpenIDConfig) discoveryDoc(ctx context.Context) (*openid.DiscoveryDoc, error) {
	// FetchDiscoveryDoc implements caching inside.
	doc, err := openid.FetchDiscoveryDoc(ctx, cfg.DiscoveryURL)
	if err != nil {
		return nil, errors.Annotate(err, "failed to fetch the discovery doc").Tag(transient.Tag).Err()
	}
	return doc, nil
}

// Method is an auth.Method implementation that uses encrypted cookies.
//
// Uses OpenID Connect to establish sessions and refresh tokens to verify
// OpenID identity provider still knows about the user.
type AuthMethod struct {
	// Configuration returns OpenID Connect configuration parameters.
	//
	// Required.
	OpenIDConfig func(ctx context.Context) (*OpenIDConfig, error)

	// Sessions keeps user sessions in some permanent storage.
	//
	// Required.
	Sessions session.Store

	// Insecure is true to allow http:// URLs and non-https cookies. Useful for
	// local development.
	Insecure bool

	// IncompatibleCookies is a list of cookies to remove when setting or clearing
	// the session cookie. It is useful to get rid of cookies from previously used
	// authentication methods.
	IncompatibleCookies []string
}

var _ interface {
	auth.Method
	auth.UsersAPI
	auth.Warmable
	auth.HasHandlers
} = (*AuthMethod)(nil)

const (
	loginURL    = "/auth/openid/login"
	logoutURL   = "/auth/openid/logout"
	callbackURL = "/auth/openid/callback"
)

// InstallHandlers installs HTTP handlers used in the login protocol.
//
// Implements auth.HasHandlers.
func (m *AuthMethod) InstallHandlers(r *router.Router, base router.MiddlewareChain) {
	r.GET(loginURL, base, m.loginHandler)
	r.GET(logoutURL, base, m.logoutHandler)
	r.GET(callbackURL, base, m.callbackHandler)
}

// Warmup prepares local caches.
//
// Implements auth.Warmable.
func (m *AuthMethod) Warmup(ctx context.Context) error {
	cfg, err := m.checkConfigured(ctx)
	if err != nil {
		return err
	}
	doc, err := cfg.discoveryDoc(ctx)
	if err != nil {
		return err
	}
	if _, err := doc.SigningKeys(ctx); err != nil {
		return err
	}
	_ = auth.GetAEAD(ctx)
	return nil
}

// Authenticate authenticates the request.
//
// Implements auth.Method.
func (m *AuthMethod) Authenticate(ctx context.Context, r *http.Request) (*auth.User, auth.Session, error) {
	encryptedCookie, _ := r.Cookie(internal.SessionCookieName)
	if encryptedCookie == nil {
		return nil, nil, nil // the method is not applicable, skip it
	}

	// Decrypt the cookie to get the session ID. Ignore undecryptable cookies.
	// This may happen if we no longer have the encryption key due to rotations
	// or we changed the cookie format. We just assume such cookies are expired.
	aead := auth.GetAEAD(ctx)
	if aead == nil {
		return nil, nil, errors.Reason("the encryption key is not configured").Err()
	}
	cookie, err := internal.DecryptSessionCookie(aead, encryptedCookie)
	if err != nil {
		logging.Warningf(ctx, "Failed to decrypt the session cookie, ignoring it: %s", err)
		return nil, nil, nil
	}
	sid := session.ID(cookie.SessionId)

	// Load the session to verify it still exists.
	session, err := m.Sessions.FetchSession(ctx, sid)
	switch {
	case err != nil:
		logging.Warningf(ctx, "Failed to fetch session %q: %s", sid, err)
		return nil, nil, errors.Reason("failed to fetch the session").Tag(transient.Tag).Err()
	case session == nil:
		logging.Warningf(ctx, "No session %q in the store, ignoring the session cookie", sid)
		return nil, nil, nil
	case session.State != sessionpb.State_STATE_OPEN:
		logging.Warningf(ctx, "Session %q is in state %q, ignoring the session cookie", sid, session.State)
		return nil, nil, nil
	}

	// Check if we need to refresh the short-lived tokens stored in the session.
	ttl := session.NextRefresh.AsTime().Sub(clock.Now(ctx))
	if internal.ShouldRefreshSession(ctx, ttl) {
		ctx := logging.SetField(ctx, "sid", sid.String())
		if ttl > 0 {
			logging.Infof(ctx, "Refreshing the session, goes stale in %s", ttl)
		} else {
			logging.Infof(ctx, "Refreshing the session, went stale %s ago", -ttl)
		}
		switch session, err = m.refreshSession(ctx, cookie, session); {
		case err != nil:
			logging.Warningf(ctx, "Failed to refresh the session: %s", err)
			return nil, nil, errors.Reason("failed to refresh the session, see server logs").Tag(transient.Tag).Err()
		case session == nil:
			return nil, nil, nil // the session is no longer valid, just ignore it
		default:
			logging.Infof(ctx, "The session was refreshed")
		}
	}

	return &auth.User{
		Identity: identity.Identity("user:" + session.Email),
		Email:    session.Email,
		Name:     session.Name,
		Picture:  session.Picture,
	}, nil, nil
}

// LoginURL returns a URL that, when visited, prompts the user to sign in,
// then redirects the user to the URL specified by dest.
//
// Implements auth.UsersAPI.
func (m *AuthMethod) LoginURL(ctx context.Context, dest string) (string, error) {
	if _, err := m.checkConfigured(ctx); err != nil {
		return "", err
	}
	return makeRedirectURL(loginURL, dest)
}

// LogoutURL returns a URL that, when visited, signs the user out,
// then redirects the user to the URL specified by dest.
//
// Implements auth.UsersAPI.
func (m *AuthMethod) LogoutURL(ctx context.Context, dest string) (string, error) {
	if _, err := m.checkConfigured(ctx); err != nil {
		return "", err
	}
	return makeRedirectURL(logoutURL, dest)
}

////////////////////////////////////////////////////////////////////////////////

var (
	// errCodeReuse is returned if the authorization code is reused.
	errCodeReuse = errors.Reason("the authorization code has already been used").Err()
	// errBadIDToken is returned if the produced ID token is not valid.
	errBadIDToken = errors.Reason("ID token validation error").Err()
	// errSessionClosed is used internally to signal the session is closed.
	errSessionClosed = errors.Reason("the session is already closed").Err()
)

// handler is one of .../login, .../logout or .../callback handlers.
type handler func(ctx context.Context, r *http.Request, rw http.ResponseWriter, cfg *OpenIDConfig, discovery *openid.DiscoveryDoc) error

// checkConfigured verifies the method is configured.
//
// Panics on API violations (i.e. coding errors) and merely returns an error if
// OpenIDConfig callback doesn't produce a valid config.
//
// Returns the resulting OpenIDConfig.
func (m *AuthMethod) checkConfigured(ctx context.Context) (*OpenIDConfig, error) {
	if m.OpenIDConfig == nil {
		panic("bad encryptedcookies.AuthMethod usage: OpenIDConfig is nil")
	}
	if m.Sessions == nil {
		panic("bad encryptedcookies.AuthMethod usage: Sessions is nil")
	}
	cfg, err := m.OpenIDConfig(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "failed to fetch OpenID config").Err()
	}
	switch {
	case cfg.DiscoveryURL == "":
		return nil, errors.Reason("bad OpenID config: no discovery URL").Err()
	case cfg.ClientID == "":
		return nil, errors.Reason("bad OpenID config: no client ID").Err()
	case cfg.ClientSecret == "":
		return nil, errors.Reason("bad OpenID config: no client secret").Err()
	case cfg.RedirectURI == "":
		return nil, errors.Reason("bad OpenID config: no redirect URI").Err()
	}
	return cfg, nil
}

// handler is a common wrapper for routes registered in InstallHandlers.
func (m *AuthMethod) handler(ctx *router.Context, cb handler) {
	cfg, err := m.checkConfigured(ctx.Context)
	if err == nil {
		var discovery *openid.DiscoveryDoc
		discovery, err = cfg.discoveryDoc(ctx.Context)
		if err == nil {
			err = cb(ctx.Context, ctx.Request, ctx.Writer, cfg, discovery)
		}
	}
	if err != nil {
		code := http.StatusBadRequest
		if transient.Tag.In(err) {
			code = http.StatusInternalServerError
		}
		http.Error(ctx.Writer, err.Error(), code)
	}
}

// loginHandler initiates the login flow.
func (m *AuthMethod) loginHandler(ctx *router.Context) {
	m.handler(ctx, func(ctx context.Context, r *http.Request, rw http.ResponseWriter, cfg *OpenIDConfig, discovery *openid.DiscoveryDoc) error {
		dest, err := normalizeURL(r.URL.Query().Get("r"))
		if err != nil {
			return errors.Annotate(err, "bad redirect URI").Err()
		}

		// Generate `state` that will be passed back to us in the callbackHandler.
		state := &encryptedcookiespb.OpenIDState{
			SessionId:    session.GenerateID(),
			Nonce:        internal.GenerateNonce(),
			CodeVerifier: internal.GenerateCodeVerifier(),
			DestHost:     r.Host,
			DestPath:     dest,
		}

		// Encrypt it using service-global AEAD, since we are going to expose it.
		aead := auth.GetAEAD(ctx)
		if aead == nil {
			return errors.Reason("the service encryption key is not configured").Err()
		}
		stateEnc, err := internal.EncryptStateB64(aead, state)
		if err != nil {
			return errors.Annotate(err, "failed to encrypt the state").Err()
		}

		// Prepare parameters for the OpenID Connect authorization endpoint.
		v := url.Values{
			"response_type":         {"code"},
			"scope":                 {"openid email profile"},
			"access_type":           {"offline"}, // want a refresh token
			"prompt":                {"consent"}, // want a NEW refresh token
			"client_id":             {cfg.ClientID},
			"redirect_uri":          {cfg.RedirectURI},
			"nonce":                 {base64.RawURLEncoding.EncodeToString(state.Nonce)},
			"code_challenge":        {internal.DeriveCodeChallenge(state.CodeVerifier)},
			"code_challenge_method": {"S256"},
			"state":                 {stateEnc},
		}

		// Finally, redirect to the OpenID provider's authorization endpoint.
		// It will eventually redirect user's browser to callbackHandler.
		http.Redirect(rw, r, discovery.AuthorizationEndpoint+"?"+v.Encode(), http.StatusFound)
		return nil
	})
}

// logoutHandler closes the session.
func (m *AuthMethod) logoutHandler(ctx *router.Context) {
	m.handler(ctx, func(ctx context.Context, r *http.Request, rw http.ResponseWriter, cfg *OpenIDConfig, discovery *openid.DiscoveryDoc) error {
		dest, err := normalizeURL(r.URL.Query().Get("r"))
		if err != nil {
			return errors.Annotate(err, "bad redirect URI").Err()
		}

		// If we have a cookie, revoke the refresh token and mark the session
		// as closed.
		if encryptedCookie, _ := r.Cookie(internal.SessionCookieName); encryptedCookie != nil {
			aead := auth.GetAEAD(ctx)
			if aead == nil {
				return errors.Reason("the encryption key is not configured").Err()
			}
			if err := m.closeSession(ctx, aead, encryptedCookie, cfg, discovery); err != nil {
				// Refresh tokens are usually observable through ID provider UI. There's
				// a list of "Such and such services have your profile information". The
				// expectation is that if the user finishes the logout flow, the service
				// will disappear from the list of authorized services. To do that, we
				// must successfully revoke the token.
				//
				// Since the logout is used *specifically* to achieve this, it is better
				// to fail it if the revocation fails, unlike the best-effort revocation
				// in /login (where the primary purpose is to get a new refresh token,
				// not to revoke the previous one).
				logging.Errorf(ctx, "An error closing the session: %s", err)
				return errors.Reason("transient error when closing the session").Tag(transient.Tag).Err()
			}
		}

		// Nuke all session cookies to get to a completely clean state.
		removeCookie(rw, r, internal.SessionCookieName)
		for _, name := range m.IncompatibleCookies {
			removeCookie(rw, r, name)
		}
		http.Redirect(rw, r, dest, http.StatusFound)
		return nil
	})
}

// callbackHandler handles a redirect from the OpenID provider.
func (m *AuthMethod) callbackHandler(ctx *router.Context) {
	m.handler(ctx, func(ctx context.Context, r *http.Request, rw http.ResponseWriter, cfg *OpenIDConfig, discovery *openid.DiscoveryDoc) error {
		q := r.URL.Query()

		// This code path is hit when user clicks "Deny" on the consent page or
		// if the OAuth client is misconfigured.
		if errorMsg := q.Get("error"); errorMsg != "" {
			return errors.Reason("%s", errorMsg).Err()
		}

		// On success we must receive the authorization code and the state.
		code := q.Get("code")
		if code == "" {
			return errors.Reason("missing `code` parameter").Err()
		}
		state := q.Get("state")
		if state == "" {
			return errors.Reason("missing `state` parameter").Err()
		}

		// Decrypt/verify `state`.
		aead := auth.GetAEAD(ctx)
		if aead == nil {
			return errors.Reason("the encryption key is not configured").Err()
		}
		statepb, err := internal.DecryptStateB64(aead, state)
		if err != nil {
			logging.Errorf(ctx, "Failed to decrypt the state: %s", err)
			return errors.Reason("bad `state` parameter").Err()
		}

		// The callback URI is hardcoded in the OAuth2 client config and must always
		// point to the default version on GAE. Yet we want to support signing-in
		// into non-default versions that have different hostnames. Do some redirect
		// dance here to pass the control to the required version if necessary
		// (so that it can set the cookie on a non-default version domain). This is
		// safe, since statepb.DestHost comes from an AEAD-encrypted state, and it
		// was produced based on "Host" header in loginHandler, which we assume
		// was verified as belonging to our service by the layer that terminates
		// TLS (e.g. GAE load balancer). Of course, if `Insecure` is true, all bets
		// are off.
		if statepb.DestHost != r.Host {
			// There's no Scheme in r.URL. Append one, otherwise url.String() returns
			// relative (broken) URL. And replace the hostname with desired one.
			url := *r.URL
			if m.Insecure {
				url.Scheme = "http"
			} else {
				url.Scheme = "https"
			}
			url.Host = statepb.DestHost
			http.Redirect(rw, r, url.String(), http.StatusFound)
			return nil
		}

		// Check there is no such session in the store. If there's, `code` has been
		// used already and we should reject this replay attempt.
		switch session, err := m.Sessions.FetchSession(ctx, statepb.SessionId); {
		case err != nil:
			logging.Errorf(ctx, "Failed to check the session: %s", err)
			return errors.Reason("transient error when checking the session").Tag(transient.Tag).Err()
		case session != nil:
			return errCodeReuse
		}

		// Exchange the authorization code for authentication tokens.
		tokens, exp, err := internal.HitTokenEndpoint(ctx, discovery, map[string]string{
			"client_id":     cfg.ClientID,
			"client_secret": cfg.ClientSecret,
			"redirect_uri":  cfg.RedirectURI,
			"grant_type":    "authorization_code",
			"code":          code,
			"code_verifier": statepb.CodeVerifier,
		})
		if err != nil {
			logging.Errorf(ctx, "Code exchange failed: %s", err) // only log on the server
			if transient.Tag.In(err) {
				return errors.Reason("transient error during code exchange").Tag(transient.Tag).Err()
			}
			return errors.Reason("fatal error during code exchange").Err()
		}

		// Verify and unpack the ID token to grab the user info and `nonce` from it.
		tok, _, err := openid.UserFromIDToken(ctx, tokens.IdToken, discovery)
		if err != nil {
			logging.Errorf(ctx, "ID token validation error: %s", err)
			if transient.Tag.In(err) {
				return transient.Tag.Apply(errBadIDToken)
			}
			return errBadIDToken
		}

		// Make sure the token was created via the expected OAuth client and used
		// the expected nonce.
		//
		// The `nonce` check essentially binds `code` to the session ID (which is
		// a true nonce here, with the state stored in the session store). The chain
		// is:
		//    1. `code` is bound to `nonce` per OpenID Connect protocol contract.
		//    2. `nonce` is bound to session ID by the signature on `state`.
		//
		// Note that this is unrelated to `code_verifier` check, which ensures that
		// `code` can't be used by someone who knows `client_secret`, but not
		// `code_verifier`.
		if tok.Aud != cfg.ClientID {
			logging.Errorf(ctx, "Bad ID token: expecting audience %q, got %q", cfg.ClientID, tok.Aud)
			return errBadIDToken
		}
		if tok.Nonce != base64.RawURLEncoding.EncodeToString(statepb.Nonce) {
			logging.Errorf(ctx, "Bad ID token: wrong nonce")
			return errBadIDToken
		}

		// Make sure we've got other required tokens.
		if tokens.AccessToken == "" {
			return errors.Reason("the ID provider didn't produce access token").Err()
		}
		if tokens.RefreshToken == "" {
			return errors.Reason("the ID provider didn't produce refresh token").Err()
		}

		// Everything looks good and we can open the session!

		// Generate per-session encryption keys, put them into the future cookie.
		cookie, sessionAEAD := internal.NewSessionCookie(statepb.SessionId)

		// Encrypt sensitive session tokens using the per-session keys.
		encryptedPrivate, err := internal.EncryptPrivate(sessionAEAD, tokens)
		if err != nil {
			logging.Errorf(ctx, "EncryptPrivate error: %s", err)
			return errors.Reason("failed to prepare the session").Err()
		}

		// Prep the session we are about to store.
		now := timestamppb.Now()
		session := &sessionpb.Session{
			State:            sessionpb.State_STATE_OPEN,
			Generation:       1,
			Created:          now,
			LastRefresh:      now,
			NextRefresh:      timestamppb.New(exp), // refresh when short-lived tokens expire
			Sub:              tok.Sub,
			Email:            tok.Email,
			Name:             tok.Name,
			Picture:          tok.Picture,
			EncryptedPrivate: encryptedPrivate,
		}

		// Actually create the new session in the store.
		err = m.Sessions.UpdateSession(ctx, statepb.SessionId, func(s *sessionpb.Session) error {
			if s.State != sessionpb.State_STATE_UNDEFINED {
				// We might be on a second try of a transaction that "failed"
				// transiently, but actually succeeded. If so, we may have stored
				// `session` already. Note that session.EncryptedPrivate is derived
				// using a random key generated just above in NewSessionCookie and not
				// exposed anywhere. There's a *very* small chance someone else managed
				// to create this session already with the exact same EncryptedPrivate.
				if proto.Equal(s, session) {
					return nil
				}
				return errCodeReuse
			}
			proto.Reset(s)
			proto.Merge(s, session)
			return nil
		})
		if err != nil {
			if err == errCodeReuse {
				return err
			}
			logging.Errorf(ctx, "Failure when storing the session: %s", err)
			return errors.Reason("failed to store the session").Tag(transient.Tag).Err()
		}

		// Best effort at properly closing the previous session. We are going to
		// override the cookie anyway.
		if existingCookie, _ := r.Cookie(internal.SessionCookieName); existingCookie != nil {
			logging.Infof(ctx, "Closing the previous session")
			if err := m.closeSession(ctx, aead, existingCookie, cfg, discovery); err != nil {
				logging.Warningf(ctx, "An error closing the previous session, ignoring: %s", err)
			}
		}

		// Encrypt the session cookie with the *global* AEAD key and set it.
		httpCookie, err := internal.EncryptSessionCookie(aead, cookie)
		if err != nil {
			logging.Errorf(ctx, "Cookie encryption error: %s", err)
			return errors.Reason("failed to prepare the cookie").Err()
		}
		httpCookie.Secure = !m.Insecure
		http.SetCookie(rw, httpCookie)
		for _, name := range m.IncompatibleCookies {
			removeCookie(rw, r, name)
		}

		// Finally redirect the user to the originally requested destination.
		http.Redirect(rw, r, statepb.DestPath, http.StatusFound)
		return nil
	})
}

// refreshSession refreshes the short-lived tokens stored in the session, thus
// checking that the refresh token (also stored there) is still valid.
//
// Returns:
//   (new session, nil) - if the session was successfully refreshed.
//   (nil, nil) - if the refresh token was revoked and/or the session is closed.
//   (nil, err) - if there was some unexpected error refreshing the session.
//
// Note that errors may contain sensitive details and should not be returned to
// the caller as is.
func (m *AuthMethod) refreshSession(ctx context.Context, cookie *encryptedcookiespb.SessionCookie, session *sessionpb.Session) (*sessionpb.Session, error) {
	// Need the discovery doc to hit the OpenID provider's endpoint.
	cfg, err := m.checkConfigured(ctx)
	if err != nil {
		return nil, err
	}
	discovery, err := cfg.discoveryDoc(ctx)
	if err != nil {
		return nil, err
	}

	// Unseal the private part of the session to get the refresh token.
	private, sessionAEAD, err := internal.UnsealPrivate(cookie, session)
	if err != nil {
		return nil, errors.Annotate(err, "failed to unseal the session").Err()
	}

	// Use the refresh token to get the new access and ID tokens. A fatal error
	// here means the refresh token is no longer valid.
	tokens, exp, err := internal.HitTokenEndpoint(ctx, discovery, map[string]string{
		"client_id":     cfg.ClientID,
		"client_secret": cfg.ClientSecret,
		"redirect_uri":  cfg.RedirectURI,
		"grant_type":    "refresh_token",
		"refresh_token": private.RefreshToken,
	})
	if err != nil {
		if transient.Tag.In(err) {
			return nil, errors.Annotate(err, "transient error when fetching new tokens").Err()
		}
		logging.Warningf(ctx, "Refresh failed, closing the session: %s", err)
	}
	stillGood := err == nil

	var tok *openid.IDToken
	var encryptedPrivate []byte
	if stillGood {
		// Grab the updated user info from the ID token.
		if tok, _, err = openid.UserFromIDToken(ctx, tokens.IdToken, discovery); err != nil {
			return nil, errors.Annotate(err, "failed to check the ID token").Err()
		}
		// Make sure we've also got the access token
		if tokens.AccessToken == "" {
			return nil, errors.Reason("the ID provider didn't produce an access token").Err()
		}
		// Reencrypt new tokens using the per-session key. The refresh token stays
		// the same.
		tokens.RefreshToken = private.RefreshToken
		if encryptedPrivate, err = internal.EncryptPrivate(sessionAEAD, tokens); err != nil {
			return nil, errors.Annotate(err, "failed to encrypt the private part of the session").Err()
		}
	}

	// Here we either successfully refreshed the session or the ID provider
	// rejected the refresh token. Either way, update the session state in
	// the storage.
	var refreshedSession *sessionpb.Session
	err = m.Sessions.UpdateSession(ctx, cookie.SessionId, func(s *sessionpb.Session) error {
		bumpGeneration(ctx, s, session.Generation)
		if s.State != sessionpb.State_STATE_OPEN {
			return errSessionClosed
		}
		s.LastRefresh = timestamppb.New(clock.Now(ctx))
		if stillGood {
			s.NextRefresh = timestamppb.New(exp)
			s.Sub = tok.Sub
			s.Email = tok.Email
			s.Name = tok.Name
			s.Picture = tok.Picture
			s.EncryptedPrivate = encryptedPrivate
		} else {
			s.State = sessionpb.State_STATE_REVOKED
			s.NextRefresh = nil
			s.Closed = timestamppb.New(clock.Now(ctx))
			s.EncryptedPrivate = nil
		}
		refreshedSession = s
		return nil
	})
	if err != nil {
		if err == errSessionClosed {
			return nil, nil
		}
		return nil, errors.Annotate(err, "failed to update the session in the storage").Err()
	}

	if stillGood {
		return refreshedSession, nil
	}
	return nil, nil
}

// closeSession closes the session and revokes the refresh token.
//
// Does nothing if the session is already closed or the cookie can't be
// decrypted.
//
// Note that errors may contain sensitive details and should not be returned to
// the caller as is.
func (m *AuthMethod) closeSession(ctx context.Context, aead tink.AEAD, encryptedCookie *http.Cookie, cfg *OpenIDConfig, discovery *openid.DiscoveryDoc) error {
	cookie, err := internal.DecryptSessionCookie(aead, encryptedCookie)
	if err != nil {
		logging.Warningf(ctx, "Failed to decrypt the session cookie, ignoring it: %s", err)
		return nil
	}
	sid := session.ID(cookie.SessionId)

	session, err := m.Sessions.FetchSession(ctx, sid)
	switch {
	case err != nil:
		return errors.Annotate(err, "failed to fetch the session").Tag(transient.Tag).Err()
	case session == nil || session.State != sessionpb.State_STATE_OPEN:
		logging.Infof(ctx, "The session is already closed")
		return nil
	}

	// Unseal the private part of the session to get the refresh token to revoke
	// it. Ignore encryption errors (they may happen if keys were replaced or
	// there was a bug in the code). We should delete the session even if we can't
	// get to its refresh token. Old refresh tokens are naturally evicted when
	// new tokens are minted.
	private, _, err := internal.UnsealPrivate(cookie, session)
	if err != nil {
		logging.Warningf(ctx, "Failed to unseal the private part of the session, ignoring the refresh token: %s", err)
	} else {
		if err := m.revokeRefreshToken(ctx, private.RefreshToken, cfg, discovery); err != nil {
			return err
		}
	}

	// Mark the session as closed in the storage.
	err = m.Sessions.UpdateSession(ctx, sid, func(s *sessionpb.Session) error {
		bumpGeneration(ctx, s, session.Generation)
		if s.State != sessionpb.State_STATE_OPEN {
			return errSessionClosed
		}
		s.State = sessionpb.State_STATE_CLOSED
		s.NextRefresh = nil
		s.Closed = timestamppb.New(clock.Now(ctx))
		s.EncryptedPrivate = nil
		return nil
	})
	if err != nil {
		if err == errSessionClosed {
			logging.Infof(ctx, "The session is already closed")
			return nil
		}
		return errors.Annotate(err, "failed to update the session in the storage").Err()
	}

	return nil
}

// revokeRefreshToken sends a request to the provider's revocation endpoint.
func (m *AuthMethod) revokeRefreshToken(ctx context.Context, token string, cfg *OpenIDConfig, discovery *openid.DiscoveryDoc) error {
	return internal.HitRevocationEndpoint(ctx, discovery, map[string]string{
		"client_id":       cfg.ClientID,
		"client_secret":   cfg.ClientSecret,
		"token":           token,
		"token_type_hint": "refresh_token",
	})
}

////////////////////////////////////////////////////////////////////////////////

// bumpGeneration bumps Generation counter in the session.
//
// It is used to detect race conditions between session update transactions.
// They should presumably be harmless, but some logging won't hurt.
func bumpGeneration(ctx context.Context, s *sessionpb.Session, expected int32) {
	if s.Generation != expected {
		logging.Warningf(ctx,
			"The session was already updated by another handler (gen %d != gen %d). Overwriting...",
			s.Generation, expected)
	}
	s.Generation++
}

// normalizeURL verifies URL is parsable and that it is relative.
func normalizeURL(dest string) (string, error) {
	if dest == "" {
		return "/", nil
	}
	u, err := url.Parse(dest)
	if err != nil {
		return "", errors.Annotate(err, "bad destination URL %q", dest).Err()
	}
	// Note: '//host/path' is a location on a server named 'host'.
	if u.IsAbs() || !strings.HasPrefix(u.Path, "/") || strings.HasPrefix(u.Path, "//") {
		return "", errors.Reason("bad absolute destination URL %q", u).Err()
	}
	// path.Clean removes trailing slash. It matters for URLs though. Keep it.
	keepSlash := strings.HasSuffix(u.Path, "/")
	u.Path = path.Clean(u.Path)
	if !strings.HasSuffix(u.Path, "/") && keepSlash {
		u.Path += "/"
	}
	if !strings.HasPrefix(u.Path, "/") {
		return "", errors.Reason("bad destination URL %q", u).Err()
	}
	return u.String(), nil
}

// makeRedirectURL is used to generate login and logout URLs.
func makeRedirectURL(base, dest string) (string, error) {
	dest, err := normalizeURL(dest)
	if err != nil {
		return "", err
	}
	if dest == "/" {
		return base, nil
	}
	v := url.Values{"r": {dest}}
	return base + "?" + v.Encode(), nil
}

// removeCookie sets a cookie to a past expiration date so that the browser can
// remove it.
//
// It also replaces the value with junk, in case the browser decides to ignore
// the expiration time.
func removeCookie(rw http.ResponseWriter, r *http.Request, cookie string) {
	if prev, err := r.Cookie(cookie); err == nil {
		cpy := *prev
		cpy.Value = "deleted"
		cpy.Path = "/"
		cpy.MaxAge = -1
		cpy.Expires = time.Unix(1, 0)
		http.SetCookie(rw, &cpy)
	}
}