package ddl

import "embed"

// SQLAssets is an embedded filesystem containing SQL DDL code needed to create the Postgres database
//
//go:embed *
var SQLAssets embed.FS
