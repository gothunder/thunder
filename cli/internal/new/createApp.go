package new

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"time"
)

type creationStep struct {
	Step string
	Done bool
}

var progressedSteps = []creationStep{
	{
		Step: "Creating app folder",
		Done: false,
	},
	{
		Step: "Adding events folder",
		Done: false,
	},
}
var progressBar = progress.New(progress.WithDefaultGradient())
var spinnerLoader = spinner.New(spinner.WithSpinner(spinner.Dot))

func createAppView(m Model) string {
	var steps string
	for i, step := range progressedSteps {
		if step.Done {
			steps += step.Step + "\n"
			continue
		}
		if i != 0 && !progressedSteps[i-1].Done {
			steps += step.Step + "\n"
			continue
		}

		steps += step.Step + " " + spinnerLoader.View() + "\n"
	}

	return steps + "\n" +
		progressBar.View()
}

func updateCreateApp(msg tea.Msg, m Model) (Model, tea.Cmd) {
	const (
		padding  = 2
		maxWidth = 80
	)

	switch msgType := msg.(type) {
	case viewChangedMsg:
		return m, tea.Batch(
			spinnerLoader.Tick,
			createRootApp(m),
		)

	case createRootAppMsg:
		return m, createRootAppMsgHandler(m)
	case createEventsMsg:
		return m, createEventsMsgHandler(m)

	case tea.WindowSizeMsg:
		progressBar.Width = msgType.Width - padding*2 - 4
		if progressBar.Width > maxWidth {
			progressBar.Width = maxWidth
		}
		return m, nil

	case tickMsg:
		if progressBar.Percent() == 1.0 {
			m.Generated = true
			return m, tea.Quit
		}

		return m, tick()

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := progressBar.Update(msg)
		progressBar = progressModel.(progress.Model)
		return m, cmd
	case spinner.TickMsg:
		var cmd tea.Cmd
		spinnerLoader, cmd = spinnerLoader.Update(msg)
		return m, cmd
	}

	return m, tea.Batch(
		spinnerLoader.Tick,
		tick(),
	)
}

type createRootAppMsg struct{}

func createRootAppMsgHandler(m Model) tea.Cmd {
	return tea.Batch(
		progressBar.IncrPercent(0.5),
		createEvents(m),
	)
}
func createRootApp(m Model) tea.Cmd {
	fn := func() tea.Msg {
		time.Sleep(time.Second * 2)

		//err := os.Mkdir(m.AppName, 0755)
		//if err != nil {
		//	panic(err)
		//}
		//if err != nil {
		//	panic(err)
		//}
		progressedSteps[0].Done = true
		return createRootAppMsg{}
	}

	return fn
}

type createEventsMsg struct{}

func createEventsMsgHandler(m Model) tea.Cmd {
	return tea.Batch(
		progressBar.IncrPercent(0.5),
	)
}
func createEvents(m Model) tea.Cmd {
	fn := func() tea.Msg {
		time.Sleep(time.Second * 2)

		//events.Create(m.AppName)
		progressedSteps[1].Done = true
		return createEventsMsg{}
	}

	return fn
}
