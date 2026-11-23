package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	UIPanel    = rl.NewColor(24, 30, 46, 232)
	UIPanel2   = rl.NewColor(40, 50, 76, 245)
	UIBorder   = rl.NewColor(255, 205, 60, 255)
	UIText     = rl.NewColor(245, 245, 240, 255)
	UIMuted    = rl.NewColor(190, 200, 215, 255)
	UIAccent   = rl.NewColor(80, 200, 255, 255)
	UIGreen    = rl.NewColor(90, 210, 110, 255)
	UIRed      = rl.NewColor(235, 80, 80, 255)
	UIShadowTx = rl.NewColor(0, 0, 0, 140)
)

func romanianCodepoints() []rune {
	r := []rune{}
	for c := rune(32); c < 127; c++ {
		r = append(r, c)
	}
	r = append(r, []rune("ăâîșțĂÂÎȘȚşţŞŢ°–—…’„”«»·•→←↑↓×")...)
	return r
}

func (g *Game) LoadFonts() {
	cps := romanianCodepoints()
	fonts := filepath.Join(os.Getenv("WINDIR"), "Fonts")
	g.Font = loadInterfaceFont(cps, 64, filepath.Join(fonts, "segoeui.ttf"), "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf")
	g.HasFont = g.Font.BaseSize > 0
	g.FontBold = loadInterfaceFont(cps, 72, filepath.Join(fonts, "segoeuib.ttf"), "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf")
	if g.FontBold.BaseSize == 0 {
		g.FontBold = g.Font
	}
}

func loadInterfaceFont(codepoints []rune, size int32, files ...string) rl.Font {
	for _, file := range files {
		if !fileExists(file) {
			continue
		}
		font := rl.LoadFontEx(file, size, codepoints)
		if font.BaseSize > 0 {
			rl.SetTextureFilter(font.Texture, rl.FilterBilinear)
			return font
		}
	}
	return rl.Font{}
}

func (g *Game) font(bold bool) rl.Font {
	if bold && g.FontBold.BaseSize > 0 {
		return g.FontBold
	}
	return g.Font
}

func (g *Game) measure(text string, size float32, bold bool) float32 {
	if !g.HasFont {
		return float32(rl.MeasureText(text, int32(size)))
	}
	return rl.MeasureTextEx(g.font(bold), text, size, 1).X
}

func (g *Game) text(text string, x, y, size float32, col rl.Color, bold bool) {
	if !g.HasFont {
		rl.DrawText(text, int32(x), int32(y), int32(size), col)
		return
	}
	f := g.font(bold)
	rl.DrawTextEx(f, text, rl.NewVector2(x+1.5, y+2), size, 1, rl.NewColor(0, 0, 0, col.A/2))
	rl.DrawTextEx(f, text, rl.NewVector2(x, y), size, 1, col)
}

func (g *Game) textCenter(text string, cx, y, size float32, col rl.Color, bold bool) {
	g.text(text, cx-g.measure(text, size, bold)/2, y, size, col, bold)
}

func panel(r rl.Rectangle, border rl.Color) {
	rl.DrawRectangleRounded(rl.NewRectangle(r.X+4, r.Y+6, r.Width, r.Height), 0.12, 10, rl.NewColor(0, 0, 0, 70))
	rl.DrawRectangleRounded(r, 0.12, 10, UIPanel)
	rl.DrawRectangleRoundedLinesEx(r, 0.12, 10, 3, border)
}

// button deseneaza un buton; intoarce (hover, click)
func (g *Game) button(r rl.Rectangle, label string, selected bool, accent rl.Color) (bool, bool) {
	m := rl.GetMousePosition()
	hover := rl.CheckCollisionPointRec(m, r)
	bg := UIPanel
	border := rl.NewColor(110, 120, 140, 255)
	if selected || hover {
		bg = UIPanel2
		border = accent
	}
	rl.DrawRectangleRounded(rl.NewRectangle(r.X+3, r.Y+5, r.Width, r.Height), 0.3, 10, rl.NewColor(0, 0, 0, 80))
	rl.DrawRectangleRounded(r, 0.3, 10, bg)
	rl.DrawRectangleRoundedLinesEx(r, 0.3, 10, 3, border)
	col := UIText
	if selected || hover {
		col = accent
	}
	size := float32(26)
	g.textCenter(label, r.X+r.Width/2, r.Y+(r.Height-size)/2-2, size, col, true)
	return hover, hover && rl.IsMouseButtonPressed(rl.MouseButtonLeft)
}

func sw() float32 { return float32(ScreenWidth) }
func sh() float32 { return float32(ScreenHeight) }

// ---------------------------------------------------------------
// FUNDAL MENIURI: lumea vazuta de sus, camera care se roteste lent
// ---------------------------------------------------------------

func (g *Game) drawMenuBackdrop() {
	a := g.Time * 0.05
	cam := rl.Camera3D{Position: v3(cosf(a)*190, 95, sinf(a)*190), Target: v3(0, 0, 0), Up: v3(0, 1, 0), Fovy: 55, Projection: rl.CameraPerspective}
	g.render3D(cam, false)
	gradientV(0, 0, ScreenWidth, ScreenHeight, rl.NewColor(10, 20, 40, 60), rl.NewColor(10, 20, 40, 170))
}

func (g *Game) drawTitle(sub string) {
	t := "BATTLE WORLDS 6.0"
	size := float32(84)
	y := float32(40)
	g.textCenter(t, sw()/2+4, y+5, size, rl.NewColor(160, 60, 20, 255), true)
	g.textCenter(t, sw()/2, y, size, LegoGoldBright, true)
	if sub != "" {
		g.textCenter(sub, sw()/2, y+92, 28, UIAccent, true)
	}
}

// ---------------------------------------------------------------
// ALEGEREA JUCATORULUI
// ---------------------------------------------------------------

func (g *Game) updatePlayerSelect() {
	g.unlockCursor()
	n := len(g.PlayerList) + 2 // + creeaza, + iesire
	if rl.IsKeyPressed(rl.KeyUp) {
		g.SelectIdx = (g.SelectIdx - 1 + n) % n
		g.Play(g.Audio.Click)
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		g.SelectIdx = (g.SelectIdx + 1) % n
		g.Play(g.Audio.Click)
	}
	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeySpace) {
		g.selectPlayerOption(g.SelectIdx)
	}
	if rl.IsKeyPressed(rl.KeyEscape) {
		g.ShouldQuit = true
	}
}

func (g *Game) selectPlayerOption(i int) {
	g.Play(g.Audio.Click)
	switch {
	case i == 0:
		g.CreateName = ""
		g.Scene = ScenePlayerCreate
	case i == len(g.PlayerList)+1:
		g.ShouldQuit = true
	default:
		idx, p := FindPlayer(g.SaveData, g.PlayerList[i-1])
		if p != nil {
			g.PlayerIdx = idx
			LoadPlayerIntoGame(p, g)
			g.hookAchievements()
			g.Scene = SceneMainMenu
			g.MenuIdx = 0
		}
	}
}

func (g *Game) drawPlayerSelect() {
	g.drawMenuBackdrop()
	g.drawTitle("Lumea blocuri cu 5 insule, poduri și tuneluri")
	n := len(g.PlayerList) + 2
	rowH := float32(64)
	pw := float32(640)
	ph := 80 + float32(n)*rowH
	if ph > sh()-220 {
		ph = sh() - 220
	}
	r := rl.NewRectangle((sw()-pw)/2, 190, pw, ph)
	panel(r, UIBorder)
	g.text("Cine se joacă?", r.X+30, r.Y+18, 32, LegoGoldBright, true)
	for i := 0; i < n; i++ {
		y := r.Y + 70 + float32(i)*rowH
		if y+rowH-8 > r.Y+r.Height {
			break
		}
		label := ""
		accent := UIAccent
		switch {
		case i == 0:
			label = "+ Jucător nou"
			accent = UIGreen
		case i == n-1:
			label = "Ieșire din joc"
			accent = UIRed
		default:
			name := g.PlayerList[i-1]
			label = name
			if _, p := FindPlayer(g.SaveData, name); p != nil {
				label = fmt.Sprintf("%s  -  nivelul %d", name, maxInt(p.Level, 1))
			}
		}
		hover, click := g.button(rl.NewRectangle(r.X+20, y, r.Width-40, rowH-10), label, g.SelectIdx == i, accent)
		if hover {
			g.SelectIdx = i
		}
		if click {
			g.selectPlayerOption(i)
		}
	}
	g.textCenter("Săgeți + ENTER sau click cu mouse-ul", sw()/2, sh()-46, 22, UIMuted, false)
}

