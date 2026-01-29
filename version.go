package otel

import "runtime/debug"

var Version = "dev"

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if dep.Path == "github.com/pink-tools/pink-otel" {
				Version = dep.Version
				return
			}
		}
	}
}
