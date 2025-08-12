module github.com/blugnu/time

// adopting go1.23 as the minimum version simplifies ticker and timer
// implementation by eliminating the need to handle or document changes
// that occurred in Go 1.23 vs previous versions

go 1.23

require github.com/blugnu/test v0.12.0
