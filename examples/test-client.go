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
:root{
  --c-p:#1B7A5E;--c-pg:#2DC98A;--c-pd:#145f48;
  --c-a:#C47B3B;
  --c-ok:#10B981;--c-warn:#F59E0B;--c-err:#EF4444;--c-info:#06B6D4;
  --bg:#F2F6F4;--bg2:#E8F0EC;--sf:#FFFFFF;--sf2:#EDF4F0;
  --bd:#C8DDD5;--tx:#0D1814;--tx2:#3D524A;--tx3:#7A9A8E;
  --nb:#0D1814;--nt:#DFF0E8;
  --cb:#0A1210;--ct:#7EFFC0;--cc:#4a9070;
  --sh-sm:0 1px 3px rgba(0,0,0,.08);
  --sh:0 4px 16px rgba(0,0,0,.10);
  --sh-lg:0 8px 32px rgba(0,0,0,.14);
  --r-s:6px;--r:10px;--r-l:16px;--r-x:24px;
  --tr:.18s ease;
}
[data-theme="dark"]{
  --bg:#0B0F0D;--bg2:#111710;--sf:#161D1A;--sf2:#1E2820;
  --bd:#2A3830;--tx:#DFF0E8;--tx2:#A0C4B4;--tx3:#5A8070;
  --sh-sm:0 1px 3px rgba(0,0,0,.3);--sh:0 4px 16px rgba(0,0,0,.4);
  --sh-lg:0 8px 32px rgba(0,0,0,.5);
}
*{box-sizing:border-box;margin:0;padding:0;}
html{scroll-behavior:smooth;}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;
  background:var(--bg);color:var(--tx);line-height:1.6;
  transition:background var(--tr),color var(--tr);}
