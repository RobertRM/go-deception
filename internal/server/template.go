package server

import "embed"

//go:embed templates/*.html
var templateFS embed.FS

//go:embed templates/icons/*
var iconsFS embed.FS
