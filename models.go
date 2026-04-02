package main

import "time"

type Screen int

const (
	MENU_SCREEN Screen = iota
	GAME_SCREEN
	DEATH_SCREEN
)

type GameStateModel struct {
	// Canvas
	TermWidth  int
	TermHeight int
	// Instance state
	Screen Screen

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
	PlayerIsAttacking bool
	PlayerDirection   Direction
	StepsRemaining    int
	MoveInput         string

	// Attack animation state
	AttackPhase     int        // 0=idle, 1=frame1, 2=frame2
	AttackSlashPos  Direction  // tile where slash renders
	MonsterFlashPos *Direction // tile of flashed monster (nil = no flash)
}

type Direction struct {
	X, Y int
}

var (
	UP    = Direction{0, -1}
	DOWN  = Direction{0, 1}
	LEFT  = Direction{-1, 0}
	RIGHT = Direction{1, 0}
)

type TurnType string

var (
	PLAYER_TURN TurnType = "PLAYER"
	WORLD_TURN  TurnType = "WORLD"
)

type TileKind rune

var (
	PLAYER            TileKind = '@'
	PLAYER_ATTACKING  TileKind = 'A'
	FLOOR             TileKind = '·'
	WALL              TileKind = '█'
	TOUCHING_WALL     TileKind = '#'
	FILLER_WALL       TileKind = '░'
	STAIRS            TileKind = '▼'
	MONSTER           TileKind = 'S'
	MONSTER_ATTACKING TileKind = 'D'
	DEAD_MONSTER      TileKind = '☠'
)

const (
	START_WIDTH  = 80
	START_HEIGHT = 24
)

type model int

type tickMsg time.Time

type Tile struct {
	Kind       TileKind // Wall, Floor, Player
	Visible    bool
	Explored   bool
	Brightness int // 0=edge, 1=mid, 2=full
	X, Y       int
}

type Entity struct {
	X, Y        int
	MoveSpeed   int // determine how much tiles per turn
	AttackSpeed int // determine attack cooldown in turns
	ViewRadius  int
	Name        string
	HP, ATK     int
	Glyph       rune
	StandingOn  TileKind
	IsAttacking bool
}

type Dungeon struct {
	Tiles    [][]Tile
	Rooms    []Room
	Monsters []Entity
	StartX   int
	StartY   int
}