.topbar{background:var(--nb);color:var(--nt);padding:0 24px;height:56px;
  display:flex;align-items:center;gap:14px;position:sticky;top:0;z-index:100;
  border-bottom:1px solid #1e2a24;}
.topbar-brand{display:flex;align-items:center;gap:10px;text-decoration:none;color:var(--nt);}
.brand-name{font-size:1rem;font-weight:700;letter-spacing:.01em;}
.brand-server{font-size:.72rem;opacity:.55;margin-top:1px;}
.topbar-nav{margin-left:auto;display:flex;align-items:center;gap:8px;}
.nav-btn{display:inline-flex;align-items:center;padding:6px 14px;
  border-radius:var(--r-s);font-size:.8rem;font-weight:600;text-decoration:none;
  color:var(--nt);background:rgba(255,255,255,.08);
  border:1px solid rgba(255,255,255,.12);transition:background var(--tr);}
.nav-btn:hover{background:rgba(255,255,255,.16);}
.nav-btn-danger{background:rgba(239,68,68,.15);border-color:rgba(239,68,68,.3);color:#fca5a5;}
.nav-btn-danger:hover{background:rgba(239,68,68,.25);}
.theme-toggle{background:rgba(255,255,255,.08);border:1px solid rgba(255,255,255,.12);
  color:var(--nt);border-radius:var(--r-s);padding:6px 10px;cursor:pointer;
  font-size:.9rem;transition:background var(--tr);}
.theme-toggle:hover{background:rgba(255,255,255,.16);}
.step-track{background:var(--sf);border-bottom:1px solid var(--bd);
  padding:14px 24px;overflow-x:auto;}
.step-pipeline{display:flex;align-items:center;gap:0;
  min-width:max-content;margin:0 auto;max-width:960px;}
.step-node{display:flex;align-items:center;flex-direction:column;gap:4px;}
.step-dot{width:32px;height:32px;border-radius:50%;display:flex;align-items:center;
  justify-content:center;font-size:.78rem;font-weight:700;
  border:2px solid var(--bd);background:var(--sf2);color:var(--tx3);
  transition:all var(--tr);}
.step-node.done .step-dot{background:var(--c-ok);border-color:var(--c-ok);color:#fff;}
.step-node.active .step-dot{background:var(--c-p);border-color:var(--c-pg);color:#fff;
  box-shadow:0 0 0 4px rgba(45,201,138,.25);}
.step-node.todo .step-dot{opacity:.6;}
.step-label{font-size:.68rem;font-weight:600;color:var(--tx3);
  white-space:nowrap;letter-spacing:.01em;}
.step-node.done .step-label{color:var(--c-ok);}
.step-node.active .step-label{color:var(--c-pg);}
.step-conn{flex:1;min-width:20px;height:2px;background:var(--bd);margin-bottom:16px;}
.container{max-width:960px;margin:28px auto;padding:0 16px;}
.row{display:grid;grid-template-columns:1fr 1fr;gap:16px;}
@media(max-width:700px){
  .row{grid-template-columns:1fr;}
  .step-label{display:none;}
}
@keyframes fadeUp{from{opacity:0;transform:translateY(12px)}to{opacity:1;transform:none}}
.card{background:var(--sf);border:1px solid var(--bd);border-radius:var(--r-l);
  padding:24px;box-shadow:var(--sh);margin-bottom:20px;animation:fadeUp .3s both;}
.card:nth-child(1){animation-delay:.05s}
.card:nth-child(2){animation-delay:.1s}
.card:nth-child(3){animation-delay:.15s}
.card:nth-child(4){animation-delay:.2s}
.card h2{font-size:1.05rem;font-weight:700;color:var(--tx);margin-bottom:6px;}
.card .sub{color:var(--tx2);font-size:.875rem;margin-bottom:18px;line-height:1.55;}
.info{background:var(--sf2);border-left:3px solid var(--c-pg);
  border-radius:0 var(--r) var(--r) 0;padding:14px 18px;
  font-size:.87rem;line-height:1.65;margin-bottom:16px;color:var(--tx2);}
.info strong{color:var(--c-pg);}
.info code{background:rgba(45,201,138,.12);color:var(--c-pg);
  padding:1px 5px;border-radius:3px;font-family:monospace;font-size:.85em;}
pre{background:var(--cb);color:var(--ct);padding:16px;border-radius:var(--r);
  font-size:.78rem;overflow-x:auto;white-space:pre-wrap;word-break:break-all;
  line-height:1.55;margin:8px 0;border:1px solid rgba(45,201,138,.1);}
.btn{display:inline-flex;align-items:center;padding:9px 22px;border-radius:var(--r);
  font-size:.875rem;font-weight:600;text-decoration:none;cursor:pointer;border:none;
  transition:transform var(--tr),box-shadow var(--tr),opacity var(--tr);}
.btn:hover{transform:translateY(-1px);box-shadow:var(--sh);}
.btn-primary{background:var(--c-p);color:#fff;}
.btn-success{background:var(--c-ok);color:#fff;}
.btn-warning{background:var(--c-warn);color:#1a1a1a;}
.btn-danger{background:var(--c-err);color:#fff;}
.btn-ghost{background:var(--sf2);color:var(--tx);border:1px solid var(--bd);}
.btn-info{background:var(--c-info);color:#fff;}
.btn-lg{padding:13px 32px;font-size:1rem;}
.tag-ok{color:#065f46;background:#d1fae5;border-radius:999px;
  padding:2px 10px;font-size:.78rem;font-weight:700;}
.tag-err{color:#7f1d1d;background:#fee2e2;border-radius:999px;
  padding:2px 10px;font-size:.78rem;font-weight:700;}
.tag-info{color:#164e63;background:#cffafe;border-radius:999px;
  padding:2px 10px;font-size:.78rem;font-weight:700;}
.tag-warn{color:#78350f;background:#fef3c7;border-radius:999px;
  padding:2px 10px;font-size:.78rem;font-weight:700;}
[data-theme="dark"] .tag-ok{color:#6ee7b7;background:rgba(16,185,129,.15);}
[data-theme="dark"] .tag-err{color:#fca5a5;background:rgba(239,68,68,.15);}
[data-theme="dark"] .tag-info{color:#67e8f9;background:rgba(6,182,212,.15);}
[data-theme="dark"] .tag-warn{color:#fcd34d;background:rgba(245,158,11,.15);}
.kv{display:flex;gap:8px;align-items:flex-start;margin:5px 0;flex-wrap:wrap;}
.kv .k{font-size:.78rem;font-weight:600;font-family:monospace;
  background:rgba(45,201,138,.12);color:var(--c-pg);
  padding:1px 8px;border-radius:var(--r-s);white-space:nowrap;}
.kv .v{font-size:.8rem;color:var(--tx);font-family:monospace;word-break:break-all;}
.step-grid{display:grid;grid-template-columns:1fr 1fr;gap:12px;margin:16px 0;}
@media(max-width:700px){.step-grid{grid-template-columns:1fr;}}
.step-item{background:var(--sf2);border:1px solid var(--bd);border-radius:var(--r);
  padding:14px;display:flex;gap:12px;align-items:flex-start;}
.step-num{width:28px;height:28px;border-radius:50%;background:var(--c-p);color:#fff;
  display:flex;align-items:center;justify-content:center;
  font-size:.78rem;font-weight:700;flex-shrink:0;}
.step-title{font-size:.85rem;font-weight:700;color:var(--tx);margin-bottom:3px;}
.step-desc{font-size:.78rem;color:var(--tx3);}
.notice{background:rgba(245,158,11,.1);border:1px solid rgba(245,158,11,.3);
  border-radius:var(--r);padding:12px 16px;font-size:.87rem;
  color:var(--c-warn);margin-bottom:16px;display:flex;gap:8px;align-items:flex-start;}
.notice-i{flex-shrink:0;}
.log-entry{border:1px solid var(--bd);border-radius:var(--r);
  margin-bottom:12px;overflow:hidden;}
.log-entry.err{border-color:rgba(239,68,68,.3);}
.log-hdr{background:var(--sf2);padding:8px 14px;font-size:.82rem;
  font-weight:600;display:flex;gap:10px;align-items:center;}
.log-entry.err .log-hdr{background:rgba(239,68,68,.06);}
.log-ts{color:var(--tx3);font-weight:400;font-family:monospace;}
.log-body{padding:10px 14px;}
hr{border:none;border-top:1px solid var(--bd);margin:16px 0;}
</style>`

const themeInitScript = "<script>(function(){var t=localStorage.getItem('oidc-theme')||'light';document.documentElement.setAttribute('data-theme',t);})()</script>"

func pageWrap(title, body string) string {
	s.mu.Lock()
	steps := buildStepPipeline()
	issuer := s.Issuer
	s.mu.Unlock()
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" data-theme="light">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>%s — OIDC Demo</title>
%s
%s
</head>
<body>
<header class="topbar">
  <a href="/" class="topbar-brand">
    <svg width="30" height="30" viewBox="0 0 30 30" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M15 2L4 7v8c0 6.5 4.7 12.6 11 14 6.3-1.4 11-7.5 11-14V7L15 2z" stroke="#2DC98A" stroke-width="1.5" fill="none"/><circle cx="15" cy="14" r="3.5" stroke="#2DC98A" stroke-width="1.5" fill="none"/><line x1="15" y1="17.5" x2="15" y2="22" stroke="#2DC98A" stroke-width="1.5" stroke-linecap="round"/></svg>
    <div><div class="brand-name">OIDC Demo</div><div class="brand-server">%s</div></div>
  </a>
  <nav class="topbar-nav">
    <a href="/log" class="nav-btn">Log</a>
    <button class="theme-toggle" onclick="toggleTheme()" id="theme-btn" aria-label="Toggle theme">🌙</button>
    <a href="/reset" class="nav-btn nav-btn-danger" onclick="return confirm('Reset all session state?')">↺ Reset</a>
  </nav>
</header>
<div class="step-track"><div class="step-pipeline">%s</div></div>
<main class="container">%s</main>
<script>
function toggleTheme(){
  var h=document.documentElement,t=h.getAttribute('data-theme')==='dark'?'light':'dark';
  h.setAttribute('data-theme',t);localStorage.setItem('oidc-theme',t);
  document.getElementById('theme-btn').textContent=t==='dark'?'☀️':'🌙';
}
(function(){
  var t=localStorage.getItem('oidc-theme')||'light';
  document.documentElement.setAttribute('data-theme',t);
  var b=document.getElementById('theme-btn');
  if(b)b.textContent=t==='dark'?'☀️':'🌙';
})();
</script>
</body></html>`,
		html.EscapeString(title), themeInitScript, pageCSS,
		html.EscapeString(issuer), steps, body)
}

var s = state // alias for lock use in pageWrap

type stepDef struct{ id, label string }

var stepDefs = []stepDef{
	{"discovery", "Discovery"},
	{"register", "Registration"},
	{"auth-code", "Auth Code"},
	{"token", "Token Exchange"},
	{"userinfo", "UserInfo"},
	{"introspect", "Introspect"},
	{"implicit", "Implicit Flow"},
	{"revoke", "Revoke"},
}

func buildStepPipeline() string {
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
	activeFound := false
	var b strings.Builder
	for i, st := range stepDefs {
		cls := "todo"
		dot := fmt.Sprintf("%d", i+1)
		if done[st.id] {
			cls = "done"
			dot = "✓"
		} else if !activeFound {
			cls = "active"
			activeFound = true
		}
		if i > 0 {
			fmt.Fprintf(&b, `<div class="step-conn"></div>`)
		}
		fmt.Fprintf(&b, `<div class="step-node %s"><div class="step-dot">%s</div><span class="step-label">%s</span></div>`,
			cls, dot, st.label)
	}
	return b.String()
}

func kv(key, val string) string {
	return fmt.Sprintf(`<div class="kv"><span class="k">%s</span><span class="v">%s</span></div>`,
		html.EscapeString(key), html.EscapeString(val))
}

func card(title, subtitle, body string) string {
	sub := ""
	if subtitle != "" {
		sub = fmt.Sprintf(`<p class="sub">%s</p>`, html.EscapeString(subtitle))
	}
	return fmt.Sprintf(`<div class="card"><h2>%s</h2>%s%s</div>`,
		html.EscapeString(title), sub, body)
}

func explain(text string) string {
	return fmt.Sprintf(`<div class="info">%s</div>`, text)
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
  <h2>Welcome to the OIDC Interactive Demo</h2>
  <p class="sub">Walk through every step of OpenID Connect — one interaction at a time.</p>
  <div class="info">
    <strong>What is OpenID Connect?</strong><br>
    OpenID Connect (OIDC) is an identity layer on top of OAuth 2.0. It lets clients verify
    the identity of an end-user and obtain basic profile information using standard HTTP flows.
  </div>
  <div class="step-grid">
    <div class="step-item"><div class="step-num">1</div><div><div class="step-title">Discovery</div><div class="step-desc">Fetch server capabilities from the well-known endpoint.</div></div></div>
    <div class="step-item"><div class="step-num">2</div><div><div class="step-title">Registration</div><div class="step-desc">Register this client dynamically at runtime (RFC 7591).</div></div></div>
    <div class="step-item"><div class="step-num">3</div><div><div class="step-title">Auth Code Flow</div><div class="step-desc">The most secure OAuth 2.0 flow, with PKCE protection.</div></div></div>
    <div class="step-item"><div class="step-num">4</div><div><div class="step-title">Token Exchange</div><div class="step-desc">Trade the authorization code for access/ID/refresh tokens.</div></div></div>
    <div class="step-item"><div class="step-num">5</div><div><div class="step-title">UserInfo</div><div class="step-desc">Fetch the authenticated user's profile claims.</div></div></div>
    <div class="step-item"><div class="step-num">6</div><div><div class="step-title">Introspection</div><div class="step-desc">Ask the server if a token is still valid (RFC 7662).</div></div></div>
    <div class="step-item"><div class="step-num">7</div><div><div class="step-title">Implicit Flow</div><div class="step-desc">Legacy browser flow — tokens arrive in the URL fragment.</div></div></div>
    <div class="step-item"><div class="step-num">8</div><div><div class="step-title">Revocation</div><div class="step-desc">Immediately invalidate a token (RFC 7009).</div></div></div>
  </div>
  <div class="notice"><span class="notice-i">ℹ️</span><span>Make sure the OIDC server is running at <strong>http://localhost:8080</strong> before you begin. Default test credentials: <strong>testuser / password123</strong></span></div>
  <a href="/step/discovery" class="btn btn-primary btn-lg">Begin → Step 1: Discovery</a>
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
      <p class="sub">HTTP %d — Discovery Document</p>
      %s
      <hr>
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
						return fmt.Sprintf(`<div class="tag-err">✗ Registration failed: %s</div>`, html.EscapeString(errMsg))
					}
					return `<div class="tag-ok">✓ Registered successfully</div><br>` +
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
      <p class="sub">Click the button below to redirect to the OIDC server's login page.
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
      <p class="sub">CSRF protection — the <code>state</code> we sent must match.</p>
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
					return `<span class="tag-ok">✓ State matches — no CSRF</span>`
				}
				return `<span class="tag-err">✗ State mismatch! Possible CSRF attack.</span>`
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
  <p class="sub">Tokens arrive in the URL <strong>fragment</strong> (after <code>#</code>).
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
    div.innerHTML = '<span class="tag-ok">✓ id_token received in fragment</span><br><br>'
      + '<div class="kv"><span class="k">id_token</span><span class="v">' + params.id_token + '</span></div>'
      + '<div class="kv"><span class="k">state</span><span class="v">' + (params.state||'') + '</span></div>'
      + (params.access_token ? '<div class="kv"><span class="k">access_token</span><span class="v">' + params.access_token + '</span></div>' : '')
      + (params.token_type ? '<div class="kv"><span class="k">token_type</span><span class="v">' + params.token_type + '</span></div>' : '');

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
    div.innerHTML = '<span class="tag-err">✗ No id_token found in fragment. '
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
  <p class="sub">The browser extracted the token from the fragment and posted it here.</p>
  %s
  %s
  <hr>
  <a href="/step/revoke" class="btn btn-success">Next → Step 8: Token Revocation</a>
  &nbsp;
  <a href="/step/implicit" class="btn btn-ghost">Repeat Implicit Flow</a>
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
						return fmt.Sprintf(`<span class="tag-err">✗ %s</span><br><br>
<a href="/step/auth-code" class="btn btn-warning">← Restart Auth Code Flow</a>`, html.EscapeString(errMsg))
					}
					return `<span class="tag-ok">✓ Tokens received!</span><br><br>` +
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
				`<span class="tag-ok">✓ Profile loaded</span><br><br>`+
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
				`<span class="tag-ok">✓ Introspection complete</span><br><br>`+
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
      <p class="sub">Click below to start the implicit flow. Log in (or reuse the existing
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
						return fmt.Sprintf(`<span class="tag-err">✗ %s</span>`, html.EscapeString(resultMsg))
					}
					return `<span class="tag-ok">✓ Token revoked!</span><br><br>` +
						`<p style="font-size:.875rem;color:var(--tx2);margin-top:8px;">The access token is now invalid. Try UserInfo again to confirm.</p>` +
						`<br><a href="/step/userinfo" class="btn btn-info" style="margin-right:8px;">↺ Re-test UserInfo (expect 401)</a>` +
						`<a href="/log" class="btn btn-ghost">📋 View Full Session Log</a>`
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
		logHTML.WriteString(`<p style="color:var(--tx3);font-style:italic">No API calls recorded yet. Complete the steps first.</p>`)
	}
	for _, e := range entries {
		entryCls, tagCls := "log-entry", "tag-ok"
		icon := "✓"
		if e.IsError {
			entryCls, tagCls = "log-entry err", "tag-err"
			icon = "✗"
		}
		fmt.Fprintf(&logHTML, `
<div class="%s">
  <div class="log-hdr">
    <span class="log-ts">%s</span>
    <span class="%s">%s %s</span>
  </div>
  <div class="log-body">
    <strong style="font-size:.83rem">Request:</strong>
    %s
    <strong style="font-size:.83rem">Response:</strong>
    %s
  </div>
</div>`, entryCls, e.Timestamp, tagCls, icon, html.EscapeString(e.Label),
			codeBlock(e.Request), codeBlock(e.Response))
	}

	body := fmt.Sprintf(`<div class="card">
  <h2>📋 Full Session Log</h2>
  <p class="sub">Every API call made during this session, in chronological order.</p>
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
