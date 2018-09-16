# Go package `mymods`

Package `mymods` is an **experimental** way to read the module versions table
for the calling executable, if it was placed there by the Go toolchain.
This uses the same technique as
[`goversion`](https://github.com/rsc/goversion), but applies it to the current
executable (as determined by `os.Executable`).

A program may wish to include version information about itself or its
dependencies somewhere in its output. For example, a CLI tool may include
version information when it's run with a `--version` command line option.

For more information, see [the `mymods` package godoc](https://godoc.org/github.com/apparentlymart/go-mymods/mymods).

This module is experimental for a number of reasons:

- The Go Modules mechanism itself is currently experimental.
- The manifest table this module reads is not documented anywhere as a stable
  mechanism that is guaranteed to exist in future versions.
- It has not been widely tested across different platforms. In particular, it
  may not work on platforms that apply exclusive locks to a running executable,
  such as on Windows.

`mymods` itself has not yet been assigned a version number since it remains
experimental. Callers should instead depend on a specific Git commit.

This module may be updated as best-practices for module version information
emerge, though the capabilities of this module may eventually (hopefully!) be
included as a standard feature of Go.
