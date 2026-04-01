package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func initialGameState() GameStateModel {
	dungeon := CreateNewDungeon()
	cx, cy := dungeon.Rooms[0].Center()
	player := Entity{
		X:          cx,
		Y:          cy,
		Name:       "Grimmm",
		StandingOn: FLOOR,
		HP:         20,
		ATK:        5,
	}

	dungeon.Tiles[player.Y][player.X] = Tile{
		Kind: PLAYER,
	}

	return GameStateModel{
		Screen:   MENU_SCREEN,
		Dungeon:  dungeon.Tiles,
		Player:   player,
		Monsters: dungeon.Monsters,
		Items:    []any{},
		Floor:    1,
		Rooms:    dungeon.Rooms,
		Turn:     PLAYER_TURN,
		Log:      []string{"Leave behind all hope thou who enter here..."},
	}
}

func main() {
	state := initialGameState()
	p := tea.NewProgram(&state)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Well, the dungeon broke: %v", err)
		os.Exit(1)
	}
}

func (m *GameStateModel) Init() tea.Cmd {
	return nil
}

func (m *GameStateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.TermHeight = msg.Height
		m.TermWidth = msg.Width
		return m, nil
	default:
		_ = msg
	}

	switch m.Screen {
	case MENU_SCREEN:
		return m.updateMenu(msg)
	case GAME_SCREEN:
		return m.updateGame(msg)
	case DEATH_SCREEN:
		return m.updateDeath(msg)
	}

	return m, nil
}

func (m *GameStateModel) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			m.Screen = GAME_SCREEN
			return m, nil
		}
	}
	return m, nil
}

func (m *GameStateModel) updateGame(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.Turn != PLAYER_TURN {
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k", "down", "j", "left", "l", "right", "h":
			moved := m.UpdatePlayerPos(msg.String())
			if moved {
				m.Turn = WORLD_TURN
				return m, tea.Tick(time.Millisecond*10, func(t time.Time) tea.Msg {
					return PlayerTurnFinished{}
				})
			}
			return m, nil
		case "t":
			m.Turn = WORLD_TURN
			return m, tea.Tick(time.Millisecond*10, func(t time.Time) tea.Msg {
				return PlayerTurnFinished{}
			})
		case "enter", "space":
			cmd := m.PlayerAttack()
			return m, cmd
		}
	case attackFinishedMsg:
		m.PlayerIsAttacking = false
		m.Dungeon[m.Player.Y][m.Player.X].Kind = PLAYER
		m.Turn = WORLD_TURN
		return m, tea.Tick(time.Millisecond*10, func(t time.Time) tea.Msg {
			return PlayerTurnFinished{}
		})
	case PlayerTurnFinished:
		m.MoveMonsters()
		if m.Player.HP <= 0 {
			m.Screen = DEATH_SCREEN
			return m, nil
		}
		m.Turn = PLAYER_TURN
		return m, nil
	}
	return m, nil
}

func (m *GameStateModel) updateDeath(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			*m = initialGameState()
			m.Screen = GAME_SCREEN
			return m, nil
		}
	}
	return m, nil
}

func (m *GameStateModel) View() tea.View {
	var screen string

	switch m.Screen {
	case MENU_SCREEN:
		screen = m.viewMenu()
	case GAME_SCREEN:
		screen = m.viewGame()
	case DEATH_SCREEN:
		screen = m.viewDeath()
	}

	resized := lipgloss.Place(m.TermWidth, m.TermHeight, lipgloss.Center, lipgloss.Center, screen)
	final := tea.NewView(resized)
	final.AltScreen = true
	return final
}

func (m *GameStateModel) viewMenu() string {
	title := titleStyle.Render(`
  ▄████  ██▀███   ██▓ ███▄ ▄███▓ ███▄ ▄███▓
 ██▒ ▀█▒▓██ ▒ ██▒▓██▒▓██▒▀█▀ ██▒▓██▒▀█▀ ██▒
▒██░▄▄▄░▓██ ░▄█ ▒▒██▒▓██    ▓██░▓██    ▓██░
░▓█  ██▓▒██▀▀█▄  ░██░▒██    ▒██ ▒██    ▒██
░▒▓███▀▒░██▓ ▒██▒░██░▒██▒   ░██▒▒██▒   ░██▒
 ░▒   ▒ ░ ▒▓ ░▒▓░░▓  ░ ▒░   ░  ░░ ▒░   ░  ░
  ░   ░   ░▒ ░ ▒░ ▒ ░░ ░      ░ ░ ░      ░
        ░   ░  ░   ░        ░          ░`)

	subtitle := subtitleStyle.Render("D U N G E O N   C R A W L E R")
	prompt := promptStyle.Render("[ Press ENTER to descend into the darkness ]")
	quit := helpStyle.Render("q: quit")

	return lipgloss.JoinVertical(lipgloss.Center, title, "", subtitle, "", "", prompt, "", quit)
}

