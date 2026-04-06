package main

import (
	"embed"

	"github.com/manzanita-research/cove/cmd"
)

//go:embed embed/Dockerfile embed/entrypoint.sh
var embedFS embed.FS

func main() {
	cmd.Execute(embedFS)
}
