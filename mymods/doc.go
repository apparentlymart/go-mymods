// Package mymods gives a Go program access to the module version metadata
// placed into its own executable image by the Go toolchain.
//
// This allows, for example, a program to discover what version of its main
// module was used to build it, which it might then choose to return as its
// own version in response to a run with a --version or similar option.
package mymods
