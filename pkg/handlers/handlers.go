package handlers

import (
	"embed"
	"html/template"

	"github.com/prasenjit-net/openid-golang/pkg/configstore"
	"github.com/prasenjit-net/openid-golang/pkg/crypto"
	"github.com/prasenjit-net/openid-golang/pkg/session"
	"github.com/prasenjit-net/openid-golang/pkg/storage"
)

// Handlers holds all HTTP handlers
type Handlers struct {
	config         *configstore.ConfigData
	storage        storage.Storage
	jwtManager     *crypto.JWTManager
	sessionManager *session.Manager
	loginTmpl      *template.Template
	consentTmpl    *template.Template
}

// minimal fallback templates used when no embed.FS is provided (e.g. tests).
const fallbackLoginTmpl = `<!DOCTYPE html><html><body>
<form method="POST" action="/login?auth_session={{.AuthSessionID}}">
{{if .ErrorMessage}}<p style="color:red">{{.ErrorMessage}}</p>{{end}}
<input name="username" required><input type="password" name="password" required>
<button type="submit">Sign In</button></form></body></html>`

const fallbackConsentTmpl = `<!DOCTYPE html><html><body>
<form method="POST" action="/consent?auth_session={{.AuthSessionID}}">
<p>{{.ClientName}} requests: {{range .Scopes}}{{.Name}} {{end}}</p>
<button name="consent" value="allow">Allow</button>
<button name="consent" value="deny">Deny</button></form></body></html>`

// NewHandlers creates a new handlers instance.
// templatesFS should contain frontend/login.html and frontend/consent.html.
// Pass an empty embed.FS (or zero value) to use minimal fallback templates (useful in tests).
func NewHandlers(store storage.Storage, jwtManager *crypto.JWTManager, cfg *configstore.ConfigData, sessionMgr *session.Manager, templatesFS embed.FS) *Handlers {
	loginTmpl := parseOrFallback(templatesFS, "frontend/login.html", fallbackLoginTmpl)
	consentTmpl := parseOrFallback(templatesFS, "frontend/consent.html", fallbackConsentTmpl)
	return &Handlers{
		config:         cfg,
		storage:        store,
		jwtManager:     jwtManager,
		sessionManager: sessionMgr,
		loginTmpl:      loginTmpl,
		consentTmpl:    consentTmpl,
	}
}

// parseOrFallback tries to parse the named file from fs; on any error it parses the fallback string.
func parseOrFallback(fsys embed.FS, name, fallback string) *template.Template {
	if tmpl, err := template.ParseFS(fsys, name); err == nil {
		return tmpl
	}
	return template.Must(template.New(name).Parse(fallback))
}

// GetStorage returns the storage instance
func (h *Handlers) GetStorage() storage.Storage {
	return h.storage
}
