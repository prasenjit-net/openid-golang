// Package main is an interactive OIDC test client demonstrating dynamic registration,
// authorization code flow (with PKCE), implicit flow, userinfo, introspection, and revocation.
// Run: go run examples/test-client.go
// Then open: http://localhost:9090
package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ─────────────────────────────────────────────
// State (single-user in-memory session)
// ─────────────────────────────────────────────

type appState struct {
	mu sync.Mutex

	// Server
	Issuer string

	// Discovery
	DiscoveryDoc map[string]interface{}

	// Registered client
	ClientID     string
	ClientSecret string
	RedirectURI  string

	// Authorization Code flow
	CodeVerifier   string // PKCE
	CodeChallenge  string
	StateParam     string
	AuthCode       string
	AuthCodeRawURL string // full URL that was built

	// Tokens (auth code flow)
	AccessToken  string
	RefreshToken string
	IDToken      string
	TokenExpiry  int

	// Implicit flow
	ImplicitStateParam string
	ImplicitIDToken    string
	ImplicitIDRawURL   string

	// UserInfo response
	UserInfoJSON string

	// Introspection response
	IntrospectJSON string

	// Revocation result
	RevocationDone bool

	// Log of API calls this session
	Log []logEntry
}

type logEntry struct {
	Timestamp string
	Label     string
	Request   string
	Response  string
	IsError   bool
}

var state = &appState{
	Issuer:      "http://localhost:8080",
	RedirectURI: "http://localhost:9090/callback",
}

// ─────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────

