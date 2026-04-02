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
		ATK:        0,
		StandingOn: FLOOR,
	}
	BRUTE Entity = Entity{
		Name:       "Brute",
		MoveSpeed:  1,
		Glyph:      'B',
		HP:         10,
		ATK:        0,
		StandingOn: FLOOR,
	}
	SHADE Entity = Entity{
		Name:       "Shade",
		MoveSpeed:  3,
		Glyph:      'W',
		HP:         10,
		ATK:        0,
		StandingOn: FLOOR,
	}
)

var possibleMonsters []Entity = []Entity{SNEK, BRUTE, SHADE}

func dungeonSize(floor int) (width, height int) {
	baseW, baseH := START_WIDTH, START_HEIGHT
	growW, growH := 10, 5
	maxW, maxH := 160, 80

	w := baseW + (floor-1)*growW
	h := baseH + (floor-1)*growH

	return min(w, maxW), min(h, maxH)
}

func CreateNewDungeon(floor int) Dungeon {
	var dg Dungeon
	switch {
	case floor <= 5:
		dg = CreateRoomLikeFloor(floor)

	case floor <= 10:
		dg = CreateCaveLikeFloor(floor)
	default:
		dg = CreateCaveLikeFloor(floor)
	}

	return dg
}

func CreateRoomLikeFloor(floor int) Dungeon {

	WIDTH, HEIGHT := dungeonSize(floor)
	tiles := make([][]Tile, HEIGHT)

	// 1. Fill with Walls
	for y := range HEIGHT {
		tiles[y] = make([]Tile, WIDTH)
		for x := range WIDTH {
			tiles[y][x] = Tile{Kind: WALL, Visible: false, X: x, Y: y}
		}
	}

	rooms := make([]Room, 0)
	maxRooms := 10 + (floor * 2)

	// 2. Generate Composite Rooms
	for i := 0; i < maxRooms*3; i++ { // Try more times to fit rooms
		w := rand.IntN(7-4) + 4
		h := rand.IntN(7-4) + 4
		x := rand.IntN(WIDTH-w-2) + 1
		y := rand.IntN(HEIGHT-h-2) + 1

		newRoom := Room{x, y, w, h}

		// Check overlap against core rooms
		if slices.ContainsFunc(rooms, newRoom.Overlaps) {
			continue
		}

		rooms = append(rooms, newRoom)

		// Carve the Main Room
		for cy := newRoom.Y; cy < newRoom.Y+newRoom.H; cy++ {
			for cx := newRoom.X; cx < newRoom.X+newRoom.W; cx++ {
				tiles[cy][cx].Kind = FLOOR
			}
		}

		// Carve a Sub-Room (The Composite overlap) to make it irregular
		subW := rand.IntN(4) + 3
		subH := rand.IntN(4) + 3
		// Offset it so it overlaps the edge of the main room
		subX := newRoom.X + rand.IntN(newRoom.W) - subW/2
		subY := newRoom.Y + rand.IntN(newRoom.H) - subH/2

		// Carve sub-room safely within map bounds
		for cy := max(1, subY); cy < min(HEIGHT-2, subY+subH); cy++ {
			for cx := max(1, subX); cx < min(WIDTH-2, subX+subW); cx++ {
				tiles[cy][cx].Kind = FLOOR
			}
		}

		if len(rooms) >= maxRooms {
			break
		}
	}

	// 3. Connect Rooms with Corridors AND Loops
	for i := 1; i < len(rooms); i++ {
		cx1, cy1 := rooms[i-1].Center()
		cx2, cy2 := rooms[i].Center()

		// Carve L-shaped corridor
		for x := min(cx1, cx2); x <= max(cx1, cx2); x++ {
			tiles[cy1][x].Kind = FLOOR
		}
		for y := min(cy1, cy2); y <= max(cy1, cy2); y++ {
			tiles[y][cx2].Kind = FLOOR
		}

		// THE LOOP MAKER: 20% chance to connect to a random older room
		if i > 2 && rand.IntN(100) < 20 {
			randPrev := rand.IntN(i - 1)
			cx3, cy3 := rooms[randPrev].Center()

			for x := min(cx2, cx3); x <= max(cx2, cx3); x++ {
				tiles[cy2][x].Kind = FLOOR
			}
			for y := min(cy2, cy3); y <= max(cy2, cy3); y++ {
				tiles[y][cx3].Kind = FLOOR
			}
		}
	}

	// 4. Corner Erosion (The "Ruined" Look)
	// If a floor tile is surrounded by walls on 3 sides (a sharp inner corner), fill it in.
	for y := 1; y < HEIGHT-1; y++ {
		for x := 1; x < WIDTH-1; x++ {
			if tiles[y][x].Kind == FLOOR {
				floorCount := 0
				if tiles[y-1][x].Kind == FLOOR {
					floorCount++
				}
				if tiles[y+1][x].Kind == FLOOR {
					floorCount++
				}
				if tiles[y][x-1].Kind == FLOOR {
					floorCount++
				}
				if tiles[y][x+1].Kind == FLOOR {
					floorCount++
				}

				// Fill in tight 1x1 dead ends and sharp corners
				if floorCount <= 1 {
					tiles[y][x].Kind = WALL
				}
			}
		}
	}

	// 5. Post-Processing: Touching Walls (Your original logic)
	dirs := [][]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for row := range HEIGHT {
		for col := range WIDTH {
			if tiles[row][col].Kind == WALL {
				for _, d := range dirs {
					newRow := row + d[0]
					newCol := col + d[1]
					if newRow < 0 || newRow >= HEIGHT || newCol < 0 || newCol >= WIDTH {
						continue
					}
					if tiles[newRow][newCol].Kind == FLOOR {
						tiles[row][col].Kind = TOUCHING_WALL
						break
					}
				}
			}
		}
	}

	// 6. Spawn Logic
	startX, startY := rooms[0].Center()
	maxDist := 0
	furthestRoomIndex := 0

	for i := 1; i < len(rooms); i++ {
		cx, cy := rooms[i].Center()
		dist := (cx-startX)*(cx-startX) + (cy-startY)*(cy-startY)

		if dist > maxDist {
			maxDist = dist
			furthestRoomIndex = i
		}
	}

	lastX, lastY := rooms[furthestRoomIndex].Center()

	tiles[lastY][lastX].Kind = STAIRS

	// 7. Monster Spawning
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

	return Dungeon{
		Tiles:    tiles,
		Monsters: monsters,
		Rooms:    rooms, // We can populate this safely for the room generator!
		StartX:   startX,
		StartY:   startY,
	}
}