func (g *Game) updatePlayerCreate() {
	g.unlockCursor()
	for {
		ch := rl.GetCharPressed()
		if ch == 0 {
			break
		}
		if len([]rune(g.CreateName)) < 16 && ch >= 32 {
			g.CreateName += string(rune(ch))
		}
	}
	if rl.IsKeyPressed(rl.KeyBackspace) && len(g.CreateName) > 0 {
		rs := []rune(g.CreateName)
		g.CreateName = string(rs[:len(rs)-1])
	}
	if rl.IsKeyPressed(rl.KeyEnter) {
		g.confirmCreate()
	}
	if rl.IsKeyPressed(rl.KeyEscape) {
		g.Scene = ScenePlayerSelect
	}
}

func (g *Game) confirmCreate() {
	name := g.CreateName
	for len(name) > 0 && name[len(name)-1] == ' ' {
		name = name[:len(name)-1]
	}
	if name == "" {
		g.say("Scrie un nume!")
		return
	}
	if _, existing := FindPlayer(g.SaveData, name); existing != nil {
		g.say("Există deja un jucător cu acest nume.")
		return
	}
	np := CreatePlayer(name)
	g.SaveData.Players = append(g.SaveData.Players, np)
	g.PlayerIdx = len(g.SaveData.Players) - 1
	g.PlayerList = append(g.PlayerList, name)
	LoadPlayerIntoGame(&g.SaveData.Players[g.PlayerIdx], g)
	g.hookAchievements()
	g.Play(g.Audio.Click)
	g.Scene = SceneDifficulty
	g.MenuIdx = -1
}

func (g *Game) drawPlayerCreate() {
	g.drawMenuBackdrop()
	g.drawTitle("Jucător nou")
	r := rl.NewRectangle((sw()-620)/2, 220, 620, 250)
	panel(r, UIBorder)
	g.text("Cum te cheamă?", r.X+30, r.Y+24, 32, LegoGoldBright, true)
	box := rl.NewRectangle(r.X+30, r.Y+80, r.Width-60, 70)
	rl.DrawRectangleRounded(box, 0.25, 8, rl.NewColor(12, 16, 26, 255))
	rl.DrawRectangleRoundedLinesEx(box, 0.25, 8, 2, UIAccent)
	name := g.CreateName
	if int(g.Time*2.5)%2 == 0 {
		name += "|"
	}
	g.text(name, box.X+20, box.Y+14, 40, UIText, true)
	_, ok := g.button(rl.NewRectangle(r.X+30, r.Y+170, 260, 58), "Gata (ENTER)", false, UIGreen)
	_, back := g.button(rl.NewRectangle(r.X+r.Width-230, r.Y+170, 200, 58), "Înapoi", false, UIRed)
	if ok {
		g.confirmCreate()
	}
	if back {
		g.Scene = ScenePlayerSelect
	}
	if g.MessageT > 0 {
		g.textCenter(g.Message, sw()/2, r.Y+r.Height+24, 26, LegoYellow, true)
	}
}

// ---------------------------------------------------------------
// MENIUL PRINCIPAL
// ---------------------------------------------------------------

var mainMenuItems = []string{"Joacă!", "Alege dificultatea", "Cum se joacă", "Schimbă jucătorul", "Ieșire"}

func (g *Game) updateMainMenu() {
	g.unlockCursor()
	n := len(mainMenuItems)
	if rl.IsKeyPressed(rl.KeyUp) {
		g.MenuIdx = (g.MenuIdx - 1 + n) % n
		g.Play(g.Audio.Click)
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		g.MenuIdx = (g.MenuIdx + 1) % n
		g.Play(g.Audio.Click)
	}
	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeySpace) {
		g.selectMainMenu(g.MenuIdx)
	}
	if rl.IsKeyPressed(rl.KeyEscape) {
		g.Scene = ScenePlayerSelect
	}
}

func (g *Game) startPlaying() {
	g.spawnEnemies(g.Level)
	g.ResetPlayerPosition()
	g.syncFog()
	g.Prog.Reveal = RevealState{}
	g.stopGPS()
	g.Prog.Sector = -1
	g.prevPos = g.Player.Pos
	if wi := g.levelWorld(g.Level); wi > 0 && !g.Prog.Visited[wi] {
		g.startGPSToWorld(wi) // lumea noua inca nevizitata: sagetile o arata
	}
	g.Player.HP = g.Player.MaxHP
	g.Scene = SceneMapExplore
	g.ShowHelp = 14
	g.LevelBanner = 3.5
	g.lockCursor()
}

func (g *Game) selectMainMenu(i int) {
	g.Play(g.Audio.Click)
	switch i {
	case 0:
		g.startPlaying()
	case 1:
		g.Scene = SceneDifficulty
	case 2:
		g.PrevScene = SceneMainMenu
		g.Scene = SceneHelp
	case 3:
		g.autosave()
		g.Scene = ScenePlayerSelect
	case 4:
		g.autosave()
		g.ShouldQuit = true
	}
}

func (g *Game) drawMainMenu() {
	g.drawMenuBackdrop()
	g.drawTitle("")
	info := fmt.Sprintf("%s   •   nivelul %d   •   %d aur   •   %s", g.Player.Name, g.Level, g.Player.Gold, difficultyOptions[g.DifficultyIdx].Name)
	g.textCenter(info, sw()/2, 140, 26, UIAccent, true)
	bw := float32(420)
	for i, it := range mainMenuItems {
		y := 200 + float32(i)*78
		acc := UIAccent
		if i == 0 {
			acc = UIGreen
		}
		if i == len(mainMenuItems)-1 {
			acc = UIRed
		}
		hover, click := g.button(rl.NewRectangle((sw()-bw)/2, y, bw, 64), it, g.MenuIdx == i, acc)
		if hover {
			g.MenuIdx = i
		}
		if click {
			g.selectMainMenu(i)
		}
	}
	done, total := g.Achieve.Count()
	g.textCenter(fmt.Sprintf("Realizări: %d din %d   •   Construcții terminate: %d", done, total, g.TotalBuilt), sw()/2, sh()-50, 22, UIMuted, false)
}

// ---------------------------------------------------------------
// DIFICULTATE
// ---------------------------------------------------------------

func (g *Game) updateDifficulty() {
	g.unlockCursor()
	n := len(difficultyOptions)
	if rl.IsKeyPressed(rl.KeyUp) {
		g.DifficultyIdx = (g.DifficultyIdx - 1 + n) % n
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		g.DifficultyIdx = (g.DifficultyIdx + 1) % n
	}
	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeySpace) {
		g.chooseDifficulty(g.DifficultyIdx)
	}
	if rl.IsKeyPressed(rl.KeyEscape) {
		g.Scene = SceneMainMenu
	}
}

func (g *Game) chooseDifficulty(i int) {
	g.DifficultyIdx = i
	g.Difficulty = difficultyOptions[i].Key
	g.applyDifficulty()
	g.autosave()
	g.Play(g.Audio.Click)
	g.startPlaying()
}

func (g *Game) drawDifficulty() {
	g.drawMenuBackdrop()
	g.drawTitle("Alege dificultatea")
	pw := float32(760)
	for i, opt := range difficultyOptions {
		y := 190 + float32(i)*104
		r := rl.NewRectangle((sw()-pw)/2, y, pw, 90)
		m := rl.GetMousePosition()
		hover := rl.CheckCollisionPointRec(m, r)
		if hover {
			g.DifficultyIdx = i
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
				g.chooseDifficulty(i)
				return
			}
		}
		sel := g.DifficultyIdx == i
		border := rl.NewColor(90, 100, 120, 255)
		if sel {
			border = opt.Color
		}
		panel(r, border)
		if sel {
			rl.DrawRectangleRounded(rl.NewRectangle(r.X+8, r.Y+8, 10, r.Height-16), 0.5, 4, opt.Color)
		}
		g.text(opt.Name, r.X+34, r.Y+10, 34, opt.Color, true)
		g.text(opt.Desc, r.X+34, r.Y+52, 22, UIText, false)
	}
	g.textCenter("Săgeți + ENTER sau click   •   ESC: înapoi", sw()/2, sh()-46, 22, UIMuted, false)
}

// ---------------------------------------------------------------
// PAUZA + AJUTOR
// ---------------------------------------------------------------

var sensNames = []string{"foarte lent", "lent", "normal", "rapid", "foarte rapid"}

func (g *Game) pauseItems() []string {
	sh := "Umbre: pornite"
	if !g.R.ShadowsOn {
		sh = "Umbre: oprite"
	}
	ms := "Mouse: " + sensNames[clampInt(g.MouseSens, 0, len(sensNames)-1)]
	return []string{"Continuă", "Salvează jocul", sh, ms, "Cum se joacă", "Meniul principal", "Ieșire din joc"}
}