func randomBase64URL(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func s256(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func (s *appState) addLog(label, request, response string, isError bool) {
	s.Log = append(s.Log, logEntry{
		Timestamp: time.Now().Format("15:04:05"),
		Label:     label,
		Request:   request,
		Response:  response,
		IsError:   isError,
	})
}

func apiPost(rawURL, contentType string, body io.Reader, authHeader string) (int, []byte, error) {
	req, err := http.NewRequest("POST", rawURL, body)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", contentType)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, b, nil
}

func apiGet(rawURL, authHeader string) (int, []byte, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return 0, nil, err
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, b, nil
}

func prettyJSON(raw []byte) string {
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		return string(raw)
	}
	return buf.String()
}

// ─────────────────────────────────────────────
// HTML rendering helpers
// ─────────────────────────────────────────────

const pageCSS = `
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
       background: #f0f2f5; color: #1a1a2e; }
.topbar { background: linear-gradient(135deg,#1a1a2e,#16213e);
          color:#fff; padding:16px 32px; display:flex; align-items:center; gap:12px; }
.topbar h1 { font-size:1.2rem; }
.topbar .badge { background:#e94560; border-radius:999px;
                 font-size:.7rem; padding:2px 8px; font-weight:700; }
.container { max-width:960px; margin:32px auto; padding:0 16px; }
.steps { display:flex; flex-wrap:wrap; gap:8px; margin-bottom:32px; }
.step-pill { padding:6px 14px; border-radius:999px; font-size:.8rem; font-weight:600;
             border:2px solid transparent; cursor:default; }
.step-pill.done   { background:#d4edda; color:#155724; border-color:#c3e6cb; }
.step-pill.active { background:#cce5ff; color:#004085; border-color:#b8daff; }
.step-pill.todo   { background:#e2e3e5; color:#6c757d; border-color:#d6d8db; }
.card { background:#fff; border-radius:12px; padding:28px;
        box-shadow:0 2px 12px rgba(0,0,0,.08); margin-bottom:24px; }
.card h2 { font-size:1.1rem; margin-bottom:8px; color:#16213e; }
.card .subtitle { color:#6c757d; font-size:.9rem; margin-bottom:20px; line-height:1.5; }
.explain { background:#f8f9fa; border-left:4px solid #007bff;
           padding:12px 16px; border-radius:0 8px 8px 0;
           font-size:.88rem; line-height:1.6; margin-bottom:20px; color:#333; }
.explain strong { color:#0056b3; }
pre { background:#1e1e1e; color:#d4d4d4; padding:16px; border-radius:8px;
      font-size:.8rem; overflow-x:auto; white-space:pre-wrap; word-break:break-all;
      line-height:1.5; margin:8px 0; }
.btn { display:inline-block; padding:10px 24px; border-radius:8px; font-size:.9rem;
       font-weight:600; text-decoration:none; cursor:pointer; border:none;
       transition:opacity .15s; }
.btn:hover { opacity:.85; }
.btn-primary   { background:#007bff; color:#fff; }
.btn-success   { background:#28a745; color:#fff; }
.btn-warning   { background:#ffc107; color:#212529; }
.btn-danger    { background:#dc3545; color:#fff; }
.btn-secondary { background:#6c757d; color:#fff; }
.btn-info      { background:#17a2b8; color:#fff; }
.row { display:flex; gap:16px; flex-wrap:wrap; }
.col { flex:1; min-width:280px; }
.kv { display:flex; gap:8px; align-items:flex-start; margin:6px 0; flex-wrap:wrap; }
.kv .key { font-weight:600; font-size:.82rem; color:#495057;
           background:#e9ecef; padding:2px 8px; border-radius:4px;
           white-space:nowrap; }
.kv .val { font-size:.82rem; color:#212529; font-family:monospace;
           word-break:break-all; }
.tag-success { color:#155724; background:#d4edda; border-radius:4px;
               padding:2px 8px; font-size:.8rem; font-weight:700; }
.tag-error   { color:#721c24; background:#f8d7da; border-radius:4px;
               padding:2px 8px; font-size:.8rem; font-weight:700; }
.log-entry { border:1px solid #dee2e6; border-radius:8px;
             margin-bottom:12px; overflow:hidden; }
.log-header { background:#f8f9fa; padding:8px 14px; font-size:.83rem;
              font-weight:600; display:flex; gap:10px; align-items:center; }
.log-body   { padding:10px 14px; }
.divider    { border:none; border-top:1px solid #dee2e6; margin:20px 0; }
.notice     { background:#fff3cd; border:1px solid #ffc107; border-radius:8px;
              padding:12px 16px; font-size:.88rem; color:#856404; margin-bottom:16px; }
</style>`

func pageWrap(title, body string) string {
	s.mu.Lock()
	steps := buildStepPills()
	s.mu.Unlock()
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>%s — OIDC Demo</title>%s</head>
<body>
<div class="topbar">
  <div>
    <h1>🔐 OIDC Interactive Demo</h1>
    <div style="font-size:.78rem;opacity:.7;margin-top:2px">Server: %s</div>
  </div>
  <span class="badge">v2</span>
  <div style="margin-left:auto;display:flex;gap:8px;">
    <a href="/" class="btn btn-secondary" style="padding:6px 14px;font-size:.8rem;">🏠 Home</a>
    <a href="/reset" class="btn btn-danger" style="padding:6px 14px;font-size:.8rem;" onclick="return confirm('Reset all state?')">↺ Reset</a>
  </div>
</div>
<div class="container">
  <div class="steps">%s</div>
  %s
</div>
</body></html>`, html.EscapeString(title), pageCSS, state.Issuer, steps, body)
}

var s = state // alias for lock use in pageWrap

type stepDef struct{ id, label string }

var stepDefs = []stepDef{
	{"discovery", "1. Discovery"},
	{"register", "2. Registration"},
	{"auth-code", "3. Auth Code"},
	{"token", "4. Token Exchange"},
	{"userinfo", "5. UserInfo"},
	{"introspect", "6. Introspect"},
	{"implicit", "7. Implicit Flow"},
	{"revoke", "8. Revoke"},
}

func buildStepPills() string {
	done := map[string]bool{
		"discovery":  state.DiscoveryDoc != nil,
		"register":   state.ClientID != "",
		"auth-code":  state.AuthCode != "",
		"token":      state.AccessToken != "",
		"userinfo":   state.UserInfoJSON != "",
		"introspect": state.IntrospectJSON != "",
		"implicit":   state.ImplicitIDToken != "",
		"revoke":     state.RevocationDone,
	}
	var b strings.Builder
	for _, st := range stepDefs {
		cls := "todo"
		if done[st.id] {
			cls = "done"
		}
		fmt.Fprintf(&b, `<span class="step-pill %s">%s</span>`, cls, st.label)
	}
	return b.String()
}

func kv(key, val string) string {
	return fmt.Sprintf(`<div class="kv"><span class="key">%s</span><span class="val">%s</span></div>`,
		html.EscapeString(key), html.EscapeString(val))
}

func card(title, subtitle, body string) string {
	return fmt.Sprintf(`<div class="card"><h2>%s</h2><p class="subtitle">%s</p>%s</div>`,
		html.EscapeString(title), subtitle, body)
}

func explain(text string) string {
	return fmt.Sprintf(`<div class="explain">%s</div>`, text)
}

func codeBlock(text string) string {
	return fmt.Sprintf("<pre>%s</pre>", html.EscapeString(text))
}

// ─────────────────────────────────────────────
// Handlers
// ─────────────────────────────────────────────

func handleHome(w http.ResponseWriter, _ *http.Request) {
	body := `
<div class="card">
  <h2>Welcome to the OIDC Interactive Demo Client</h2>
  <p class="subtitle">This tool walks you through every step of OpenID Connect, one interaction at a time.
  Each step shows you the exact HTTP request, the server response, and explains <em>why</em> it matters.</p>

  <div class="explain">
    <strong>What is OpenID Connect?</strong><br>
    OpenID Connect (OIDC) is an identity layer on top of OAuth 2.0. It lets clients verify
    the identity of an end-user and obtain basic profile information using standard HTTP flows.
    <br><br>
    <strong>What you will explore here:</strong>
    <ol style="margin-left:18px;margin-top:6px;line-height:2">
      <li><strong>Discovery</strong> — Fetch the server's capabilities from the well-known endpoint</li>
      <li><strong>Dynamic Registration</strong> — Register this client with the server at runtime (RFC 7591)</li>
      <li><strong>Authorization Code Flow</strong> — The most secure OAuth 2.0 flow, with PKCE</li>
      <li><strong>Token Exchange</strong> — Trade the authorization code for access/ID/refresh tokens</li>
      <li><strong>UserInfo Endpoint</strong> — Fetch the authenticated user's profile</li>
      <li><strong>Token Introspection</strong> — Ask the server if a token is still valid (RFC 7662)</li>
      <li><strong>Implicit Flow</strong> — Legacy browser flow where tokens arrive in the URL fragment</li>
      <li><strong>Token Revocation</strong> — Invalidate a token immediately (RFC 7009)</li>
    </ol>
  </div>

  <div class="notice">
    ℹ️ Make sure the OIDC server is running at <strong>http://localhost:8080</strong> before you begin.
    Default test credentials: <strong>testuser / password123</strong>
  </div>

  <a href="/step/discovery" class="btn btn-primary">Begin → Step 1: Discovery</a>
</div>`
	_, _ = fmt.Fprint(w, pageWrap("Home", body))
}

func handleDiscovery(w http.ResponseWriter, _ *http.Request) {
	issuer := state.Issuer
	discoveryURL := issuer + "/.well-known/openid-configuration"

	status, body, err := apiGet(discoveryURL, "")
	var respStr string
	if err != nil {
		respStr = "ERROR: " + err.Error()
		state.mu.Lock()
		state.addLog("Discovery", "GET "+discoveryURL, respStr, true)
		state.mu.Unlock()
	} else {
		respStr = prettyJSON(body)
		var doc map[string]interface{}
		if json.Unmarshal(body, &doc) == nil {
			state.mu.Lock()
			state.DiscoveryDoc = doc
			state.addLog("Discovery", "GET "+discoveryURL, fmt.Sprintf("HTTP %d\n%s", status, respStr), false)
			state.mu.Unlock()
		}
	}

	content := fmt.Sprintf(`
%s
<div class="row">
  <div class="col">
    %s
    %s
    %s
  </div>
  <div class="col">
    <div class="card">
      <h2>Server Response</h2>
      <p class="subtitle">HTTP %d — Discovery Document</p>
      %s
      <hr class="divider">
      <a href="/step/register" class="btn btn-success">Next → Step 2: Dynamic Registration</a>
    </div>
  </div>
</div>`,
		card("Step 1: OpenID Connect Discovery",
			"We ask the server what it supports before doing anything else.",
			explain(`<strong>Why Discovery?</strong><br>
Instead of hard-coding endpoint URLs, OIDC clients fetch the <em>Discovery Document</em>
from <code>/.well-known/openid-configuration</code>. This document advertises every endpoint
(authorization, token, userinfo, …), supported algorithms, scopes, and features.
Clients can adapt automatically if the server configuration changes.`),
		),
		card("Request", "", fmt.Sprintf("%s\n%s",
			explain(`<strong>Method:</strong> GET (no authentication needed — it's public)<br>
<strong>URL:</strong> <code>%s</code>`+discoveryURL+`</code>`),
			codeBlock("GET "+discoveryURL),
		)),
		"",
		"",
		status,
		codeBlock(respStr),
	)
	_, _ = fmt.Fprint(w, pageWrap("Discovery", content))
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	issuer := state.Issuer

	// Discover registration endpoint
	regEndpoint := issuer + "/register"
	state.mu.Lock()
	if state.DiscoveryDoc != nil {
		if ep, ok := state.DiscoveryDoc["registration_endpoint"].(string); ok && ep != "" {
			regEndpoint = ep
		}
	}
	state.mu.Unlock()

	reqBody := map[string]interface{}{
		"client_name":                "OIDC Demo Client (dynamic)",
		"redirect_uris":              []string{"http://localhost:9090/callback"},
		"grant_types":                []string{"authorization_code", "implicit"},
		"response_types":             []string{"code", "id_token", "id_token token"},
		"scope":                      "openid profile email",
		"token_endpoint_auth_method": "client_secret_basic",
		"application_type":           "web",
		"contacts":                   []string{"demo@example.com"},
	}
	reqJSON, _ := json.MarshalIndent(reqBody, "", "  ")

	status, respBytes, err := apiPost(regEndpoint, "application/json",
		bytes.NewReader(reqJSON), "")

	var result map[string]interface{}
	var errMsg string
	if err != nil {
		errMsg = "HTTP call failed: " + err.Error()
	} else if err2 := json.Unmarshal(respBytes, &result); err2 != nil {
		errMsg = "JSON parse error: " + err2.Error()
	} else if e, ok := result["error"].(string); ok {
		errMsg = e + ": " + fmt.Sprint(result["error_description"])
	} else {
		state.mu.Lock()
		state.ClientID = fmt.Sprint(result["client_id"])
		state.ClientSecret = fmt.Sprint(result["client_secret"])
		state.addLog("Dynamic Registration",
			fmt.Sprintf("POST %s\n%s", regEndpoint, string(reqJSON)),
			fmt.Sprintf("HTTP %d\n%s", status, prettyJSON(respBytes)), false)
		state.mu.Unlock()
	}

	if errMsg != "" {
		state.mu.Lock()
		state.addLog("Dynamic Registration",
			fmt.Sprintf("POST %s\n%s", regEndpoint, string(reqJSON)),
			fmt.Sprintf("HTTP %d — ERROR: %s", status, errMsg), true)
		state.mu.Unlock()
	}

	_ = r
	content := fmt.Sprintf(`
<div class="row">
  <div class="col">
    %s
    %s
  </div>
  <div class="col">
    %s
  </div>
</div>`,
		card("Step 2: Dynamic Client Registration",
			"Register this client with the OIDC server at runtime — no manual config needed.",
			explain(`<strong>RFC 7591 — OAuth 2.0 Dynamic Client Registration</strong><br>
Normally a developer registers a client in a dashboard and gets a <code>client_id</code> and
<code>client_secret</code>. Dynamic Registration lets the client register itself programmatically.<br><br>
We POST a JSON metadata document to the <code>registration_endpoint</code>. The server responds
with the assigned <code>client_id</code>, <code>client_secret</code>, and echoes back the
configuration it accepted.`)+
				func() string {
					if errMsg != "" {
						return fmt.Sprintf(`<div class="tag-error">✗ Registration failed: %s</div>`, html.EscapeString(errMsg))
					}
					return `<div class="tag-success">✓ Registered successfully</div><br>` +
						kv("client_id", state.ClientID) +
						kv("client_secret", state.ClientSecret) +
						`<br><a href="/step/auth-code" class="btn btn-success">Next → Step 3: Authorization Code Flow</a>`
				}(),
		),
		card("Registration Request", "", fmt.Sprintf("%s%s",
			explain(`We send the client's metadata: its name, allowed redirect URIs, grant types, and scopes.
The server validates this and assigns credentials.`),
			codeBlock(fmt.Sprintf("POST %s\nContent-Type: application/json\n\n%s", regEndpoint, string(reqJSON))),
		)),
		card("Server Response", fmt.Sprintf("HTTP %d", status),
			codeBlock(prettyJSON(respBytes))),
	)
	_, _ = fmt.Fprint(w, pageWrap("Dynamic Registration", content))
}

func handleAuthCode(w http.ResponseWriter, _ *http.Request) {
	state.mu.Lock()
	clientID := state.ClientID
	redirectURI := state.RedirectURI
	issuer := state.Issuer

	// PKCE
	verifier := randomBase64URL(64)
	challenge := s256(verifier)
	stateParam := randomBase64URL(16)
	state.CodeVerifier = verifier
	state.CodeChallenge = challenge
	state.StateParam = stateParam

	authEndpoint := issuer + "/authorize"
	if state.DiscoveryDoc != nil {
		if ep, ok := state.DiscoveryDoc["authorization_endpoint"].(string); ok && ep != "" {
			authEndpoint = ep
		}
	}

	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "openid profile email")
	params.Set("state", stateParam)
	params.Set("code_challenge", challenge)
	params.Set("code_challenge_method", "S256")
	params.Set("nonce", randomBase64URL(16))
	authURL := authEndpoint + "?" + params.Encode()
	state.AuthCodeRawURL = authURL
	state.mu.Unlock()

	content := fmt.Sprintf(`
<div class="row">
  <div class="col">
    %s
  </div>
  <div class="col">
    %s
    %s
    <div class="card">
      <h2>👆 Your Turn!</h2>
      <p class="subtitle">Click the button below to redirect to the OIDC server's login page.
      Log in with <strong>testuser / password123</strong>, approve the consent, and you will
      be redirected back here with an authorization code.</p>
      <a href="%s" class="btn btn-primary" style="font-size:1rem;padding:14px 32px;">
        🚀 Start Authorization Code Flow
      </a>
    </div>
  </div>
</div>`,
		card("Step 3: Authorization Code Flow with PKCE",
			"The most secure OAuth 2.0 flow — the browser never touches the tokens directly.",
			explain(`<strong>How it works:</strong><br>
1. We redirect the user's browser to the authorization server's <code>/authorize</code> endpoint<br>
2. The user logs in and consents<br>
3. The server redirects back to our <code>redirect_uri</code> with a short-lived <em>authorization code</em><br>
4. Our server (not the browser) exchanges the code for tokens at the <code>/token</code> endpoint<br><br>
<strong>PKCE (Proof Key for Code Exchange — RFC 7636)</strong><br>
PKCE prevents authorization code interception attacks. We generate a random
<code>code_verifier</code>, hash it to get <code>code_challenge</code> (S256 = SHA-256 + base64url),
and send only the challenge to the server. When exchanging the code, we send the original verifier —
only <em>we</em> can prove we started this flow.`),
		),
		card("PKCE Values (generated just now)", "",
			kv("code_verifier (secret, never sent first)", verifier)+
				kv("code_challenge (SHA-256 of verifier, sent)", challenge)+
				kv("state (CSRF protection)", stateParam),
		),
		card("Authorization URL", "This is the URL the browser will visit",
			codeBlock(authURL),
		),
		authURL,
	)
	_, _ = fmt.Fprint(w, pageWrap("Authorization Code Flow", content))
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	// Implicit flow? The fragment never reaches the server.
	// We check query params for auth code, or render JS for implicit.

	code := r.URL.Query().Get("code")
	errParam := r.URL.Query().Get("error")

	if errParam != "" {
		errDesc := r.URL.Query().Get("error_description")
		body := card("⚠️ Authorization Error",
			"The authorization server returned an error.",
			explain(fmt.Sprintf(`<strong>Error:</strong> %s<br><strong>Description:</strong> %s<br><br>
Common causes: user denied consent, invalid parameters, or expired session.`,
				html.EscapeString(errParam), html.EscapeString(errDesc))),
		)
		_, _ = fmt.Fprint(w, pageWrap("Error", body))
		return
	}

	if code != "" {
		// Authorization Code flow callback
		incomingState := r.URL.Query().Get("state")
		state.mu.Lock()
		state.AuthCode = code
		expectedState := state.StateParam
		state.mu.Unlock()

		stateOK := incomingState == expectedState

		body := fmt.Sprintf(`
<div class="row">
  <div class="col">
    %s
  </div>
  <div class="col">
    %s
    <div class="card">
      <h2>State Validation</h2>
      <p class="subtitle">CSRF protection — the <code>state</code> we sent must match.</p>
      %s
      %s
    </div>
  </div>
</div>`,
			card("✅ Authorization Code Received!",
				"The OIDC server redirected back with a short-lived code.",
				explain(`<strong>What just happened?</strong><br>
The user logged in, granted consent, and the server issued an <em>authorization code</em>.
This code is short-lived (typically 60 seconds) and single-use. It must be exchanged
immediately for tokens at the <code>/token</code> endpoint — and only from your <em>server</em>
(never from the browser).<br><br>
The code itself carries no user data. It is meaningless to anyone who intercepts it
because they would also need the <code>code_verifier</code> (PKCE) and the client secret.`)+
					kv("code", code)+
					kv("state", incomingState)+
					`<br><a href="/step/token" class="btn btn-success">Next → Step 4: Token Exchange</a>`,
			),
			card("Redirect URI Parameters", "", codeBlock(r.URL.RawQuery)),
			func() string {
				if stateOK {
					return `<span class="tag-success">✓ State matches — no CSRF</span>`
				}
				return `<span class="tag-error">✗ State mismatch! Possible CSRF attack.</span>`
			}(),
			"",
		)
		_, _ = fmt.Fprint(w, pageWrap("Code Received", body))
		return
	}

	// Implicit flow — tokens in fragment: render JS page
	body := fmt.Sprintf(`
<div class="card">
  <h2>⚡ Implicit Flow Callback</h2>
  <p class="subtitle">Tokens arrive in the URL <strong>fragment</strong> (after <code>#</code>).
  Fragments are <em>never sent to the server</em> — only JavaScript can read them.</p>
  %s
  <div id="result" style="margin-top:16px;"></div>
  <div id="next" style="margin-top:16px;"></div>
</div>
<script>
(function() {
  var fragment = window.location.hash.substring(1);
  var params = {};
  fragment.split('&').forEach(function(p) {
    var kv = p.split('=');
    params[decodeURIComponent(kv[0])] = decodeURIComponent(kv.slice(1).join('='));
  });
  var div = document.getElementById('result');
  if (params.id_token) {
    div.innerHTML = '<span class="tag-success">✓ id_token received in fragment</span><br><br>'
      + '<div class="kv"><span class="key">id_token</span><span class="val">' + params.id_token + '</span></div>'
      + '<div class="kv"><span class="key">state</span><span class="val">' + (params.state||'') + '</span></div>'
      + (params.access_token ? '<div class="kv"><span class="key">access_token</span><span class="val">' + params.access_token + '</span></div>' : '')
      + (params.token_type ? '<div class="kv"><span class="key">token_type</span><span class="val">' + params.token_type + '</span></div>' : '');

    // Save to server
    var form = document.createElement('form');
    form.method = 'POST';
    form.action = '/implicit-save';
    ['id_token','access_token','state','token_type'].forEach(function(k) {
      if (params[k]) {
        var inp = document.createElement('input');
        inp.type = 'hidden'; inp.name = k; inp.value = params[k];
        form.appendChild(inp);
      }
    });
    document.body.appendChild(form);
    form.submit();
  } else {
    div.innerHTML = '<span class="tag-error">✗ No id_token found in fragment. '
      + 'Fragment was: ' + (fragment||'(empty)') + '</span>';
    document.getElementById('next').innerHTML = '<a href="/step/implicit" class="btn btn-warning">← Try Implicit Flow again</a>';
  }
})();
</script>`,
		explain(`<strong>Security Note:</strong> In the implicit flow the <code>id_token</code> (and optionally
<code>access_token</code>) are embedded directly in the redirect URI fragment. Fragments stay in the
browser — they are <em>never</em> sent to the redirect server. JavaScript must extract them.
This is why the implicit flow is considered <em>less secure</em> and deprecated for new applications
in favour of the authorization code flow with PKCE.`),
	)
	_, _ = fmt.Fprint(w, pageWrap("Implicit Callback", body))
}

func handleImplicitSave(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	idToken := r.FormValue("id_token")
	state.mu.Lock()
	state.ImplicitIDToken = idToken
	state.addLog("Implicit Flow",
		"Browser extracted fragment tokens",
		fmt.Sprintf("id_token received (%d chars)", len(idToken)), false)
	state.mu.Unlock()
	http.Redirect(w, r, "/step/implicit-result", http.StatusSeeOther)
}

func handleImplicitResult(w http.ResponseWriter, _ *http.Request) {
	state.mu.Lock()
	tok := state.ImplicitIDToken
	state.mu.Unlock()

	body := fmt.Sprintf(`
<div class="card">
  <h2>⚡ Implicit Flow — ID Token Received</h2>
  <p class="subtitle">The browser extracted the token from the fragment and posted it here.</p>
  %s
  %s
  <hr class="divider">
  <a href="/step/revoke" class="btn btn-success">Next → Step 8: Token Revocation</a>
  &nbsp;
  <a href="/step/implicit" class="btn btn-secondary">Repeat Implicit Flow</a>
</div>`,
		explain(`The <code>id_token</code> is a signed JWT. You can decode it to inspect the user's identity claims.
The token is self-contained — no server lookup needed to verify it (just check the signature).`),
		codeBlock(tok),
	)
	_, _ = fmt.Fprint(w, pageWrap("Implicit Result", body))
}

func handleToken(w http.ResponseWriter, _ *http.Request) {
	state.mu.Lock()
	code := state.AuthCode
	verifier := state.CodeVerifier
	clientID := state.ClientID
	clientSecret := state.ClientSecret
	redirectURI := state.RedirectURI
	issuer := state.Issuer

	tokenEndpoint := issuer + "/token"
	if state.DiscoveryDoc != nil {
		if ep, ok := state.DiscoveryDoc["token_endpoint"].(string); ok && ep != "" {
			tokenEndpoint = ep
		}
	}
	state.mu.Unlock()

	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", code)
	formData.Set("redirect_uri", redirectURI)
	formData.Set("code_verifier", verifier)

	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(clientID+":"+clientSecret))
	reqStr := fmt.Sprintf("POST %s\nAuthorization: Basic <base64(%s:...)>\nContent-Type: application/x-www-form-urlencoded\n\n%s",
		tokenEndpoint, clientID, formData.Encode())

	status, respBytes, err := apiPost(tokenEndpoint, "application/x-www-form-urlencoded",
		strings.NewReader(formData.Encode()), authHeader)

	var tokenResp map[string]interface{}
	var errMsg string
	if err != nil {
		errMsg = "HTTP error: " + err.Error()
	} else if e := json.Unmarshal(respBytes, &tokenResp); e != nil {
		errMsg = "JSON error: " + e.Error()
	} else if errCode, ok := tokenResp["error"].(string); ok {
		errMsg = errCode + ": " + fmt.Sprint(tokenResp["error_description"])
	} else {
		state.mu.Lock()
		state.AccessToken = fmt.Sprint(tokenResp["access_token"])
		state.IDToken = fmt.Sprint(tokenResp["id_token"])
		if rt, ok := tokenResp["refresh_token"].(string); ok {
			state.RefreshToken = rt
		}
		if exp, ok := tokenResp["expires_in"].(float64); ok {
			state.TokenExpiry = int(exp)
		}
		state.addLog("Token Exchange", reqStr,
			fmt.Sprintf("HTTP %d\n%s", status, prettyJSON(respBytes)), false)
		state.mu.Unlock()
	}

	if errMsg != "" {
		state.mu.Lock()
		state.addLog("Token Exchange", reqStr, fmt.Sprintf("ERROR: %s", errMsg), true)
		state.mu.Unlock()
	}

	state.mu.Lock()
	at := state.AccessToken
	idt := state.IDToken
	state.mu.Unlock()

	content := fmt.Sprintf(`
<div class="row">
  <div class="col">
    %s
  </div>
  <div class="col">
    %s
    %s
  </div>
</div>`,
		card("Step 4: Token Exchange",
			"Exchange the authorization code for real tokens.",
			explain(`<strong>What happens at the token endpoint?</strong><br>
We POST four key pieces of data to <code>/token</code>:<br>
• <code>grant_type=authorization_code</code> — tells the server what we're doing<br>
• <code>code</code> — the authorization code from Step 3<br>
• <code>redirect_uri</code> — must match exactly what was used in the authorize request<br>
• <code>code_verifier</code> — the PKCE secret; the server hashes it and compares to the challenge<br><br>
The server authenticates <em>us</em> (the client) via HTTP Basic auth using <code>client_id:client_secret</code>.
If everything matches, it returns three tokens:<br>
• <strong>access_token</strong>: a bearer token for API calls<br>
• <strong>id_token</strong>: a signed JWT with the user's identity<br>
• <strong>refresh_token</strong>: a long-lived token to get new access tokens`)+
				func() string {
					if errMsg != "" {
						return fmt.Sprintf(`<span class="tag-error">✗ %s</span><br><br>
<a href="/step/auth-code" class="btn btn-warning">← Restart Auth Code Flow</a>`, html.EscapeString(errMsg))
					}
					return `<span class="tag-success">✓ Tokens received!</span><br><br>` +
						kv("access_token (first 40 chars)", at[:min(40, len(at))]+"…") +
						kv("id_token (first 40 chars)", idt[:min(40, len(idt))]+"…") +
						`<br><a href="/step/userinfo" class="btn btn-success">Next → Step 5: UserInfo</a>`
				}(),
		),
		card("Token Request", "",
			codeBlock(reqStr),
		),
		card("Token Response", fmt.Sprintf("HTTP %d", status),
			codeBlock(prettyJSON(respBytes)),
		),
	)
	_, _ = fmt.Fprint(w, pageWrap("Token Exchange", content))
}

func handleUserInfo(w http.ResponseWriter, _ *http.Request) {
	state.mu.Lock()
	at := state.AccessToken
	issuer := state.Issuer
	userInfoEndpoint := issuer + "/userinfo"
	if state.DiscoveryDoc != nil {
		if ep, ok := state.DiscoveryDoc["userinfo_endpoint"].(string); ok && ep != "" {
			userInfoEndpoint = ep
		}
	}
	state.mu.Unlock()

	status, respBytes, err := apiGet(userInfoEndpoint, "Bearer "+at)
	reqStr := fmt.Sprintf("GET %s\nAuthorization: Bearer %s…", userInfoEndpoint, at[:min(20, len(at))])
	if err != nil {
		state.mu.Lock()
		state.addLog("UserInfo", reqStr, "ERROR: "+err.Error(), true)
		state.mu.Unlock()
	} else {
		state.mu.Lock()
		state.UserInfoJSON = prettyJSON(respBytes)
		state.addLog("UserInfo", reqStr, fmt.Sprintf("HTTP %d\n%s", status, state.UserInfoJSON), false)
		state.mu.Unlock()
	}

	state.mu.Lock()
	uiJSON := state.UserInfoJSON
	state.mu.Unlock()

	content := fmt.Sprintf(`
<div class="row">
  <div class="col">
    %s
  </div>
  <div class="col">
    %s
    %s
  </div>
</div>`,
		card("Step 5: UserInfo Endpoint",
			"Fetch the authenticated user's profile using the access token.",
			explain(`<strong>UserInfo Endpoint (OIDC Core §5.3)</strong><br>
After obtaining an <code>access_token</code>, the client can call the <code>/userinfo</code>
endpoint to retrieve profile claims for the authenticated user. The claims returned depend
on the <em>scopes</em> that were granted (e.g., <code>profile</code> gives name/picture,
<code>email</code> gives email address).<br><br>
The access token is sent as a <strong>Bearer token</strong> in the Authorization header.
The server validates the token, looks up the user, and returns the appropriate claims as JSON.
This is an alternative to embedding all claims in the ID token.`)+
				`<span class="tag-success">✓ Profile loaded</span><br><br>`+
				`<a href="/step/introspect" class="btn btn-success">Next → Step 6: Token Introspection</a>`,
		),
		card("Request", "", codeBlock(reqStr)),
		card("UserInfo Response", fmt.Sprintf("HTTP %d", status), codeBlock(uiJSON)),
	)
	_, _ = fmt.Fprint(w, pageWrap("UserInfo", content))
}

func handleIntrospect(w http.ResponseWriter, _ *http.Request) {
	state.mu.Lock()
	at := state.AccessToken
	clientID := state.ClientID
	clientSecret := state.ClientSecret
	issuer := state.Issuer
	introspectEndpoint := issuer + "/introspect"
	if state.DiscoveryDoc != nil {
		if ep, ok := state.DiscoveryDoc["introspection_endpoint"].(string); ok && ep != "" {
			introspectEndpoint = ep
		}
	}
	state.mu.Unlock()

	formData := url.Values{}
	formData.Set("token", at)
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(clientID+":"+clientSecret))
	reqStr := fmt.Sprintf("POST %s\nAuthorization: Basic <client_credentials>\n\ntoken=%s…",
		introspectEndpoint, at[:min(20, len(at))])

	status, respBytes, err := apiPost(introspectEndpoint, "application/x-www-form-urlencoded",
		strings.NewReader(formData.Encode()), authHeader)

	if err != nil {
		state.mu.Lock()
		state.addLog("Introspection", reqStr, "ERROR: "+err.Error(), true)
		state.mu.Unlock()
	} else {
		state.mu.Lock()
		state.IntrospectJSON = prettyJSON(respBytes)
		state.addLog("Introspection", reqStr, fmt.Sprintf("HTTP %d\n%s", status, state.IntrospectJSON), false)
		state.mu.Unlock()
	}

	state.mu.Lock()
	intrJSON := state.IntrospectJSON
	state.mu.Unlock()

	content := fmt.Sprintf(`
<div class="row">
  <div class="col">
    %s
  </div>
  <div class="col">
    %s
    %s
  </div>
</div>`,
		card("Step 6: Token Introspection (RFC 7662)",
			"Ask the server: is this token still valid?",
			explain(`<strong>Why Introspect?</strong><br>
Resource servers (APIs) that receive an access token need to verify it. Rather than parsing
and validating a JWT locally, they can call the <code>/introspect</code> endpoint. The server
checks its database, verifies the token hasn't been revoked or expired, and returns metadata:<br>
• <code>active: true/false</code> — is the token valid right now?<br>
• <code>scope</code>, <code>sub</code>, <code>client_id</code>, <code>exp</code> — token metadata<br><br>
This is essential for <em>opaque tokens</em> (non-JWTs) and for detecting revoked tokens
even if they haven't expired yet.`)+
				`<span class="tag-success">✓ Introspection complete</span><br><br>`+
				`<a href="/step/implicit" class="btn btn-success">Next → Step 7: Implicit Flow</a>`,
		),
		card("Request", "", codeBlock(reqStr)),
		card("Introspection Response", fmt.Sprintf("HTTP %d", status), codeBlock(intrJSON)),
	)
	_, _ = fmt.Fprint(w, pageWrap("Introspection", content))
}

func handleImplicit(w http.ResponseWriter, _ *http.Request) {
	state.mu.Lock()
	clientID := state.ClientID
	redirectURI := state.RedirectURI
	issuer := state.Issuer

	authEndpoint := issuer + "/authorize"
	if state.DiscoveryDoc != nil {
		if ep, ok := state.DiscoveryDoc["authorization_endpoint"].(string); ok && ep != "" {
			authEndpoint = ep
		}
	}

	implicitState := randomBase64URL(16)
	nonce := randomBase64URL(16)
	state.ImplicitStateParam = implicitState
	state.mu.Unlock()

	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "id_token")
	params.Set("scope", "openid profile email")
	params.Set("state", implicitState)
	params.Set("nonce", nonce)
	implicitURL := authEndpoint + "?" + params.Encode()

	state.mu.Lock()
	state.ImplicitIDRawURL = implicitURL
	state.mu.Unlock()

	content := fmt.Sprintf(`
<div class="row">
  <div class="col">
    %s
  </div>
  <div class="col">
    %s
    <div class="card">
      <h2>👆 Your Turn!</h2>
      <p class="subtitle">Click below to start the implicit flow. Log in (or reuse the existing
      session), grant consent, and observe how the <code>id_token</code> lands in the
      URL fragment — <em>never on the server</em>.</p>
      <a href="%s" class="btn btn-warning" style="font-size:1rem;padding:14px 32px;">
        ⚡ Start Implicit Flow
      </a>
    </div>
  </div>
</div>`,
		card("Step 7: Implicit Flow (Legacy)",
			"Tokens delivered directly to the browser — no server-side token exchange.",
			explain(`<strong>Implicit Flow (OIDC Core §3.2)</strong><br>
In the implicit flow, tokens are returned directly in the redirect URI <em>fragment</em>
(the part after <code>#</code>). No separate token exchange step is needed.<br><br>
<strong>response_type=id_token</strong> returns only an ID token.<br>
<strong>response_type=token id_token</strong> returns both an access token and ID token.<br><br>
<strong>⚠️ Deprecation Note:</strong> The implicit flow is <em>no longer recommended</em> by
OAuth 2.0 Security Best Current Practices (RFC 9700). Fragments can leak in browser history,
referrer headers, and logs. Use <strong>authorization code + PKCE</strong> instead.<br><br>
<strong>Nonce (required for implicit)</strong>: A random value we include in the request;
the server must embed it in the ID token. When we receive the token, we verify the nonce
matches — preventing replay attacks.`)+
				kv("response_type", "id_token")+
				kv("nonce", nonce)+
				kv("state", implicitState),
		),
		card("Authorization URL", "", codeBlock(implicitURL)),
		implicitURL,
	)
	_, _ = fmt.Fprint(w, pageWrap("Implicit Flow", content))
}

