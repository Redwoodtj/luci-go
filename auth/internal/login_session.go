// Copyright 2022 The LUCI Authors.
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

//go:build !copybara
// +build !copybara

package internal

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"go.chromium.org/luci/auth/loginsessionspb"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/system/terminal"
	"go.chromium.org/luci/grpc/prpc"
)

var spinnerChars = []rune("⣾⣽⣻⢿⡿⣟⣯⣷")

type loginSessionTokenProvider struct {
	loginSessionsHost string
	clientID          string
	clientSecret      string
	scopes            []string
	transport         http.RoundTripper
	cacheKey          CacheKey
}

func init() {
	NewLoginSessionTokenProvider = func(ctx context.Context, loginSessionsHost, clientID, clientSecret string, scopes []string, transport http.RoundTripper) (TokenProvider, error) {
		return &loginSessionTokenProvider{
			loginSessionsHost: loginSessionsHost,
			clientID:          clientID,
			clientSecret:      clientSecret,
			scopes:            scopes,
			transport:         transport,
			// Reuse the same key as userAuthTokenProvider to share refresh tokens
			// between two methods. They are compatible.
			cacheKey: CacheKey{
				Key:    fmt.Sprintf("user/%s", clientID),
				Scopes: scopes,
			},
		}, nil
	}
}

func (p *loginSessionTokenProvider) RequiresInteraction() bool {
	return true
}

func (p *loginSessionTokenProvider) Lightweight() bool {
	return false
}

func (p *loginSessionTokenProvider) Email() string {
	// We don't know the email before user logs in.
	return UnknownEmail
}

func (p *loginSessionTokenProvider) CacheKey(ctx context.Context) (*CacheKey, error) {
	return &p.cacheKey, nil
}

func (p *loginSessionTokenProvider) MintToken(ctx context.Context, base *Token) (*Token, error) {
	// It is never correct to use this login method on bots or from scripts.
	// Check if stdout is really a terminal a real user can interact with.
	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
		return nil, errors.Reason("interactive login flow requires the stdout to be attached to a terminal").Err()
	}

	// The list of scopes is displayed on the consent page as well, but show it
	// in the terminal too, for clarity.
	fmt.Println("Getting a refresh token with following OAuth scopes:")
	for _, scope := range p.scopes {
		fmt.Printf("  * %s\n", scope)
	}
	fmt.Println()

	// pRPC client to use for interacting with the sessions server.
	httpClient := &http.Client{Transport: p.transport}
	sessions := loginsessionspb.NewLoginSessionsClient(&prpc.Client{
		C:    httpClient,
		Host: p.loginSessionsHost,
	})

	// Generate a code verifier (a random string) and corresponding challenge for
	// PKCE protocol. They are used to make sure only us can exchange the
	// authorization code for tokens (because only we know the code verifier).
	codeVerifier := generateCodeVerifier()
	codeChallenge := deriveCodeChallenge(codeVerifier)

	// Collect some information about the running environment to show to the user
	// so they can understand better what is invoking the login session. Both are
	// best effort and optional (and easily spoofable).
	executable, _ := os.Executable()
	hostname, _ := os.Hostname()

	// Start a new login session that we'll ask the user to complete.
	session, err := sessions.CreateLoginSession(ctx, &loginsessionspb.CreateLoginSessionRequest{
		OauthClientId:          p.clientID,
		OauthScopes:            p.scopes,
		OauthS256CodeChallenge: codeChallenge,
		ExecutableName:         filepath.Base(executable),
		ClientHostname:         hostname,
	})
	if err != nil {
		return nil, errors.Annotate(err, "failed to create the login session").Err()
	}
	fmt.Printf(
		"Visit this link to complete the login flow in the browser. Do not share it with anyone!\n\n%s\n\n",
		session.LoginFlowUrl,
	)

	// Start a goroutine that renders the most recent confirmation code.
	animationCtrl := startAnimation(ctx)
	defer animationCtrl("", 0)
	animationCtrl(session.ConfirmationCode, session.ConfirmationCodeRefresh.AsDuration())

	// Start polling the session until it moves to a finished state.
	sessionID := session.Id
	sessionPassword := session.Password
	confirmationCode := session.ConfirmationCode
	for session.State == loginsessionspb.LoginSession_PENDING {
		// Use the poll interval suggested by the server, unless it is super small.
		sleep := session.PollInterval.AsDuration()
		if sleep < time.Second {
			sleep = time.Second
		}
		if clock.Sleep(ctx, sleep).Incomplete() {
			return nil, ctx.Err()
		}
		session, err = sessions.GetLoginSession(ctx, &loginsessionspb.GetLoginSessionRequest{
			LoginSessionId:       sessionID,
			LoginSessionPassword: sessionPassword,
		})
		if err != nil {
			return nil, errors.Annotate(err, "failed to poll the login session").Err()
		}
		// Send new confirmation code to the loop that renders it.
		if confirmationCode != session.ConfirmationCode {
			confirmationCode = session.ConfirmationCode
			animationCtrl(confirmationCode, session.ConfirmationCodeRefresh.AsDuration())
		}
	}
	animationCtrl("", 0)

	// A session can expire or fail (if the user cancels it).
	if session.State != loginsessionspb.LoginSession_SUCCEEDED {
		fmt.Printf("Login session failed with status %s!\n", session.State)
		if session.OauthError != "" {
			fmt.Printf("OAuth error: %s\n", session.OauthError)
		}
		return nil, errors.Reason("the login flow failed").Err()
	}

	// We've got the authorization code and the redirect URL needed to complete
	// the flow on our side. Note that we need to use codeVerifier here that only
	// we know.
	tok, err := (&oauth2.Config{
		ClientID:     p.clientID,
		ClientSecret: p.clientSecret,
		RedirectURL:  session.OauthRedirectUrl,
		Endpoint:     google.Endpoint,
	}).Exchange(ctx,
		session.OauthAuthorizationCode,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
	if err != nil {
		return nil, err
	}

	// Convert the token to the format we want, parse ID token into claims.
	// The `nonce` claim must match the login session ID per how Login Sessions
	// protocol works. And `aud` claim must match the OAuth client ID per how
	// OAuth protocol works.
	switch processed, claims, err := processProviderReply(ctx, tok, ""); {
	case err != nil:
		return nil, err
	case claims == nil:
		return nil, errors.Reason("the authorization provider didn't return ID token").Err()
	case claims.Nonce != sessionID:
		return nil, errors.Reason("invalid `nonce` in the ID token %q, expecting %q", claims.Nonce, sessionID).Err()
	case claims.Aud != p.clientID:
		return nil, errors.Reason("invalid `aud` in the ID token %q, expecting %q", claims.Aud, p.clientID).Err()
	default:
		return processed, nil
	}
}