func (g *Game) updatePause() {
	g.unlockCursor()
	n := len(g.pauseItems())
	if rl.IsKeyPressed(rl.KeyUp) {
		g.PauseIdx = (g.PauseIdx - 1 + n) % n
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		g.PauseIdx = (g.PauseIdx + 1) % n
	}
	if rl.IsKeyPressed(rl.KeyEnter) {
		g.selectPause(g.PauseIdx)
	}
	if g.PauseIdx == 3 && (rl.IsKeyPressed(rl.KeyLeft) || rl.IsKeyPressed(rl.KeyRight)) {
		if rl.IsKeyPressed(rl.KeyLeft) {
			g.MouseSens = clampInt(g.MouseSens-1, 0, len(mouseSensLevels)-1)
		} else {
			g.MouseSens = clampInt(g.MouseSens+1, 0, len(mouseSensLevels)-1)
		}
		g.Play(g.Audio.Click)
	}
	if rl.IsKeyPressed(rl.KeyEscape) || rl.IsKeyPressed(rl.KeyP) {
		g.Scene = g.PrevScene
		g.lockCursor()
	}
}

func (g *Game) selectPause(i int) {
	g.Play(g.Audio.Click)
	switch i {
	case 0:
		g.Scene = g.PrevScene
		g.lockCursor()
	case 1:
		g.autosave()
		g.sayIcon(IconSave, "Jocul a fost salvat!")
	case 2:
		g.R.ShadowsOn = !g.R.ShadowsOn && g.R.ShadowRT.ID > 0
	case 3:
		g.MouseSens = (g.MouseSens + 1) % len(mouseSensLevels)
		g.autosave()
	case 4:
		g.Scene = SceneHelp
	case 5:
		g.autosave()
		g.Scene = SceneMainMenu
	case 6:
		g.autosave()
		g.ShouldQuit = true
	}
}

func (g *Game) drawPause() {
	g.render3D(g.Camera, true)
	rl.DrawRectangle(0, 0, ScreenWidth, ScreenHeight, rl.NewColor(10, 14, 26, 150))
	g.textCenter("PAUZĂ", sw()/2, 50, 60, LegoGoldBright, true)
	items := g.pauseItems()
	bw := float32(440)
	gap := minf(70, (sh()-200)/float32(len(items)))
	for i, it := range items {
		y := 140 + float32(i)*gap
		acc := UIAccent
		if i == len(items)-1 {
			acc = UIRed
		}
		hover, click := g.button(rl.NewRectangle((sw()-bw)/2, y, bw, gap-12), it, g.PauseIdx == i, acc)
		if hover {
			g.PauseIdx = i
		}
		if click {
			g.selectPause(i)
		}
	}
	if g.PauseIdx == 3 {
		g.textCenter("Click sau săgeți stânga / dreapta: cât de repede se rotește camera", sw()/2, sh()-96, 22, UIMuted, false)
	}
	if g.MessageT > 0 {
		g.textCenter(g.Message, sw()/2, sh()-60, 26, LegoYellow, true)
	}
}

func (g *Game) updateHelp() {
	g.unlockCursor()
	if rl.IsKeyPressed(rl.KeyEscape) || rl.IsKeyPressed(rl.KeyEnter) || rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		if g.PrevScene == SceneMapExplore {
			g.Scene = ScenePauseMenu
		} else {
			g.Scene = SceneMainMenu
		}
	}
}

var helpLines = [][2]string{
	{"W A S D / săgeți", "mergi (SHIFT = alergi); înoți oriunde între insule"},
	{"Mouse", "rotești camera; ținta e mereu în mijlocul ecranului"},
	{"Click stânga", "pistol de departe, sabie de aproape; mașini, bărci, avioane, pescăruși (dau aur)"},
	{"Sabia la șantier", "lovește șantierul blocuri ca să pui piese (100 aur)"},
	{"SPAȚIU / click dreapta", "sari; pe trambulină: apasă sau ține SPAȚIU ca să sari mai sus"},
	{"E", "mașină, cabina telefonică (hartă, costume, arme, LUMI), trenul magic, roată, lift, trambulină"},
	{"Nivel nou", "învingi monștrii (vin pe rând), apoi Șeful (coroana): se deschide o insulă nouă"},
	{"Lume nouă", "după ultima insulă vine trenul magic la gară: urci și pleci în altă lume"},
	{"Eroii lumilor", "Vampirul: ține SPAȚIU în aer = planezi; Eroul Zburător și Vrăjitorul: ține SPAȚIU = zbori"},
	{"Q (sau F)", "Băiatul-Păianjen: te tragi cu pânza spre locul țintit (pe acoperișuri)"},
	{"Stele", "monștrii, construcția și 5 piese aurii pe fiecare insulă; stelele aduc arme noi"},
	{"C", "în roată și în montagne russe: camera din ochii aventurierului / din spate"},
	{"M sau TAB", "harta mare (cu Marele Tunel și ieșirile lui)"},
	{"Rotița mouse-ului", "camera mai aproape / mai departe (în roată, foarte departe)"},
	{"H", "arată / ascunde ajutorul"},
	{"ESC sau P", "pauză (salvare, umbre, viteza mouse-ului, ieșire)"},
	{"F11", "ecran complet"},
}

func (g *Game) drawHelp() {
	if g.PrevScene == SceneMapExplore {
		g.render3D(g.Camera, true)
		rl.DrawRectangle(0, 0, ScreenWidth, ScreenHeight, rl.NewColor(10, 14, 26, 170))
	} else {
		g.drawMenuBackdrop()
	}
	pw := minf(980, sw()-40)
	lh := minf(46, (sh()-180)/float32(len(helpLines)))
	ph := 100 + float32(len(helpLines))*lh
	r := rl.NewRectangle((sw()-pw)/2, (sh()-ph)/2, pw, ph)
	panel(r, UIBorder)
	g.text("Cum se joacă", r.X+30, r.Y+20, 36, LegoGoldBright, true)
	for i, l := range helpLines {
		y := r.Y + 78 + float32(i)*lh
		g.text(l[0], r.X+30, y, 22, UIAccent, true)
		g.text(l[1], r.X+300, y, 22, UIText, false)
	}
	g.textCenter("Click sau ESC pentru a reveni", sw()/2, r.Y+r.Height+16, 22, UIMuted, false)
}

// ---------------------------------------------------------------
// NIVEL TERMINAT / AI FOST INVINS
// ---------------------------------------------------------------

func (g *Game) updateLevelComplete() {
	g.unlockCursor()
	g.updateParticles()
	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeySpace) || rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		g.spawnEnemies(g.Level)
		g.Player.HP = g.Player.MaxHP
		g.Scene = SceneMapExplore
		g.LevelBanner = 3.5
		g.lockCursor()
	}
}

func (g *Game) drawLevelComplete() {
	g.render3D(g.Camera, true)
	rl.DrawRectangle(0, 0, ScreenWidth, ScreenHeight, rl.NewColor(10, 20, 40, 160))
	for i := 0; i < 5; i++ {
		a := g.Time*0.8 + float32(i)*1.256
		DrawStar(sw()/2+cosf(a)*260, 210+sinf(a)*50, 22, LegoGoldBright)
	}
	DrawTrophy(sw()/2, 210, 120)
	g.textCenter(fmt.Sprintf("NIVELUL %d TERMINAT!", g.Level-1), sw()/2, 300, 64, LegoGoldBright, true)
	g.textCenter(fmt.Sprintf("Bravo, %s! Ai eliberat toate cele 5 insule!", g.Player.Name), sw()/2, 380, 30, UIText, true)
	g.textCenter(fmt.Sprintf("Aur: %d   •   Argint: %d   •   Scor: %d", g.Player.Gold, g.Player.Silver, g.Score), sw()/2, 430, 26, LegoYellow, false)
	drawIcon(IconMouse, sw()/2-60, 520, 44, rl.White)
	g.drawKeycap("ENTER", sw()/2+40, 520, 44, true, 255)
	g.hint(fmt.Sprintf("Click sau ENTER pentru nivelul %d", g.Level), sw()/2, 556, 255)
}

func (g *Game) updateGameOver() {
	g.unlockCursor()
	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeySpace) || rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		g.Player.HP = g.Player.MaxHP
		g.ResetPlayerPosition()
		g.Invuln = 4
		for i := range g.Enemies {
			g.Enemies[i].Stun = 3
			g.Enemies[i].Windup = 0
		}
		g.Projectiles = g.Projectiles[:0]
		g.Scene = SceneMapExplore
		g.sayIcon(IconHeart, "Ai reînviat în piața centrală. Hai din nou!")
		g.lockCursor()
	}
}

