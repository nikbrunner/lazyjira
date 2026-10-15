package main

import (
	"runtime/debug"
	"testing"
)

func TestResolveVersion(t *testing.T) {
	t.Parallel()
	vcs := []debug.BuildSetting{{Key: "vcs.revision", Value: "abc"}}
	other := []debug.BuildSetting{{Key: "GOOS", Value: "linux"}}
	cases := []struct {
		name    string
		ldflags string
		info    *debug.BuildInfo
		ok      bool
		want    string
	}{
		{"ldflags wins", "1.2.3", &debug.BuildInfo{Main: debug.Module{Version: "v0.13.0"}}, true, "1.2.3"},
		{"proxy install", "dev", &debug.BuildInfo{Main: debug.Module{Version: "v0.13.0"}, Settings: other}, true, "0.13.0"},
		{"local build with vcs settings", "dev", &debug.BuildInfo{Main: debug.Module{Version: "v0.13.0"}, Settings: vcs}, true, "dev"},
		{"devel", "dev", &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, true, "dev"},
		{"empty main version", "dev", &debug.BuildInfo{}, true, "dev"},
		{"build info unavailable", "dev", nil, false, "dev"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertEqual(t, "resolveVersion", resolveVersion(tc.ldflags, tc.info, tc.ok), tc.want)
		})
	}
}
