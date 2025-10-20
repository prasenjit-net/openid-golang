package ui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed dist
var adminUI embed.FS

//go:embed setup.html
var setupHTML embed.FS

// GetAdminUI returns an http.FileSystem for the embedded admin UI
func GetAdminUI() http.FileSystem {
	subFS, err := fs.Sub(adminUI, "dist")
	if err != nil {
		panic(err)
	}
	return http.FS(subFS)
}

// GetAdminFS returns the embedded FS for the admin UI
func GetAdminFS() fs.FS {
	subFS, err := fs.Sub(adminUI, "dist")
	if err != nil {
		panic(err)
	}
	return subFS
}

// GetSetupHTML returns the setup wizard HTML content with placeholders replaced
func GetSetupHTML(storageInfo string) (string, error) {
	content, err := setupHTML.ReadFile("setup.html")
	if err != nil {
		return "", err
	}

	html := string(content)
	html = strings.Replace(html, "{{STORAGE_INFO}}", storageInfo, 1)

	return html, nil
}
