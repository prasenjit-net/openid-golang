package main

import "embed"

//go:embed frontend/dist
var adminUIFS embed.FS

//go:embed frontend/setup.html
var setupHTMLFS embed.FS

//go:embed frontend/login.html frontend/consent.html
var oidcTemplatesFS embed.FS
