package main

import "math/rand/v2"

func (m *GameStateModel) MoveMonsters() {
	for i := range m.Monsters {
		monster := &m.Monsters[i]
		for range monster.MoveSpeed {
			if monster.HP > 0 {
				newPos := m.greedyMove(Direction{m.Player.X, m.Player.Y}, Direction{monster.X, monster.Y})

				if newPos.X == monster.X && newPos.Y == monster.Y {
					if manhatann(Direction{monster.X, monster.Y}, Direction{m.Player.X, m.Player.Y}) == 1 {
						m.Player.HP -= monster.ATK
						m.Dungeon[monster.Y][monster.X].Kind = MONSTER
					}
					continue
				}

				// Restore what was underneath
				m.Dungeon[monster.Y][monster.X].Kind = monster.StandingOn

				// Save what's at the destination
				monster.StandingOn = m.Dungeon[newPos.Y][newPos.X].Kind

				// Move
				monster.X = newPos.X
				monster.Y = newPos.Y
				m.Dungeon[newPos.Y][newPos.X].Kind = MONSTER
			}
		}
	}
}

func (m *GameStateModel) greedyMove(targetPos, enemyPos Direction) Direction {
	dirs := []Direction{UP, DOWN, LEFT, RIGHT}

	bestX, bestY := enemyPos.X, enemyPos.Y
	bestDist := manhatann(enemyPos, targetPos)
	if bestDist <= 1 {
		return Direction{enemyPos.X, enemyPos.Y}
	}

	for _, d := range dirs {
		newX := enemyPos.X + d.X
		newY := enemyPos.Y + d.Y

		if newX < 0 || newY < 0 || newX >= len(m.Dungeon[0]) || newY >= len(m.Dungeon) {
			continue
		}

		switch m.Dungeon[newY][newX].Kind {
		case TOUCHING_WALL, WALL, PLAYER, DEAD_MONSTER, MONSTER:
			continue
		case FLOOR:
			dist := manhatann(Direction{newX, newY}, targetPos)
			if dist < bestDist {
				bestDist = dist
				bestX, bestY = newX, newY
			}
		}
	}

	if bestX == enemyPos.X && bestY == enemyPos.Y {
		rand.Shuffle(len(dirs), func(i, j int) { dirs[i], dirs[j] = dirs[j], dirs[i] })
		for _, d := range dirs {
			newX := enemyPos.X + d.X
			newY := enemyPos.Y + d.Y

			if newX < 0 || newY < 0 || newX >= START_WIDTH || newY >= START_HEIGHT {
				continue
			}

			if m.Dungeon[newY][newX].Kind == FLOOR {
				return Direction{newX, newY}
			}
		}
	}

	return Direction{bestX, bestY}
}

func manhatann(enemyPos, targetPos Direction) int {
	return abs(enemyPos.X-targetPos.X) + abs(enemyPos.Y-targetPos.Y)
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
