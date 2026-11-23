package main

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"os"
	"path/filepath"
	"strings"
)

var uiShotHook func(*Game)

func (g *Game) useUI6() bool { return true }
func u6s() float32           { return minf(sw()/1000, sh()/640) }
func (g *Game) u6bg() {
	g.drawMenuBackdrop()
	gradientV(0, 0, ScreenWidth, ScreenHeight, rl.NewColor(15, 48, 90, 100), rl.NewColor(9, 19, 54, 205))
}
func (g *Game) u6title() {
	s := u6s()
	g.textCenter("BATTLE WORLDS 6.0", sw()/2, 18*s, 50*s, LegoGoldBright, true)
	g.textCenter("Lumile Battle", sw()/2, 72*s, 23*s, UIText, true)
}
func (g *Game) u6card(r rl.Rectangle, selected bool, col rl.Color) (bool, bool) {
	hover := rl.CheckCollisionPointRec(rl.GetMousePosition(), r)
	if hover {
		selected = true
	}
	border := col
	if !selected {
		border = rl.NewColor(120, 164, 194, 255)
	}
	rl.DrawRectangleRounded(rl.NewRectangle(r.X+3, r.Y+5, r.Width, r.Height), .25, 10, rl.NewColor(4, 12, 35, 100))
	rl.DrawRectangleRounded(r, .25, 10, rl.NewColor(25, 48, 82, 235))
	rl.DrawRectangleRoundedLinesEx(r, .25, 10, 3*u6s(), border)
	return hover, hover && rl.IsMouseButtonPressed(rl.MouseButtonLeft)
}
func (g *Game) u6icon(k int, x, y, s float32) {
	switch k {
	case 0:
		rl.DrawTriangle(vv(x-s*.28, y-s*.42), vv(x-s*.28, y+s*.42), vv(x+s*.42, y), UIGreen)
	case 1:
		rl.DrawRectangleRounded(rl.NewRectangle(x-s*.1, y-s*.42, s*.2, s*.84), .4, 6, UIGreen)
		rl.DrawRectangleRounded(rl.NewRectangle(x-s*.42, y-s*.1, s*.84, s*.2), .4, 6, UIGreen)
	case 2:
		drawIcon(IconExit, x, y, s, UIRed)
	case 3:
		DrawStar(x, y, s*.48, LegoGoldBright)
	case 4:
		g.textCenter("?", x, y-s*.52, s, UIText, true)
	case 5:
		rl.DrawCircleV(vv(x-s*.2, y-s*.12), s*.18, LegoYellow)
		rl.DrawCircleV(vv(x+s*.2, y-s*.12), s*.18, UIAccent)
		rl.DrawCircleV(vv(x-s*.2, y+s*.25), s*.22, LegoYellow)
		rl.DrawCircleV(vv(x+s*.2, y+s*.25), s*.22, UIAccent)
	case 6:
		drawIcon(IconSave, x, y, s, UIText)
	case 7:
		rl.DrawCircleV(vv(x, y), s*.23, LegoGoldBright)
		for i := 0; i < 8; i++ {
			a := float32(i) * .7854
			rl.DrawLineEx(vv(x+cosf(a)*s*.3, y+sinf(a)*s*.3), vv(x+cosf(a)*s*.45, y+sinf(a)*s*.45), s*.06, LegoGoldBright)
		}
	case 8:
		drawIcon(IconMouse, x, y, s, UIText)
	case 9:
		drawIcon(IconWorld, x, y, s, UIText)
	case 10:
		drawIcon(IconCheck, x, y, s, UIGreen)
	}
}
func (g *Game) u6button(r rl.Rectangle, label string, k int, sel bool, col rl.Color) (bool, bool) {
	h, c := g.u6card(r, sel, col)
	s := minf(r.Height*.52, r.Width*.36)
	g.u6icon(k, r.X+r.Width/2, r.Y+r.Height*.37, s)
	g.textCenter(label, r.X+r.Width/2, r.Y+r.Height*.75, minf(19*u6s(), r.Height*.16), UIText, true)
	return h, c
}
func (g *Game) drawPlayerSelect6() {
	g.u6bg()
	g.u6title()
	s := u6s()
	n := len(g.PlayerList) + 2
	cols := 3
	rows := (n + cols - 1) / cols
	w := minf(250*s, (sw()-80*s)/3)
	h := minf(130*s, (sh()-155*s)/float32(rows))
	start := (sw() - 3*w - 20*s) / 2
	for i := 0; i < n; i++ {
		r := rl.NewRectangle(start+float32(i%cols)*(w+10*s), 120*s+float32(i/cols)*h, w, h-8*s)
		label := "Jucător nou"
		k := 1
		col := UIGreen
		if i == n-1 {
			label = "Ieșire"
			k = 2
			col = UIRed
		}
		if i > 0 && i < n-1 {
			name := g.PlayerList[i-1]
			label = name
			if _, p := FindPlayer(g.SaveData, name); p != nil {
				label = fmt.Sprintf("%s • nivelul %d", name, maxInt(p.Level, 1))
			}
			k = -1
			col = UIAccent
		}
		var hover, click bool
		if k >= 0 {
			hover, click = g.u6button(r, label, k, g.SelectIdx == i, col)
		} else {
			hover, click = g.u6card(r, g.SelectIdx == i, col)
			x, y := r.X+r.Width/2, r.Y+r.Height*.35
			rl.DrawCircleV(vv(x, y), r.Height*.22, Realms[(i-1)%len(Realms)].Main)
			initial := []rune(strings.ToUpper(g.PlayerList[i-1]))
			if len(initial) > 0 {
				g.textCenter(string(initial[0]), x, y-r.Height*.17, r.Height*.29, UIText, true)
			}
			g.textCenter(label, x, r.Y+r.Height*.72, minf(18*s, r.Height*.16), UIText, true)
		}
		if hover {
			g.SelectIdx = i
		}
		if click {
			g.selectPlayerOption(i)
		}
	}
}
func (g *Game) drawPlayerCreate6() {
	g.u6bg()
	g.u6title()
	s := u6s()
	r := rl.NewRectangle(sw()/2-300*s, 150*s, 600*s, 350*s)
	panel(r, UIAccent)
	g.textCenter("Cum te cheamă?", sw()/2, r.Y+25*s, 28*s, UIText, true)
	box := rl.NewRectangle(r.X+30*s, r.Y+75*s, r.Width-60*s, 65*s)
	rl.DrawRectangleRounded(box, .2, 8, rl.NewColor(10, 25, 50, 255))
	name := g.CreateName
	if int(g.Time*2.5)%2 == 0 {
		name += "|"
	}
	g.text(name, box.X+15*s, box.Y+10*s, 35*s, UIText, true)
	_, ok := g.u6button(rl.NewRectangle(r.X+40*s, r.Y+165*s, 230*s, 150*s), "Gata", 10, false, UIGreen)
	_, back := g.u6button(rl.NewRectangle(r.X+330*s, r.Y+165*s, 230*s, 150*s), "Înapoi", 2, false, UIRed)
	if ok {
		g.confirmCreate()
	}
	if back {
		g.Scene = ScenePlayerSelect
	}
	if g.MessageT > 0 {
		g.textCenter(g.Message, sw()/2, r.Y+r.Height+10*s, 22*s, LegoYellow, true)
	}
}
func (g *Game) drawMainMenu6() {
	g.u6bg()
	g.u6title()
	s := u6s()
	g.textCenter(fmt.Sprintf("%s • nivelul %d • %d aur • %s", g.Player.Name, g.Level, g.Player.Gold, difficultyOptions[clampInt(g.DifficultyIdx, 0, len(difficultyOptions)-1)].Name), sw()/2, 105*s, 19*s, UIText, true)
	w := 150 * s
	gap := 12 * s
	x := (sw() - (5*w + 4*gap)) / 2
	for i, label := range mainMenuItems {
		k := []int{0, 3, 4, 5, 2}[i]
		col := UIAccent
		if i == 0 {
			col = UIGreen
		}
		if i == 4 {
			col = UIRed
		}
		h, c := g.u6button(rl.NewRectangle(x+float32(i)*(w+gap), 155*s, w, 150*s), label, k, g.MenuIdx == i, col)
		if h {
			g.MenuIdx = i
		}
		if c {
			g.selectMainMenu(i)
		}
	}
	for i, r := range Realms {
		cx := sw()/2 + (float32(i)-float32(len(Realms)-1)/2)*105*s
		cy := 410 * s
		rl.DrawCircleV(vv(cx, cy), 38*s, r.Main)
		rl.DrawRing(vv(cx, cy), 33*s, 38*s, 0, 360, 32, r.Accent)
		if !g.realmOpen(i) {
			drawIcon(IconLock, cx, cy, 42*s, rl.White)
		}
		if i == curRealm {
			DrawStar(cx, cy-53*s, 13*s, LegoGoldBright)
		}
	}
	done, total := g.Achieve.Count()
	g.textCenter(fmt.Sprintf("Realizări: %d din %d • Construcții: %d", done, total, g.TotalBuilt), sw()/2, sh()-40*s, 18*s, UIText, false)
}
func (g *Game) drawDifficulty6() {
	g.u6bg()
	g.u6title()
	s := u6s()
	n := len(difficultyOptions)
	w := minf(210*s, (sw()-40*s)/float32(n))
	for i, opt := range difficultyOptions {
		r := rl.NewRectangle((sw()-float32(n)*w)/2+float32(i)*w, 170*s, w-10*s, 270*s)
		h, c := g.u6card(r, g.DifficultyIdx == i, opt.Color)
		if h {
			g.DifficultyIdx = i
		}
		if c {
			g.chooseDifficulty(i)
			return
		}
		count := maxInt(i, 1)
		if count > 3 {
			count = 1
		}
		for j := 0; j < count; j++ {
			DrawStar(r.X+r.Width/2+(float32(j)-float32(count-1)/2)*42*s, r.Y+75*s, 20*s, opt.Color)
		}
		g.textCenter(opt.Name, r.X+r.Width/2, r.Y+135*s, 23*s, opt.Color, true)
		line := ""
		y := r.Y + 175*s
		for _, word := range strings.Fields(opt.Desc) {
			next := strings.TrimSpace(line + " " + word)
			if line != "" && g.measure(next, 14*s, false) > r.Width-20*s {
				g.textCenter(line, r.X+r.Width/2, y, 14*s, UIText, false)
				y += 19 * s
				line = word
			} else {
				line = next
			}
		}
		g.textCenter(line, r.X+r.Width/2, y, 14*s, UIText, false)
	}
}
func (g *Game) drawPause6() {
	g.render3D(g.Camera, true)
	gradientV(0, 0, ScreenWidth, ScreenHeight, rl.NewColor(15, 40, 80, 180), rl.NewColor(9, 19, 50, 220))
	s := u6s()
	g.textCenter("PAUZĂ", sw()/2, 22*s, 48*s, LegoGoldBright, true)
	items := g.pauseItems()
	w, h := 180*s, 145*s
	x := (sw() - 4*w - 30*s) / 2
	for i, label := range items {
		k := []int{0, 6, 7, 8, 4, 9, 2}[i]
		col := UIAccent
		if i == 6 {
			col = UIRed
		}
		hover, click := g.u6button(rl.NewRectangle(x+float32(i%4)*(w+10*s), 105*s+float32(i/4)*(h+12*s), w, h), label, k, g.PauseIdx == i, col)
		if hover {
			g.PauseIdx = i
		}
		if click {
			g.selectPause(i)
		}
	}
	if g.MessageT > 0 {
		g.textCenter(g.Message, sw()/2, sh()-45*s, 22*s, LegoYellow, true)
	}
}
func (g *Game) drawGameOver6() {
	g.render3D(g.Camera, true)
	rl.DrawRectangle(0, 0, ScreenWidth, ScreenHeight, rl.NewColor(25, 20, 60, 190))
	s := u6s()
	g.textCenter("AI FOST ÎNVINS!", sw()/2, 100*s, 48*s, UIRed, true)
	g.textCenter("Eroii se ridică mereu!", sw()/2, 165*s, 25*s, UIText, true)
	drawHeart(sw()/2, 280*s, 110*s, UIRed)
	g.u6icon(0, sw()/2, 410*s, 80*s)
	g.textCenter("Click sau ENTER ca să reînvii", sw()/2, 490*s, 22*s, UIText, true)
}
func (g *Game) drawLoading6(msg string, pct float32) {
	rl.BeginDrawing()
	gradientV(0, 0, ScreenWidth, ScreenHeight, rl.NewColor(92, 185, 245, 255), rl.NewColor(28, 70, 150, 255))
	s := u6s()
	g.u6title()
	left, right := sw()*.15, sw()*.85
	y := sh() * .65
	rl.DrawLineEx(vv(left, y), vv(right, y), 8*s, UIText)
	for i := 0; i < 15; i++ {
		x := left + (right-left)*float32(i)/14
		rl.DrawLineEx(vv(x, y-8*s), vv(x, y+8*s), 4*s, LegoGoldBright)
	}
	drawIcon(IconTrain, left+(right-left)*clampf(pct, 0, 1), y-45*s, 90*s, rl.White)
	g.textCenter(msg, sw()/2, y+45*s, 21*s, UIText, true)
	rl.EndDrawing()
}
func init() {
	if dir := os.Getenv("BATTLE_UI6_SHOTS"); dir != "" {
		uiShotHook = func(g *Game) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				panic(err)
			}
			screens := []struct {
				name string
				draw func()
			}{{"select", g.drawPlayerSelect6}, {"create", g.drawPlayerCreate6}, {"main", g.drawMainMenu6}, {"difficulty", g.drawDifficulty6}, {"pause", g.drawPause6}, {"gameover", g.drawGameOver6}}
			for _, v := range screens {
				for i := 0; i < 2; i++ {
					rl.BeginDrawing()
					v.draw()
					rl.EndDrawing()
				}
				rl.TakeScreenshot(filepath.Join(dir, v.name+".png"))
			}
			for i := 0; i < 2; i++ {
				g.drawLoading6("Pregătim lumile...", .55)
			}
			rl.TakeScreenshot(filepath.Join(dir, "loading.png"))
			os.Exit(0)
		}
	}
}
