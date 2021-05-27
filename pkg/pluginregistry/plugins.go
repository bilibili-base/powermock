package pluginregistry

// Plugin defines the most basic requirements for plug-ins
// All plugins should implement this interface
type Plugin interface {
	Name() string
}
