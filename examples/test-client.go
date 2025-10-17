package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// Simple test client to verify the OpenID Connect flow
func main() {
	// Configuration
	clientID := "your-client-id"           // Get this from seed output
	clientSecret := "your-client-secret"   // Get this from seed output
	redirectURI := "http://localhost:9090/callback"
	issuer := "http://localhost:8080"

	fmt.Println("OpenID Connect Test Client")
	fmt.Println("===========================")
	fmt.Println()
	fmt.Printf("Make sure the OpenID server is running at %s\n", issuer)
	fmt.Println("Run the seed script first: go run scripts/seed.go")
	fmt.Println()
	fmt.Printf("Client ID: %s\n", clientID)
	fmt.Printf("Client Secret: %s\n", clientSecret)
	fmt.Println()

	// Start callback server
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		
		if code == "" {
			errorCode := r.URL.Query().Get("error")
			errorDesc := r.URL.Query().Get("error_description")
			fmt.Fprintf(w, "Error: %s - %s", errorCode, errorDesc)
			return
		}

		fmt.Printf("\n✓ Received authorization code: %s\n", code)
		fmt.Printf("✓ State: %s\n", state)
		
		fmt.Fprintf(w, `
			<html>
			<body>
				<h2>Authorization Successful!</h2>
				<p>Authorization code received: <code>%s</code></p>
				<p>State: <code>%s</code></p>
				<p>Now you can exchange this code for tokens using the /token endpoint.</p>
				<pre>
curl -X POST %s/token \
  -u "%s:%s" \
  -d "grant_type=authorization_code&code=%s&redirect_uri=%s"
				</pre>
			</body>
			</html>
		`, code, state, issuer, clientID, clientSecret, code, redirectURI)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		authURL := fmt.Sprintf("%s/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=openid%%20profile%%20email&state=test123",
			issuer, clientID, redirectURI)
		
		fmt.Fprintf(w, `
			<html>
			<body>
				<h2>OpenID Connect Test Client</h2>
				<p><a href="%s">Click here to authorize</a></p>
				<p>Or visit this URL:</p>
				<pre>%s</pre>
				<hr>
				<p><strong>Test credentials:</strong></p>
				<ul>
					<li>Username: testuser</li>
					<li>Password: password123</li>
				</ul>
			</body>
			</html>
		`, authURL, authURL)
	})

	fmt.Println("Starting test client on http://localhost:9090")
	fmt.Println("Visit http://localhost:9090 to start the authorization flow")
	fmt.Println()

	server := &http.Server{
		Addr:         ":9090",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}
