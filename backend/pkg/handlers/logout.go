package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/prasenjit-net/openid-golang/pkg/models"
	"github.com/prasenjit-net/openid-golang/pkg/session"
)

// Logout handles the OIDC front-channel logout endpoint
// Implements OpenID Connect Front-Channel Logout 1.0
func (h *Handlers) Logout(c echo.Context) error {
	// Get parameters
	// idTokenHint is optional and not currently validated
	// idTokenHint := c.QueryParam("id_token_hint")
	postLogoutRedirectURI := c.QueryParam("post_logout_redirect_uri")
	state := c.QueryParam("state")

	// Get current user session
	userSession := session.GetUserSession(c)
	if userSession == nil {
		// No active session - redirect to post_logout_redirect_uri if provided
		if postLogoutRedirectURI != "" {
			return h.redirectAfterLogout(c, postLogoutRedirectURI, state)
		}
		return c.JSON(http.StatusOK, map[string]string{
			"message": "No active session",
		})
	}

	// Get all clients associated with this session for front-channel logout
	sessionClients, err := h.storage.GetSessionClientsBySessionID(userSession.ID)
	if err != nil {
		return jsonError(c, http.StatusInternalServerError, ErrorServerError, "Failed to retrieve session clients")
	}

	// Delete the user session
	if err := h.sessionManager.DeleteUserSession(c, userSession.ID); err != nil {
		return jsonError(c, http.StatusInternalServerError, ErrorServerError, "Failed to delete session")
	}

	// Delete session-client associations
	if err := h.storage.DeleteSessionClientsBySessionID(userSession.ID); err != nil {
		// Log error but don't fail the logout
		c.Logger().Warnf("Failed to delete session clients: %v", err)
	}

	// If there are clients to notify, render an HTML page with iframes
	if len(sessionClients) > 0 {
		return h.renderLogoutPage(c, sessionClients, postLogoutRedirectURI, state)
	}

	// No clients to notify - redirect directly
	if postLogoutRedirectURI != "" {
		return h.redirectAfterLogout(c, postLogoutRedirectURI, state)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Logout successful",
	})
}

// renderLogoutPage renders an HTML page with iframes for front-channel logout
func (h *Handlers) renderLogoutPage(c echo.Context, sessionClients []*models.SessionClient, postLogoutRedirectURI, state string) error {
	// Build logout notification URLs for each client
	var iframes []LogoutIframe
	for _, sc := range sessionClients {
		client, err := h.storage.GetClientByID(sc.ClientID)
		if err != nil || client == nil || client.FrontchannelLogoutURI == "" {
			continue
		}

		logoutURL := client.FrontchannelLogoutURI

		// Add sid parameter if client requires session ID
		if client.FrontchannelLogoutSessionRequired {
			separator := "?"
			if strings.Contains(logoutURL, "?") {
				separator = "&"
			}
			logoutURL = fmt.Sprintf("%s%ssid=%s", logoutURL, separator, url.QueryEscape(sc.Sid))

			// Also add iss parameter per spec
			logoutURL = fmt.Sprintf("%s&iss=%s", logoutURL, url.QueryEscape(h.config.Issuer))
		}

		iframes = append(iframes, LogoutIframe{
			URL: logoutURL,
		})
	}

	// Prepare redirect URL if provided
	redirectURL := ""
	if postLogoutRedirectURI != "" {
		redirectURL = postLogoutRedirectURI
		if state != "" {
			separator := "?"
			if strings.Contains(redirectURL, "?") {
				separator = "&"
			}
			redirectURL = fmt.Sprintf("%s%sstate=%s", redirectURL, separator, url.QueryEscape(state))
		}
	}

	// Render HTML page
	tmpl := template.Must(template.New("logout").Parse(logoutPageTemplate))
	data := LogoutPageData{
		Iframes:     iframes,
		RedirectURL: redirectURL,
	}

	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response().WriteHeader(http.StatusOK)
	return tmpl.Execute(c.Response().Writer, data)
}

// redirectAfterLogout redirects to post_logout_redirect_uri with state parameter
func (h *Handlers) redirectAfterLogout(c echo.Context, postLogoutRedirectURI, state string) error {
	redirectURL := postLogoutRedirectURI
	if state != "" {
		separator := "?"
		if strings.Contains(redirectURL, "?") {
			separator = "&"
		}
		redirectURL = fmt.Sprintf("%s%sstate=%s", redirectURL, separator, url.QueryEscape(state))
	}
	return c.Redirect(http.StatusFound, redirectURL)
}

// LogoutIframe represents an iframe for front-channel logout
type LogoutIframe struct {
	URL string
}

// LogoutPageData holds data for the logout page template
type LogoutPageData struct {
	Iframes     []LogoutIframe
	RedirectURL string
}

// logoutPageTemplate is the HTML template for the logout page with iframes
const logoutPageTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Logging out...</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background-color: #f5f5f5;
        }
        .container {
            text-align: center;
            padding: 2rem;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .spinner {
            border: 4px solid #f3f3f3;
            border-top: 4px solid #3498db;
            border-radius: 50%;
            width: 40px;
            height: 40px;
            animation: spin 1s linear infinite;
            margin: 0 auto 1rem;
        }
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
        iframe {
            display: none;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="spinner"></div>
        <h2>Logging out...</h2>
        <p>Please wait while we complete the logout process.</p>
    </div>
    
    {{range .Iframes}}
    <iframe src="{{.URL}}"></iframe>
    {{end}}
    
    <script>
        // Wait for all iframes to load, then redirect
        var iframes = document.querySelectorAll('iframe');
        var loadedCount = 0;
        var timeout = 5000; // 5 seconds timeout
        
        function checkComplete() {
            if (loadedCount >= iframes.length) {
                redirectToDestination();
            }
        }
        
        function redirectToDestination() {
            {{if .RedirectURL}}
            window.location.href = '{{.RedirectURL}}';
            {{else}}
            document.querySelector('.container').innerHTML = '<h2>Logout Complete</h2><p>You have been logged out successfully.</p>';
            {{end}}
        }
        
        // Set up iframe load handlers
        iframes.forEach(function(iframe) {
            iframe.onload = function() {
                loadedCount++;
                checkComplete();
            };
            iframe.onerror = function() {
                loadedCount++;
                checkComplete();
            };
        });
        
        // Fallback timeout in case some iframes don't load
        setTimeout(redirectToDestination, timeout);
        
        // If no iframes, redirect immediately
        if (iframes.length === 0) {
            redirectToDestination();
        }
    </script>
</body>
</html>
`
