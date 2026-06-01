package runtime

import "embed"

//go:embed c_lib/io.c c_lib/tensor.c c_lib/tensor.h
var RuntimeFS embed.FS
