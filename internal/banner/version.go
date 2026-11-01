package banner

import "runtime/debug"

func init() {
	// Release binaries get AppVersion injected via -ldflags. For other build
	// paths (`go install ...@version`, source builds), fall back to module
	// build info so the reported version still matches the tag.
	if AppVersion != "dev" {
		return
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		AppVersion = v
		return
	}
	// Source build without a module version: expose the VCS revision.
	var revision, modified string
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value
		}
	}
	if revision != "" {
		if len(revision) > 12 {
			revision = revision[:12]
		}
		AppVersion = "dev-" + revision
		if modified == "true" {
			AppVersion += "-dirty"
		}
	}
}
