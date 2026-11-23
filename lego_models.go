package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	LegoRed        = rl.NewColor(220, 30, 30, 255)
	LegoBlue       = rl.NewColor(30, 110, 220, 255)
	LegoYellow     = rl.NewColor(255, 210, 10, 255)
	LegoGreen      = rl.NewColor(40, 180, 50, 255)
	LegoOrange     = rl.NewColor(255, 120, 20, 255)
	LegoWhite      = rl.NewColor(245, 245, 250, 255)
	LegoBlack      = rl.NewColor(35, 35, 40, 255)
	LegoCyan       = rl.NewColor(40, 200, 230, 255)
	LegoPink       = rl.NewColor(255, 130, 180, 255)
	LegoPurple     = rl.NewColor(150, 50, 200, 255)
	LegoLime       = rl.NewColor(160, 220, 40, 255)
	LegoDarkGrey   = rl.NewColor(80, 85, 90, 255)
	LegoLightGrey  = rl.NewColor(170, 175, 180, 255)
	LegoBrown      = rl.NewColor(120, 70, 40, 255)
	LegoGold       = rl.NewColor(255, 215, 0, 255)
	LegoGoldBright = rl.NewColor(255, 235, 50, 255)
	LegoTeal       = rl.NewColor(20, 160, 150, 255)
	LegoMagenta    = rl.NewColor(210, 30, 120, 255)
)

type LegoBrick struct {
	RelPos   rl.Vector3
	Size     rl.Vector3
	Color    rl.Color
	StudCols int
	StudRows int
	OrderIdx int
}

type LegoTheme int

const (
	ThemeHouse LegoTheme = iota
	ThemeDino
	ThemeCar
	ThemeCartoon
	ThemeCastle
	ThemeRobot
	ThemeArenaLetter
)

type LegoItemDefinition struct {
	ID          int
	Name        string
	Theme       LegoTheme
	Bricks      []LegoBrick
	BaseScale   float32
	Description string
	Icon        string
}

