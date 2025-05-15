package resources

import "embed"

//go:embed *.sql
var QueryFiles embed.FS
