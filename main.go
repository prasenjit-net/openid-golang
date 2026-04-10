package main

import (
	"github.com/prasenjit-net/openid-golang/cmd"
)

func main() {
	cmd.SetEmbeds(adminUIFS, publicFS)
	cmd.Execute()
}
