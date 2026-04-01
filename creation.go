package main

import (
	"math/rand/v2"
	"slices"
)

// Possible Monsters
var (
	SNEK Entity = Entity{
		Name:       "Snek",
		Glyph:      'S',
		MoveSpeed:  2,
		HP:         10,
		ATK:        2,
		StandingOn: FLOOR,
	}
	BRUTE Entity = Entity{
		Name:       "Brute",
		MoveSpeed:  1,
		Glyph:      'B',
		HP:         10,
		ATK:        3,
		StandingOn: FLOOR,
	}
	SHADE Entity = Entity{
		Name:       "Shade",
		MoveSpeed:  3,
		Glyph:      'W',
		HP:         10,
		ATK:        1,
		StandingOn: FLOOR,
	}
)

var possibleMonsters []Entity = []Entity{SNEK, BRUTE, SHADE}

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

		for range numMonsters {
			mx := room.X + rand.IntN(room.W)
			my := room.Y + rand.IntN(room.H)

			if tiles[my][mx].Kind == FLOOR {
				tiles[my][mx].Kind = MONSTER

				entry := rand.IntN(len(possibleMonsters))
				mon := possibleMonsters[entry]
				mon.X = mx
				mon.Y = my

				monsters = append(monsters, mon)
			}
		}
	}

	dungeon.Monsters = monsters

	return dungeon
}
