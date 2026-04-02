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
	dungeon := CreateNewDungeon(0)
	cx, cy := dungeon.Rooms[0].Center()
	player := Entity{
		X:          cx,
		Y:          cy,
		Name:       "Grimmm",
		MoveSpeed:  1,
		StandingOn: FLOOR,
		HP:         20,
		ATK:        5,
	}

	dungeon.Tiles[player.Y][player.X] = Tile{
		Kind: PLAYER,
	}

	return GameStateModel{
		Screen:          MENU_SCREEN,
		Dungeon:         dungeon.Tiles,
		Player:          player,
		Monsters:        dungeon.Monsters,
		Items:           []any{},
		Floor:           1,
		Rooms:           dungeon.Rooms,
		Turn:            PLAYER_TURN,
		PlayerDirection: DOWN,
		Log:             []string{"Leave behind all hope thou who enter here..."},
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
		if m.Turn != PLAYER_TURN || m.StepsRemaining > 0 {
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k", "down", "j", "left", "l", "right", "h":
			m.MoveInput = msg.String()
			m.StepsRemaining = m.Player.MoveSpeed

			// Take the first step immediately
			endTile, moved := m.MovePlayerOneStep(m.MoveInput)
			m.StepsRemaining--

			if !moved {
				m.StepsRemaining = 0
				return m, nil
			}

			if endTile == STAIRS {
				m.StepsRemaining = 0
				m.HandleStairs()
				return m, nil
			}

			// If more steps, chain the next one
			if m.StepsRemaining > 0 {
				return m, tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
					return MovementStepMsg{}
				})
			}

			// All steps done, pass to world turn
			m.Turn = WORLD_TURN
			return m, tea.Tick(time.Millisecond*10, func(t time.Time) tea.Msg {
				return PlayerTurnFinished{}
			})
		// case "shift + up", "K", "shift + down", "J", "shift + left", "L", "shift +right", "H":
		// 	m.MoveInput = msg.String()
		// 	m.StepsRemaining = m.Player.MoveSpeed + 1
		//
		// 	// Take the first step immediately
		// 	endTile, moved := m.MovePlayerOneStep(m.MoveInput)
		// 	m.StepsRemaining--
		//
		// 	if !moved {
		// 		m.StepsRemaining = 0
		// 		return m, nil
		// 	}
		//
		// 	if endTile == STAIRS {
		// 		m.StepsRemaining = 0
		// 		m.HandleStairs()
		// 		return m, nil
		// 	}
		//
		// 	// If more steps, chain the next one
		// 	if m.StepsRemaining > 0 {
		// 		return m, tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		// 			return MovementStepMsg{}
		// 		})
		// 	}
		//
		// 	// All steps done, pass to world turn
		// 	m.Turn = WORLD_TURN
		// 	return m, tea.Tick(time.Millisecond*10, func(t time.Time) tea.Msg {
		// 		return PlayerTurnFinished{}
		// 	})
		case "t":
			m.Turn = WORLD_TURN
			return m, tea.Tick(time.Millisecond*10, func(t time.Time) tea.Msg {
				return PlayerTurnFinished{}
			})
		case "enter", "space":
			cmd := m.PlayerAttack()
			return m, cmd
		}
	case MovementStepMsg:
		endTile, moved := m.MovePlayerOneStep(m.MoveInput)
		m.StepsRemaining--

		if !moved || endTile == STAIRS {
			m.StepsRemaining = 0
		}

		if endTile == STAIRS {
			m.HandleStairs()
			return m, nil
		}

		// More steps remaining, keep going
		if m.StepsRemaining > 0 && moved {
			return m, tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
				return MovementStepMsg{}
			})
		}

		// Done moving, world turn
		m.Turn = WORLD_TURN
		return m, tea.Tick(time.Millisecond*10, func(t time.Time) tea.Msg {
			return PlayerTurnFinished{}
		})
	case attackPhaseMsg:
		switch msg.Phase {
		case 1:
			m.AttackPhase = 2
			return m, tea.Tick(time.Millisecond*60, func(t time.Time) tea.Msg {
				return attackPhaseMsg{Phase: 2}
			})
		case 2:
			m.AttackPhase = 0
			m.MonsterFlashPos = nil
			m.PlayerIsAttacking = false
			m.Dungeon[m.Player.Y][m.Player.X].Kind = PLAYER
			m.Turn = WORLD_TURN
			return m, tea.Tick(time.Millisecond*10, func(t time.Time) tea.Msg {
				return PlayerTurnFinished{}
			})
		}
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

