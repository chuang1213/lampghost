package ghost

import (
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/Catizard/lampghost/internel/difftable"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// sessionState is used to track which model is focused
type sessionState uint

const (
	defaultTime              = time.Minute
	levelView   sessionState = iota
	songView
)

var (
	modelStyle = lipgloss.NewStyle().
			Align(lipgloss.Left, lipgloss.Center).
			BorderStyle(lipgloss.HiddenBorder())
	focusedModelStyle = lipgloss.NewStyle().
				Align(lipgloss.Left, lipgloss.Center).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("69"))
	listStyle = lipgloss.NewStyle().Margin(1, 2)
)

type mainModel struct {
	levelList   list.Model
	songList    list.Model
	state       sessionState
	choice      item
	w           int
	h           int
	songListMap map[string]list.Model
}

func buildLevelList(dth *difftable.DiffTableHeader, diffTable []difftable.DiffTable) ([]list.Item, string) {
	// convert diffTable to list items
	levels := make(map[string]interface{})
	for _, v := range diffTable {
		levels[v.Level] = new(interface{})
	}
	if len(levels) == 0 {
		panic("tableHeader.json file corrupted, no level found")
	}
	sortedLevels := make([]string, 0)
	for level := range levels {
		sortedLevels = append(sortedLevels, level)
	}
	sort.Slice(sortedLevels, func(i, j int) bool {
		ll := sortedLevels[i]
		rr := sortedLevels[j]
		ill, errL := strconv.Atoi(ll)
		irr, errR := strconv.Atoi(rr)
		if errL == nil && errR == nil {
			return ill < irr
		}
		return ll < rr
	})
	items := make([]list.Item, 0)
	for _, v := range sortedLevels {
		title := dth.Symbol + " " + v
		n := item{
			title: title,
			desc:  "発狂BMS難易度表" + title,
		}
		items = append(items, n)
	}
	return items, sortedLevels[0]
}

func buildSongList(diffTable []difftable.DiffTable, defaultLevel string) []list.Item {

}

func newModel(dth *difftable.DiffTableHeader, diffTable []difftable.DiffTable) mainModel {
	m := mainModel{state: levelView}
	// build level list
	levelItems, defaultLevel := buildLevelList(dth, diffTable)
	m.levelList = list.New(levelItems, list.NewDefaultDelegate(), 0, 0)
	m.levelList.Title = "Levels"
	m.levelList.SetShowHelp(false)
	m.levelList.SetShowStatusBar(false)
	m.levelList.KeyMap.NextPage.Unbind()
	m.levelList.KeyMap.PrevPage.Unbind()
	// TODO: build song list
	songItems := buildSongList(diffTable, defaultLevel)
	m.songList = list.New(songItems, list.NewDefaultDelegate(), 0, 0)

	return m
}

func (m mainModel) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			if m.state == levelView {
				m.state = songView
			} else {
				m.state = levelView
			}
		case "enter":
			if m.state == levelView {
				i, ok := m.levelList.SelectedItem().(item)
				if ok {
					m.choice = i
				}
				m.state = songView
			}
		}
		switch m.state {
		// update whichever model is focused
		case levelView:
			m.levelList, cmd = m.levelList.Update(msg)
			cmds = append(cmds, cmd)
		default:
			m.songList, cmd = m.songList.Update(msg)
			cmds = append(cmds, cmd)
		}
	case tea.WindowSizeMsg:
		h, v := listStyle.GetFrameSize()
		m.w = msg.Width
		m.h = msg.Height
		m.levelList.SetSize(msg.Width-h, msg.Height-v)
		m.songList.SetSize(msg.Width-h, msg.Height-v)
	}

	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	var s string
	// model := m.currentFocusedModel()
	if m.state == levelView {
		s += lipgloss.JoinHorizontal(lipgloss.Top, m.levelList.View(), m.songList.View())
	} else {
		s += lipgloss.JoinHorizontal(lipgloss.Top, m.levelList.View(), m.songList.View())
	}
	// s += helpStyle.Render(fmt.Sprintf("\ntab: focus next • n: new %s • q: exit\n", model))
	return s
}

// Open lamp ghost tui application.
// The terminal would be split into 2 pieces:
// left is the specified difficult table's levels
// right is the related song list and lamp status
func OpenGhostTui(dth *difftable.DiffTableHeader, dt []difftable.DiffTable, songData []SongData, scoreLog []ScoreLog) {
	if _, err := tea.NewProgram(newModel(dth, dt)).Run(); err != nil {
		log.Fatal(err)
	}
}
