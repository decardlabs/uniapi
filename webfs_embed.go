//go:build !external_static

package main

import (
	"embed"
	"io/fs"
)

// embeddedWebFS stores bundled frontend assets for default builds.
//
//go:embed web/build/*
var embeddedWebFS embed.FS

var webFS fs.FS = embeddedWebFS
