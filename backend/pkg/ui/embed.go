package ui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:dist
var adminUI embed.FS

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
