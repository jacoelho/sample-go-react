package client

import (
	"embed"
)

//go:generate yarn build

//go:embed build/*
var Content embed.FS