const (
	viewPortH = 40
	viewPortW = 60
)

func (m *GameStateModel) viewGame() string {
	// ---------------------------------------------------------
	// 1. BUILD THE MAP
	// ---------------------------------------------------------
	// Camera settings:
	vpW := min(viewPortW, len(m.Dungeon[0]), m.TermWidth-30)
	vpH := min(viewPortH, len(m.Dungeon), m.TermHeight-8)
	camX := m.Player.X - vpW/2
	camY := m.Player.Y - vpH/2

	camX = max(0, min(camX, len(m.Dungeon[0])-vpW))
	camY = max(0, min(camY, len(m.Dungeon)-vpH))

	mapSb := strings.Builder{}
	for y := camY; y < camY+vpH; y++ {
		for x := camX; x < camX+vpW; x++ {
			tile := m.Dungeon[y][x]
			char := string(tile.Kind)

			// Slash overlay / monster flash during attack animation
			if m.AttackPhase > 0 && x == m.AttackSlashPos.X && y == m.AttackSlashPos.Y {
				if m.MonsterFlashPos != nil && x == m.MonsterFlashPos.X && y == m.MonsterFlashPos.Y {
					// Monster hit — render flashed monster glyph
					glyph := string(tile.Kind)
					for _, mon := range m.Monsters {
						if mon.X == x && mon.Y == y {
							glyph = string(mon.Glyph)
						}
					}
					mapSb.WriteString(monsterFlashStyle.Render(glyph))
				} else {
					// Empty space — render directional slash
					var slashGlyph string
					switch m.PlayerDirection {
					case UP, DOWN:
						slashGlyph = "|"
					case LEFT, RIGHT:
						slashGlyph = "-"
					}
					mapSb.WriteString(slashStyle.Render(slashGlyph))
				}
				continue
			}

			// Fog of War stub
			if !tile.Explored {
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
				glyph := "S"
				for _, mon := range m.Monsters {
					if mon.X == tile.X && mon.Y == tile.Y {
						glyph = string(mon.Glyph)
					}
				}
				mapSb.WriteString(monsterStyle.Render(glyph))
			case TOUCHING_WALL:
				mapSb.WriteString(touchingWallStyle.Render(char))
			case FILLER_WALL:
				mapSb.WriteString(fillerWallStyle.Render(char))
			case WALL:
				mapSb.WriteString(wallStyle.Render(char))
			case FLOOR:
				mapSb.WriteString(floorStyle.Render(char))
			case DEAD_MONSTER:
				if m.MonsterFlashPos != nil && tile.X == m.MonsterFlashPos.X && tile.Y == m.MonsterFlashPos.Y {
					mapSb.WriteString(monsterFlashStyle.Render(char))
				} else {
					mapSb.WriteString(monsterStyle.Render(char))
				}
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
	fmt.Fprintf(&sidebarSb, "Floor:  %d\n", m.Floor)
	fmt.Fprintf(&sidebarSb, "Floor Size:  %d x %d\n", len(m.Dungeon), len(m.Dungeon[0]))
	fmt.Fprintf(&sidebarSb, "Turn:   %s\n", m.Turn)
	fmt.Fprintf(&sidebarSb, "Health: %s / 20\n", hpRendered)
	fmt.Fprintf(&sidebarSb, "Attack: %d\n\n", m.Player.ATK)

	sidebarSb.WriteString(headerStyle.Render("Inventory") + "\n")
	sidebarSb.WriteString("- (Empty)\n")

	sidebarPanel := sidebarStyle.Render(sidebarSb.String())

	// ---------------------------------------------------------
	// 3. BUILD THE MESSAGE LOG
	// ---------------------------------------------------------
	// ---------------------------------------------------------
	// 3. ASSEMBLE TOP SECTION
	// ---------------------------------------------------------
	topSection := lipgloss.JoinHorizontal(lipgloss.Top, mapPanel, sidebarPanel)

	// ---------------------------------------------------------
	// 4. BUILD THE MESSAGE LOG (sized to match top section)
	// ---------------------------------------------------------
	logCount := len(m.Log)
	start := max(0, logCount-3)
	recentLogs := strings.Join(m.Log[start:], "\n")
	logPanel := logBoxStyle.Width(lipgloss.Width(topSection) - 4).Render(recentLogs)
	fullScreen := lipgloss.JoinVertical(lipgloss.Left, topSection, logPanel)
	helpText := helpStyle.Render(" movement: h j k l / arrows • q: quit • enter/space: attack • t: wait")

	return lipgloss.JoinVertical(lipgloss.Left, fullScreen, helpText)
}
