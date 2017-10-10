package version

import (
	"fmt"
	"runtime"
)

var (
	Version        = "none"
	Branch         = "none"
	BuildTimestamp = "none"
	GoVersion      = runtime.Version()
)

func String() string {
	return fmt.Sprintf("version: %s, branch: %s, build timestamp: %s, go version: %s",
		Version,
		Branch,
		BuildTimestamp,
		GoVersion)
}
