package main

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// V 6.0: SCHIMBAREA LUMII MARI
//   - curRealm = lumea in care e Aventurierul; Sectors / Worlds / NumWorlds si
//     marginea hartii se schimba odata cu ea;
//   - fiecare lume isi pastreaza progresul (nivel, insule vizitate, piese
//     aurii, santiere) in RealmState; costumele si armele sunt comune;
//   - lumile deja construite raman in memorie (calatorie rapida inapoi).
// ====================================================================

var curRealm int

var (
	islandSectors []SectorInfo
	islandWorlds  []WorldDef
	MapEdgeX      = mapEdgeIslandX
	MapEdgeZ      = mapEdgeIslandZ
)

// RealmState: progresul dintr-o lume mare.
type RealmState struct {
	Seen      bool
	Level     int
	Visited   [MaxZones]bool
	Bricks    [MaxZones][BricksPerWorld]bool
	BeachPads [2]bool
	RodeWheel bool
	RodeCoast bool
	Stages    []int
	Upgrades  []int
	Stars     int // stelele stranse aici (pentru armele deblocate oriunde)
}

func init() {
	islandSectors = append([]SectorInfo{}, Sectors...)
	islandWorlds = append([]WorldDef{}, Worlds...)
}

func (g *Game) ensureRealmStates() {
	for len(g.RealmStates) < len(Realms) {
		g.RealmStates = append(g.RealmStates, RealmState{Level: 1})
	}
}

// otherRealmStars: stelele din celelalte lumi (armele se deblocheaza cu toate stelele).
func (g *Game) otherRealmStars() int {
	n := 0
	for i := range g.RealmStates {
		if i != curRealm {
			n += g.RealmStates[i].Stars
		}
	}
	return n
}

// setRealmGlobals: sectoarele, lumile-insule si marginea hartii pentru lumea ri.
func setRealmGlobals(ri int, cw *CityWorld) {
	curRealm = ri
	if ri == 0 {
		Sectors = append([]SectorInfo{}, islandSectors...)
		Worlds = append([]WorldDef{}, islandWorlds...)
		NumWorlds = len(Worlds)
		MapEdgeX, MapEdgeZ = mapEdgeIslandX, mapEdgeIslandZ
		for _, w := range Worlds {
			Sectors[w.Sector].Name = w.Name
		}
		return
	}
	R := &Realms[ri]
	Sectors = Sectors[:0:0]
	Worlds = Worlds[:0:0]
	for i, z := range R.Zones {
		if i >= len(realmZoneLayout) {
			break
		}
		c := realmZoneLayout[i]
		in := rl.Vector2Normalize(rl.NewVector2(-c.X, -c.Y))
		Sectors = append(Sectors, SectorInfo{Name: z.Name, Center: c, Inner: in})
		oi := -1
		if i == 0 {
			oi = outfitOfRealm(ri)
		}
		Worlds = append(Worlds, WorldDef{Sector: i, Name: z.Name, Short: z.Short, Main: z.Main, Accent: z.Accent, Outfit: oi})
	}
	NumWorlds = len(Worlds)
	MapEdgeX, MapEdgeZ = RealmEdge, RealmEdge
}

// realmSectorAt: zona in care e punctul (lumile de uscat).
func realmSectorAt(x, z float32) int {
	best, bd := -1, float32(1e9)
	for i, s := range Sectors {
		d := sq(x-s.Center.X) + sq(z-s.Center.Y)
		if d < bd {
			best, bd = i, d
		}
	}
	if bd > sq(90) {
		return -1
	}
	return best
}

// realmZoneR: raza zonei (pentru monstri, piese aurii, nori)
func realmZoneR(si int) float32 {
	if si == 0 {
		return 58
	}
	return 62
}

// saveRealmState: progresul lumii curente intra in RealmStates.
func (g *Game) saveRealmState() {
	g.ensureRealmStates()
	rs := &g.RealmStates[curRealm]
	rs.Seen = true
	rs.Level = g.Level
	rs.Visited = g.Prog.Visited
	rs.Bricks = g.Prog.Bricks
	rs.BeachPads = g.Prog.BeachPads
	rs.RodeWheel = g.Prog.RodeWheel
	rs.RodeCoast = g.Prog.RodeCoast
	if g.City != nil {
		rs.Stages = rs.Stages[:0]
		rs.Upgrades = rs.Upgrades[:0]
		for i := range g.City.LegoSites {
			rs.Stages = append(rs.Stages, g.City.LegoSites[i].CurrentStage)
			rs.Upgrades = append(rs.Upgrades, g.City.LegoSites[i].UpgradeLevel)
		}
	}
	rs.Stars = g.realmStars()
}

