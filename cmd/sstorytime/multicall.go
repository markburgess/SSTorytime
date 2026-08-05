package main

import (
	"path/filepath"
	"strings"
)

// multiCall maps an install name (argv0 basename) to a root subcommand name.
// Each subcommand registers itself from init() via registerMultiCall.
var multiCall = map[string]string{}

func registerMultiCall(cmdName string, names ...string) {
	for _, name := range names {
		if name == "" {
			continue
		}
		multiCall[name] = cmdName
	}
}

// applyMultiCall rewrites argv for busybox-style invocation.
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
	if len(args) > 1 && args[1] == sub {
		return args
	}
	out := make([]string, 0, len(args)+1)
	out = append(out, args[0], sub)
	out = append(out, args[1:]...)
	return out
}
