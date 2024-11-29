package tui

import "github.com/charmbracelet/bubbles/v2/key"

type KeyMap struct {
	Up           key.Binding
	Down         key.Binding
	Left         key.Binding
	Right        key.Binding
	Quit         key.Binding
	Enter        key.Binding
	Filter       key.Binding
	ClearFilter  key.Binding
	NextTab      key.Binding
	PrevTab      key.Binding
	FullHelpKeys [][]key.Binding
}

var Keys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "move left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "move right"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "pick selected item"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter/fuzzy find items"),
	),
	ClearFilter: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "clear filter"),
	),
	NextTab: key.NewBinding(
		key.WithKeys("n", "tab"),
		key.WithHelp("n/tab", "switch to the next tab"),
	),
	PrevTab: key.NewBinding(
		key.WithKeys("p", "shift+tab"),
		key.WithHelp("p/shift+tab", "switch to the previous tab"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q/ctrl+c", "quit"),
	),
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit}
}

func (k *KeyMap) SetFullHelpKeys(keys [][]key.Binding) {
	k.FullHelpKeys = keys
}

func (k *KeyMap) DefaultFullHelpKeys() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right, k.Enter},
		{k.Quit, k.NextTab, k.PrevTab, k.Filter, k.ClearFilter},
	}
}

func (k *KeyMap) DefaultFullHelpKeysSingleTab() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right, k.Enter},
		{k.Quit, k.Filter, k.ClearFilter},
	}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k KeyMap) FullHelp() [][]key.Binding {
	return k.FullHelpKeys
}