// loadRealmState: progresul lumii curente din RealmStates (dupa ce lumea e construita).
func (g *Game) loadRealmState() {
	g.ensureRealmStates()
	rs := &g.RealmStates[curRealm]
	g.Level = maxInt(rs.Level, 1)
	g.Prog.Visited = rs.Visited
	g.Prog.Bricks = rs.Bricks
	g.Prog.BeachPads = rs.BeachPads
	g.Prog.RodeWheel = rs.RodeWheel
	g.Prog.RodeCoast = rs.RodeCoast
	if g.City != nil {
		for i := range g.City.LegoSites {
			s := &g.City.LegoSites[i]
			s.CurrentStage = 0
			s.UpgradeLevel = 1
			if i < len(rs.Stages) {
				s.CurrentStage = clampInt(rs.Stages[i], 0, s.TotalStages)
			}
			if i < len(rs.Upgrades) && rs.Upgrades[i] > 1 {
				s.UpgradeLevel = clampInt(rs.Upgrades[i], 1, 5)
			}
			s.TargetScale = float32(s.UpgradeLevel)
			s.CurrentScale = s.TargetScale
			s.HammerActive = false
		}
	}
	g.syncFog()
	g.Prog.LastStars = g.totalStars()
}

// cityFor: lumea construita (din memorie sau noua, cu ecran de incarcare).
func (g *Game) cityFor(ri int) *CityWorld {
	if g.cityCache == nil {
		g.cityCache = map[int]*CityWorld{}
	}
	if cw := g.cityCache[ri]; cw != nil {
		return cw
	}
	step := 0
	cw := InitRealmWorld(ri, func(msg string) {
		step++
		bootLog("%s", msg)
		g.drawLoading(msg, float32(step)/6)
	})
	g.cityCache[ri] = cw
	return cw
}

// applyRealmLook: cerul, ceata si lumina lumii.
func (g *Game) applyRealmLook(ri int) {
	r := g.R
	if g.islandLook == nil {
		g.islandLook = &[6][3]float32{r.ZenithCol, r.HorizonCol, r.SunCol, r.SkyCol, r.GroundCol, r.FogCol}
		g.islandFogD = r.FogDensity
	}
	if ri == 0 {
		l := g.islandLook
		r.ZenithCol, r.HorizonCol, r.SunCol, r.SkyCol, r.GroundCol, r.FogCol = l[0], l[1], l[2], l[3], l[4], l[5]
		r.FogDensity = g.islandFogD
		r.applyStaticUniforms()
		return
	}
	R := &Realms[ri]
	f3 := func(c rl.Color) [3]float32 {
		return [3]float32{float32(c.R) / 255, float32(c.G) / 255, float32(c.B) / 255}
	}
	r.ZenithCol = f3(R.SkyTop)
	r.HorizonCol = f3(R.SkyHorizon)
	r.SunCol = f3(R.SunColor)
	sky := f3(lerpColor(R.SkyTop, R.SkyHorizon, 0.5))
	for k := range sky {
		sky[k] *= R.Ambient * 1.25
	}
	r.SkyCol = sky
	r.GroundCol = f3(shade(R.Grass, 0.6))
	r.FogCol = f3(R.FogColor)
	r.FogDensity = g.islandFogD * R.FogDensity
	r.applyStaticUniforms()
}

// switchRealm: Aventurierul ajunge in lumea ri (progresul celei vechi se pastreaza).
func (g *Game) switchRealm(ri int) {
	if ri < 0 || ri >= len(Realms) {
		return
	}
	g.ensureRealmStates()
	g.arriveFirst = !g.RealmStates[ri].Seen
	if g.Drive.Active {
		g.exitCar()
	}
	g.Ride = RideState{}
	g.saveRealmState()
	var cw *CityWorld
	if ri == 0 {
		cw = g.cityCache[0]
	} else {
		setRealmGlobals(ri, nil)
		cw = g.cityFor(ri)
	}
	bootLog("switch: lume construita")
	g.City = cw
	setRealmGlobals(ri, cw)
	nav = cw.Nav
	g.applyRealmLook(ri)
	bootLog("switch: lumina")
	freeFogMeshes()
	buildFogMeshes(cw.W)
	bootLog("switch: nori")
	g.placeGoldBricks()
	bootLog("switch: piese aurii")
	g.loadRealmState()
	bootLog("switch: progres")
	g.Enemies = g.Enemies[:0]
	g.Projectiles = g.Projectiles[:0]
	g.Particles = g.Particles[:0]
	g.AimEnemy, g.AimCar, g.Attacker = -1, -1, -1
	g.AimKind, g.AimIdx = TgtNone, -1
	g.Prog.Reveal = RevealState{}
	g.stopGPS()
	g.Prog.Sector = -1
	g.spawnEnemies(g.Level)
	door := cw.StationDoor()
	g.Player.Pos = door
	g.Player.VelY = 0
	g.Swimming = false
	g.prevPos = g.Player.Pos
	g.Invuln = 3
	bootLog("lume: %s", Realms[ri].Name)
	g.autosave()
}

