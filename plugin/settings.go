package plugin

// Settings holds a server's plugin settings. This it not specific per plugin, but applied across all plugins.
type Settings struct {
	// Folder is the directory where all the plugin data folders will reside in. It is usually "./plugins".
	Folder string
}
