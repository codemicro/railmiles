package webAssets

import "embed"

//go:generate npm run build

//go:embed public/*
var Public embed.FS