// GetLegoBlueprints returns the rich library with distinct unique models & colors
func GetLegoBlueprints() []LegoItemDefinition {
	return []LegoItemDefinition{
		// 1. T-REX (Lime & Green)
		{
			ID:          0,
			Name:        "T-Rex Dino",
			Theme:       ThemeDino,
			BaseScale:   1.0,
			Description: "Regele Dinozaurilor blocuri!",
			Icon:        "🦕",
			Bricks: []LegoBrick{
				{RelPos: rl.NewVector3(-0.6, 0.4, 0), Size: rl.NewVector3(0.8, 0.8, 0.8), Color: LegoLime, OrderIdx: 1},
				{RelPos: rl.NewVector3(0.6, 0.4, 0), Size: rl.NewVector3(0.8, 0.8, 0.8), Color: LegoLime, OrderIdx: 2},
				{RelPos: rl.NewVector3(0, 1.2, 0), Size: rl.NewVector3(1.8, 1.2, 1.6), Color: LegoGreen, OrderIdx: 3},
				{RelPos: rl.NewVector3(0, 1.2, -1.2), Size: rl.NewVector3(0.8, 0.8, 1.2), Color: LegoGreen, OrderIdx: 4},
				{RelPos: rl.NewVector3(0, 1.6, -2.0), Size: rl.NewVector3(0.6, 0.6, 1.0), Color: LegoLime, OrderIdx: 5},
				{RelPos: rl.NewVector3(0, 2.2, 0.4), Size: rl.NewVector3(1.2, 1.2, 1.2), Color: LegoGreen, OrderIdx: 6},
				{RelPos: rl.NewVector3(0, 3.0, 1.0), Size: rl.NewVector3(1.6, 1.2, 1.8), Color: LegoLime, OrderIdx: 7},
				{RelPos: rl.NewVector3(0, 2.6, 1.8), Size: rl.NewVector3(1.4, 0.6, 1.2), Color: LegoGreen, OrderIdx: 8},
				{RelPos: rl.NewVector3(-0.85, 3.2, 1.2), Size: rl.NewVector3(0.3, 0.4, 0.4), Color: LegoWhite, OrderIdx: 9},
				{RelPos: rl.NewVector3(0.85, 3.2, 1.2), Size: rl.NewVector3(0.3, 0.4, 0.4), Color: LegoWhite, OrderIdx: 10},
				{RelPos: rl.NewVector3(-0.9, 3.2, 1.3), Size: rl.NewVector3(0.2, 0.2, 0.2), Color: LegoBlack, OrderIdx: 11},
				{RelPos: rl.NewVector3(0.9, 3.2, 1.3), Size: rl.NewVector3(0.2, 0.2, 0.2), Color: LegoBlack, OrderIdx: 12},
				{RelPos: rl.NewVector3(-0.7, 1.5, 0.8), Size: rl.NewVector3(0.4, 0.4, 0.6), Color: LegoYellow, OrderIdx: 13},
				{RelPos: rl.NewVector3(0.7, 1.5, 0.8), Size: rl.NewVector3(0.4, 0.4, 0.6), Color: LegoYellow, OrderIdx: 14},
				{RelPos: rl.NewVector3(0, 3.7, 0.8), Size: rl.NewVector3(0.6, 0.6, 0.6), Color: LegoGold, OrderIdx: 15},
			},
		},
		// 2. TRICERATOPS (Cyan & Gold)
		{
			ID:          1,
			Name:        "Triceratops Dino",
			Theme:       ThemeDino,
			BaseScale:   1.0,
			Description: "Dinozaurul puternic cu 3 coarne!",
			Icon:        "🦏",
			Bricks: []LegoBrick{
				{RelPos: rl.NewVector3(-0.7, 0.4, 0.8), Size: rl.NewVector3(0.6, 0.8, 0.6), Color: LegoTeal, OrderIdx: 1},
				{RelPos: rl.NewVector3(0.7, 0.4, 0.8), Size: rl.NewVector3(0.6, 0.8, 0.6), Color: LegoTeal, OrderIdx: 2},
				{RelPos: rl.NewVector3(-0.7, 0.4, -0.8), Size: rl.NewVector3(0.6, 0.8, 0.6), Color: LegoTeal, OrderIdx: 3},
				{RelPos: rl.NewVector3(0.7, 0.4, -0.8), Size: rl.NewVector3(0.6, 0.8, 0.6), Color: LegoTeal, OrderIdx: 4},
				{RelPos: rl.NewVector3(0, 1.1, 0), Size: rl.NewVector3(1.8, 1.2, 2.2), Color: LegoCyan, OrderIdx: 5},
				{RelPos: rl.NewVector3(0, 1.0, -1.4), Size: rl.NewVector3(0.6, 0.6, 1.0), Color: LegoTeal, OrderIdx: 6},
				{RelPos: rl.NewVector3(0, 1.7, 1.2), Size: rl.NewVector3(2.4, 1.6, 0.4), Color: LegoOrange, OrderIdx: 7},    // Shield Frill
				{RelPos: rl.NewVector3(0, 1.3, 1.6), Size: rl.NewVector3(1.2, 1.0, 1.2), Color: LegoCyan, OrderIdx: 8},      // Head
				{RelPos: rl.NewVector3(0, 1.2, 2.3), Size: rl.NewVector3(0.8, 0.6, 0.8), Color: LegoTeal, OrderIdx: 9},      // Snout
				{RelPos: rl.NewVector3(0, 1.7, 2.4), Size: rl.NewVector3(0.3, 0.7, 0.3), Color: LegoWhite, OrderIdx: 10},    // Nose Horn
				{RelPos: rl.NewVector3(-0.6, 2.2, 1.4), Size: rl.NewVector3(0.3, 0.9, 0.3), Color: LegoWhite, OrderIdx: 11}, // Horn L
				{RelPos: rl.NewVector3(0.6, 2.2, 1.4), Size: rl.NewVector3(0.3, 0.9, 0.3), Color: LegoWhite, OrderIdx: 12},  // Horn R
				{RelPos: rl.NewVector3(-0.5, 1.5, 1.7), Size: rl.NewVector3(0.2, 0.2, 0.2), Color: LegoBlack, OrderIdx: 13}, // Eye L
				{RelPos: rl.NewVector3(0.5, 1.5, 1.7), Size: rl.NewVector3(0.2, 0.2, 0.2), Color: LegoBlack, OrderIdx: 14},  // Eye R
			},
		},
		// 3. BRACHIOSAURUS (Purple & Pink Long Neck)
		{
			ID:          2,
			Name:        "Brachiosaurus",
			Theme:       ThemeDino,
			BaseScale:   1.1,
			Description: "Uriesul cu gat lung si bland!",
			Icon:        "🦕",
			Bricks: []LegoBrick{
				{RelPos: rl.NewVector3(-0.7, 0.5, 0.8), Size: rl.NewVector3(0.7, 1.0, 0.7), Color: LegoPurple, OrderIdx: 1},
				{RelPos: rl.NewVector3(0.7, 0.5, 0.8), Size: rl.NewVector3(0.7, 1.0, 0.7), Color: LegoPurple, OrderIdx: 2},
				{RelPos: rl.NewVector3(-0.7, 0.5, -0.8), Size: rl.NewVector3(0.7, 1.0, 0.7), Color: LegoPurple, OrderIdx: 3},
				{RelPos: rl.NewVector3(0.7, 0.5, -0.8), Size: rl.NewVector3(0.7, 1.0, 0.7), Color: LegoPurple, OrderIdx: 4},
				{RelPos: rl.NewVector3(0, 1.4, 0), Size: rl.NewVector3(2.0, 1.4, 2.4), Color: LegoMagenta, OrderIdx: 5},
				{RelPos: rl.NewVector3(0, 1.2, -1.6), Size: rl.NewVector3(0.8, 0.6, 1.2), Color: LegoPurple, OrderIdx: 6},
				{RelPos: rl.NewVector3(0, 2.2, 1.0), Size: rl.NewVector3(0.9, 1.2, 0.9), Color: LegoMagenta, OrderIdx: 7},
				{RelPos: rl.NewVector3(0, 3.2, 1.2), Size: rl.NewVector3(0.8, 1.2, 0.8), Color: LegoPurple, OrderIdx: 8},
				{RelPos: rl.NewVector3(0, 4.2, 1.4), Size: rl.NewVector3(0.7, 1.2, 0.7), Color: LegoMagenta, OrderIdx: 9},
				{RelPos: rl.NewVector3(0, 5.0, 1.8), Size: rl.NewVector3(1.0, 0.8, 1.4), Color: LegoPink, OrderIdx: 10},
				{RelPos: rl.NewVector3(-0.4, 5.1, 1.9), Size: rl.NewVector3(0.2, 0.2, 0.2), Color: LegoBlack, OrderIdx: 11},
				{RelPos: rl.NewVector3(0.4, 5.1, 1.9), Size: rl.NewVector3(0.2, 0.2, 0.2), Color: LegoBlack, OrderIdx: 12},
				{RelPos: rl.NewVector3(0, 5.5, 1.6), Size: rl.NewVector3(0.5, 0.4, 0.5), Color: LegoGold, OrderIdx: 13},
			},
		},
		// 4. BLUEY PUPPY (Blue & Cyan)
		{
			ID:          3,
			Name:        "Bluey Puppy",
			Theme:       ThemeCartoon,
			BaseScale:   1.0,
			Description: "Prietenul cel mai vesel si jucaus!",
			Icon:        "🐶",
			Bricks: []LegoBrick{
				{RelPos: rl.NewVector3(-0.5, 0.4, 0), Size: rl.NewVector3(0.7, 0.8, 0.7), Color: LegoCyan, OrderIdx: 1},
				{RelPos: rl.NewVector3(0.5, 0.4, 0), Size: rl.NewVector3(0.7, 0.8, 0.7), Color: LegoCyan, OrderIdx: 2},
				{RelPos: rl.NewVector3(0, 1.2, 0), Size: rl.NewVector3(1.6, 1.2, 1.2), Color: LegoBlue, OrderIdx: 3},
				{RelPos: rl.NewVector3(0, 1.2, 0.5), Size: rl.NewVector3(1.0, 0.9, 0.4), Color: LegoCyan, OrderIdx: 4},
				{RelPos: rl.NewVector3(-0.9, 1.2, 0), Size: rl.NewVector3(0.4, 0.9, 0.5), Color: LegoCyan, OrderIdx: 5},
				{RelPos: rl.NewVector3(0.9, 1.2, 0), Size: rl.NewVector3(0.4, 0.9, 0.5), Color: LegoCyan, OrderIdx: 6},
				{RelPos: rl.NewVector3(0, 2.3, 0), Size: rl.NewVector3(1.8, 1.4, 1.4), Color: LegoBlue, OrderIdx: 7},
				{RelPos: rl.NewVector3(0, 2.0, 0.7), Size: rl.NewVector3(1.0, 0.6, 0.6), Color: LegoYellow, OrderIdx: 8},
				{RelPos: rl.NewVector3(0, 2.2, 1.0), Size: rl.NewVector3(0.4, 0.3, 0.3), Color: LegoBlack, OrderIdx: 9},
				{RelPos: rl.NewVector3(-0.4, 2.5, 0.7), Size: rl.NewVector3(0.4, 0.4, 0.2), Color: LegoWhite, OrderIdx: 10},
				{RelPos: rl.NewVector3(0.4, 2.5, 0.7), Size: rl.NewVector3(0.4, 0.4, 0.2), Color: LegoWhite, OrderIdx: 11},
				{RelPos: rl.NewVector3(-0.7, 3.2, 0), Size: rl.NewVector3(0.5, 0.7, 0.5), Color: LegoBlue, OrderIdx: 12},
				{RelPos: rl.NewVector3(0.7, 3.2, 0), Size: rl.NewVector3(0.5, 0.7, 0.5), Color: LegoBlue, OrderIdx: 13},
				{RelPos: rl.NewVector3(0, 1.0, -0.8), Size: rl.NewVector3(0.4, 0.4, 0.8), Color: LegoCyan, OrderIdx: 14},
			},
		},
		// 5. BINGO PUPPY (Orange & Yellow)
		{
			ID:          4,
			Name:        "Bingo Puppy",
			Theme:       ThemeCartoon,
			BaseScale:   0.9,
			Description: "Sora mai mica si dulce a lui Bluey!",
			Icon:        "🦊",
			Bricks: []LegoBrick{
				{RelPos: rl.NewVector3(-0.45, 0.35, 0), Size: rl.NewVector3(0.6, 0.7, 0.6), Color: LegoYellow, OrderIdx: 1},
				{RelPos: rl.NewVector3(0.45, 0.35, 0), Size: rl.NewVector3(0.6, 0.7, 0.6), Color: LegoYellow, OrderIdx: 2},
				{RelPos: rl.NewVector3(0, 1.0, 0), Size: rl.NewVector3(1.5, 1.0, 1.1), Color: LegoOrange, OrderIdx: 3},
				{RelPos: rl.NewVector3(0, 1.0, 0.45), Size: rl.NewVector3(0.9, 0.8, 0.3), Color: LegoYellow, OrderIdx: 4},
				{RelPos: rl.NewVector3(-0.8, 1.0, 0), Size: rl.NewVector3(0.35, 0.8, 0.45), Color: LegoOrange, OrderIdx: 5},
				{RelPos: rl.NewVector3(0.8, 1.0, 0), Size: rl.NewVector3(0.35, 0.8, 0.45), Color: LegoOrange, OrderIdx: 6},
				{RelPos: rl.NewVector3(0, 2.0, 0), Size: rl.NewVector3(1.6, 1.2, 1.2), Color: LegoOrange, OrderIdx: 7},
				{RelPos: rl.NewVector3(0, 1.7, 0.6), Size: rl.NewVector3(0.9, 0.5, 0.5), Color: LegoYellow, OrderIdx: 8},
				{RelPos: rl.NewVector3(0, 1.9, 0.9), Size: rl.NewVector3(0.3, 0.25, 0.25), Color: LegoBlack, OrderIdx: 9},
				{RelPos: rl.NewVector3(-0.35, 2.2, 0.6), Size: rl.NewVector3(0.35, 0.35, 0.2), Color: LegoWhite, OrderIdx: 10},
				{RelPos: rl.NewVector3(0.35, 2.2, 0.6), Size: rl.NewVector3(0.35, 0.35, 0.2), Color: LegoWhite, OrderIdx: 11},
				{RelPos: rl.NewVector3(-0.6, 2.8, 0), Size: rl.NewVector3(0.45, 0.6, 0.45), Color: LegoOrange, OrderIdx: 12},
				{RelPos: rl.NewVector3(0.6, 2.8, 0), Size: rl.NewVector3(0.45, 0.6, 0.45), Color: LegoOrange, OrderIdx: 13},
				{RelPos: rl.NewVector3(0, 0.8, -0.7), Size: rl.NewVector3(0.35, 0.35, 0.7), Color: LegoYellow, OrderIdx: 14},
			},
		},
		// 6. TURBO RACE CAR (Flame Red & Gold)
		{
			ID:          5,
			Name:        "Turbo Race Car",
			Theme:       ThemeCar,
			BaseScale:   1.1,
			Description: "Masina de curse ultra-rapida!",
			Icon:        "🏎️",
			Bricks: []LegoBrick{
				{RelPos: rl.NewVector3(-0.9, 0.4, 1.0), Size: rl.NewVector3(0.5, 0.8, 0.8), Color: LegoBlack, OrderIdx: 1},
				{RelPos: rl.NewVector3(0.9, 0.4, 1.0), Size: rl.NewVector3(0.5, 0.8, 0.8), Color: LegoBlack, OrderIdx: 2},
				{RelPos: rl.NewVector3(-0.9, 0.4, -1.0), Size: rl.NewVector3(0.5, 0.8, 0.8), Color: LegoBlack, OrderIdx: 3},
				{RelPos: rl.NewVector3(0.9, 0.4, -1.0), Size: rl.NewVector3(0.5, 0.8, 0.8), Color: LegoBlack, OrderIdx: 4},
				{RelPos: rl.NewVector3(0, 0.5, 0), Size: rl.NewVector3(1.6, 0.4, 3.0), Color: LegoDarkGrey, OrderIdx: 5},
				{RelPos: rl.NewVector3(0, 0.9, 0), Size: rl.NewVector3(1.8, 0.6, 2.8), Color: LegoRed, OrderIdx: 6},
				{RelPos: rl.NewVector3(0, 1.3, -0.2), Size: rl.NewVector3(1.4, 0.6, 1.4), Color: LegoCyan, OrderIdx: 7},
				{RelPos: rl.NewVector3(0, 1.7, -0.2), Size: rl.NewVector3(1.2, 0.3, 1.2), Color: LegoRed, OrderIdx: 8},
				{RelPos: rl.NewVector3(0, 0.9, 1.4), Size: rl.NewVector3(1.4, 0.4, 0.6), Color: LegoGold, OrderIdx: 9},
				{RelPos: rl.NewVector3(-0.5, 1.0, 1.6), Size: rl.NewVector3(0.3, 0.3, 0.2), Color: LegoWhite, OrderIdx: 10},
				{RelPos: rl.NewVector3(0.5, 1.0, 1.6), Size: rl.NewVector3(0.3, 0.3, 0.2), Color: LegoWhite, OrderIdx: 11},
				{RelPos: rl.NewVector3(-0.6, 1.4, -1.3), Size: rl.NewVector3(0.3, 0.7, 0.3), Color: LegoDarkGrey, OrderIdx: 12},
				{RelPos: rl.NewVector3(0.6, 1.4, -1.3), Size: rl.NewVector3(0.3, 0.7, 0.3), Color: LegoDarkGrey, OrderIdx: 13},
				{RelPos: rl.NewVector3(0, 1.8, -1.4), Size: rl.NewVector3(2.0, 0.2, 0.6), Color: LegoGold, OrderIdx: 14},
			},
		},
		// 7. MONSTER TRUCK (Lime Green & Black)
		{
			ID:          6,
			Name:        "Monster Truck",
			Theme:       ThemeCar,
			BaseScale:   1.2,
			Description: "Camion urias cu roti masive!",
			Icon:        "🚙",
			Bricks: []LegoBrick{
				{RelPos: rl.NewVector3(-1.2, 0.6, 1.2), Size: rl.NewVector3(0.8, 1.2, 1.2), Color: LegoBlack, OrderIdx: 1},
				{RelPos: rl.NewVector3(1.2, 0.6, 1.2), Size: rl.NewVector3(0.8, 1.2, 1.2), Color: LegoBlack, OrderIdx: 2},
				{RelPos: rl.NewVector3(-1.2, 0.6, -1.2), Size: rl.NewVector3(0.8, 1.2, 1.2), Color: LegoBlack, OrderIdx: 3},
				{RelPos: rl.NewVector3(1.2, 0.6, -1.2), Size: rl.NewVector3(0.8, 1.2, 1.2), Color: LegoBlack, OrderIdx: 4},
				{RelPos: rl.NewVector3(0, 0.9, 0), Size: rl.NewVector3(1.8, 0.6, 3.2), Color: LegoDarkGrey, OrderIdx: 5},
				{RelPos: rl.NewVector3(0, 1.5, 0), Size: rl.NewVector3(2.0, 0.8, 3.0), Color: LegoLime, OrderIdx: 6},
				{RelPos: rl.NewVector3(0, 2.1, -0.2), Size: rl.NewVector3(1.8, 0.8, 1.8), Color: LegoCyan, OrderIdx: 7},
				{RelPos: rl.NewVector3(0, 2.6, -0.2), Size: rl.NewVector3(1.8, 0.2, 1.8), Color: LegoLime, OrderIdx: 8},
				{RelPos: rl.NewVector3(0, 1.6, 1.5), Size: rl.NewVector3(1.6, 0.5, 0.6), Color: LegoYellow, OrderIdx: 9},
				{RelPos: rl.NewVector3(-0.7, 1.7, 1.7), Size: rl.NewVector3(0.3, 0.3, 0.2), Color: LegoWhite, OrderIdx: 10},
				{RelPos: rl.NewVector3(0.7, 1.7, 1.7), Size: rl.NewVector3(0.3, 0.3, 0.2), Color: LegoWhite, OrderIdx: 11},
				{RelPos: rl.NewVector3(0, 2.8, -0.2), Size: rl.NewVector3(1.2, 0.3, 0.4), Color: LegoOrange, OrderIdx: 12}, // Roof lights
			},
		},
		// 8. ROBO-TITAN (Gold & Cobalt Blue)
		{
			ID:          7,
			Name:        "Robo-Titan",
			Theme:       ThemeRobot,
			BaseScale:   1.0,
			Description: "Robotul de otel cu laser si scut!",
			Icon:        "🤖",
			Bricks: []LegoBrick{
				{RelPos: rl.NewVector3(-0.6, 0.5, 0), Size: rl.NewVector3(0.7, 1.0, 0.9), Color: LegoDarkGrey, OrderIdx: 1},
				{RelPos: rl.NewVector3(0.6, 0.5, 0), Size: rl.NewVector3(0.7, 1.0, 0.9), Color: LegoDarkGrey, OrderIdx: 2},
				{RelPos: rl.NewVector3(0, 1.4, 0), Size: rl.NewVector3(1.8, 1.0, 1.2), Color: LegoGold, OrderIdx: 3},
				{RelPos: rl.NewVector3(0, 2.2, 0), Size: rl.NewVector3(2.0, 1.4, 1.4), Color: LegoBlue, OrderIdx: 4},
				{RelPos: rl.NewVector3(0, 2.2, 0.7), Size: rl.NewVector3(0.8, 0.8, 0.2), Color: LegoCyan, OrderIdx: 5},
				{RelPos: rl.NewVector3(-1.2, 2.2, 0), Size: rl.NewVector3(0.6, 1.2, 0.7), Color: LegoGold, OrderIdx: 6},
				{RelPos: rl.NewVector3(1.2, 2.2, 0), Size: rl.NewVector3(0.6, 1.2, 0.7), Color: LegoGold, OrderIdx: 7},
				{RelPos: rl.NewVector3(-1.3, 1.4, 0.3), Size: rl.NewVector3(0.5, 0.8, 0.5), Color: LegoDarkGrey, OrderIdx: 8},
				{RelPos: rl.NewVector3(1.3, 1.4, 0.3), Size: rl.NewVector3(0.5, 0.8, 0.5), Color: LegoDarkGrey, OrderIdx: 9},
				{RelPos: rl.NewVector3(-1.3, 1.1, 0.9), Size: rl.NewVector3(0.8, 0.8, 0.3), Color: LegoCyan, OrderIdx: 10},
				{RelPos: rl.NewVector3(1.3, 1.1, 1.1), Size: rl.NewVector3(0.4, 0.4, 1.2), Color: LegoYellow, OrderIdx: 11},
				{RelPos: rl.NewVector3(0, 3.2, 0), Size: rl.NewVector3(1.2, 1.0, 1.1), Color: LegoGold, OrderIdx: 12},
				{RelPos: rl.NewVector3(0, 3.2, 0.5), Size: rl.NewVector3(0.9, 0.3, 0.3), Color: LegoCyan, OrderIdx: 13},
				{RelPos: rl.NewVector3(0, 3.9, 0), Size: rl.NewVector3(0.2, 0.6, 0.2), Color: LegoRed, OrderIdx: 14},
			},
		},
	}
}

