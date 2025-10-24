package client

import (
	"fmt"
	"runtime"
	"strings"
)

func buildUserAgent() string {
	// runtime.Version() returns "go1.22.0", strip the "go" prefix for consistency with other clients
	goVersion := strings.TrimPrefix(runtime.Version(), "go")

	return fmt.Sprintf(
		"thingsdiary-client/%s (go/%s; %s; %s)",
		Version,
		goVersion,
		runtime.GOOS,
		runtime.GOARCH,
	)
}