func handleRevoke(w http.ResponseWriter, _ *http.Request) {
	state.mu.Lock()
	at := state.AccessToken
	clientID := state.ClientID
	clientSecret := state.ClientSecret
	issuer := state.Issuer
	revokeEndpoint := issuer + "/revoke"
	if state.DiscoveryDoc != nil {
		if ep, ok := state.DiscoveryDoc["revocation_endpoint"].(string); ok && ep != "" {
			revokeEndpoint = ep
		}
	}
	state.mu.Unlock()

	formData := url.Values{}
	formData.Set("token", at)
	formData.Set("token_type_hint", "access_token")
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(clientID+":"+clientSecret))
	reqStr := fmt.Sprintf("POST %s\nAuthorization: Basic <client_credentials>\n\ntoken=%s…&token_type_hint=access_token",
		revokeEndpoint, at[:min(20, len(at))])

	status, respBytes, err := apiPost(revokeEndpoint, "application/x-www-form-urlencoded",
		strings.NewReader(formData.Encode()), authHeader)

	var resultMsg string
	var isErr bool
	if err != nil {
		resultMsg = "HTTP error: " + err.Error()
		isErr = true
	} else if status == http.StatusOK {
		state.mu.Lock()
		state.RevocationDone = true
		state.mu.Unlock()
		resultMsg = fmt.Sprintf("HTTP %d — Token revoked successfully.", status)
	} else {
		resultMsg = fmt.Sprintf("HTTP %d — %s", status, string(respBytes))
		isErr = true
	}

	state.mu.Lock()
	state.addLog("Token Revocation", reqStr, resultMsg, isErr)
	state.mu.Unlock()

	content := fmt.Sprintf(`
<div class="row">
  <div class="col">
    %s
  </div>
  <div class="col">
    %s
    %s
  </div>
</div>`,
		card("Step 8: Token Revocation (RFC 7009)",
			"Immediately invalidate an access or refresh token.",
			explain(`<strong>RFC 7009 — OAuth 2.0 Token Revocation</strong><br>
When a user logs out or a client is uninstalled, tokens should be revoked so they cannot
be used again — even if they haven't expired. The client POSTs the token to <code>/revoke</code>,
authenticated with its client credentials. The server invalidates the token immediately.<br><br>
After revocation:<br>
• Calls to <code>/userinfo</code> with this token will return <code>401 Unauthorized</code><br>
• Introspection will return <code>active: false</code><br><br>
Both <strong>access tokens</strong> and <strong>refresh tokens</strong> can be revoked.
Revoking a refresh token may also revoke all associated access tokens.`)+
				func() string {
					if isErr {
						return fmt.Sprintf(`<span class="tag-error">✗ %s</span>`, html.EscapeString(resultMsg))
					}
					return `<span class="tag-success">✓ Token revoked!</span><br><br>` +
						`<p style="font-size:.9rem;color:#555;margin-top:8px;">The access token is now invalid. Try UserInfo again to confirm.</p>` +
						`<br><a href="/step/userinfo" class="btn btn-info" style="margin-right:8px;">↺ Re-test UserInfo (expect 401)</a>` +
						`<a href="/log" class="btn btn-secondary">📋 View Full Session Log</a>`
				}(),
		),
		card("Revocation Request", "", codeBlock(reqStr)),
		card("Response", "", codeBlock(resultMsg)),
	)
	_, _ = fmt.Fprint(w, pageWrap("Token Revocation", content))
}

