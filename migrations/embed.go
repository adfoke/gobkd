package migrations

import "embed"

// Files keeps SQL migrations in lexical order.
//
//go:embed *.sql
var Files embed.FS
