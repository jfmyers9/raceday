package ui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jfmyers/tmux-raceday/internal/f1"
	"github.com/jfmyers/tmux-raceday/internal/nascar"
	"github.com/jfmyers/tmux-raceday/internal/series"
	"github.com/jfmyers/tmux-raceday/internal/weather"
)

var (
	headerStyle = lipgloss.NewStyle().Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("236"))

	rowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	favStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("202"))

	flagStyles = map[int]lipgloss.Style{
		nascar.FlagGreen:   lipgloss.NewStyle().Foreground(lipgloss.Color("46")),
		nascar.FlagCaution: lipgloss.NewStyle().Foreground(lipgloss.Color("226")),
		nascar.FlagRed:     lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		nascar.FlagWhite:   lipgloss.NewStyle().Foreground(lipgloss.Color("15")),
	}

	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("248"))
)

type tickMsg time.Time
type feedMsg *nascar.LiveFeed
type weatherMsg *weather.Conditions
type weatherTickMsg time.Time
type errMsg error

// Series modes
const (
	SeriesNASCAR = iota
	SeriesF1
)

// View modes
const (
	ViewLeaderboard = iota
	ViewSchedule
	ViewEntryList
	ViewStandings
	ViewF1Leaderboard
	ViewF1Schedule
)

type Model struct {
	feed       *nascar.LiveFeed
	race       *nascar.Race // schedule data for weekend view
	standings  []nascar.PointsEntry
	weather    *weather.Conditions
	favDriver  string
	cursor     int
	offset     int
	width      int
	height     int
	err        error
	sortCol    int
	sortAsc    bool
	searchMode bool
	searchTerm string
	quitting   bool
	activeView   int
	activeSeries int
	seriesLocked bool // true once user manually switches series
	f1Live       *series.LiveState
	f1Schedule   []series.Race
}

func NewModel(driverNum int) Model {
	fav := ""
	if driverNum > 0 {
		fav = strconv.Itoa(driverNum)
	}
	return Model{
		favDriver: fav,
		sortCol:   0, // position
		sortAsc:   true,
	}
}

type scheduleMsg *nascar.Race
type standingsMsg []nascar.PointsEntry
type f1LiveStateMsg struct{ state *series.LiveState }
type f1ScheduleMsg struct{ races []series.Race }

func (m Model) Init() tea.Cmd {
	return tea.Batch(fetchFeed, fetchSchedule, fetchStandings, fetchF1Live, fetchF1Schedule, tickCmd(), weatherTickCmd())
}

func fetchSchedule() tea.Msg {
	year := time.Now().Year()
	races, err := nascar.FetchCupSchedule(year)
	if err != nil {
		return errMsg(err)
	}
	race := nascar.NextRace(races)
	if race == nil {
		return scheduleMsg(nil)
	}
	return scheduleMsg(race)
}

func fetchStandings() tea.Msg {
	entries, err := nascar.FetchStandings()
	if err != nil {
		return errMsg(err)
	}
	return standingsMsg(entries)
}

func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func fetchFeed() tea.Msg {
	feed, err := nascar.FetchLiveFeed()
	if err != nil {
		return errMsg(err)
	}
	return feedMsg(feed)
}

func fetchF1Live() tea.Msg {
	s := f1.NewSeries()
	state, _ := s.FetchLiveState()
	return f1LiveStateMsg{state: state}
}

func fetchF1Schedule() tea.Msg {
	s := f1.NewSeries()
	races, _ := s.FetchSchedule(time.Now().Year())
	return f1ScheduleMsg{races: races}
}

func weatherTickCmd() tea.Cmd {
	return tea.Tick(5*time.Minute, func(t time.Time) tea.Msg {
		return weatherTickMsg(t)
	})
}

func fetchWeatherCmd(lat, lon float64) tea.Cmd {
	return func() tea.Msg {
		if lat == 0 && lon == 0 {
			return weatherMsg(nil)
		}
		cond, err := weather.FetchCurrent(lat, lon)
		if err != nil {
			return weatherMsg(nil)
		}
		return weatherMsg(cond)
	}
}