func (g *Game) drawGameOver() {
	g.render3D(g.Camera, true)
	rl.DrawRectangle(0, 0, ScreenWidth, ScreenHeight, rl.NewColor(40, 10, 10, 170))
	g.textCenter("AI FOST ÎNVINS!", sw()/2, 220, 68, UIRed, true)
	g.textCenter("Nu-i nimic, eroii se ridică mereu!", sw()/2, 310, 30, UIText, true)
	drawHeart(sw()/2, 420, 90, UIRed)
	drawIcon(IconMouse, sw()/2-60, 520, 44, rl.White)
	g.drawKeycap("ENTER", sw()/2+40, 520, 44, true, 255)
	g.hint("Click sau ENTER ca să reînvii în piața centrală", sw()/2, 556, 255)
}

// ---------------------------------------------------------------
// HUD
// ---------------------------------------------------------------

func drawHeart(cx, cy, s float32, col rl.Color) {
	rl.DrawCircleV(rl.NewVector2(cx-s*0.25, cy-s*0.1), s*0.3, col)
	rl.DrawCircleV(rl.NewVector2(cx+s*0.25, cy-s*0.1), s*0.3, col)
	rl.DrawTriangle(rl.NewVector2(cx-s*0.54, cy), rl.NewVector2(cx, cy+s*0.55), rl.NewVector2(cx+s*0.54, cy), col)
}

func drawCoin(cx, cy, r float32, col rl.Color) {
	rl.DrawCircleV(rl.NewVector2(cx, cy), r, shade(col, 0.75))
	rl.DrawCircleV(rl.NewVector2(cx, cy), r*0.8, col)
	rl.DrawCircleV(rl.NewVector2(cx-r*0.25, cy-r*0.25), r*0.25, rl.NewColor(255, 255, 255, 120))
}

// linie cu contur inchis (se vede pe orice fundal)
func outlinedLine(a, b rl.Vector2, thick float32, col rl.Color) {
	rl.DrawLineEx(a, b, thick+3, rl.NewColor(10, 12, 20, 200))
	rl.DrawLineEx(a, b, thick, col)
}

// drawCrosshair: tinta mare, cu contur, in mijlocul ecranului; rosie pe inamic
func (g *Game) drawCrosshair() {
	if g.Weapon == WeaponNone || g.Fade > 0 {
		return
	}
	cx, cy := sw()/2, sh()/2
	onEnemy := g.AimEnemy >= 0 || g.AimCar >= 0 || g.AimKind != TgtNone
	col := rl.NewColor(255, 255, 255, 240)
	if onEnemy {
		col = rl.NewColor(255, 70, 60, 255)
	}
	if g.Weapon == WeaponGun {
		gap := float32(7)
		ln := float32(13)
		if g.ShootCD > 0.1 {
			gap += 4 * (g.ShootCD - 0.1) / 0.1
		}
		outlinedLine(rl.NewVector2(cx-gap-ln, cy), rl.NewVector2(cx-gap, cy), 3, col)
		outlinedLine(rl.NewVector2(cx+gap, cy), rl.NewVector2(cx+gap+ln, cy), 3, col)
		outlinedLine(rl.NewVector2(cx, cy-gap-ln), rl.NewVector2(cx, cy-gap), 3, col)
		outlinedLine(rl.NewVector2(cx, cy+gap), rl.NewVector2(cx, cy+gap+ln), 3, col)
		rl.DrawCircleV(rl.NewVector2(cx, cy), 3.5, rl.NewColor(10, 12, 20, 220))
		rl.DrawCircleV(rl.NewVector2(cx, cy), 2.2, col)
		if onEnemy {
			rl.DrawRing(rl.NewVector2(cx, cy), 22, 25, 0, 360, 36, rl.NewColor(255, 70, 60, 200))
		}
	} else {
		// sabia: arc de cerc (lovitura de aproape)
		rl.DrawRing(rl.NewVector2(cx, cy), 15, 19, 200, 340, 24, rl.NewColor(10, 12, 20, 200))
		rl.DrawRing(rl.NewVector2(cx, cy), 16, 18, 205, 335, 24, col)
		rl.DrawCircleV(rl.NewVector2(cx, cy), 3, col)
	}
	if onEnemy {
		bw := float32(80)
		var pct float32
		if g.AimEnemy >= 0 && g.AimEnemy < len(g.Enemies) {
			e := &g.Enemies[g.AimEnemy]
			pct = clampf(float32(e.HP/e.MaxHP), 0, 1)
		} else if g.AimCar >= 0 {
			pct = clampf(g.City.Cars[g.AimCar].HP/CarMaxHP, 0, 1)
		} else {
			pct = g.aimThingHP()
		}
		if pct >= 0 {
			rl.DrawRectangleRounded(rl.NewRectangle(cx-bw/2, cy+32, bw, 9), 0.5, 6, rl.NewColor(20, 10, 10, 220))
			rl.DrawRectangleRounded(rl.NewRectangle(cx-bw/2, cy+32, bw*pct, 9), 0.5, 6, UIRed)
		}
	}
}

func (g *Game) drawWeaponBadge() {
	if g.Weapon == WeaponNone {
		return
	}
	r := rl.NewRectangle(sw()-200, sh()-86, 180, 66)
	panel(r, UIBorder)
	name := GunNames[clampInt(g.Prog.Gun, 0, NumGuns-1)]
	hint := "trage de departe"
	if g.Weapon == WeaponSword {
		name = "Sabie laser"
		hint = "lovește de aproape"
	}
	// iconita
	ix, iy := r.X+34, r.Y+33
	if g.Weapon == WeaponGun {
		drawGunPicture(g.Prog.Gun, ix, iy, 44, 255)
	} else {
		rl.DrawLineEx(rl.NewVector2(ix-14, iy+14), rl.NewVector2(ix+12, iy-14), 5, LegoCyan)
		rl.DrawLineEx(rl.NewVector2(ix-14, iy+3), rl.NewVector2(ix-3, iy+14), 4, LegoGold)
	}
	g.text(name, r.X+62, r.Y+14, 17, LegoGoldBright, true)
	g.text(hint, r.X+62, r.Y+38, 14, UIMuted, false)
}