func CreateCaveLikeFloor(floor int) Dungeon {
	WIDTH, HEIGHT := dungeonSize(floor)
	tiles := make([][]Tile, HEIGHT)

	// 1. Fill the entire map with Walls
	for y := range HEIGHT {
		tiles[y] = make([]Tile, WIDTH)
		for x := range WIDTH {
			tiles[y][x] = Tile{
				Kind:    WALL,
				Visible: false,
				X:       x,
				Y:       y,
			}
		}
	}
	currX, currY := WIDTH/2, HEIGHT/2
	dirs := [][]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	maxTunnels := 30 + (floor * 10)

	for range maxTunnels {
		dir := dirs[rand.IntN(len(dirs))]

		tunnelLength := rand.IntN(7) + 4

		for range tunnelLength {
			nextX := currX + dir[1]
			nextY := currY + dir[0]

			if nextX > 2 && nextX < WIDTH-3 && nextY > 2 && nextY < HEIGHT-3 {
				currX = nextX
				currY = nextY

				tiles[currY][currX].Kind = FLOOR
			} else {
				break
			}
		}

		brushRadius := rand.IntN(2) + 2

		for dy := -brushRadius; dy <= brushRadius; dy++ {
			for dx := -brushRadius; dx <= brushRadius; dx++ {
				carveX := currX + dx
				carveY := currY + dy

				if carveX > 1 && carveX < WIDTH-2 && carveY > 1 && carveY < HEIGHT-2 {
					if (dx*dx)+(dy*dy) <= (brushRadius * brushRadius) {
						tiles[carveY][carveX].Kind = FLOOR
					}
				}
			}
		}
	}

	startX, startY := WIDTH/2, HEIGHT/2
	maxDist := 0
	stairX, stairY := startX, startY

	for y := 1; y < HEIGHT-1; y++ {
		for x := 1; x < WIDTH-1; x++ {
			if tiles[y][x].Kind == FLOOR {
				dist := (x-startX)*(x-startX) + (y-startY)*(y-startY)
				if dist > maxDist {
					maxDist = dist
					stairX, stairY = x, y
				}
			}
		}
	}

	tiles[startY][startX].Kind = FLOOR
	tiles[stairY][stairX].Kind = STAIRS

	// 4. Post-Processing: Walls (Adapted from your original code)
	for row := range HEIGHT {
		for col := range WIDTH {
			if tiles[row][col].Kind == WALL {
				for _, d := range dirs {
					newRow := row + d[0]
					newCol := col + d[1]

					if newRow < 0 || newRow >= HEIGHT || newCol < 0 || newCol >= WIDTH {
						continue
					}

					if tiles[newRow][newCol].Kind == FLOOR {
						tiles[row][col].Kind = TOUCHING_WALL
						break // Only need one adjacent floor to be a touching wall
					}
				}
			}
		}
	}

	// 5. Monster Spawning Logic
	// Since we no longer have a `rooms` array, we just scatter monsters randomly
	monsters := make([]Entity, 0)
	numMonsters := 5 + (floor * 2) // Scale monsters with floor depth

	for range numMonsters {
		// Try a few times to find a valid floor tile
		for range 50 {
			mx := rand.IntN(WIDTH-2) + 1
			my := rand.IntN(HEIGHT-2) + 1

			// Don't spawn monsters right on top of the player spawn (center)
			distFromCenter := (mx-WIDTH/2)*(mx-WIDTH/2) + (my-HEIGHT/2)*(my-HEIGHT/2)

			if tiles[my][mx].Kind == FLOOR && distFromCenter > 25 {
				tiles[my][mx].Kind = MONSTER
				entry := rand.IntN(len(possibleMonsters))
				mon := possibleMonsters[entry]
				mon.X = mx
				mon.Y = my

				monsters = append(monsters, mon)
				break // Found a spot, move to the next monster
			}
		}
	}
	dungeon := Dungeon{
		Tiles:    tiles,
		Monsters: monsters,
		Rooms:    []Room{},
		StartX:   WIDTH / 2,
		StartY:   HEIGHT / 2,
	}

	return dungeon
}
