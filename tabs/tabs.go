package tabs

type (
	Tab             string
	InstalledAction string
)

const (
	INSTALL   Tab = "Install"
	RUNNING   Tab = "Running"
	INSTALLED Tab = "Installed"

	CHAT   InstalledAction = "Chat"
	UPDATE InstalledAction = "Update"
	DELETE InstalledAction = "Delete"
)
