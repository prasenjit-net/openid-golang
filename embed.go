package main

import "embed"

//go:embed frontend/dist
var adminUIFS embed.FS

//go:embed public
var publicFS embed.FS