func (m Model) weatherCoords() (float64, float64) {
	if m.feed != nil && m.feed.TrackID != 0 {
		lat, lon, ok := nascar.TrackCoords(m.feed.TrackID)
		if ok {
			return lat, lon
		}
	}
	if m.f1Live != nil {
		return m.f1Live.Lat, m.f1Live.Lon
	}
	if m.race != nil && m.race.TrackID != 0 {
		lat, lon, ok := nascar.TrackCoords(m.race.TrackID)
		if ok {
			return lat, lon
		}
	}
	return 0, 0
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		return m, tea.Batch(fetchFeed, fetchF1Live, tickCmd())

	case feedMsg:
		feed := (*nascar.LiveFeed)(msg)
		m.feed = feed
		m.err = nil
		if feed == nil || !feed.IsLiveCupRace() {
			if m.f1Live == nil {
				m.weather = nil
			}
		} else if m.weather == nil && feed.TrackID != 0 {
			lat, lon, ok := nascar.TrackCoords(feed.TrackID)
			if ok {
				return m, fetchWeatherCmd(lat, lon)
			}
		}

	case weatherMsg:
		m.weather = msg

	case weatherTickMsg:
		lat, lon := m.weatherCoords()
		if lat != 0 && lon != 0 && m.hasLiveRace() {
			return m, tea.Batch(fetchWeatherCmd(lat, lon), weatherTickCmd())
		}
		return m, weatherTickCmd()

	case scheduleMsg:
		m.race = msg

	case standingsMsg:
		m.standings = msg

	case f1LiveStateMsg:
		m.f1Live = msg.state
		if !m.seriesLocked {
			m.autoDetectSeries()
		}

	case f1ScheduleMsg:
		m.f1Schedule = msg.races

	case errMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.searchMode {
			return m.handleSearchKey(msg)
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		m.quitting = true
		return m, tea.Quit
	case key.Matches(msg, keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
		if m.cursor < m.offset {
			m.offset = m.cursor
		}
	case key.Matches(msg, keys.Down):
		max := m.vehicleCount() - 1
		if m.cursor < max {
			m.cursor++
		}
		visible := m.visibleRows()
		if m.cursor >= m.offset+visible {
			m.offset = m.cursor - visible + 1
		}
	case key.Matches(msg, keys.Search):
		m.searchMode = true
		m.searchTerm = ""
	case key.Matches(msg, keys.Tab):
		m.sortCol = (m.sortCol + 1) % 4
	case key.Matches(msg, keys.GotoFav):
		m.jumpToFavorite()
	case key.Matches(msg, keys.SwitchSeries):
		m.switchSeries()
	case key.Matches(msg, keys.View1):
		if m.activeSeries == SeriesF1 {
			m.activeView = ViewF1Leaderboard
		} else {
			m.activeView = ViewLeaderboard
		}
	case key.Matches(msg, keys.View2):
		if m.activeSeries == SeriesF1 {
			m.activeView = ViewF1Schedule
		} else {
			m.activeView = ViewSchedule
		}
	case key.Matches(msg, keys.View3):
		if m.activeSeries == SeriesNASCAR {
			m.activeView = ViewEntryList
		}
	case key.Matches(msg, keys.View4):
		if m.activeSeries == SeriesNASCAR {
			m.activeView = ViewStandings
		}
	}
	return m, nil
}

func (m *Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.searchMode = false
		m.searchTerm = ""
	case tea.KeyEnter:
		m.searchMode = false
		m.findDriver(m.searchTerm)
	case tea.KeyBackspace:
		if len(m.searchTerm) > 0 {
			m.searchTerm = m.searchTerm[:len(m.searchTerm)-1]
		}
	default:
		if len(msg.Runes) > 0 {
			m.searchTerm += string(msg.Runes)
		}
	}
	return m, nil
}