func (g *Game) drawHUD() {
	p := &g.Player
	// lovit: marginile ecranului se inrosesc
	if g.HurtFlash > 0 {
		a := uint8(160 * clampf(g.HurtFlash/0.45, 0, 1))
		t := int32(70)
		rl.DrawRectangle(0, 0, ScreenWidth, t, rl.NewColor(220, 30, 30, a/2))
		rl.DrawRectangle(0, ScreenHeight-t, ScreenWidth, t, rl.NewColor(220, 30, 30, a/2))
		rl.DrawRectangle(0, 0, t, ScreenHeight, rl.NewColor(220, 30, 30, a/2))
		rl.DrawRectangle(ScreenWidth-t, 0, t, ScreenHeight, rl.NewColor(220, 30, 30, a/2))
	}
	// panoul jucatorului
	r := rl.NewRectangle(18, 18, 360, 132)
	panel(r, UIBorder)
	g.text(p.Name, r.X+18, r.Y+10, 30, LegoGoldBright, true)
	// nivelul: o stea cu cifra (textul cu dificultatea ramane mic)
	drawStar(r.X+30, r.Y+56, 13, LegoGoldBright)
	g.text(fmt.Sprintf("%d", g.Level), r.X+50, r.Y+44, 22, LegoGoldBright, true)
	g.text(fmt.Sprintf("nivelul  •  %s", difficultyOptions[g.DifficultyIdx].Name), r.X+78, r.Y+48, 15, UIMuted, false)
	drawHeart(r.X+30, r.Y+84, 26, UIRed)
	bar := rl.NewRectangle(r.X+52, r.Y+72, 220, 22)
	rl.DrawRectangleRounded(bar, 0.5, 8, rl.NewColor(60, 20, 24, 255))
	pct := clampf(float32(p.HP/p.MaxHP), 0, 1)
	hc := UIGreen
	if pct < 0.3 {
		hc = UIRed
	} else if pct < 0.6 {
		hc = LegoYellow
	}
	if pct > 0 {
		rl.DrawRectangleRounded(rl.NewRectangle(bar.X, bar.Y, bar.Width*pct, bar.Height), 0.5, 8, hc)
	}
	g.text(fmt.Sprintf("%.0f", p.HP), bar.X+bar.Width+10, bar.Y, 17, UIMuted, false)
	drawCoin(r.X+30, r.Y+114, 11, LegoGold)
	g.text(fmt.Sprintf("%d", p.Gold), r.X+46, r.Y+102, 22, LegoYellow, true)
	drawCoin(r.X+170, r.Y+114, 10, LegoLightGrey)
	g.text(fmt.Sprintf("%d", p.Silver), r.X+186, r.Y+102, 22, UIText, true)

	// inamici ramasi: cate un monstru pentru fiecare (cei invinsi = gri, bifati)
	left := g.regularLeft()
	total := len(g.Enemies)
	if bi := g.bossIndex(); bi >= 0 {
		total-- // Seful are pictograma lui (coroana), la capatul randului
	}
	if total > 0 {
		is := float32(30)
		if total > 14 {
			is = 420 / float32(total)
		}
		tw := float32(total)*(is+6) + 30 + 50
		er := rl.NewRectangle((sw()-tw)/2, 16, tw, 50)
		panel(er, UIRed)
		for k := 0; k < total; k++ {
			x := er.X + 18 + is/2 + float32(k)*(is+6)
			if k < left {
				drawIcon(IconMonster, x, er.Y+27, is, rl.NewColor(230, 70, 60, 255))
			} else {
				drawIcon(IconMonster, x, er.Y+27, is, rl.NewColor(90, 96, 110, 255))
				drawIcon(IconCheck, x, er.Y+27, is*0.7, UIGreen)
			}
		}
		// Seful: coroana gri (inca nu a venit), aurie care pulseaza (e aici)
		if bi := g.bossIndex(); bi >= 0 {
			x := er.X + er.Width - 34
			e := &g.Enemies[bi]
			switch {
			case e.Alive:
				s := 40 + 4*sinf(g.Time*6)
				drawIcon(IconCrown, x, er.Y+25, s, LegoGoldBright)
			default:
				drawIcon(IconCrown, x, er.Y+25, 36, rl.NewColor(110, 110, 125, 255))
			}
		}
		hintTxt := fmt.Sprintf("Monștri rămași: %d / %d, apoi vine Șeful (coroana)", left, total)
		if g.Prog.BossOut {
			hintTxt = "A venit Șeful! Învinge-l (coloana aurie)"
		}
		g.hint(hintTxt, sw()/2, er.Y+er.Height+6, 220)
	}
	g.drawStarsHUD()

	g.drawMinimap()

	if g.ZoneBanner > 0 && g.Zone >= 0 {
		a := uint8(255 * clampf(g.ZoneBanner, 0, 1))
		name := Sectors[g.Zone].Name
		nw := g.measure(name, 30, true)
		iconBubble(IconIsland, sw()/2-nw/2-34, 112, 50, LegoGoldBright, a)
		g.textCenter(name, sw()/2+10, 96, 30, rl.NewColor(255, 235, 90, a), true)
	}
	if g.LevelBanner > 0 {
		g.LevelBanner -= g.FrameTime
		a := uint8(255 * clampf(g.LevelBanner, 0, 1))
		cy := sh() * 0.27
		drawStar(sw()/2, cy, 58, fade(LegoGoldBright, a))
		g.textCenter(fmt.Sprintf("%d", g.Level), sw()/2, cy-26, 48, fade(rl.NewColor(120, 60, 10, 255), a), true)
		for k := 0; k < 5; k++ {
			drawIcon(IconMonster, sw()/2-96+float32(k)*48, cy+82, 32, fade(rl.NewColor(230, 70, 60, 255), a))
		}
		g.hint(g.levelBannerText(), sw()/2, cy+104, a)
	}

	g.drawCrosshair()
	g.drawWeaponBadge()
	g.drawRideHUD()
	g.drawDriveHUD()

	if g.MessageT > 0 {
		a := uint8(255 * clampf(g.MessageT, 0, 1))
		if g.MessageIcon != IconNone {
			// mesaj desenat: pictograma mare, textul mic dedesubt (pentru parinte)
			// sus, sub bara cu inamicii: nu acopera tinta si nici pe Aventurierul
			my := float32(178)
			k := 1 + 0.15*clampf((g.MessageT-2.7)/0.3, 0, 1)
			iconBubble(g.MessageIcon, sw()/2, my, 62*k, LegoGoldBright, a)
			g.hint(g.Message, sw()/2, my+44, a)
		} else {
			mw := g.measure(g.Message, 24, true) + 44
			mr := rl.NewRectangle((sw()-mw)/2, sh()-196, mw, 44)
			rl.DrawRectangleRounded(mr, 0.4, 8, rl.NewColor(20, 26, 40, uint8(float32(a)*0.85)))
			g.textCenter(g.Message, sw()/2, mr.Y+9, 24, rl.NewColor(255, 240, 150, a), true)
		}
	}

	// santierul de langa: sabia construieste. Totul se vede ca desen:
	// sabia care loveste -> bara din caramizi blocuri (cate piese mai sunt) -> moneda (cat costa).
	if g.NearSite >= 0 && g.Ride.Kind == RideNone && !g.Drive.Active {
		site := &g.City.LegoSites[g.NearSite]
		name := cleanName(site.Blueprint.Name)
		pw := float32(620)
		ph := float32(104)
		pr := rl.NewRectangle((sw()-pw)/2, sh()-156, pw, ph)
		panel(pr, LegoGold)
		sx, sy := pr.X+50, pr.Y+42
		sw0 := sinf(g.Time * 7)
		rl.DrawCircleV(vv(sx, sy), 34, icoDark)
		drawIcon(IconSword, sx+sw0*3, sy-sw0*3, 52, rl.White)
		cost := 100
		var hintTxt string
		if site.CurrentStage < site.TotalStages {
			drawBrickBar(pr.X+104, pr.Y+28, pw-104-96, 30, site.CurrentStage, site.TotalStages, LegoYellow, g.Time)
			hintTxt = fmt.Sprintf("Lovește cu sabia: piesa %d din %d (100 aur)  •  %s", site.CurrentStage+1, site.TotalStages, name)
		} else if site.UpgradeLevel < 5 {
			cost = 250 * site.UpgradeLevel
			drawIcon(IconGrow, pr.X+150, pr.Y+40, 64, LegoYellow)
			drawDots(pr.X+300, pr.Y+42, 14, g.UpgradeHits, 5, UIGreen)
			// marimea: 5 cuburi tot mai mari, cele atinse colorate
			for k := 1; k <= 5; k++ {
				s := 8 + float32(k)*3.4
				x := pr.X + 412 + float32(k-1)*25
				c := rl.NewColor(70, 76, 96, 255)
				if k <= site.UpgradeLevel {
					c = LegoYellow
				} else if k == site.UpgradeLevel+1 {
					c = fade(UIGreen, uint8(150+100*sinf(g.Time*6)))
				}
				rl.DrawRectangleRounded(rl.NewRectangle(x-s/2, pr.Y+56-s, s, s), 0.2, 4, c)
			}
			hintTxt = fmt.Sprintf("Lovește de 5 ori: %s crește la %dx (%d aur)  •  acum e %dx", name, site.UpgradeLevel+1, cost, site.UpgradeLevel)
		}
		cx := pr.X + pw - 52
		ic := IconCoin
		if g.Player.Gold < int64(cost) {
			ic = IconNoCoin
		}
		drawIcon(ic, cx, pr.Y+36, 44, rl.White)
		g.textCenter(fmt.Sprintf("%d", cost), cx, pr.Y+58, 17, LegoYellow, true)
		g.hint(hintTxt, sw()/2, pr.Y+ph-24, 255)
	} else if g.Prompt != "" && !g.Drive.Active {
		// ce poate face aici: tasta E (desenata) -> pictograma actiunii; textul mic dedesubt
		hw := g.measure(g.Prompt, 16, false) + 40
		pw := maxf(260, hw)
		ph := float32(104)
		pr := rl.NewRectangle((sw()-pw)/2, sh()-156, pw, ph)
		panel(pr, UIAccent)
		if g.PromptKind != PromptNone {
			g.drawKeycap("E", sw()/2-72, pr.Y+42, 50, true, 255)
			drawIcon(IconArrowRight, sw()/2-14, pr.Y+44, 26, UIText)
			iconBubble(g.PromptIcon, sw()/2+52, pr.Y+42, 66, UIAccent, 255)
		} else {
			iconBubble(IconCoaster, sw()/2-40, pr.Y+42, 62, UIAccent, 255)
			iconBubble(g.PromptIcon, sw()/2+40, pr.Y+42, 62, UIAccent, 255)
		}
		g.hint(g.Prompt, sw()/2, pr.Y+ph-24, 255)
	}

	if g.ShowHelp > 0 {
		a := uint8(230 * clampf(g.ShowHelp, 0, 1))
		lines := helpLines[:8]
		hr := rl.NewRectangle(18, sh()-40-float32(len(lines))*28, 560, 20+float32(len(lines))*28)
		rl.DrawRectangleRounded(hr, 0.1, 8, rl.NewColor(20, 26, 40, uint8(float32(a)*0.85)))
		for i, l := range lines {
			g.text(l[0], hr.X+16, hr.Y+12+float32(i)*28, 19, rl.NewColor(120, 210, 255, a), true)
			g.text(l[1], hr.X+200, hr.Y+12+float32(i)*28, 17, rl.NewColor(240, 240, 240, a), false)
		}
	} else if !g.Drive.Active {
		// H = ajutor, M = harta (taste desenate)
		g.drawKeycap("H", 40, sh()-40, 30, false, 200)
		g.textCenter("?", 72, sh()-54, 26, rl.NewColor(255, 255, 255, 200), true)
		g.drawKeycap("M", 116, sh()-40, 30, false, 200)
		drawIcon(IconMap, 152, sh()-40, 28, rl.White)
	}
	if g.ShowMap {
		g.drawBigMap()
	}
}

