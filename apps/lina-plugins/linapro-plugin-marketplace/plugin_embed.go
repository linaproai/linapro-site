// Package marketplace embeds source-plugin resources used by the backend
// registration package. Plugin-full builds import the backend package, so this
// package only owns static assets and shared identity constants.
package marketplace

import (
	"embed"
	"io/fs"
)

// PluginID is the stable plugin identifier shared by plugin.yaml and source
// registration.
const PluginID = "linapro-plugin-marketplace"

// embeddedFiles contains the source-plugin manifest and owned resources that
// the host scans during builtin source-plugin discovery.
//
//go:embed plugin.yaml README.md README.zh-CN.md all:backend all:frontend all:manifest all:hack
var embeddedFiles embed.FS

// EmbeddedFiles returns the read-only source-plugin resource filesystem.
func EmbeddedFiles() fs.FS {
	return embeddedFiles
}
