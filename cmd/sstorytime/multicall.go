package main

import (
	"path/filepath"
	"strings"
)

// multiCall maps an install name (argv0 basename) to a root subcommand name.
// Each subcommand registers itself from init() via registerMultiCall.
// Invoking a symlink/hardlink named e.g. N4L is equivalent to "sstorytime n4l …".
var multiCall = map[string]string{}

// registerMultiCall binds busybox-style binary names to a cobra subcommand.
// cmdName is the Use of a root-level command (e.g. "n4l", "graph-report").
// names are argv0 basenames (e.g. "N4L", "searchN4L").
func registerMultiCall(cmdName string, names ...string) {
	for _, name := range names {
		if name == "" {
			continue
		}
		multiCall[name] = cmdName
	}
}

// applyMultiCall rewrites argv for busybox-style invocation.
// If the program is invoked as one of the registered names, the matching
// subcommand is inserted so Cobra sees "sstorytime <sub> …".
func applyMultiCall(args []string) []string {
	if len(args) == 0 {
		return args
	}
	base := filepath.Base(args[0])
	base = strings.TrimSuffix(base, ".exe")
	sub, ok := multiCall[base]
	if !ok {
		return args
	}
	// Already "N4L n4l …" or user passed the subcommand explicitly.
	if len(args) > 1 && args[1] == sub {
		return args
	}
	out := make([]string, 0, len(args)+1)
	out = append(out, args[0], sub)
	out = append(out, args[1:]...)
	return out
}
