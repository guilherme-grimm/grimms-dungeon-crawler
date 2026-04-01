package main

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
