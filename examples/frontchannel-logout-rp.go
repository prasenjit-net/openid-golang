// Example RP with Front-Channel Logout Support
//
// This example demonstrates how to implement an OpenID Connect Relying Party
// that supports front-channel logout notifications from the OP.
//
// Usage:
//   go run examples/frontchannel-logout-rp.go
//
// Then visit http://localhost:3000 to test the logout flow.

package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"
)

// Session represents a user session with sid tracking
type Session struct {
	ID        string
	UserID    string
	Sid       string    // Session ID from ID token (for logout)
	CreatedAt time.Time
}

// SessionStore manages user sessions
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session // sessionID -> Session
}

// NewSessionStore creates a new session store
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*Session),
	}
}

// Create creates a new session
func (s *SessionStore) Create(userID, sid string) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := &Session{
		ID:        generateID(),
		UserID:    userID,
		Sid:       sid,
		CreatedAt: time.Now(),
	}
	s.sessions[session.ID] = session
	return session
}

// FindBySid finds all sessions with the given sid
func (s *SessionStore) FindBySid(sid string) []*Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Session
	for _, session := range s.sessions {
		if session.Sid == sid {
			result = append(result, session)
		}
	}
	return result
}

// Delete deletes a session by ID
func (s *SessionStore) Delete(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
}

// List returns all active sessions
func (s *SessionStore) List() []*Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Session
	for _, session := range s.sessions {
		result = append(result, session)
	}
	return result
}

var (
	sessionStore = NewSessionStore()
	opIssuer     = "https://auth.example.com" // Replace with your OP issuer
)

// generateID generates a random ID
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", frontChannelLogoutHandler)
	http.HandleFunc("/sessions", sessionsHandler)

	log.Println("Starting RP server on http://localhost:3000")
	log.Println("Front-channel logout endpoint: http://localhost:3000/logout")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

// homeHandler displays the home page
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("home").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Front-Channel Logout RP Example</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        .button { display: inline-block; padding: 10px 20px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; }
        .session { padding: 10px; margin: 10px 0; background: #f0f0f0; border-radius: 4px; }
        h1 { color: #333; }
        code { background: #f4f4f4; padding: 2px 6px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>Front-Channel Logout RP Example</h1>
    
    <h2>About</h2>
    <p>This example demonstrates how a Relying Party (RP) can implement OpenID Connect Front-Channel Logout.</p>
    
    <h2>Configuration</h2>
    <p>
        <strong>Front-Channel Logout URI:</strong> <code>http://localhost:3000/logout</code><br>
        <strong>Expected Issuer:</strong> <code>{{ .Issuer }}</code>
    </p>
    
    <h2>Actions</h2>
    <p>
        <a href="/login?sid=test-sid-{{ .Random }}" class="button">Simulate Login</a>
        <a href="/sessions" class="button">View Sessions</a>
    </p>
    
    <h2>How to Test</h2>
    <ol>
        <li>Register this RP with your OpenID Provider with:<br>
            <code>frontchannel_logout_uri: http://localhost:3000/logout</code><br>
            <code>frontchannel_logout_session_required: true</code>
        </li>
        <li>Perform a login flow to create a session with a real sid</li>
        <li>Initiate logout at the OP: <code>GET /logout</code></li>
        <li>The OP will load <code>http://localhost:3000/logout?iss={{ .Issuer }}&sid=YOUR_SID</code> in an iframe</li>
        <li>This RP will terminate the session and respond</li>
        <li>Check /sessions to verify the session was terminated</li>
    </ol>
    
    <h2>Current Sessions</h2>
    <p>{{ .SessionCount }} active session(s)</p>
    {{ if .Sessions }}
    {{ range .Sessions }}
    <div class="session">
        <strong>Session ID:</strong> {{ .ID }}<br>
        <strong>User ID:</strong> {{ .UserID }}<br>
        <strong>sid:</strong> {{ .Sid }}<br>
        <strong>Created:</strong> {{ .CreatedAt }}
    </div>
    {{ end }}
    {{ end }}
</body>
</html>
    `))

	data := map[string]interface{}{
		"Issuer":       opIssuer,
		"Random":       generateID()[:8],
		"Sessions":     sessionStore.List(),
		"SessionCount": len(sessionStore.List()),
	}

	tmpl.Execute(w, data)
}

// loginHandler simulates a login (in reality, this would be an OAuth callback)
func loginHandler(w http.ResponseWriter, r *http.Request) {
	sid := r.URL.Query().Get("sid")
	if sid == "" {
		sid = generateID()
	}

	// Create a session with the sid from the ID token
	session := sessionStore.Create("user-"+generateID()[:8], sid)

	log.Printf("Created session: ID=%s, UserID=%s, sid=%s", session.ID, session.UserID, session.Sid)

	http.Redirect(w, r, "/", http.StatusFound)
}

// frontChannelLogoutHandler handles front-channel logout notifications from the OP
//
// This endpoint:
// - Accepts GET requests (loaded in an iframe)
// - Receives iss and sid parameters
// - Validates the issuer
// - Finds and terminates sessions with matching sid
// - Returns a simple HTML response
func frontChannelLogoutHandler(w http.ResponseWriter, r *http.Request) {
	iss := r.URL.Query().Get("iss")
	sid := r.URL.Query().Get("sid")

	log.Printf("Front-channel logout request: iss=%s, sid=%s", iss, sid)

	// Validate issuer (IMPORTANT: Always validate the issuer in production)
	if iss != "" && iss != opIssuer {
		log.Printf("ERROR: Invalid issuer: %s (expected %s)", iss, opIssuer)
		http.Error(w, "Invalid issuer", http.StatusBadRequest)
		return
	}

	// Find and terminate all sessions with this sid
	if sid != "" {
		sessions := sessionStore.FindBySid(sid)
		for _, session := range sessions {
			log.Printf("Terminating session: ID=%s, UserID=%s, sid=%s", session.ID, session.UserID, session.Sid)
			sessionStore.Delete(session.ID)
		}
		log.Printf("Terminated %d session(s) for sid=%s", len(sessions), sid)
	} else {
		log.Println("WARNING: No sid parameter provided in logout request")
	}

	// Return a simple HTML page (loaded in iframe, so user won't see it)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Logged Out</title>
</head>
<body>
    <p>Session terminated successfully.</p>
</body>
</html>
    `)
}

// sessionsHandler displays all active sessions
func sessionsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("sessions").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Active Sessions</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        .session { padding: 15px; margin: 10px 0; background: #f0f0f0; border-radius: 4px; }
        h1 { color: #333; }
        .button { display: inline-block; padding: 10px 20px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; margin-top: 20px; }
    </style>
</head>
<body>
    <h1>Active Sessions</h1>
    
    <p><strong>Total:</strong> {{ .Count }} session(s)</p>
    
    {{ if .Sessions }}
    {{ range .Sessions }}
    <div class="session">
        <strong>Session ID:</strong> {{ .ID }}<br>
        <strong>User ID:</strong> {{ .UserID }}<br>
        <strong>sid (for logout):</strong> {{ .Sid }}<br>
        <strong>Created:</strong> {{ .CreatedAt }}
    </div>
    {{ end }}
    {{ else }}
    <p>No active sessions</p>
    {{ end }}
    
    <a href="/" class="button">Back to Home</a>
</body>
</html>
    `))

	data := map[string]interface{}{
		"Sessions": sessionStore.List(),
		"Count":    len(sessionStore.List()),
	}

	tmpl.Execute(w, data)
}
