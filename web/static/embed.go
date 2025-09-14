package static

import "embed"

// StaticFiles contains the web dashboard static assets.
//
//go:embed *
var StaticFiles embed.FS
