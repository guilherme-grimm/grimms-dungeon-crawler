package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// STYLES

// --- STYLES ---
// --- STYLES ---
var (
	// Entity Styles
	playerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFAA")).Bold(true)
	attackStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA00")).Bold(true)
	monsterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0044")).Bold(true)
	stairStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Bold(true)

	// Environment Styles
	wallStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#444455"))
	floorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#222222"))

	// Health Colors
	hpHighStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
	hpMedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Bold(true)
	hpLowStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true).Blink(true)

	// Layout Containers
	mapBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#777777")).
			Padding(0, 1)

	sidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#555555")).
			Width(22). // Fixed width for the sidebar
			Padding(0, 1)

	logBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444455")).
			Padding(0, 1)

	// Text Styles
	headerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Bold(true).Underline(true)
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Italic(true)
)

type TurnType string

var (
	PLAYER_TURN TurnType = "PLAYER"
	WORLD_TURN  TurnType = "WORLD"
)

type TileKind rune

var (
	PLAYER           TileKind = '@'
	PLAYER_ATTACKING TileKind = 'A'
	FLOOR            TileKind = '·'
	WALL             TileKind = '#'
	TOUCHING_WALL    TileKind = '□'
	FILLER_WALL      TileKind = '█'
	PASSAGE_WALL     TileKind = '='
	STAIRS           TileKind = '▼'
	MONSTER          TileKind = 'S'
)

const (
	WIDTH  = 80
	HEIGHT = 24
)

type model int

type tickMsg time.Time

type Tile struct {
	Kind     TileKind // Wall, Floor, Player
	Visible  bool
	Explored bool
	X, Y     int
}

type Entity struct {
	X, Y    int
	Glyph   rune
	Name    string
	HP, ATK int
}

type Dungeon struct {
	Tiles    [][]Tile
	Rooms    []Room
	Monsters []Entity
}

type GameStateModel struct {
	// Instance state
	IsPlaying bool

	//Game states
	Dungeon  [][]Tile
	Player   Entity
	Monsters []Entity
	Items    []any
	Floor    int
	Log      []string
	Rooms    []Room
	Turn     TurnType

	//Player state
	IsAttacking bool
}

type Room struct {
	X, Y, W, H int
}

func (r Room) Center() (int, int) {
	return r.X + r.W/2, r.Y + r.H/2
}

func (r Room) Overlaps(other Room) bool {
	return !(r.X+r.W+1 <= other.X ||
		other.X+other.W+1 <= r.X ||
		r.Y+r.H <= other.Y ||
		other.Y+other.H+1 <= r.Y)
}

func main() {
	state := initialGameState()
	p := tea.NewProgram(&state)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Well, the dungeon broke: %v", err)
		os.Exit(1)
	}
}

/*
Creates a new dungeon floor
*/
func CreateNewDungeon() Dungeon {
	tiles := make([][]Tile, HEIGHT)
	// Fill with Walls
	for y := range HEIGHT {
		tiles[y] = make([]Tile, WIDTH)
		for x := range WIDTH {
			kind := WALL

			tiles[y][x] = Tile{
				Kind:    kind,
				Visible: true,
				X:       x,
				Y:       y,
			}
		}
	}

	rooms := make([]Room, 0)

	for i := 0; i <= 40; i++ {
		w := rand.IntN(8-5) + 5
		h := rand.IntN(7-4) + 4
		x := rand.IntN(WIDTH-w-1) + 1
		y := rand.IntN(HEIGHT-h-1) + 1

		newRoom := Room{x, y, w, h}

		if slices.ContainsFunc(rooms, newRoom.Overlaps) {
			continue
		}

		if len(rooms) >= 10 {
			break
		}

		rooms = append(rooms, newRoom)
	}

	for _, room := range rooms {
		for y := room.Y; y < room.Y+room.H; y++ {
			for x := room.X; x < room.X+room.W; x++ {
				tiles[y][x].Kind = FLOOR
			}
		}
	}

	for i := 1; i < len(rooms); i++ {
		cx1, cy1 := rooms[i-1].Center()
		cx2, cy2 := rooms[i].Center()

		for x := min(cx1, cx2); x <= max(cx1, cx2); x++ {
			tiles[cy1][x].Kind = FLOOR
		}
		for y := min(cy1, cy2); y <= max(cy1, cy2); y++ {
			tiles[y][cx2].Kind = FLOOR
		}
	}

	// up, down, left, right
	dirs := [][]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	// POST PROCESSING
	for row := range HEIGHT {
		for col := range WIDTH {
			curr := tiles[row][col]
			if curr.Kind == WALL {
				for _, d := range dirs {
					newRow := row + d[0]
					newCol := col + d[1]

					if newRow < 0 || newRow >= HEIGHT || newCol < 0 || newCol >= WIDTH {
						continue
					}
					next := tiles[newRow][newCol]
					if next.Kind == FLOOR {
						tiles[row][col].Kind = TOUCHING_WALL
					}
					switch next.Kind {
					case FLOOR:
						tiles[row][col].Kind = TOUCHING_WALL
					case WALL:
						chance := rand.Int() * 100
						if chance/2 == 0 {
							tiles[row][col].Kind = FILLER_WALL
						}
					}
				}
			}
		}
	}

	dungeon := Dungeon{
		Tiles: tiles,
		Rooms: rooms,
	}
	lastRoom := rooms[len(rooms)-1]

	sx, sy := lastRoom.Center()

	tiles[sy][sx].Kind = STAIRS

	// DOC: Monster spawning logic
	monsters := make([]Entity, 0)
	for i := 1; i < len(rooms); i++ {
		room := rooms[i]
		numMonsters := rand.IntN(3) + 1

		for j := 0; j < numMonsters; j++ {
			mx := room.X + rand.IntN(room.W)
			my := room.Y + rand.IntN(room.H)

			if tiles[my][mx].Kind == FLOOR {
				tiles[my][mx].Kind = MONSTER

				monsters = append(monsters, Entity{
					X:     mx,
					Y:     my,
					Glyph: rune(MONSTER),
					Name:  "Snek",
					HP:    10,
					ATK:   2,
				})
			}
		}
	}

	dungeon.Monsters = monsters

	return dungeon
}

