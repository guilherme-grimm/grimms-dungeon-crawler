package main

type PlayerTurnFinished struct{}
type WorldTurnFinished struct{}
type attackPhaseMsg struct{ Phase int }
type MovementStepMsg struct{}