// drawRideHUD: inaltimea in roata / montagne russe, ritmul trambulinei (totul desenat)
func (g *Game) drawRideHUD() {
	p := &g.Player
	switch g.Ride.Kind {
	case RideWheel, RideCoaster:
		// coloana de inaltime cu Aventurierul care urca (stanga)
		top := float32(60)
		if g.Ride.Kind == RideCoaster {
			top = 32
		}
		g.drawHeightBar(p.Pos.Y, top, LegoCyan, -1)
		// tastele (jos): mouse = privesti, C = camera, E = cobori
		y := sh() - 84
		items := []struct {
			key  string
			icon IconKind
		}{{"", IconMouse}, {"C", IconCamera}, {"E", IconExit}}
		if g.Ride.First {
			items[1].icon = IconEye
		}
		if g.Ride.Kind == RideCoaster {
			items = items[:2]
		}
		w := float32(len(items))*120 + 20
		rl.DrawRectangleRounded(rl.NewRectangle((sw()-w)/2, y-36, w, 72), 0.3, 8, rl.NewColor(20, 26, 40, 190))
		for k, it := range items {
			x := (sw()-w)/2 + 70 + float32(k)*120
			if it.key == "" {
				drawIcon(IconMouse, x-24, y, 40, rl.White)
				drawIcon(IconEye, x+22, y, 36, rl.White)
			} else {
				g.drawKeycap(it.key, x-22, y, 40, false, 255)
				drawIcon(it.icon, x+24, y, 36, rl.White)
			}
		}
		hint := "Mouse = privești în jur  •  C = camera din ochii aventurierului  •  E = cobori"
		if g.Ride.First {
			hint = "Mouse = privești în jur  •  C = camera din spate  •  E = cobori"
		}
		if g.Ride.Kind == RideCoaster {
			hint = strings.Replace(hint, "  •  E = cobori", "", 1)
			hint += fmt.Sprintf("  •  %.0f km/h", g.City.Coaster.V*3.6)
		}
		g.hint(hint, sw()/2, y+42, 255)
	case RideTrampoline:
		pd := &g.City.W.Pads[g.Ride.Idx]
		h := p.Pos.Y - pd.MatY()
		g.drawHeightBar(maxf(h, 0), pd.MaxApex, pd.Color, g.PadApex)
		// tasta SPATIU desenata: verde si pulsand = apasa acum
		window := g.PadPress > 0 || p.VelY < 0
		cx, cy := sw()/2, sh()-222
		g.drawSpaceKey(cx, cy, 46, window, 255)
		msg := "Așteaptă..."
		if window {
			msg = "ACUM! Apasă SPAȚIU (sau ține-l apăsat)"
		} else {
			drawIcon(IconWait, cx+112, cy, 30, rl.NewColor(200, 205, 220, 255))
		}
		g.hint(msg, cx, cy+34, 255)
		if g.PadBoostMsg > 0 {
			a := uint8(255 * clampf(g.PadBoostMsg/0.4, 0, 1))
			for k := -1; k <= 1; k++ {
				drawIcon(IconArrowUp, cx+float32(k)*56, cy-86-absf(float32(k))*-10, 52, fade(UIGreen, a))
			}
			drawStar(cx-110, cy-96, 16, fade(LegoGoldBright, a))
			drawStar(cx+110, cy-96, 16, fade(LegoGoldBright, a))
		}
	}
}

// drawHeightBar: coloana de inaltime (stanga ecranului) cu un Aventurierul mic care urca.
// mark >= 0: linia aurie = pana unde va sari.
func (g *Game) drawHeightBar(h, max float32, col rl.Color, mark float32) {
	bx, by, bh := float32(30), float32(178), minf(300, sh()-340)
	rl.DrawRectangleRounded(rl.NewRectangle(bx-4, by-4, 34, bh+8), 0.5, 6, rl.NewColor(16, 20, 32, 210))
	k := clampf(h/max, 0, 1)
	rl.DrawRectangleRounded(rl.NewRectangle(bx+3, by+bh-bh*k, 20, bh*k), 0.5, 6, col)
	if mark >= 0 {
		ka := clampf(mark/max, 0, 1)
		rl.DrawRectangle(int32(bx-6), int32(by+bh-bh*ka), 38, 4, LegoGoldBright)
	}
	// Aventurierul (capul galben) in dreptul inaltimii
	hy := by + bh - bh*k
	rl.DrawCircleV(vv(bx+48, hy), 11, LegoYellow)
	rl.DrawCircleV(vv(bx+44, hy-2), 2, icoDark)
	rl.DrawCircleV(vv(bx+52, hy-2), 2, icoDark)
	drawIcon(IconArrowUp, bx+13, by-26, 26, rl.White)
	g.textCenter(fmt.Sprintf("%.0f m", h), bx+13, by+bh+10, 17, UIMuted, false)
}

// ---------------------------------------------------------------
// HARTA MICA (centrata pe Aventurierul) si HARTA MARE (tasta M)
// ---------------------------------------------------------------