func handleLog(w http.ResponseWriter, _ *http.Request) {
	state.mu.Lock()
	entries := make([]logEntry, len(state.Log))
	copy(entries, state.Log)
	state.mu.Unlock()

	var logHTML strings.Builder
	if len(entries) == 0 {
		logHTML.WriteString(`<p style="color:#6c757d">No API calls recorded yet. Complete the steps first.</p>`)
	}
	for _, e := range entries {
		cls := "tag-success"
		icon := "✓"
		if e.IsError {
			cls = "tag-error"
			icon = "✗"
		}
		fmt.Fprintf(&logHTML, `
<div class="log-entry">
  <div class="log-header">
    <span class="timestamp" style="color:#6c757d;font-weight:normal">%s</span>
    <span class="%s">%s %s</span>
  </div>
  <div class="log-body">
    <strong style="font-size:.83rem">Request:</strong>
    %s
    <strong style="font-size:.83rem">Response:</strong>
    %s
  </div>
</div>`, e.Timestamp, cls, icon, html.EscapeString(e.Label),
			codeBlock(e.Request), codeBlock(e.Response))
	}

	body := fmt.Sprintf(`<div class="card">
  <h2>📋 Full Session Log</h2>
  <p class="subtitle">Every API call made during this session, in chronological order.</p>
  %s
</div>`, logHTML.String())
	_, _ = fmt.Fprint(w, pageWrap("Session Log", body))
}

