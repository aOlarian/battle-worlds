package main

import (
	"fmt"
	"os"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			crashLog(r)
			os.Exit(2)
		}
	}()
	debug := ParseDebug()
	if debug == nil {
		showEnterprise()
	}
	bootLog("start")
	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagMsaa4xHint | rl.FlagVsyncHint)
	rl.SetTraceLogLevel(rl.LogWarning)
	if debug != nil {
		rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagMsaa4xHint)
		ScreenWidth, ScreenHeight = 1000, 640
	}
	rl.InitWindow(ScreenWidth, ScreenHeight, WindowTitle)
	if !rl.IsWindowReady() {
		fmt.Println("Nu am putut deschide fereastra jocului (placa video nu suporta OpenGL 3.3 sau nu exista ecran).")
		return
	}
	defer rl.CloseWindow()
	bootLog("fereastra ok, OpenGL %d", rl.GetVersion())
	rl.SetExitKey(0)
	// cate cadre pe secunda poate monitorul (60, 120, 144...): mouse-ul raspunde imediat
	fps := int32(rl.GetMonitorRefreshRate(rl.GetCurrentMonitor()))
	if fps < 50 || fps > 360 {
		fps = 60
	}
	rl.SetTargetFPS(fps)
	ScreenWidth = int32(rl.GetScreenWidth())
	ScreenHeight = int32(rl.GetScreenHeight())

	if debug != nil {
		fmt.Println("fereastra ok")
	}
	g := NewGame()
	g.LoadFonts()
	g.drawLoading("Pornim motorul grafic...", 0.03)
	g.Audio = InitAudio()
	bootLog("audio: placa de sunet=%v, sunete=%d, ambianta=%v", g.Audio.ready, len(g.Audio.all), g.Audio.hasMusic)
	defer g.Audio.CloseAudio()
	if debug != nil {
		fmt.Println("audio ok")
	}
	g.R = NewRenderer()
	if os.Getenv("BATTLE_GALLERY") != "" {
		buildWeaponMeshes()
		RunGallery(g)
		return
	}
	step := 0
	start := time.Now()
	g.City = InitCityWorld(func(msg string) {
		step++
		bootLog("%s", msg)
		g.drawLoading(msg, float32(step)/13)
	})
	fmt.Printf("Lumea construita in %v\n", time.Since(start).Round(time.Millisecond))
	buildFogMeshes(g.City.W)
	buildGoldBrickMesh()
	g.placeGoldBricks()
	buildNav(g.City)
	g.City.Nav = nav
	g.cityCache = map[int]*CityWorld{0: g.City}
	g.ensureRealmStates()
	buildTrainMesh()
	if debug != nil {
		fmt.Printf("navigare: %d noduri\n", len(nav.N))
	}
	g.SaveData = LoadPlayers()
	if g.SaveData.MouseSens > 0 {
		g.MouseSens = clampInt(g.SaveData.MouseSens-1, 0, len(mouseSensLevels)-1)
	}
	for _, p := range g.SaveData.Players {
		g.PlayerList = append(g.PlayerList, p.Name)
	}
	g.hookAchievements()
	g.ResetPlayerPosition()
	if uiShotHook != nil {
		uiShotHook(g)
	}
	bootLog("lumea gata, intram in joc")
	frames := 0
	if debug != nil {
		g.Debug = debug
		debug.Setup(g)
	}

	for !rl.WindowShouldClose() && !g.ShouldQuit {
		if rl.IsKeyPressed(rl.KeyF11) || ((rl.IsKeyDown(rl.KeyLeftAlt) || rl.IsKeyDown(rl.KeyRightAlt)) && rl.IsKeyPressed(rl.KeyEnter)) {
			rl.ToggleBorderlessWindowed()
		}
		ScreenWidth = int32(rl.GetScreenWidth())
		ScreenHeight = int32(rl.GetScreenHeight())
		g.Update()
		if g.Debug == nil || !g.Debug.FixedStep || g.Debug.Frame%90 == 0 {
			g.Draw()
			if frames++; frames == 1 || frames == 120 {
				bootLog("cadrul %d desenat", frames)
			}
		}
		if g.Debug != nil && g.Debug.AfterFrame(g) {
			break
		}
	}
	if g.Scene != ScenePlayerSelect && g.Scene != ScenePlayerCreate {
		g.autosave()
	}
}

func (g *Game) hookAchievements() {
	g.Achieve.OnUnlock = func(a *Achievement) {
		g.Play(g.Audio.Achievement)
	}
}

