package proto

const (
	topicNameSpace            = "core"
	ListAvailablePluginsTopic = topicNameSpace + ".pluginlist.available"
	ListInstalledPluginsTopic = topicNameSpace + ".pluginlist.installed"
	InstallPluginTopic        = topicNameSpace + ".plugin.install"
	UninstallPluginTopic      = topicNameSpace + ".plugin.uninstall"
	CreatePlugin              = topicNameSpace + ".plugin.create"
)
