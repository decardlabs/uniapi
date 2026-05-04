//go:build external_static

package main

import (
	"io/fs"
	"os"
)

var webFS fs.FS = os.DirFS(".")
