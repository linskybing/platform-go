package builtin

import "github.com/linskybing/platform-go/internal/plugin"

// Init registers all built-in plugins.
func Init() {
	plugin.Register(&PreemptionPlugin{})
}
