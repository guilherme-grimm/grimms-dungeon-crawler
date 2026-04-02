package main

import "math"

func (m *GameStateModel) UpdateVisibility() {
	height := len(m.Dungeon)
	width := len(m.Dungeon[0])

	for y := range height {
		for x := range width {
			m.Dungeon[y][x].Visible = false
		}
	}

	steps := 360

	for i := range steps {
		angle := float64(i) * (2.0 * math.Pi / float64(steps))
		dx := math.Cos(angle)
		dy := math.Sin(angle)

		for dist := range m.Player.ViewRadius {
			tx := m.Player.X + int(math.Round(dx*float64(dist)))
			ty := m.Player.Y + int(math.Round(dy*float64(dist)))

			if tx < 0 || tx >= width || ty < 0 || ty >= height {
				break
			}

			m.Dungeon[ty][tx].Visible = true
			m.Dungeon[ty][tx].Explored = true

			// Brightness falloff: full → mid → edge
			radius := m.Player.ViewRadius
			switch {
			case dist < radius/3:
				m.Dungeon[ty][tx].Brightness = 2
			case dist < radius*2/3:
				m.Dungeon[ty][tx].Brightness = 1
			default:
				m.Dungeon[ty][tx].Brightness = 0
			}

			if m.Dungeon[ty][tx].Kind == WALL || m.Dungeon[ty][tx].Kind == TOUCHING_WALL {
				break
			}
		}
	}
}
