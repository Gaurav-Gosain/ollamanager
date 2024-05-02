package tui

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render
	StatusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F1F1F1")).
			Background(lipgloss.Color("#8839ef")).
			Bold(true).
			Padding(0, 1)
	checkMark = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓ ")
)

const (
	padding  = 2
	maxWidth = 80
)

// Response represents the response structure from the API
type Response struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

type progressErrMsg struct{ err error }

func finalPause() tea.Cmd {
	return tea.Tick(time.Millisecond*750, func(_ time.Time) tea.Msg {
		return nil
	})
}

type InstallModel struct {
	Err           error
	status        string
	rawStatus     string
	Spinner       spinner.Model
	Progress      progress.Model
	isDownloading bool
}

var spinners = []spinner.Spinner{
	spinner.Line,
	spinner.Dot,
	spinner.MiniDot,
	spinner.Jump,
	spinner.Pulse,
	spinner.Points,
	spinner.Globe,
	spinner.Moon,
	spinner.Monkey,
	spinner.Meter,
	spinner.Hamburger,
}

func InitSpinner() spinner.Model {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	randomSpinner := spinners[rand.Intn(len(spinners))]
	s.Spinner = randomSpinner
	return s
}

func (m InstallModel) Init() tea.Cmd {
	return m.Spinner.Tick
}

func (m InstallModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.Err = errors.New("user quit mid download :(")
			return m, tea.Quit
		default:
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.Progress.Width = msg.Width - padding*2 - 4
		if m.Progress.Width > maxWidth {
			m.Progress.Width = maxWidth
		}
		return m, nil

	case progressErrMsg:
		m.Err = msg.err
		return m, tea.Quit

	case Response:
		var cmds []tea.Cmd

		if m.rawStatus != msg.Status {
			m.Spinner = InitSpinner()
			cmds = append(cmds, m.Spinner.Tick)
			if m.status != "" {
				cmds = append(cmds, tea.Println(strings.Repeat(" ", padding), checkMark, m.status))
			}
			switch msg.Status {
			case "pulling manifest":
				m.status = "Pulling manifest..."
			case "verifying sha256 digest":
				m.status = "Verifying sha256 digest..."
			case "writing manifest":
				m.status = "Writing manifest..."
			case "removing any unused layers":
				m.status = "Removing any unused layers..."
			case "success":
				m.status = "Success!"
				cmds = append(cmds, tea.Sequence(finalPause(), tea.Quit))
			default:
				m.status = "Downloading... (" + msg.Status + ") "
			}
		}

		m.isDownloading = msg.Total != 0
		m.rawStatus = msg.Status

		if m.isDownloading {
			progress := float64(msg.Completed) / float64(msg.Total)
			cmds = append(cmds, m.Progress.SetPercent(progress))
		}

		return m, tea.Batch(cmds...)

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.Progress.Update(msg)
		m.Progress = progressModel.(progress.Model)
		return m, cmd

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd

	default:
		return m, nil
	}
}

func (m InstallModel) View() string {
	if m.Err != nil {
		// utils.PrintError(m.err, true)
		return ""
	}

	if m.rawStatus == "success" {
		return ""
	}

	pad := strings.Repeat(" ", padding)

	status := StatusStyle.SetString(m.status).String()

	if m.isDownloading {
		return fmt.Sprintf(
			"\n%s%s  %s\n\n%s%s\n\n",
			pad, m.Spinner.View(), status,
			pad, m.Progress.View(),
		)
	}

	return fmt.Sprintf(
		"\n%s%s  %s\n\n",
		pad, m.Spinner.View(), status,
	)
}
