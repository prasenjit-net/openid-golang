package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/prasenjit-net/openid-golang/pkg/configstore"
)

// BootstrapHandler handles the initial setup wizard
type BootstrapHandler struct {
	configStore configstore.ConfigStore
}

// NewBootstrapHandler creates a new bootstrap handler
func NewBootstrapHandler(configStore configstore.ConfigStore) *BootstrapHandler {
	return &BootstrapHandler{
		configStore: configStore,
	}
}

// SetupRequest represents the initial setup request
type SetupRequest struct {
	Issuer string `json:"issuer"`
}

// SetupResponse represents the setup response
type SetupResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// CheckInitialized checks if the application is already initialized
func (h *BootstrapHandler) CheckInitialized(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	initialized, err := h.configStore.IsInitialized(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"initialized": initialized,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Initialize performs the initial setup with minimal config (just issuer)
func (h *BootstrapHandler) Initialize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Check if already initialized
	initialized, err := h.configStore.IsInitialized(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if initialized {
		http.Error(w, "Already initialized", http.StatusBadRequest)
		return
	}

	// Parse request
	var req SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Issuer == "" {
		http.Error(w, "Issuer URL is required", http.StatusBadRequest)
		return
	}

	// Initialize with minimal config
	if err := configstore.InitializeMinimalConfig(ctx, h.configStore, req.Issuer); err != nil {
		http.Error(w, "Failed to initialize: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SetupResponse{
		Success: true,
		Message: "Setup completed successfully. Reloading...",
	})
}

// ServeSetupWizard serves a simple HTML setup wizard
func (h *BootstrapHandler) ServeSetupWizard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>OpenID Server Setup</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            max-width: 500px;
            width: 100%%;
            padding: 40px;
        }
        h1 { 
            color: #333;
            margin-bottom: 10px;
            font-size: 28px;
        }
        .subtitle {
            color: #666;
            margin-bottom: 30px;
            font-size: 14px;
        }
        .form-group { 
            margin-bottom: 20px;
        }
        label { 
            display: block;
            margin-bottom: 8px;
            font-weight: 600;
            color: #333;
            font-size: 14px;
        }
        input { 
            width: 100%%;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 14px;
            transition: border-color 0.3s;
        }
        input:focus {
            outline: none;
            border-color: #667eea;
        }
        button { 
            width: 100%%;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            padding: 14px;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            font-size: 16px;
            font-weight: 600;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        button:hover { 
            transform: translateY(-2px);
            box-shadow: 0 10px 20px rgba(102, 126, 234, 0.4);
        }
        button:disabled {
            opacity: 0.6;
            cursor: not-allowed;
            transform: none;
        }
        .message { 
            margin-top: 20px;
            padding: 12px;
            border-radius: 8px;
            font-size: 14px;
            display: none;
        }
        .message.error {
            background: #fee;
            color: #c33;
            border: 1px solid #fcc;
        }
        .message.success {
            background: #efe;
            color: #3c3;
            border: 1px solid #cfc;
        }
        .example {
            font-size: 12px;
            color: #999;
            margin-top: 4px;
        }
        .storage-info {
            background: #f8f9fa;
            padding: 12px;
            border-radius: 8px;
            margin-bottom: 20px;
            font-size: 13px;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 OpenID Server Setup</h1>
        <p class="subtitle">Let's get your OpenID Connect server configured!</p>
        
        <div class="storage-info">
            <strong>Storage:</strong> <span id="storageType">Detecting...</span>
        </div>
        
        <form id="setupForm">
            <div class="form-group">
                <label for="issuer">Issuer URL *</label>
                <input 
                    type="url" 
                    id="issuer" 
                    name="issuer" 
                    placeholder="https://id.example.com" 
                    required
                    autocomplete="off"
                >
                <div class="example">This is your OpenID provider's public URL</div>
            </div>

            <button type="submit" id="submitBtn">Initialize Server</button>
            <div id="message" class="message"></div>
        </form>
    </div>

    <script>
        // Detect storage type on load
        async function detectStorage() {
            const mongoEnv = '%s' || 'Not configured';
            const storageEl = document.getElementById('storageType');
            
            if (mongoEnv !== 'Not configured') {
                storageEl.innerHTML = '<strong>MongoDB</strong> (from environment)';
            } else {
                storageEl.innerHTML = '<strong>JSON file</strong> (data/config.json)';
            }
        }

        detectStorage();

        document.getElementById('setupForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const submitBtn = document.getElementById('submitBtn');
            const messageDiv = document.getElementById('message');
            const issuer = document.getElementById('issuer').value.trim();
            
            // Validate
            if (!issuer) {
                messageDiv.className = 'message error';
                messageDiv.style.display = 'block';
                messageDiv.textContent = 'Please enter an issuer URL';
                return;
            }

            // Disable button
            submitBtn.disabled = true;
            submitBtn.textContent = 'Initializing...';
            messageDiv.style.display = 'none';

            try {
                const response = await fetch('/api/setup/initialize', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ issuer })
                });

                const result = await response.json();
                
                if (response.ok) {
                    messageDiv.className = 'message success';
                    messageDiv.style.display = 'block';
                    messageDiv.textContent = '✓ ' + result.message;
                    
                    // Reload after 2 seconds
                    setTimeout(() => {
                        window.location.reload();
                    }, 2000);
                } else {
                    throw new Error(result.message || response.statusText);
                }
            } catch (error) {
                messageDiv.className = 'message error';
                messageDiv.style.display = 'block';
                messageDiv.textContent = '✗ Setup failed: ' + error.message;
                submitBtn.disabled = false;
                submitBtn.textContent = 'Initialize Server';
            }
        });
    </script>
</body>
</html>`, getMongoEnvIndicator())
}

func getMongoEnvIndicator() string {
	if os.Getenv("MONGODB_URI") != "" {
		return "MONGODB_URI"
	}
	return ""
}