func (g *Game) Draw() {
	rl.BeginDrawing()
	rl.ClearBackground(rl.NewColor(190, 225, 250, 255))
	switch g.Scene {
	case ScenePlayerSelect:
		g.drawPlayerSelect6()
	case ScenePlayerCreate:
		g.drawPlayerCreate6()
	case SceneMainMenu:
		g.drawMainMenu6()
	case SceneDifficulty:
		g.drawDifficulty6()
	case SceneMapExplore:
		g.render3D(g.Camera, true)
		g.drawWorldOverlays()
		g.drawHUD()
		g.drawReveal()
		g.drawTrainOverlay()
	case SceneWinnerGlory:
		g.drawLevelComplete()
	case SceneGameOver:
		g.drawGameOver6()
	case ScenePauseMenu:
		g.drawPause6()
	case SceneHelp:
		g.drawHelp()
	case SceneBooth:
		g.render3D(g.Camera, true)
		g.drawBooth()
	}
	if g.Fade > 0 {
		rl.DrawRectangle(0, 0, ScreenWidth, ScreenHeight, rl.NewColor(8, 10, 20, uint8(255*clampf(g.Fade, 0, 1))))
	}
	g.Achieve.DrawPopup(g)
	if g.Debug != nil && g.Debug.ShowFPS {
		rl.DrawFPS(10, ScreenHeight-26)
	}
	rl.EndDrawing()
}

// render3D deseneaza lumea: umbre, cer, oras, personaje, apa, efecte.
func (g *Game) render3D(cam rl.Camera3D, actors bool) {
	r := g.R
	focus := g.Player.Pos
	if g.Scene == ScenePlayerSelect || g.Scene == SceneMainMenu || g.Scene == SceneDifficulty || g.Scene == ScenePlayerCreate || (g.Scene == SceneHelp && g.PrevScene != SceneMapExplore) {
		focus = cam.Target
		actors = false
	}
	indoor := float32(0)
	if actors && g.City.W.InTunnel(cam.Position.X, cam.Position.Z, cam.Position.Y) >= 0 {
		indoor = 1
	}
	r.BeginFrame(cam, g.Time)
	r.SetIndoor(indoor)
	depth := func(mesh rl.Mesh, m rl.Matrix, _ rl.Color) { rl.DrawMesh(mesh, r.MatDepth, m) }
	r.ShadowPass(focus, func(rl.Material) {
		g.City.DrawOpaque(r, depth, g.Time, true, focus)
		g.drawTrain(depth)
		if actors {
			g.drawActors(depth, false)
		}
	})
	rl.BeginMode3D(cam)
	r.DrawSky(cam)
	g.drawNightSky(cam)
	r.BindShadow()
	lit := func(mesh rl.Mesh, m rl.Matrix, tint rl.Color) { r.DrawLit(mesh, m, tint) }
	g.City.DrawOpaque(r, lit, g.Time, false, focus)
	g.drawTrain(lit)
	if actors {
		g.drawActors(lit, true)
	}
	if actors {
		g.drawGoldBricks()
	}
	if g.City.Water.VertexCount > 0 {
		rl.DrawMesh(g.City.Water, r.MatWater, rl.MatrixIdentity())
	}
	g.City.DrawTransparent(r, lit, g.Time, g.NearSite)
	if actors {
		g.City.DrawBeacons(r, g.Time, g.NearSite)
		g.drawGoldBrickGlow()
		g.drawGPS()
		g.drawBeam()
		g.drawZip()
		g.drawSurfboard()
		g.drawSnow()
		g.drawEffects()
	}
	g.drawFog()
	rl.EndMode3D()
}

