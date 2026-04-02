package main

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
)

func (m *GameStateModel) PlayerAttack() tea.Cmd {
	m.PlayerIsAttacking = true
	m.Dungeon[m.Player.Y][m.Player.X].Kind = TileKind(PLAYER_ATTACKING)

	attackPosX := m.Player.X + m.PlayerDirection.X
	attackPosY := m.Player.Y + m.PlayerDirection.Y

	// Set up attack animation state
	m.AttackPhase = 1
	m.AttackSlashPos = Direction{attackPosX, attackPosY}

	m.Log = append(m.Log, fmt.Sprintf("Player attacked at: %v", Direction{attackPosX, attackPosY}))
	switch m.Dungeon[attackPosY][attackPosX].Kind {
	case MONSTER:
		m.Log = append(m.Log, "Is a monster!")
		m.MonsterFlashPos = &Direction{attackPosX, attackPosY}
		for i := range m.Monsters {
			monster := &m.Monsters[i]
			if monster.X == attackPosX && monster.Y == attackPosY {
				monster.HP -= m.Player.ATK
				m.Log = append(m.Log, fmt.Sprintf("ouch! %v lost %v/%v", monster.Name, monster.HP, monster.HP+m.Player.ATK))

				if monster.HP <= 0 {
					m.Log = append(m.Log, "Killed!")
					m.Dungeon[monster.Y][monster.X].Kind = DEAD_MONSTER
				}
				break
			}
		}
	default:
		m.Log = append(m.Log, "IT's not a monster")
	}

	return tea.Tick(time.Millisecond*60, func(t time.Time) tea.Msg {
		return attackPhaseMsg{Phase: 1}
	})
}

// MovePlayerOneStep moves the player one tile in the given direction.
// Returns the destination tile kind and whether the move succeeded.
func (m *GameStateModel) MovePlayerOneStep(move string) (TileKind, bool) {
	newX, newY := m.Player.X, m.Player.Y

	switch move {
	case "up", "k":
		m.PlayerDirection = UP
		newY--
	case "down", "j":
		m.PlayerDirection = DOWN
		newY++
	case "left", "h":
		m.PlayerDirection = LEFT
		newX--
	case "right", "l":
		m.PlayerDirection = RIGHT
		newX++
	}

	if newX <= 0 || newX >= len(m.Dungeon[0]) || newY <= 0 || newY >= len(m.Dungeon) {
		return WALL, false
	}

	destinationKind := m.Dungeon[newY][newX].Kind

	switch destinationKind {
	case WALL, FILLER_WALL, TOUCHING_WALL:
		m.Log = append(m.Log, "You can't pass through wall, you are not Kitty Pride.")
		return destinationKind, false
	case MONSTER:
		m.Log = append(m.Log, "No, you're not able to ram into Dead or Alive monsters, yet.")
		return destinationKind, false
	}

	// Restore what was underneath
	m.Dungeon[m.Player.Y][m.Player.X].Kind = m.Player.StandingOn

	// Save what's at the destination
	m.Player.StandingOn = m.Dungeon[newY][newX].Kind

	// Move the player
	m.Player.X = newX
	m.Player.Y = newY
	m.Dungeon[newY][newX].Kind = PLAYER

	return destinationKind, true
}

// HandleStairs transitions to a new dungeon floor.
func (m *GameStateModel) HandleStairs() {
	dungeon := CreateNewDungeon(m.Floor + 1)
	cx, cy := dungeon.Rooms[0].Center()

	m.Player.X = cx
	m.Player.Y = cy
	m.Dungeon = dungeon.Tiles
	m.Rooms = dungeon.Rooms
	m.Monsters = dungeon.Monsters
	m.Floor++

	m.Dungeon[cy][cx] = Tile{Kind: PLAYER}
	m.Player.StandingOn = FLOOR

	m.Log = append(m.Log, "Another one enters the pit of despair...")
}
