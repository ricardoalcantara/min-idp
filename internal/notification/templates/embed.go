package templates

import "embed"

//go:embed *.html *.txt
var Files embed.FS
