package tabs

type (
	Tab          string
	ManageAction string
)

const (
	INSTALL Tab = "Install"
	MONITOR Tab = "Monitor"
	MANAGE  Tab = "Manage"

	CHAT   ManageAction = "Chat"
	UPDATE ManageAction = "Update"
	DELETE ManageAction = "Delete"
)
