// Package buildinfo owns the version injected into release artifacts.
package buildinfo

// Version is replaced with the repository VERSION value through -ldflags.
var Version = "dev"