func handleReset(w http.ResponseWriter, r *http.Request) {
	state.mu.Lock()
	// Reset fields individually — never overwrite the mutex itself, or
	// assigning a fresh struct would replace the locked mutex and panic on Unlock.
	state.DiscoveryDoc = nil
	state.ClientID = ""
	state.ClientSecret = ""
	state.CodeVerifier = ""
	state.CodeChallenge = ""
	state.StateParam = ""
	state.AuthCode = ""
	state.AuthCodeRawURL = ""
	state.AccessToken = ""
	state.RefreshToken = ""
	state.IDToken = ""
	state.TokenExpiry = 0
	state.ImplicitStateParam = ""
	state.ImplicitIDToken = ""
	state.ImplicitIDRawURL = ""
	state.UserInfoJSON = ""
	state.IntrospectJSON = ""
	state.RevocationDone = false
	state.Log = nil
	state.mu.Unlock()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ─────────────────────────────────────────────
// Main
// ─────────────────────────────────────────────

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/step/discovery", handleDiscovery)
	mux.HandleFunc("/step/register", handleRegister)
	mux.HandleFunc("/step/auth-code", handleAuthCode)
	mux.HandleFunc("/step/token", handleToken)
	mux.HandleFunc("/step/userinfo", handleUserInfo)
	mux.HandleFunc("/step/introspect", handleIntrospect)
	mux.HandleFunc("/step/implicit", handleImplicit)
	mux.HandleFunc("/step/implicit-result", handleImplicitResult)
	mux.HandleFunc("/step/revoke", handleRevoke)
	mux.HandleFunc("/callback", handleCallback)
	mux.HandleFunc("/implicit-save", handleImplicitSave)
	mux.HandleFunc("/log", handleLog)
	mux.HandleFunc("/reset", handleReset)

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  🔐 OIDC Interactive Demo Client")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Open: http://localhost:9090")
	fmt.Println("  Server must be running at: http://localhost:8080")
	fmt.Println("  Test credentials: testuser / password123")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	server := &http.Server{
		Addr:         ":9090",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