func (m *GameStateModel) viewDeath() string {
	skull := deathStyle.Render(`

         .AMMMMMMMMMMA.
       .AV. :::.:.:.::MA.
      A' :..        : .:'A
     A'..              . 'A.
    A' :.    :::::::::  : :'A
    M  .    :::.:.:.:::  . .M
    M  :   ::.:.....::.:   .M
    V : :.::.:........:.:  :V
   A  A:    ..:...:...:.   A A
  .V  MA:.....:M.::.::. .:AM.M
 A'  .VMMMMMMMMM:.:AMMMMMMMV: A
:M .  .'VMMMMMMV.:A 'VMMMMV .:M:
 V.:.  ..'VMMMV.:AM..'VMV' .: V
  V.  .:. .....:AMMA. . .:. .V
   VMM...: ...:.MMMM.: .: MMV
       'VM: . ..M.:M..:::M'
         'M::. .:.... .::M
          M:.  :. .... ..M
          V:  M:. M. :M .V
          'V.:M.. M. :M.V'
             'V.:.V.:.V`)

	message := deathMessageStyle.Render("YOU HAVE PERISHED")
	stats := helpStyle.Render(fmt.Sprintf("Reached floor %d", m.Floor))
	prompt := promptStyle.Render("[ Press ENTER to try again ]")
	quit := helpStyle.Render("q: quit")

	return lipgloss.JoinVertical(lipgloss.Center, skull, "", message, "", stats, "", prompt, "", quit)
}

func (m *GameStateModel) viewGame() string {
	// ---------------------------------------------------------
	// 1. BUILD THE MAP
	// ---------------------------------------------------------
	mapSb := strings.Builder{}
	for _, row := range m.Dungeon {
		for _, tile := range row {
			char := string(tile.Kind)

			// UI Trick: We can stub the Fog of War here!
			// If you later set Explored to false upon generation, this will hide them.
			if !tile.Explored {
				// Uncomment this when you want to enable pure darkness
				// mapSb.WriteString(" ")
				// continue
			}

			switch tile.Kind {
			case PLAYER:
				mapSb.WriteString(playerStyle.Render(char))
			case PLAYER_ATTACKING:
				mapSb.WriteString(attackStyle.Render(char))
			case MONSTER_ATTACKING:
				mapSb.WriteString(attackStyle.Render(char))
			case MONSTER:
				mapSb.WriteString(monsterStyle.Render(char))
			case TOUCHING_WALL:
				mapSb.WriteString(touchingWallStyle.Render(char))
			case WALL, FILLER_WALL, PASSAGE_WALL:
				mapSb.WriteString(wallStyle.Render(char))
			case FLOOR:
				mapSb.WriteString(floorStyle.Render(char))
			case STAIRS:
				mapSb.WriteString(stairStyle.Render(char))
			default:
				mapSb.WriteString(char)
			}
		}
		mapSb.WriteString("\n")
	}
	mapPanel := mapBoxStyle.Render(mapSb.String())

	// ---------------------------------------------------------
	// 2. BUILD THE SIDEBAR (Stats & Info)
	// ---------------------------------------------------------
	var hpRendered string
	if m.Player.HP > 10 {
		hpRendered = hpHighStyle.Render(fmt.Sprintf("%d", m.Player.HP))
	} else if m.Player.HP > 5 {
		hpRendered = hpMedStyle.Render(fmt.Sprintf("%d", m.Player.HP))
	} else {
		hpRendered = hpLowStyle.Render(fmt.Sprintf("%d", m.Player.HP))
	}

	sidebarSb := strings.Builder{}
	sidebarSb.WriteString(headerStyle.Render(m.Player.Name) + "\n\n")
	sidebarSb.WriteString(fmt.Sprintf("Floor:  %d\n", m.Floor))
	sidebarSb.WriteString(fmt.Sprintf("Turn:   %s\n", m.Turn))
	sidebarSb.WriteString(fmt.Sprintf("Health: %s / 20\n", hpRendered))
	sidebarSb.WriteString(fmt.Sprintf("Attack: %d\n\n", m.Player.ATK))

	sidebarSb.WriteString(headerStyle.Render("Inventory") + "\n")
	sidebarSb.WriteString("- (Empty)\n")

	sidebarPanel := sidebarStyle.Render(sidebarSb.String())

	// ---------------------------------------------------------
	// 3. BUILD THE MESSAGE LOG
	// ---------------------------------------------------------
	logCount := len(m.Log)
	start := max(0, logCount-3)
	recentLogs := strings.Join(m.Log[start:], "\n")
	logPanel := logBoxStyle.Width(WIDTH + 26).Render(recentLogs)

	// ---------------------------------------------------------
	// 4. ASSEMBLE THE FINAL UI
	// ---------------------------------------------------------
	topSection := lipgloss.JoinHorizontal(lipgloss.Top, mapPanel, sidebarPanel)
	fullScreen := lipgloss.JoinVertical(lipgloss.Left, topSection, logPanel)
	helpText := helpStyle.Render(" movement: h j k l / arrows • q: quit • enter/space: attack • t: wait")

	return lipgloss.JoinVertical(lipgloss.Left, fullScreen, helpText)
}