var keys = struct {
	Quit         key.Binding
	Up           key.Binding
	Down         key.Binding
	Search       key.Binding
	Tab          key.Binding
	GotoFav      key.Binding
	SwitchSeries key.Binding
	View1        key.Binding
	View2        key.Binding
	View3        key.Binding
	View4        key.Binding
}{
	Quit:         key.NewBinding(key.WithKeys("q", "ctrl+c")),
	Up:           key.NewBinding(key.WithKeys("k", "up")),
	Down:         key.NewBinding(key.WithKeys("j", "down")),
	Search:       key.NewBinding(key.WithKeys("/")),
	Tab:          key.NewBinding(key.WithKeys("tab")),
	GotoFav:      key.NewBinding(key.WithKeys("f")),
	SwitchSeries: key.NewBinding(key.WithKeys("s")),
	View1:        key.NewBinding(key.WithKeys("1")),
	View2:        key.NewBinding(key.WithKeys("2")),
	View3:        key.NewBinding(key.WithKeys("3")),
	View4:        key.NewBinding(key.WithKeys("4")),
}

func (m *Model) jumpToFavorite() {
	if m.feed == nil || m.favDriver == "" {
		return
	}
	vehicles := m.sortedVehicles()
	for i, v := range vehicles {
		if v.VehicleNumber == m.favDriver {
			m.cursor = i
			visible := m.visibleRows()
			m.offset = m.cursor - visible/2
			if m.offset < 0 {
				m.offset = 0
			}
			return
		}
	}
}

func (m *Model) findDriver(term string) {
	if m.feed == nil {
		return
	}
	term = strings.ToLower(term)
	vehicles := m.sortedVehicles()
	for i, v := range vehicles {
		if strings.Contains(strings.ToLower(v.Driver.FullName), term) ||
			v.VehicleNumber == term {
			m.cursor = i
			visible := m.visibleRows()
			m.offset = m.cursor - visible/2
			if m.offset < 0 {
				m.offset = 0
			}
			return
		}
	}
}

func (m Model) hasLiveRace() bool {
	return (m.feed != nil && m.feed.IsLiveCupRace()) || m.f1Live != nil
}

func (m *Model) autoDetectSeries() {
	if m.f1Live != nil {
		m.activeSeries = SeriesF1
		m.activeView = ViewF1Leaderboard
	} else if m.feed != nil && m.feed.IsLiveCupRace() {
		m.activeSeries = SeriesNASCAR
		m.activeView = ViewLeaderboard
	}
}

func (m *Model) switchSeries() {
	m.seriesLocked = true
	if m.activeSeries == SeriesNASCAR {
		m.activeSeries = SeriesF1
		m.activeView = ViewF1Leaderboard
	} else {
		m.activeSeries = SeriesNASCAR
		m.activeView = ViewLeaderboard
	}
}

func (m Model) vehicleCount() int {
	if m.feed == nil {
		return 0
	}
	return len(m.feed.Vehicles)
}

func (m Model) visibleRows() int {
	// header bar (3 lines) + column header (1) + status bar (1)
	return m.height - 5
}