func (g *Game) drawMapContents(x0, y0, w, h, scale float32, cx, cz float32, labels bool) {
	if g.City.Realm > 0 {
		g.drawRealmMap(x0, y0, w, h, scale, cx, cz)
		return
	}
	toMap := func(x, z float32) rl.Vector2 {
		return rl.NewVector2(x0+w/2+(x-cx)*scale, y0+h/2+(z-cz)*scale)
	}
	world := g.City.W
	rl.BeginScissorMode(int32(x0), int32(y0), int32(w), int32(h))
	rl.DrawRectangle(int32(x0), int32(y0), int32(w), int32(h), rl.NewColor(60, 150, 210, 255))
	land := rl.NewColor(120, 196, 90, 255)
	for _, l := range world.Lands {
		switch l.Kind {
		case 0:
			c := toMap(l.C.X, l.C.Y)
			hu, hv := (l.HalfU-2)*scale, (l.HalfV-2)*scale
			rl.DrawRectangleRounded(rl.NewRectangle(c.X-hu, c.Y-hv, 2*hu, 2*hv), 0.35, 8, land)
		case 1:
			rl.DrawCircleV(toMap(l.C.X, l.C.Y), (l.R-2)*scale, land)
		}
	}
	rl.DrawCircleV(toMap(0, 0), 40*scale, rl.NewColor(226, 214, 190, 255))
	for _, m := range world.Mountains {
		a := toMap(m.World(-MountAuIn+6, 0))
		b := toMap(m.World(MountAuOut-6, 0))
		rl.DrawLineEx(a, b, 26*scale, rl.NewColor(150, 130, 110, 255))
		rl.DrawLineEx(toMap(m.World(-PortalIn, 0)), toMap(m.World(PortalOut, 0)), maxf(3, 3*scale), m.Neon)
	}
	for i, b := range world.Bridges {
		col := rl.NewColor(240, 240, 240, 255)
		if i < 4 {
			col = b.Color
			if g.BridgesCrossed[i] {
				col = LegoGoldBright
			}
		}
		rl.DrawLineEx(toMap(b.A.X, b.A.Y), toMap(b.B.X, b.B.Y), maxf(5, 13*scale), rl.NewColor(40, 40, 50, 255))
		rl.DrawLineEx(toMap(b.A.X, b.A.Y), toMap(b.B.X, b.B.Y), maxf(3, 9*scale), col)
	}
	// Marele Tunel: linie intrerupta albastra + iesirile (verde)
	gt := &world.Corridors[0]
	for s := float32(0); s < gt.Len; s += 16 {
		x0s, z0s := gt.Point(s, 0)
		x1s, z1s := gt.Point(minf(s+9, gt.Len), 0)
		rl.DrawLineEx(toMap(x0s, z0s), toMap(x1s, z1s), maxf(3, 6*scale), rl.NewColor(30, 60, 160, 255))
	}
	for i := 1; i < len(world.Corridors); i++ {
		c := &world.Corridors[i]
		x, z := c.Point(c.Len, 0)
		rl.DrawCircleV(toMap(x, z), maxf(4, 4*scale), UIGreen)
	}
	for _, l := range world.Lifts {
		rl.DrawCircleV(toMap(l.Top.X, l.Top.Z), maxf(4, 4*scale), UIGreen)
	}
	for _, x := range []float32{-392, 392} {
		rl.DrawCircleV(toMap(x, 0), maxf(4, 5*scale), UIGreen)
	}
	// atractii
	for _, fw := range g.City.Wheels {
		c := toMap(fw.Pos.X, fw.Pos.Z)
		circleLines(c, maxf(5, fw.R*scale*0.5), LegoPink)
	}
	if c := g.City.Coaster; c != nil {
		for i := 0; i < len(c.Pts); i += 12 {
			a := c.Pts[i]
			b := c.Pts[(i+12)%len(c.Pts)]
			rl.DrawLineEx(toMap(a.X, a.Z), toMap(b.X, b.Z), 2, LegoRed)
		}
	}
	for _, pd := range world.Pads {
		rl.DrawCircleV(toMap(pd.Pos.X, pd.Pos.Z), maxf(2.5, pd.R*scale), pd.Color)
	}
	for _, v := range g.City.VendorPods {
		s := &g.City.LegoSites[v.SlotIndex]
		c := toMap(s.Position.X, s.Position.Z)
		col := LegoGoldBright
		if s.CurrentStage >= s.TotalStages {
			col = UIGreen
		}
		DrawStar(c.X, c.Y, 6, col)
	}
	// V 5.0: lumile inchise stau sub nori (si muntele / tunelul spre ele)
	for wi := 1; wi < NumWorlds; wi++ {
		f := g.Prog.MapFog[wi]
		if f < 0.01 {
			continue
		}
		si := Worlds[wi].Sector
		a := uint8(255 * clampf(f*1.2, 0, 1))
		c := Sectors[si].Center
		r := float32(78)
		if si >= SecWest {
			r = 92
		}
		cc := toMap(c.X, c.Y)
		drawCloudBlob(cc.X, cc.Y, r*scale, a, g.Time)
		if si < 4 {
			mx, mz := world.Mountains[si].C.X, world.Mountains[si].C.Y
			mc := toMap(mx, mz)
			drawCloudBlob(mc.X, mc.Y, 30*scale, a, g.Time+3)
		}
		if f > 0.6 && scale > 0.25 {
			drawIcon(IconLock, cc.X, cc.Y-4, maxf(18, 34*scale*2), fade(rl.White, a))
		}
	}
	for i := range g.Enemies {
		e := &g.Enemies[i]
		if !e.Alive {
			continue
		}
		c := toMap(e.Pos.X, e.Pos.Z)
		pulse := 3.5 + 1.2*sinf(g.Time*6)
		rl.DrawCircleV(c, pulse+1, rl.NewColor(60, 0, 0, 255))
		rl.DrawCircleV(c, pulse, UIRed)
	}
	if labels {
		for wi, wd := range Worlds {
			if !g.worldOpen(wi) {
				continue
			}
			s := Sectors[wd.Sector]
			c := toMap(s.Center.X, s.Center.Y)
			g.textCenter(wd.Name, c.X, c.Y-12, 18, rl.NewColor(30, 40, 30, 255), true)
		}
		c := toMap(0, 22)
		g.textCenter("Marele Tunel de sub Mare", c.X, c.Y, 16, rl.NewColor(20, 40, 120, 255), true)
	}
	// Aventurierul: sageata in directia camerei
	pp := toMap(g.Player.Pos.X, g.Player.Pos.Z)
	fx, fz := -sinf(g.CamYaw), -cosf(g.CamYaw)
	tip := rl.NewVector2(pp.X+fx*10, pp.Y+fz*10)
	l := rl.NewVector2(pp.X-fz*6-fx*6, pp.Y+fx*6-fz*6)
	rr := rl.NewVector2(pp.X+fz*6-fx*6, pp.Y-fx*6-fz*6)
	rl.DrawTriangle(tip, l, rr, LegoWhite)
	rl.DrawTriangle(tip, rr, l, LegoWhite)
	rl.DrawCircleV(pp, 3.5, LegoBlue)
	rl.EndScissorMode()
}

func (g *Game) drawMinimap() {
	size := float32(240)
	x0 := sw() - size - 18
	y0 := float32(18)
	frame := rl.NewRectangle(x0-6, y0-6, size+12, size+46)
	panel(frame, UIBorder)
	g.drawMapContents(x0, y0, size, size, size/380, g.Player.Pos.X, g.Player.Pos.Z, false)
	zone := "Pe apă"
	if curRealm > 0 {
		zone = "Pe drum"
	}
	if g.Zone >= 0 {
		zone = Sectors[g.Zone].Name
	}
	if g.InTunnel >= 0 {
		zone = g.City.W.TunnelName(g.InTunnel)
	} else if g.OnBridge >= 0 {
		zone = g.City.W.Bridges[g.OnBridge].Name
	}
	if g.measure(zone, 22, true) > size {
		g.textCenter(zone, x0+size/2, y0+size+8, 17, LegoGoldBright, true)
	} else {
		g.textCenter(zone, x0+size/2, y0+size+6, 22, LegoGoldBright, true)
	}
}

func (g *Game) drawBigMap() {
	rl.DrawRectangle(0, 0, ScreenWidth, ScreenHeight, rl.NewColor(10, 14, 26, 160))
	w := sw() - 80
	h := sh() - 120
	scale := minf(w/880, h/500)
	mw, mh := 880*scale, 500*scale
	x0, y0 := (sw()-mw)/2, (sh()-mh)/2+20
	panel(rl.NewRectangle(x0-8, y0-48, mw+16, mh+56), UIBorder)
	g.textCenter("Harta lumii  •  M = închide", sw()/2, y0-40, 28, LegoGoldBright, true)
	g.drawMapContents(x0, y0, mw, mh, scale, 0, 0, true)
}

// ---------------------------------------------------------------
// INDICATOARE PESTE LUMEA 3D (bare de viata, tinta, numere)
// ---------------------------------------------------------------

