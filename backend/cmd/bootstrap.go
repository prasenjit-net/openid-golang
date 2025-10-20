package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prasenjit-net/openid-golang/pkg/configstore"
	"github.com/prasenjit-net/openid-golang/pkg/handlers"
	"github.com/spf13/cobra"
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Start bootstrap server for initial setup",
	Long: `Start a minimal server that serves the setup wizard.
This is used when the application is not yet configured.`,
	Run: runBootstrap,
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)
}

func runBootstrap(cmd *cobra.Command, args []string) {
	// Create default JSON config store for bootstrap
	configStore := configstore.DefaultJSONConfigStore()

	ctx := context.Background()

	// Check if already initialized
	initialized, err := configStore.IsInitialized(ctx)
	if err != nil {
		log.Fatalf("Failed to check initialization status: %v", err)
	}

	if initialized {
		log.Println("⚠️  Application is already initialized!")
		log.Println("Run 'openid-server serve' to start the server.")
		os.Exit(0)
	}

	// Create bootstrap handler
	bootstrapHandler := handlers.NewBootstrapHandler(configStore)

	// Setup routes
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/bootstrap/status", bootstrapHandler.CheckInitialized)
	mux.HandleFunc("/api/bootstrap/initialize", bootstrapHandler.Initialize)

	// Serve setup wizard UI
	// TODO: Serve the React setup wizard UI
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>OpenID Server Setup</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: bold; }
        input, select { width: 100%%; padding: 8px; box-sizing: border-box; }
        button { background: #007bff; color: white; padding: 10px 20px; border: none; cursor: pointer; }
        button:hover { background: #0056b3; }
        .error { color: red; margin-top: 10px; }
        .success { color: green; margin-top: 10px; }
    </style>
</head>
<body>
    <h1>🚀 OpenID Connect Server Setup</h1>
    <p>Welcome! Please configure your OpenID Connect server.</p>
    
    <form id="setupForm">
        <div class="form-group">
            <label>Storage Backend:</label>
            <select id="storageType" name="storage_type" onchange="toggleStorageFields()">
                <option value="json">JSON File</option>
                <option value="mongodb">MongoDB</option>
            </select>
        </div>

        <div id="jsonFields">
            <div class="form-group">
                <label>JSON File Path:</label>
                <input type="text" name="json_file_path" value="data/openid.json">
            </div>
        </div>

        <div id="mongoFields" style="display:none;">
            <div class="form-group">
                <label>MongoDB URI:</label>
                <input type="text" name="mongo_uri" placeholder="mongodb://localhost:27017">
            </div>
            <div class="form-group">
                <label>Database Name:</label>
                <input type="text" name="mongo_database" value="openid">
            </div>
        </div>

        <div class="form-group">
            <label>Issuer URL:</label>
            <input type="text" name="issuer" placeholder="https://id.example.com" required>
        </div>

        <div class="form-group">
            <label>Server Host:</label>
            <input type="text" name="host" value="0.0.0.0">
        </div>

        <div class="form-group">
            <label>Server Port:</label>
            <input type="number" name="port" value="8080">
        </div>

        <button type="submit">Initialize Server</button>
        <div id="message"></div>
    </form>

    <script>
        function toggleStorageFields() {
            const storageType = document.getElementById('storageType').value;
            document.getElementById('jsonFields').style.display = storageType === 'json' ? 'block' : 'none';
            document.getElementById('mongoFields').style.display = storageType === 'mongodb' ? 'block' : 'none';
        }

        document.getElementById('setupForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const data = {
                storage_type: formData.get('storage_type'),
                issuer: formData.get('issuer'),
                host: formData.get('host'),
                port: parseInt(formData.get('port')),
                jwt_expiry_minutes: 60,
                refresh_enabled: true
            };

            if (data.storage_type === 'json') {
                data.json_file_path = formData.get('json_file_path');
            } else {
                data.mongo_uri = formData.get('mongo_uri');
                data.mongo_database = formData.get('mongo_database');
            }

            try {
                const response = await fetch('/api/bootstrap/initialize', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });

                const result = await response.json();
                const messageDiv = document.getElementById('message');
                
                if (response.ok) {
                    messageDiv.className = 'success';
                    messageDiv.textContent = '✓ ' + result.message + ' - Please restart the server with: openid-server serve';
                } else {
                    messageDiv.className = 'error';
                    messageDiv.textContent = '✗ Setup failed: ' + (result.message || response.statusText);
                }
            } catch (error) {
                document.getElementById('message').className = 'error';
                document.getElementById('message').textContent = '✗ Error: ' + error.message;
            }
        });
    </script>
</body>
</html>`)
	})

	// Create server
	host := "0.0.0.0"
	port := 8080
	addr := fmt.Sprintf("%s:%d", host, port)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("\nShutting down bootstrap server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Bootstrap server shutdown error: %v", err)
		}
	}()

	// Start server
	log.Printf("🚀 Bootstrap server starting on http://%s", addr)
	log.Printf("📝 Open your browser and navigate to http://localhost:%d to complete setup", port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Bootstrap server error: %v", err)
	}
}