// arrivedInRealm: Aventurierul coboara din tren in lumea noua: primeste eroul lumii.
func (g *Game) arrivedInRealm() {
	ri := curRealm
	R := &Realms[ri]
	g.LevelBanner = 3
	g.ZoneBanner = 0
	if oi := outfitOfRealm(ri); oi >= 0 && ri > 0 {
		first := g.arriveFirst || !g.Prog.OwnOutfit[oi]
		g.arriveFirst = false
		g.Prog.OwnOutfit[oi] = true
		if first {
			g.wearOutfit(oi)
			g.sayIcon(IconOutfit, fmt.Sprintf("Bun venit în %s! Acum ești %s.", R.Name, Outfits[oi].Name))
			g.Play(g.Audio.Achievement)
			return
		}
	}
	g.sayIcon(IconWorld, "Bun venit în "+R.Name+"!")
}

// realmDone: toate insulele lumii au fost trecute (si mai e o lume dupa).
func (g *Game) realmDone() bool { return g.Level > NumWorlds }

// onRealmComplete: Seful ultimei insule a fost invins: trenul vine dupa Aventurierul.
func (g *Game) onRealmComplete() {
	next := curRealm + 1
	if next >= len(Realms) {
		g.sayIcon(IconCrown, "Ai terminat toate lumile! Ești campionul!")
		return
	}
	if g.UnlockedRealm < next {
		g.UnlockedRealm = next
	}
	g.callTrain(next)
	st := g.City.StationDoor()
	g.startGPS(st, -1)
	g.sayIcon(IconTrain, "Lume nouă! Trenul magic te așteaptă la gară. Urmează săgețile!")
}

// ---------------- harta unei lumi de uscat ----------------