func (p *loginSessionTokenProvider) RefreshToken(ctx context.Context, prev, base *Token) (*Token, error) {
	return refreshToken(ctx, prev, base, &oauth2.Config{
		ClientID:     p.clientID,
		ClientSecret: p.clientSecret,
		Endpoint:     google.Endpoint,
	})
}

// generateCodeVerifier generates a random string used as a code_verifier in
// the PKCE protocol.
//
// See https://tools.ietf.org/html/rfc7636.
func generateCodeVerifier() string {
	blob := make([]byte, 50)
	if _, err := rand.Read(blob); err != nil {
		panic(fmt.Sprintf("failed to generate code verifier: %s", err))
	}
	return base64.RawURLEncoding.EncodeToString(blob)
}

// deriveCodeChallenge derives code_challenge from the code_verifier.
func deriveCodeChallenge(codeVerifier string) string {
	codeVerifierS256 := sha256.Sum256([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(codeVerifierS256[:])
}

// startAnimation starts background rendering of the most recent confirmation
// code (plus some cute spinner animation).
//
// Returns a function that can be used to control the animation. Passing it
// a non-empty string would replace the confirmation code. Passing it an empty
// string would stop the animation.
func startAnimation(ctx context.Context) (ctrl func(string, time.Duration)) {
	type codeAndExp struct {
		code string        // the code itself
		exp  time.Time     // moment it expires
		life time.Duration // its lifetime when we got it
	}

	spinCh := make(chan codeAndExp)
	spinDone := false

	spinWG := sync.WaitGroup{}
	spinWG.Add(1)

	fmt.Printf("When asked, use this confirmation code (it refreshes with time):\n\n")
	fmt.Printf("\n\n\n\n") // allocate lines we'll keep overriding in the loop

	// ESC code sequence helpers.
	up := func(lines int) { fmt.Printf("\033[%dA", lines) }
	line := func(msg string, args ...interface{}) { fmt.Printf("\r\033[2K"+msg+"\n", args...) }

	go func() {
		defer spinWG.Done()

		current := codeAndExp{}
		round := 0

	loop:
		for {
			select {
			case code, ok := <-spinCh:
				if !ok {
					break loop
				}
				current = code
			case res := <-clock.After(ctx, 100*time.Millisecond):
				if res.Err != nil {
					break loop
				}
			}

			// Wait until we get the first confirmation code before rendering
			// anything. This should be fast, since we already know it by the time
			// the goroutine starts, so we just wait for local goroutines to
			// "synchronize".
			if current.code == "" {
				continue
			}

			spinner := string(spinnerChars[round%len(spinnerChars)])
			round += 1

			// Calculate a portion of code's lifetime left.
			ratio := float32(time.Until(current.exp).Seconds() / current.life.Seconds())

			// Convert it into a number of progress bar characters to print.
			total := len(current.code)
			filled := int(ratio*float32(total)) + 1
			if filled < 0 {
				filled = 0
			} else if filled > total {
				filled = total
			}

			// Redraw everything.
			up(4)
			line("%s", current.code)
			line("%s%s", strings.Repeat("─", filled), strings.Repeat(" ", total-filled))
			line("")
			line("Waiting for the login flow to complete in the browser %s", spinner)
		}

		// Redraw the last line replacing the spinner with "Done".
		up(1)
		line("Waiting for the login flow to complete in the browser. Done!\n")
	}()

	return func(code string, exp time.Duration) {
		if !spinDone {
			if code == "" {
				spinDone = true
				close(spinCh)
				spinWG.Wait()
			} else {
				spinCh <- codeAndExp{code, time.Now().Add(exp), exp}
			}
		}
	}
}