func (g *Game) drawWorldOverlays() {
	cam := g.Camera
	fwd := rl.Vector3Normalize(rl.Vector3Subtract(cam.Target, cam.Position))
	onScreen := func(p rl.Vector3) (rl.Vector2, bool) {
		if rl.Vector3DotProduct(rl.Vector3Subtract(p, cam.Position), fwd) < 0.5 {
			return rl.Vector2{}, false
		}
		s := rl.GetWorldToScreen(p, cam)
		return s, s.X > -50 && s.X < sw()+50 && s.Y > -50 && s.Y < sh()+50
	}
	nearest := -1
	nd := float32(1e9)
	for i := range g.Enemies {
		e := &g.Enemies[i]
		if !e.Alive {
			continue
		}
		d := rl.Vector3Distance(e.Pos, g.Player.Pos)
		if d < nd {
			nd, nearest = d, i
		}
		if d > 45 {
			// V 5.0: de departe, monstrul desenat in varful coloanei de lumina
			if s, ok := onScreen(v3(e.Pos.X, e.Pos.Y+3.6*e.Scale+10, e.Pos.Z)); ok {
				k, ring := IconMonster, UIRed
				if e.Boss {
					k, ring = IconCrown, LegoGoldBright
				}
				iconBubble(k, s.X, s.Y, 34, ring, 230)
			}
			continue
		}
		s, ok := onScreen(v3(e.Pos.X, e.Pos.Y+3.3*e.Scale, e.Pos.Z))
		if !ok {
			continue
		}
		bw := float32(90)
		rl.DrawRectangleRounded(rl.NewRectangle(s.X-bw/2-2, s.Y-2, bw+4, 14), 0.5, 6, rl.NewColor(20, 10, 10, 220))
		pct := clampf(float32(e.HP/e.MaxHP), 0, 1)
		if pct > 0 {
			rl.DrawRectangleRounded(rl.NewRectangle(s.X-bw/2, s.Y, bw*pct, 10), 0.5, 6, UIRed)
		}
		if d < 25 {
			g.textCenter(e.Name, s.X, s.Y-28, 20, UIText, true)
		}
		if e.Windup > 0 {
			g.textCenter("!", s.X, s.Y-60, 40, LegoYellow, true)
		}
	}
	// sageata spre cel mai apropiat inamic (cand nu e pe ecran)
	if nearest >= 0 && g.Ride.Kind == RideNone {
		e := &g.Enemies[nearest]
		target := v3(e.Pos.X, e.Pos.Y+2.2*e.Scale, e.Pos.Z)
		s, ok := onScreen(target)
		margin := float32(70)
		if !(ok && s.X > margin && s.X < sw()-margin && s.Y > margin && s.Y < sh()-margin) {
			to := rl.Vector3Subtract(target, cam.Position)
			right := rl.Vector3Normalize(rl.Vector3CrossProduct(fwd, v3(0, 1, 0)))
			ax := rl.Vector3DotProduct(to, right)
			az := rl.Vector3DotProduct(to, fwd)
			ang := atan2f(ax, az)
			cx, cy := sw()/2, sh()/2
			rad := minf(sw(), sh())*0.5 - 90
			px := clampf(cx+sinf(ang)*rad*1.3, 60, sw()-60)
			py := clampf(cy-cosf(ang)*rad, 80, sh()-80)
			dx, dy := sinf(ang), -cosf(ang)
			tip := rl.NewVector2(px+dx*22, py+dy*22)
			l := rl.NewVector2(px-dy*14-dx*8, py+dx*14-dy*8)
			r := rl.NewVector2(px+dy*14-dx*8, py-dx*14-dy*8)
			rl.DrawTriangle(tip, l, r, UIRed)
			rl.DrawTriangle(tip, r, l, UIRed)
			drawIcon(IconMonster, px, py+34, 26, rl.NewColor(230, 70, 60, 255))
			g.textCenter(fmt.Sprintf("%.0f m", nd), px, py+50, 14, rl.NewColor(255, 240, 170, 210), false)
		}
	}
	// santierele neterminate din apropiere: marcaj auriu cu distanta
	p := g.Player.Pos
	for i := range g.City.LegoSites {
		st := &g.City.LegoSites[i]
		if st.CurrentStage >= st.TotalStages {
			continue
		}
		d := rl.Vector3Distance(st.Position, p)
		if d > 70 || d < 9 {
			continue
		}
		s, ok := onScreen(v3(st.Position.X, st.Position.Y+maxf(st.Meshes.Height*st.Scale(), 2)+4, st.Position.Z))
		if !ok {
			continue
		}
		// caramida blocuri care sare + cat e gata (bara), distanta mica
		bob := sinf(g.Time*3+float32(i)) * 4
		rl.DrawCircleV(vv(s.X, s.Y+bob), 24, icoDark)
		drawBrick(s.X, s.Y+bob+2, 30, LegoGoldBright)
		bw := float32(54)
		k := float32(st.CurrentStage) / float32(maxInt(st.TotalStages, 1))
		rl.DrawRectangleRounded(rl.NewRectangle(s.X-bw/2-2, s.Y+28, bw+4, 11), 0.5, 4, icoDark)
		rl.DrawRectangleRounded(rl.NewRectangle(s.X-bw/2, s.Y+30, bw*maxf(k, 0.04), 7), 0.5, 4, LegoYellow)
		g.textCenter(fmt.Sprintf("%.0f m", d), s.X, s.Y+42, 14, rl.NewColor(255, 240, 170, 200), false)
	}
	for _, f := range g.Floats {
		s, ok := onScreen(f.Pos)
		if !ok {
			continue
		}
		a := uint8(255 * clampf(f.Life/0.6, 0, 1))
		col := f.Color
		col.A = a
		if strings.HasSuffix(f.Text, " aur") {
			// aur: moneda desenata + cifra
			num := strings.TrimSuffix(f.Text, " aur")
			tw := g.measure(num, f.Size, true)
			drawCoin(s.X-tw/2-f.Size*0.45, s.Y+f.Size*0.5, f.Size*0.42, fade(LegoGold, a))
			g.textCenter(num, s.X+f.Size*0.2, s.Y, f.Size, col, true)
			continue
		}
		g.textCenter(f.Text, s.X, s.Y, f.Size, col, true)
	}
}

// ---------------------------------------------------------------
// ECRAN DE INCARCARE
// ---------------------------------------------------------------

func (g *Game) drawLoading(msg string, pct float32) {
	if os.Getenv("BATTLE_DEBUG") != "" {
		fmt.Println("incarcare:", msg)
	}
	if g.useUI6() {
		g.drawLoading6(msg, pct)
		return
	}
	rl.BeginDrawing()
	rl.ClearBackground(rl.NewColor(30, 90, 170, 255))
	gradientV(0, 0, ScreenWidth, ScreenHeight, rl.NewColor(80, 160, 240, 255), rl.NewColor(20, 60, 130, 255))
	g.textCenter("BATTLE WORLDS 6.0", sw()/2, sh()/2-120, 80, LegoGoldBright, true)
	g.textCenter(msg, sw()/2, sh()/2, 30, UIText, true)
	bw := float32(500)
	rl.DrawRectangleRounded(rl.NewRectangle((sw()-bw)/2, sh()/2+60, bw, 24), 0.5, 8, rl.NewColor(10, 20, 40, 200))
	rl.DrawRectangleRounded(rl.NewRectangle((sw()-bw)/2, sh()/2+60, bw*clampf(pct, 0.02, 1), 24), 0.5, 8, LegoGoldBright)
	rl.EndDrawing()
}

// circleLines: cerc conturat (DrawCircleLinesV are un bug in raylib-go purego)
func circleLines(c rl.Vector2, r float32, col rl.Color) {
	rl.DrawRing(c, r-1.5, r+1.5, 0, 360, 40, col)
}

// gradientV: degrade vertical din benzi (DrawRectangleGradientV opreste unele drivere)
func gradientV(x, y, w, h int32, top, bottom rl.Color) {
	n := int32(48)
	for i := int32(0); i < n; i++ {
		y0 := y + h*i/n
		y1 := y + h*(i+1)/n
		rl.DrawRectangle(x, y0, w, y1-y0, mixColor(top, bottom, (float32(i)+0.5)/float32(n)))
	}
}

// drawDriveHUD: vitezometrul si tastele cand Aventurierul conduce
func (g *Game) drawDriveHUD() {
	if !g.Drive.Active {
		return
	}
	kmh := absf(g.Drive.Speed) * 3.6
	// vitezometru cu ac (se vede fara sa citesti cifrele)
	cx, cy := sw()-84, sh()-150
	rl.DrawCircleV(rl.NewVector2(cx, cy), 64, rl.NewColor(16, 20, 34, 220))
	rl.DrawRing(rl.NewVector2(cx, cy), 54, 62, 135, 315, 30, rl.NewColor(80, 200, 110, 255))
	rl.DrawRing(rl.NewVector2(cx, cy), 54, 62, 315, 375, 12, rl.NewColor(250, 200, 60, 255))
	rl.DrawRing(rl.NewVector2(cx, cy), 54, 62, 375, 405, 8, rl.NewColor(240, 80, 60, 255))
	k := clampf(kmh/(driveTurbo*3.6), 0, 1)
	a := (135 + 270*k) * math.Pi / 180
	outlinedLine(vv(cx, cy), vv(cx+cosf(a)*50, cy+sinf(a)*50), 5, rl.White)
	rl.DrawCircleV(vv(cx, cy), 8, LegoRed)
	g.textCenter(fmt.Sprintf("%.0f km/h", kmh), cx, cy+22, 17, UIMuted, false)
	// tastele masinii, desenate, fiecare cu ce face
	keys := []struct {
		k    string
		icon IconKind
		on   bool
	}{
		{"W", IconArrowUp, rl.IsKeyDown(rl.KeyW)}, {"S", IconArrowDown, rl.IsKeyDown(rl.KeyS)},
		{"A", IconArrowLeft, rl.IsKeyDown(rl.KeyA)}, {"D", IconArrowRight, rl.IsKeyDown(rl.KeyD)},
		{"SHIFT", IconTurbo, rl.IsKeyDown(rl.KeyLeftShift)}, {"Q", IconHorn, rl.IsKeyDown(rl.KeyQ)}, {"E", IconExit, false},
	}
	x := sw()/2 - 288
	y := sh() - 66
	rl.DrawRectangleRounded(rl.NewRectangle(x-40, y-54, 576+80, 100), 0.25, 8, rl.NewColor(20, 26, 40, 190))
	for i, kk := range keys {
		kx := x + float32(i)*96
		if kk.k == "SHIFT" {
			kx += 6
		}
		drawIcon(kk.icon, kx, y-30, 26, rl.White)
		g.drawKeycap(kk.k, kx, y+4, 34, kk.on, 255)
	}
	g.hint("W = accelerezi  •  S = frânezi / înapoi  •  A / D = virezi  •  SHIFT = turbo  •  Q = claxon  •  E = cobori", sw()/2, y+28, 255)
}