func (g *Game) drawRealmMap(x0, y0, w, h, scale float32, cx, cz float32) {
	scale *= 0.84
	toMap := func(x, z float32) rl.Vector2 {
		return rl.NewVector2(x0+w/2+(x-cx)*scale, y0+h/2+(z-cz)*scale)
	}
	R := &Realms[curRealm]
	cw := g.City
	rl.BeginScissorMode(int32(x0), int32(y0), int32(w), int32(h))
	rl.DrawRectangle(int32(x0), int32(y0), int32(w), int32(h), shade(R.Dirt, 0.8))
	a := toMap(-RealmEdge, -RealmEdge)
	b := toMap(RealmEdge, RealmEdge)
	rl.DrawRectangleRounded(rl.NewRectangle(a.X, a.Y, b.X-a.X, b.Y-a.Y), 0.06, 8, R.Grass)
	for _, r := range cw.Roads {
		rl.DrawLineEx(toMap(r.A.X, r.A.Y), toMap(r.B.X, r.B.Y), maxf(4, 2*r.HalfW*scale), R.Path)
	}
	for i, zn := range cw.Zones {
		c := toMap(zn.C.X, zn.C.Y)
		zd := R.Zones[i]
		rl.DrawCircleV(c, zn.R*scale, fade(lerpColor(zd.Main, LegoWhite, 0.35), 200))
		rl.DrawCircleV(c, RealmPlazaR*scale, R.Path)
	}
	for _, v := range cw.VendorPods {
		s := &cw.LegoSites[v.SlotIndex]
		c := toMap(s.Position.X, s.Position.Z)
		col := LegoGoldBright
		if s.CurrentStage >= s.TotalStages {
			col = UIGreen
		}
		DrawStar(c.X, c.Y, 6, col)
	}
	st := toMap(cw.Station.X, cw.Station.Z)
	drawIcon(IconTrain, st.X, st.Y, maxf(18, 26*scale*3), rl.White)
	for wi := 1; wi < NumWorlds; wi++ {
		f := g.Prog.MapFog[wi]
		if f < 0.01 {
			continue
		}
		al := uint8(255 * clampf(f*1.2, 0, 1))
		c := Sectors[wi].Center
		cc := toMap(c.X, c.Y)
		drawCloudBlob(cc.X, cc.Y, 80*scale, al, g.Time)
		if f > 0.6 && scale > 0.25 {
			drawIcon(IconLock, cc.X, cc.Y-4, maxf(18, 34*scale*2), fade(rl.White, al))
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
	pp := toMap(g.Player.Pos.X, g.Player.Pos.Z)
	fx, fz := -sinf(g.CamYaw), -cosf(g.CamYaw)
	tip := rl.NewVector2(pp.X+fx*10, pp.Y+fz*10)
	l := rl.NewVector2(pp.X-fz*6-fx*6, pp.Y+fx*6-fz*6)
	rr := rl.NewVector2(pp.X+fz*6-fx*6, pp.Y-fx*6-fz*6)
	rl.DrawTriangle(tip, l, rr, LegoWhite)
	rl.DrawTriangle(tip, rr, l, LegoWhite)
	rl.DrawCircleV(pp, 3.5, LegoBlue)
	rl.EndScissorMode()
	_ = math.Pi
}

// ---------------- gara din lumea 0 (insulele) ----------------

// islandStationSpot: un loc plat si liber de ~34 x 14 m, cat mai aproape de centru.
func islandStationSpot(w *World) rl.Vector3 {
	ok := func(x, z float32) bool {
		along, toHub := stationAxes(v3(x, 0, z))
		for u := float32(-17); u <= 17; u += 2.5 {
			for v := float32(-7); v <= 3; v += 2 {
				px, pz := x+along.X*u+toHub.X*v, z+along.Y*u+toHub.Y*v
				if w.LandSD(px, pz) > -7 || w.onRoad(px, pz) || !w.CanStand(px, pz, 0.3, 1.2) {
					return false
				}
				gy, k := w.GroundAt(px, pz, 2)
				if k == SurfWater || absf(gy) > 0.7 {
					return false
				}
			}
		}
		return true
	}
	cands := []rl.Vector2{}
	for _, s := range islandSectors {
		for ri := 0; ri < 12; ri++ {
			r := 10 + float32(ri)*5
			for ai := 0; ai < 36; ai++ {
				a := float32(ai) / 36 * 2 * math.Pi
				cands = append(cands, rl.NewVector2(s.Center.X+cosf(a)*r, s.Center.Y+sinf(a)*r))
			}
		}
	}
	best := rl.NewVector2(WestX, 30)
	bd := float32(1e12)
	for _, c := range cands {
		d := c.X*c.X + c.Y*c.Y
		if d >= bd || !ok(c.X, c.Y) {
			continue
		}
		bd, best = d, c
	}
	gy, _ := w.GroundAt(best.X, best.Y, 2)
	return v3(best.X, gy, best.Y)
}

// backToIslands: inapoi in lumea 0 fara sa salvam (la schimbarea jucatorului).
func (g *Game) backToIslands() {
	if curRealm == 0 || g.cityCache == nil || g.cityCache[0] == nil {
		return
	}
	cw := g.cityCache[0]
	g.City = cw
	setRealmGlobals(0, cw)
	nav = cw.Nav
	g.applyRealmLook(0)
	freeFogMeshes()
	buildFogMeshes(cw.W)
	g.placeGoldBricks()
	g.Train = TrainState{}
}

// drawNightSky: luna mare si stelele, in lumile de noapte (dupa cer, inainte de lume).
func (g *Game) drawNightSky(cam rl.Camera3D) {
	if !Realms[curRealm].Night {
		return
	}
	r := g.R
	sd := rl.Vector3Normalize(v3(r.SunDir.X, r.SunDir.Y, r.SunDir.Z))
	if sd.Y < 0.35 {
		sd = rl.Vector3Normalize(v3(sd.X, 0.35, sd.Z))
	}
	rl.DisableDepthMask()
	c := rl.Vector3Add(cam.Position, rl.Vector3Scale(sd, 700))
	rl.DrawSphereEx(c, 70, 18, 18, rl.NewColor(255, 248, 215, 60))
	rl.DrawSphereEx(c, 46, 18, 18, rl.NewColor(255, 250, 225, 255))
	for k, o := range [][3]float32{{12, 8, 9}, {-14, -6, 7}, {4, -16, 6}, {-6, 14, 5}} {
		_ = k
		right := rl.Vector3Normalize(rl.Vector3CrossProduct(sd, v3(0, 1, 0)))
		up := rl.Vector3CrossProduct(right, sd)
		p := rl.Vector3Add(rl.Vector3Subtract(c, rl.Vector3Scale(sd, 40)), rl.Vector3Add(rl.Vector3Scale(right, o[0]), rl.Vector3Scale(up, o[1])))
		rl.DrawSphereEx(p, o[2], 10, 10, rl.NewColor(225, 220, 195, 255))
	}
	// stelele: puncte fixe pe bolta (aceleasi la fiecare cadru)
	for i := 0; i < 260; i++ {
		fi := float32(i)
		a := fi * 2.399963
		y := 0.12 + float32(math.Mod(float64(fi*0.6180339), 1))*0.88
		rr := sqrtf(1 - y*y)
		d := v3(cosf(a)*rr, y, sinf(a)*rr)
		p := rl.Vector3Add(cam.Position, rl.Vector3Scale(d, 800))
		s := float32(1.2 + float32(i%3)*0.6)
		tw := uint8(170 + 85*sinf(g.Time*2+fi))
		rl.DrawCube(p, s, s, s, rl.NewColor(255, 255, 240, tw))
	}
	rl.EnableDepthMask()
}