func (m Model) sortedVehicles() []nascar.Vehicle {
	if m.feed == nil {
		return nil
	}
	vehicles := make([]nascar.Vehicle, len(m.feed.Vehicles))
	copy(vehicles, m.feed.Vehicles)

	sort.SliceStable(vehicles, func(i, j int) bool {
		var less bool
		switch m.sortCol {
		case 0: // position
			less = vehicles[i].RunningPosition < vehicles[j].RunningPosition
		case 1: // car number
			ni, _ := strconv.Atoi(vehicles[i].VehicleNumber)
			nj, _ := strconv.Atoi(vehicles[j].VehicleNumber)
			less = ni < nj
		case 2: // delta
			less = vehicles[i].Delta < vehicles[j].Delta
		case 3: // laps led
			less = vehicles[i].TotalLapsLed() > vehicles[j].TotalLapsLed()
		}
		if !m.sortAsc {
			return !less
		}
		return less
	})
	return vehicles
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var content string
	switch m.activeView {
	case ViewF1Leaderboard:
		content = renderF1LeaderboardView(m.f1Live, m.width)
	case ViewF1Schedule:
		content = renderF1ScheduleView(m.f1Schedule, m.width)
	case ViewSchedule:
		content = renderScheduleView(m.race, m.weather, m.width)
	case ViewEntryList:
		content = renderEntryListView(m.feed, m.favDriver, m.width)
	case ViewStandings:
		content = renderStandingsView(m.standings, m.favDriver, m.width)
	default:
		if m.feed == nil {
			if m.err != nil {
				content = fmt.Sprintf("Error: %v\nPress q to quit.", m.err)
			} else {
				content = "Loading race data..."
			}
		} else if !m.feed.IsLiveCupRace() {
			content = "No live Cup Series race."
		} else {
			content = m.renderLeaderboard()
		}
	}

	return content + "\n" + m.renderStatusBar()
}

