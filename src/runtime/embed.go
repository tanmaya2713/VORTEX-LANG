package runtime

import "embed"

//go:embed io.c tensor.c tensor.h
var RuntimeFS embed.FS
