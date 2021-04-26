package client

import (
	"embed"
	"io/fs"
)

//go:generate yarn build

//go:embed build/*
var content embed.FS

func Contents() (fs.FS, error) {
	return fs.Sub(content, "build")
}
