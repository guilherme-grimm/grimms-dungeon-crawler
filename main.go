package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"slices"
	"time"

	tea "charm.land/bubbletea/v2"
)

type TileKind rune

var (
	PLAYER        TileKind = '@'
	FLOOR         TileKind = '·'
	WALL          TileKind = '#'
	TOUCHING_WALL TileKind = '□'
	FILLER_WALL   TileKind = '█'
	PASSAGE_WALL  TileKind = '='
	STAIRS        TileKind = '▼'
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
	Tiles [][]Tile
	Rooms []Room
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
	p := tea.NewProgram(initialGameState())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Well, the dungeon broke: %v", err)
		os.Exit(1)
	}
}

func CreateNewDungeon() Dungeon {
	tiles := make([][]Tile, HEIGHT)
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
		Monsters: []Entity{},
		Items:    []any{},
		Floor:    1,
		Rooms:    dungeon.Rooms,
		Log:      []string{"Leave behind all hope thou who enter here..."},
	}
}

func (m GameStateModel) Init() tea.Cmd {
	return nil
}

func (m GameStateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k", "down", "j", "left", "l", "right", "h":
			m.UpdatePlayerPos(msg.String())
			return m, nil
		case "enter", "space":
			m.PlayerAttack()
			return m, nil
		}
	}
	return m, nil
}

func (m *GameStateModel) PlayerAttack() {
	m.IsAttacking = true
	time.Sleep(time.Millisecond * 20)
}

func (m *GameStateModel) UpdatePlayerPos(move string) {
	newX, newY := m.Player.X, m.Player.Y

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
		return
	}

	currKind := m.Dungeon[newY][newX].Kind

	if currKind == WALL || currKind == FILLER_WALL || currKind == TOUCHING_WALL {
		return
	}

	m.Dungeon[m.Player.Y][m.Player.X] = Tile{Kind: FLOOR}

	m.Player.X = newX
	m.Player.Y = newY

	m.Dungeon[newY][newX] = Tile{Kind: PLAYER}
}

func (m GameStateModel) View() tea.View {
	var s string

	for _, row := range m.Dungeon {
		for _, tile := range row {
			s += string(tile.Kind)
		}

		s += "\n"
	}

	return tea.NewView(s)
}