func (m Model) renderLeaderboard() string {
	var b strings.Builder

	b.WriteString(m.renderTitle())
	b.WriteString("\n\n")

	b.WriteString(m.renderColumnHeaders())
	b.WriteString("\n")

	vehicles := m.sortedVehicles()
	visible := m.visibleRows()
	if visible < 1 {
		visible = 10
	}

	end := m.offset + visible
	if end > len(vehicles) {
		end = len(vehicles)
	}

	for i := m.offset; i < end; i++ {
		v := vehicles[i]
		b.WriteString(m.renderRow(v, i == m.cursor))
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) renderTitle() string {
	flag := nascar.FlagSymbol(m.feed.FlagState)
	flagName := flagStateName(m.feed.FlagState)

	title := fmt.Sprintf("%s %s — %s", flag, m.feed.RunName, m.feed.TrackName)
	title = titleStyle.Render(title)

	stageLapsLeft := m.feed.Stage.FinishAtLap - m.feed.LapNumber
	if stageLapsLeft < 0 {
		stageLapsLeft = 0
	}

	info := fmt.Sprintf("Lap %d/%d | Stage %d (%d laps to stage end) | %s | %d cautions (%d laps) | %d lead changes",
		m.feed.LapNumber, m.feed.LapsInRace,
		m.feed.Stage.StageNum, stageLapsLeft,
		flagName,
		m.feed.NumberOfCautions, m.feed.NumberOfCautionLaps, m.feed.NumberOfLeadChanges)

	if m.weather != nil {
		info += fmt.Sprintf(" | %.0f°F (feels %.0f°F) %s %.0fmph %s",
			m.weather.Temp, m.weather.FeelsLike,
			weather.Symbol(m.weather.WeatherCode),
			m.weather.WindSpeed,
			weather.WindDirectionArrow(m.weather.WindDirection))
	}

	return title + "\n" + dimStyle.Render(info)
}

func flagStateName(state int) string {
	switch state {
	case nascar.FlagGreen:
		return "GREEN"
	case nascar.FlagCaution:
		return "CAUTION"
	case nascar.FlagRed:
		return "RED"
	case nascar.FlagWhite:
		return "WHITE"
	case nascar.FlagFinished:
		return "CHECKERED"
	default:
		return "UNKNOWN"
	}
}

func (m Model) renderColumnHeaders() string {
	cols := []struct {
		label string
		width int
	}{
		{"POS", 4},
		{"#", 4},
		{"DRIVER", 22},
		{"DELTA", 8},
		{"LAPS", 5},
		{"LED", 4},
		{"SPEED", 7},
		{"PITS", 5},
		{"TIRE", 5},
		{"+/-", 4},
		{"MOV", 4},
		{"STATUS", 7},
	}

	sortIndicators := []string{"POS", "#", "DELTA", "LED"}

	var parts []string
	for i, c := range cols {
		label := c.label
		if i < len(sortIndicators) && i == m.sortCol {
			if m.sortAsc {
				label += "▲"
			} else {
				label += "▼"
			}
		}
		parts = append(parts, fmt.Sprintf("%-*s", c.width, label))
	}

	return headerStyle.Render(strings.Join(parts, " "))
}

func (m Model) renderRow(v nascar.Vehicle, selected bool) string {
	diff := v.RunningPosition - v.StartingPosition
	diffStr := " "
	if diff > 0 {
		diffStr = fmt.Sprintf("+%d", diff)
	} else if diff < 0 {
		diffStr = fmt.Sprintf("%d", diff)
	}

	statusStr := "RUN"
	if !v.IsOnTrack {
		statusStr = "OUT"
	}
	if v.IsOnDVP {
		statusStr = "DVP"
	}
	if v.Status != 1 {
		statusStr = "DNF"
	}

	pitCount := v.PitCount()
	tireLaps := ""
	if m.feed != nil {
		tl := v.LapsSinceLastPit(m.feed.LapNumber)
		tireLaps = strconv.Itoa(tl)
	}

	movStr := " "
	if v.PassingDiff > 0 {
		movStr = fmt.Sprintf("+%d", v.PassingDiff)
	} else if v.PassingDiff < 0 {
		movStr = fmt.Sprintf("%d", v.PassingDiff)
	}

	row := fmt.Sprintf("%-4d %-4s %-22s %7.3f %-5d %-4d %6.1f %-5d %-5s %-4s %-4s %-7s",
		v.RunningPosition,
		v.VehicleNumber,
		truncate(v.Driver.FullName, 22),
		v.Delta,
		v.LapsCompleted,
		v.TotalLapsLed(),
		v.LastLapSpeed,
		pitCount,
		tireLaps,
		diffStr,
		movStr,
		statusStr,
	)

	isFav := v.VehicleNumber == m.favDriver

	switch {
	case selected && isFav:
		return lipgloss.NewStyle().
			Background(lipgloss.Color("58")).
			Foreground(lipgloss.Color("226")).
			Bold(true).
			Render(row)
	case selected:
		return lipgloss.NewStyle().
			Background(lipgloss.Color("237")).
			Foreground(lipgloss.Color("15")).
			Render(row)
	case isFav:
		return favStyle.Render(row)
	case v.Status != 1 || v.IsOnDVP:
		return dimStyle.Render(row)
	default:
		return rowStyle.Render(row)
	}
}

func (m Model) renderStatusBar() string {
	var viewTabs []string
	if m.activeSeries == SeriesF1 {
		viewTabs = []string{"1:Race", "2:Calendar"}
		tabToView := []int{ViewF1Leaderboard, ViewF1Schedule}
		for i := range viewTabs {
			if m.activeView == tabToView[i] {
				viewTabs[i] = "[" + viewTabs[i] + "]"
			}
		}
	} else {
		viewTabs = []string{"1:Race", "2:Schedule", "3:Entry", "4:Standings"}
		for i := range viewTabs {
			if i == m.activeView {
				viewTabs[i] = "[" + viewTabs[i] + "]"
			}
		}
	}

	seriesLabel := "NASCAR"
	if m.activeSeries == SeriesF1 {
		seriesLabel = "F1"
	}

	left := fmt.Sprintf("[%s] %s  s:series  q:quit", seriesLabel, strings.Join(viewTabs, " "))
	if m.activeView == ViewLeaderboard {
		left += "  j/k:scroll  /:search  tab:sort  f:fav"
	}
	if m.searchMode {
		left = fmt.Sprintf("Search: %s█", m.searchTerm)
	}

	right := ""
	if m.activeSeries == SeriesNASCAR && m.favDriver != "" && m.feed != nil {
		if v := m.feed.FindDriver(m.favDriver); v != nil {
			right = fmt.Sprintf("#%s %s P%d", v.VehicleNumber, v.Driver.LastName, v.RunningPosition)
		}
	}

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 1
	}

	return statusBarStyle.Render(left + strings.Repeat(" ", gap) + right)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