func (g *Game) drawActors(draw DrawFn, lit bool) {
	p := &g.Player
	blink := g.Invuln > 0.7 && g.Scene == SceneMapExplore && int(g.Time*10)%2 == 0
	// cu camera din ochii aventurierului (in atractii) nu il desenam
	hideSelf := g.Ride.First && (g.Ride.Kind == RideCoaster || g.Ride.Kind == RideWheel)
	if g.Train.Carry && (g.Train.Phase == TrainDepart || g.Train.Phase == TrainFadeOut || g.Train.Phase == TrainArriveNew) {
		hideSelf = true
	}
	if g.Drive.Active {
		// Aventurierul sta la volan (se vede prin geamuri)
		c := &g.City.Cars[g.Drive.Car]
		sy, cy := sinf(c.Yaw), cosf(c.Yaw)
		lx, lz := float32(-0.42), float32(-0.35)
		seat := v3(c.Pos.X+cy*lx+sy*lz, c.Pos.Y+0.2, c.Pos.Z-sy*lx+cy*lz)
		DrawCharacter(draw, p.Meshes, seat, c.Yaw, 0.5, Pose{Sit: true}, rl.White)
	} else if !blink && !hideSelf {
		tint := rl.White
		// cand camera e lipita de Aventurierul, il facem putin transparent (sa se vada tinta)
		if lit && g.CamBoom < 2.2 && g.Ride.Kind == RideNone {
			tint = rl.NewColor(255, 255, 255, 110)
		}
		DrawCharacter(draw, p.Meshes, p.Pos, p.Yaw, 1.0, p.Pose, tint)
	}
	for i := range g.Enemies {
		e := &g.Enemies[i]
		if !e.Alive {
			continue
		}
		if lit && !g.R.SphereVisible(v3(e.Pos.X, e.Pos.Y+1.5, e.Pos.Z), 3) {
			continue
		}
		if lit {
			switch {
			case e.Flash > 0:
				g.R.SetFlash(0.55)
			case e.Windup > 0:
				// isi pregateste lovitura: palpaie, ca Aventurierul sa aiba timp sa se fereasca
				g.R.SetFlash(0.25 + 0.2*sinf(g.Time*30))
			}
		}
		pose := e.Pose
		DrawCharacter(draw, e.Meshes, e.Pos, e.Yaw, e.Scale, pose, rl.White)
		if lit && (e.Flash > 0 || e.Windup > 0) {
			g.R.SetFlash(0)
		}
	}
}

// efecte fara iluminare: proiectile, particule, fascicul spre tinta
func (g *Game) drawEffects() {
	for _, pr := range g.Projectiles {
		if pr.Enemy {
			rl.DrawSphereEx(pr.Position, 0.38, 8, 8, rl.NewColor(255, 230, 120, 255))
			rl.DrawSphereEx(pr.Position, 0.7, 8, 8, rl.NewColor(255, 110, 30, 120))
			continue
		}
		drawPlayerShot(&pr)
	}
	for _, pt := range g.Particles {
		a := clampf(pt.Life/pt.MaxLife, 0, 1)
		c := pt.Color
		c.A = uint8(float32(c.A) * a)
		s := pt.Size * (0.5 + 0.5*a)
		rl.DrawCube(pt.Position, s, s, s, c)
	}
	if g.Scene == SceneMapExplore {
		// semnul de deasupra celui mai apropiat inamic (se vede si peste cladiri)
		nearest := -1
		nd := float32(1e9)
		for i := range g.Enemies {
			if g.Enemies[i].Alive {
				if d := rl.Vector3Distance(g.Enemies[i].Pos, g.Player.Pos); d < nd {
					nd, nearest = d, i
				}
			}
		}
		// V 5.0: fiecare monstru are deasupra o coloana de lumina inalta, groasa
		// cat sa se vada si de la 200 m (latimea creste cu distanta)
		rl.DisableDepthMask()
		for i := range g.Enemies {
			e := &g.Enemies[i]
			if !e.Alive {
				continue
			}
			d := rl.Vector3Distance(e.Pos, g.Camera.Position)
			if d < 8 {
				continue
			}
			top := e.Pos.Y + 3.6*e.Scale + sinf(g.Time*4)*0.3
			w := maxf(0.35, d*0.007)
			pulse := 0.75 + 0.25*sinf(g.Time*3+float32(i))
			col := rl.NewColor(255, 60, 50, uint8(120*pulse))
			hgt := float32(70)
			if e.Boss {
				col = rl.NewColor(255, 190, 40, uint8(150*pulse))
				hgt = 110
				w *= 1.6
			}
			rl.DrawCylinder(v3(e.Pos.X, top+1.6, e.Pos.Z), w*0.5, w, hgt, 8, col)
			rl.DrawCylinder(v3(e.Pos.X, top+1.6, e.Pos.Z), w*1.6, w*2.6, hgt, 8, rl.NewColor(col.R, col.G, col.B, uint8(35*pulse)))
		}
		rl.EnableDepthMask()
		if nearest >= 0 && nd > 6 {
			e := &g.Enemies[nearest]
			top := e.Pos.Y + 3.6*e.Scale + sinf(g.Time*4)*0.3
			rl.DrawCylinder(v3(e.Pos.X, top, e.Pos.Z), 0.75, 0.0, 1.0, 10, rl.NewColor(255, 210, 40, 255))
			rl.DrawCylinder(v3(e.Pos.X, top+1.0, e.Pos.Z), 0.25, 0.25, 1.0, 8, UIRed)
		}
	}
}