func initialGameState() GameStateModel {
	dungeon := CreateNewDungeon()
	cx, cy := dungeon.Rooms[0].Center()
	player := Entity{
		X:     cx,
		Y:     cy,
		Glyph: '@',
		Name:  "Grimmm",
		HP:    20,
		ATK:   5,
	}

	dungeon.Tiles[player.Y][player.X] = Tile{
		Kind: PLAYER,
	}

	return GameStateModel{
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

func (m *GameStateModel) Init() tea.Cmd {
	return nil
}

func (m *GameStateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				return m, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
					return PlayerTurnFinished{}
				})

			}
			return m, nil
		case "enter", "space":
			cmd := m.PlayerAttack()
			return m, cmd
		}
	case attackFinishedMsg:
		m.IsAttacking = false
		m.Dungeon[m.Player.Y][m.Player.X].Kind = PLAYER
		return m, nil
	case PlayerTurnFinished:
		m.Turn = PLAYER_TURN
		return m, nil
	default:
		return m, nil
	}

	return m, nil
}

type PlayerTurnFinished struct{}
type WorldTurnFinished struct{}

type attackFinishedMsg struct{}

func (m *GameStateModel) PlayerAttack() tea.Cmd {
	m.IsAttacking = true

	m.Dungeon[m.Player.Y][m.Player.X].Kind = TileKind(PLAYER_ATTACKING) // Assuming you defined this

	return tea.Tick(time.Millisecond*150, func(t time.Time) tea.Msg {
		return attackFinishedMsg{}
	})
}

func (m *GameStateModel) UpdatePlayerPos(move string) bool {
	m.Turn = PLAYER_TURN
	endTile, moved := m.MovePlayer(move)
	switch endTile {
	case STAIRS:
		dungeon := CreateNewDungeon()
		cx, cy := dungeon.Rooms[0].Center()

		m.Player.X = cx
		m.Player.Y = cy
		m.Dungeon = dungeon.Tiles
		m.Rooms = dungeon.Rooms
		m.Monsters = dungeon.Monsters
		m.Floor++

		m.Dungeon[cy][cx] = Tile{Kind: PLAYER}
	}

	return moved
}

func (m *GameStateModel) MovePlayer(move string) (TileKind, bool) {
	newX, newY := m.Player.X, m.Player.Y
	m.Turn = PLAYER_TURN

	switch move {
	case "up", "k":
		newY--
	case "down", "j":
		newY++
	case "left", "h":
		newX--
	case "right", "l":
		newX++
	}

	if newX <= 0 || newX >= WIDTH-1 || newY <= 0 || newY >= HEIGHT-1 {
		return WALL, false
	}

	destinationKind := m.Dungeon[newY][newX].Kind

	if destinationKind == WALL || destinationKind == FILLER_WALL || destinationKind == TOUCHING_WALL {
		return destinationKind, false
	}

	m.Dungeon[m.Player.Y][m.Player.X] = Tile{Kind: FLOOR}

	m.Player.X = newX
	m.Player.Y = newY

	m.Dungeon[newY][newX] = Tile{Kind: PLAYER}

	return destinationKind, true
}

func (m *GameStateModel) View() tea.View {
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
			case MONSTER:
				mapSb.WriteString(monsterStyle.Render(char))
			case WALL, FILLER_WALL, TOUCHING_WALL, PASSAGE_WALL:
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
	// Wrap the raw map text in our bordered box
	mapPanel := mapBoxStyle.Render(mapSb.String())

	// ---------------------------------------------------------
	// 2. BUILD THE SIDEBAR (Stats & Info)
	// ---------------------------------------------------------
	// Dynamic HP Coloring
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
	// Grab the last 3 messages from the log array to avoid overflowing the screen
	logCount := len(m.Log)
	start := max(0, logCount-3)

	// Join them with newlines and fill the box width
	recentLogs := strings.Join(m.Log[start:], "\n")

	// Set the log box width to match the map + sidebar combined (approx 106 chars)
	logPanel := logBoxStyle.Width(WIDTH + 26).Render(recentLogs)

	// ---------------------------------------------------------
	// 4. ASSEMBLE THE FINAL UI
	// ---------------------------------------------------------
	// Join the Map and Sidebar side-by-side
	topSection := lipgloss.JoinHorizontal(lipgloss.Top, mapPanel, sidebarPanel)

	// Join the Top Section with the Log underneath
	fullScreen := lipgloss.JoinVertical(lipgloss.Left, topSection, logPanel)

	// Add the help text at the very bottom outside the boxes
	helpText := helpStyle.Render(" movement: h j k l / arrows • q: quit • enter/space: attack")
	finalUI := lipgloss.JoinVertical(lipgloss.Left, fullScreen, helpText)

	return tea.NewView(finalUI)
}
