package akasha

import "embed"

//go:embed all:templates/*
var TemplateFS embed.FS

//go:embed all:static/css/* all:static/js/* all:static/*.svg
var StaticFS embed.FS