// GetArenaLetterBlueprints builds the block-letter construction sites in the plaza.
func GetArenaLetterBlueprints() []LegoItemDefinition {
	glyphs := map[rune][5]string{
		'A': {".###.", "#...#", "#####", "#...#", "#...#"},
		'R': {"####.", "#...#", "####.", "#.#..", "#..##"},
		'E': {"#####", "#....", "####.", "#....", "#####"},
		'N': {"#...#", "##..#", "#.#.#", "#..##", "#...#"},
	}
	colors := []rl.Color{LegoGold, LegoRed, LegoBlue, LegoGreen, LegoOrange}
	letters := make([]LegoItemDefinition, 0, len("ARENA"))
	for i, ch := range "ARENA" {
		item := LegoItemDefinition{ID: 100 + i, Name: "Litera [ " + string(ch) + " ]", Theme: ThemeArenaLetter, BaseScale: 1, Icon: "ABC"}
		for row := 4; row >= 0; row-- {
			for column, cell := range glyphs[ch][row] {
				if cell == '#' {
					item.Bricks = append(item.Bricks, LegoBrick{
						RelPos: rl.NewVector3(float32(column-2)*.62, float32(4-row)+.5, 0),
						Size:   rl.NewVector3(.62, 1, .62), Color: colors[i], OrderIdx: len(item.Bricks) + 1,
					})
				}
			}
		}
		letters = append(letters, item)
	}
	return letters
}
