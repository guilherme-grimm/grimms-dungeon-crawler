package main

import "charm.land/lipgloss/v2"

// --- STYLES ---
var (
	// Entity Styles
	playerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFAA")).Bold(true)
	attackStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA00")).Bold(true)
	monsterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0044")).Bold(true)
	stairStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Bold(true)

	// Fog of War — explored but not visible (dim)
	dimWallStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#0a0a10"))
	dimTouchingWallStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#2a3338"))
	dimFloorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#151515"))
	dimStairStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#666620"))
	dimMonsterStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#661122"))
	dimStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("#111111"))

	// Mid-brightness — edge of vision radius
	midWallStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#0d0d14"))
	midTouchingWallStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#3d4d55"))
	midFloorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#1a1a1a"))
	midMonsterStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#992233"))
	midStairStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#999930")).Bold(true)

	// Attack Effect Styles
	monsterFlashStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	slashStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFDD44")).Bold(true)

	// Environment Styles
	wallStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#111118"))
	touchingWallStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#556677")).Bold(true)
	fillerWallStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#1a1a22"))
	floorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#222222"))

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

	// Menu Styles
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0044")).Bold(true)
	subtitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")).Bold(true)
	promptStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))

	// Death Screen Styles
	deathStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	deathMessageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0044")).Bold(true)
)
