package board

var moveperms = [][]string{
	{"up", "down", "right", "left"},
	{"up", "down", "left", "right"},
	{"up", "right", "down", "left"},
	{"up", "right", "left", "down"},
	{"up", "left", "right", "down"},
	{"up", "left", "down", "right"},
	{"down", "up", "right", "left"},
	{"down", "up", "left", "right"},
	{"down", "right", "up", "left"},
	{"down", "right", "left", "up"},
	{"down", "left", "right", "up"},
	{"down", "left", "up", "right"},
	{"right", "down", "up", "left"},
	{"right", "down", "left", "up"},
	{"right", "up", "down", "left"},
	{"right", "up", "left", "down"},
	{"right", "left", "up", "down"},
	{"right", "left", "down", "up"},
	{"left", "down", "right", "up"},
	{"left", "down", "up", "right"},
	{"left", "right", "down", "up"},
	{"left", "right", "up", "down"},
	{"left", "up", "right", "down"},
	{"left", "up", "down", "right"},
}

const perms = 24
