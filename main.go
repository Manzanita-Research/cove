package main

import (
	"embed"

	"github.com/manzanita-research/claudebox/cmd"
)

//go:embed embed/Dockerfile
var embedFS embed.FS

func main() {
	cmd.Execute(embedFS)
}
