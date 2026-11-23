package main

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// V 5.0: PROGRESUL PE LUMI
//   - nivelul N deschide primele N lumi (Worlds); celelalte stau sub nori
//     si nu se poate intra in ele (nici pe jos, nici inot, nici cu masina);
//   - la trecerea nivelului: harta mare 5 s, insula noua apare din nor,
//     apoi sageti GPS pe jos pana la cabina telefonica a lumii noi;
//   - la prima intrare intr-o lume: costumul ei;
//   - 3 stele pe lume: monstrii (cu Seful), constructia, 5 piese aurii.
// ====================================================================

const BricksPerWorld = 5

// pragurile de stele pentru arme (index = arma)
var GunStars = [NumGuns]int{0, 2, 5, 8, 12, 16}

type RevealState struct {
	Active bool
	T      float32
	World  int
}

const (
	revealConfetti = float32(1.4) // confetti inainte de harta
	revealMapIn    = float32(0.7) // harta apare
	revealHold     = float32(5.0) // harta ramane (insula noua iese din nor)
	revealMapOut   = float32(0.9) // harta dispare
)

type ProgState struct {
	Fog       [MaxZones]float32 // 1 = sub nori in lume, 0 = deschisa
	MapFog    [MaxZones]float32 // acelasi lucru pe harta
	Visited   [MaxZones]bool
	OwnOutfit []bool
	Outfit    int
	Gun       int
	Bricks    [MaxZones][BricksPerWorld]bool
	BeachPads [2]bool // steaua 2 pe plaja: ambele trambuline
	RodeWheel bool    // steaua 2 in parc: roata uriasa ...
	RodeCoast bool    // ... si montagne russe
	Reveal    RevealState
	GPS       GPSState
	BossOut   bool
	LockMsgT  float32
	LastStars int
	Sector    int
	Sparkle   float32 // efect la schimbarea costumului
}

func newProgState() ProgState {
	p := ProgState{OwnOutfit: make([]bool, len(Outfits)), Sector: -1}
	p.OwnOutfit[0] = true
	for i := range p.Fog {
		p.Fog[i] = 1
		p.MapFog[i] = 1
	}
	return p
}

// cheat: modul de test al parintelui: totul deblocat (lumi, insule, arme, costume).
func (g *Game) cheat() bool { return g.Difficulty == "cheat" }

func (g *Game) worldsOpen() int {
	if g.cheat() {
		return NumWorlds
	}
	return clampInt(g.Level, 1, NumWorlds)
}

func (g *Game) worldOpen(wi int) bool { return wi >= 0 && wi < g.worldsOpen() }

func (g *Game) sectorOpen(si int) bool {
	wi := WorldOfSector(si)
	return wi < 0 || g.worldOpen(wi)
}

// syncFog: norii se potrivesc cu nivelul (la incarcare, fara animatie).
func (g *Game) syncFog() {
	for i := range g.Prog.Fog {
		f := float32(1)
		if g.worldOpen(i) {
			f = 0
		}
		g.Prog.Fog[i] = f
		g.Prog.MapFog[i] = f
	}
}

// lockedAt: punctul e intr-o lume inca inchisa (sau pe drumul spre ea)?
func (g *Game) lockedAt(x, z, y float32) bool {
	if g.Debug != nil && g.Debug.NoLocks {
		return false
	}
	if s := SectorAt(x, z); s >= 0 && !g.sectorOpen(s) {
		return true
	}
	w := g.City.W
	for i := range w.Mountains {
		if g.sectorOpen(i) {
			continue
		}
		m := &w.Mountains[i]
		u := x*m.Dir.X + z*m.Dir.Y
		v := x*m.Perp.X + z*m.Perp.Y
		if u > CenterR+4 && u < 205 && absf(v) < 36 {
			return true
		}
	}
	if y < -2.5 { // Marele Tunel: jumatatea spre o insula inchisa
		if x < -(CenterR+26) && !g.sectorOpen(SecWest) {
			return true
		}
		if x > CenterR+26 && !g.sectorOpen(SecPark) {
			return true
		}
	}
	for _, s := range []struct {
		x  float32
		si int
	}{{-VentX, SecWest}, {VentX, SecPark}} {
		if !g.sectorOpen(s.si) && sq(x-s.x)+z*z < sq(VentR+8) {
			return true
		}
	}
	return false
}

func (g *Game) lockBump() {
	if g.Prog.LockMsgT <= 0 {
		g.sayIcon(IconLock, "Insula se deschide la nivelul următor: învinge monștrii!")
		g.Prog.LockMsgT = 4
	}
}

// ---------------------------------------------------------------
// NORII de peste lumile inchise (un mesh pe lume, construit o data)
// ---------------------------------------------------------------

var fogMeshes [MaxZones]rl.Mesh
var fogCenters [MaxZones]rl.Vector3
var fogReady bool

func buildFogMeshes(w *World) {
	if c, ok := fogCache[curRealm]; ok {
		fogMeshes, fogCenters = c.m, c.c
		fogRealm = curRealm
		fogReady = true
		return
	}
	fogRealm = curRealm
	fogMeshes = [MaxZones]rl.Mesh{}
	rng := rand.New(rand.NewSource(55))
	puff := func(mb *MeshBuilder, c rl.Vector3, r float32) {
		for k := 0; k < 3; k++ {
			o := v3((rng.Float32()-0.5)*r*0.9, (rng.Float32()-0.3)*r*0.35, (rng.Float32()-0.5)*r*0.9)
			rr := r * (0.65 + rng.Float32()*0.35)
			mb.EllipsoidF(rl.Vector3Add(c, o), v3(rr, rr*0.72, rr), 6, 10, func(n rl.Vector3) rl.Color {
				k := 0.86 + 0.14*clampf(n.Y*0.5+0.5, 0, 1)
				return rl.NewColor(uint8(250*k), uint8(252*k), 255, 255)
			})
		}
	}
	for wi, wd := range Worlds {
		if wi == 0 {
			continue
		}
		mb := NewMB()
		si := wd.Sector
		c := Sectors[si].Center
		rx, rz := float32(84), float32(84)
		switch si {
		case SecWest:
			rx, rz = 76, 88
		case SecPark:
			rx, rz = 80, 100
		}
		if curRealm > 0 {
			rx, rz = realmZoneR(si)+10, realmZoneR(si)+10
		}
		fogCenters[wi] = v3(c.X, 0, c.Y)
		// zid de nori pe marginea insulei (doua randuri) + capac
		n := 34
		for i := 0; i < n; i++ {
			a := float32(i) / float32(n) * 2 * math.Pi
			puff(mb, v3(c.X+cosf(a)*rx, 6+rng.Float32()*4, c.Y+sinf(a)*rz), 13)
			if i%2 == 0 {
				puff(mb, v3(c.X+cosf(a+0.09)*(rx-8), 22+rng.Float32()*5, c.Y+sinf(a+0.09)*(rz-8)), 14)
			}
		}
		for i := 0; i < 16; i++ {
			a := rng.Float32() * 2 * math.Pi
			d := sqrtf(rng.Float32()) * 0.7
			puff(mb, v3(c.X+cosf(a)*rx*d, 34+rng.Float32()*8, c.Y+sinf(a)*rz*d), 18)
		}
		// perdea de nori la intrarea pe podul scurt spre muntele lumii
		if si < 4 && curRealm == 0 {
			m := &w.Mountains[si]
			for k := -2; k <= 2; k++ {
				x, z := m.World(CenterR+7-m.uOffset(), float32(k)*6)
				puff(mb, v3(x, 3+absf(float32(k)), z), 6)
				x, z = m.World(CenterR+12-m.uOffset(), float32(k)*6)
				puff(mb, v3(x, 9, z), 6)
			}
		}
		// Marele Tunel si insulita cu lift
		if (si == SecWest || si == SecPark) && curRealm == 0 {
			sx := float32(-1)
			if si == SecPark {
				sx = 1
			}
			gt := &w.Corridors[0]
			s, _ := gt.Local(sx*(CenterR+27), 0)
			fy := gt.FloorC(s)
			for k := -1; k <= 1; k++ {
				puff(mb, v3(sx*(CenterR+28), fy+2.2, float32(k)*3.5), 3.4)
				puff(mb, v3(sx*(CenterR+31), fy+3.6, float32(k)*3.5), 3.2)
			}
			for i := 0; i < 6; i++ {
				a := float32(i) / 6 * 2 * math.Pi
				puff(mb, v3(sx*VentX+cosf(a)*9, 5, sinf(a)*9), 9)
			}
			puff(mb, v3(sx*VentX, 14, 0), 12)
		}
		fogMeshes[wi] = mb.Build()
	}
	fogReady = true
}

// uOffset: Mountain.World(u, v) are u masurat de la centrul muntelui (IsletD).
func (m *Mountain) uOffset() float32 { return IsletD }

func (g *Game) drawFog() {
	if !fogReady {
		return
	}
	for wi := 1; wi < NumWorlds; wi++ {
		f := g.Prog.Fog[wi]
		if f < 0.01 {
			continue
		}
		c := fogCenters[wi]
		s := 1 + (1-f)*0.55 // se umfla si se risipeste
		m := rl.MatrixMultiply(rl.MatrixTranslate(-c.X, 0, -c.Z), rl.MatrixScale(s, 1+(1-f)*0.8, s))
		m = rl.MatrixMultiply(m, rl.MatrixTranslate(c.X, (1-f)*18, c.Z))
		a := uint8(255 * clampf(f*1.3, 0, 1))
		g.R.DrawLit(fogMeshes[wi], m, rl.NewColor(255, 255, 255, a))
	}
}

// ---------------------------------------------------------------
// ACTUALIZARE: lumi vizitate, nori, dezvaluire
// ---------------------------------------------------------------

func (g *Game) updateProgress(dt float32) {
	pr := &g.Prog
	if pr.LockMsgT > 0 {
		pr.LockMsgT -= dt
	}
	if pr.Sparkle > 0 {
		pr.Sparkle -= dt
	}
	// norii lumilor deschise se risipesc (in afara de cea care inca se dezvaluie pe harta)
	for i := range pr.Fog {
		if !g.worldOpen(i) || (pr.Reveal.Active && pr.Reveal.World == i) {
			continue
		}
		if pr.Fog[i] > 0 {
			pr.Fog[i] = maxf(0, pr.Fog[i]-dt/3.2)
		}
		if pr.MapFog[i] > 0 && !pr.Reveal.Active {
			pr.MapFog[i] = 0
		}
	}
	p := &g.Player
	si := SectorAt(p.Pos.X, p.Pos.Z)
	if si != pr.Sector {
		pr.Sector = si
		if wi := WorldOfSector(si); wi >= 0 && g.worldOpen(wi) {
			g.enterWorld(wi)
		}
	}
	g.updateGPS(dt)
	g.updateBricks(dt)
	g.checkStars()
}

func (g *Game) enterWorld(wi int) {
	pr := &g.Prog
	if pr.GPS.Active && pr.GPS.World == wi && !pr.GPS.FromInside {
		g.stopGPS()
	}
	if pr.Visited[wi] {
		return
	}
	pr.Visited[wi] = true
	oi := Worlds[wi].Outfit
	if oi >= 0 && oi < len(pr.OwnOutfit) && !pr.OwnOutfit[oi] {
		pr.OwnOutfit[oi] = true
		if wi > 0 {
			g.wearOutfit(oi)
			g.sayIcon(IconOutfit, "Costum nou: "+Outfits[oi].Name+"! Altele găsești în cabina telefonică.")
			g.Play(g.Audio.Achievement)
		}
	}
	g.ZoneBanner = 4
}

func (g *Game) wearOutfit(oi int) {
	if oi < 0 || oi >= len(Outfits) {
		return
	}
	g.Prog.Outfit = oi
	g.Player.Meshes = GetCharMeshes(Outfits[oi].Style)
	g.Prog.Sparkle = 1.2
	p := g.Player.Pos
	for n := 0; n < 50; n++ {
		c := []rl.Color{LegoGoldBright, LegoWhite, LegoCyan, LegoPink}[n%4]
		a := float32(n) * 0.7
		g.spawnParticle(v3(p.X+cosf(a)*0.9, p.Y+0.2+float32(n%10)*0.25, p.Z+sinf(a)*0.9), v3(cosf(a)*1.5, 2+rand.Float32()*3, sinf(a)*1.5), c, 0.18, 1.2)
	}
}

// startReveal: s-a trecut nivelul si se deschide lumea wi.
func (g *Game) startReveal(wi int) {
	pr := &g.Prog
	pr.Reveal = RevealState{Active: true, World: wi}
	if wi >= 0 {
		pr.Fog[wi] = 1
		pr.MapFog[wi] = 1
	}
	g.stopGPS()
	p := g.Player.Pos
	for n := 0; n < 120; n++ {
		c := []rl.Color{LegoRed, LegoYellow, LegoBlue, LegoGreen, LegoPink, LegoWhite}[n%6]
		g.spawnParticle(v3(p.X, p.Y+2.5, p.Z), v3((rand.Float32()-0.5)*14, rand.Float32()*12+4, (rand.Float32()-0.5)*14), c, 0.28, 2.2)
	}
}

func revealTotal() float32 { return revealConfetti + revealMapIn + revealHold + revealMapOut }

// revealMapAlpha: cat de vizibila e harta de dezvaluire acum (0..1).
func (g *Game) revealMapAlpha() float32 {
	r := &g.Prog.Reveal
	if !r.Active || r.World < 0 {
		return 0
	}
	t := r.T - revealConfetti
	switch {
	case t < 0:
		return 0
	case t < revealMapIn:
		return t / revealMapIn
	case t < revealMapIn+revealHold:
		return 1
	default:
		return clampf(1-(t-revealMapIn-revealHold)/revealMapOut, 0, 1)
	}
}

func (g *Game) updateReveal(dt float32) {
	pr := &g.Prog
	r := &pr.Reveal
	r.T += dt
	if r.World < 0 {
		if r.T >= revealConfetti {
			r.Active = false
			g.spawnEnemies(g.Level)
			g.LevelBanner = 3
		}
		return
	}
	t := r.T - revealConfetti - revealMapIn
	if t > 0.6 {
		// insula noua iese din nor pe harta in ~2.5 s
		pr.MapFog[r.World] = clampf(1-(t-0.6)/2.5, 0, 1)
	}
	if r.T >= revealTotal() {
		r.Active = false
		pr.MapFog[r.World] = 0
		g.spawnEnemies(g.Level)
		g.LevelBanner = 3
		g.startGPSToWorld(r.World)
	}
}

// ---------------------------------------------------------------
// STELELE
// ---------------------------------------------------------------

func (g *Game) starMonsters(wi int) bool { return g.Level > wi+1 }

func (g *Game) starBuild(wi int) bool {
	si := Worlds[wi].Sector
	if curRealm > 0 {
		si += 100 // fara cazurile speciale ale insulelor; santierele se cauta mai jos
	}
	switch si {
	case SecWest:
		return g.Prog.BeachPads[0] && g.Prog.BeachPads[1]
	case SecPark:
		return g.Prog.RodeWheel && g.Prog.RodeCoast
	}
	any := false
	for i := range g.City.LegoSites {
		s := &g.City.LegoSites[i]
		if SectorAt(s.Position.X, s.Position.Z) != si%100 {
			continue
		}
		any = true
		if s.CurrentStage < s.TotalStages {
			return false
		}
	}
	return any
}

func (g *Game) bricksFound(wi int) int {
	n := 0
	for _, b := range g.Prog.Bricks[wi] {
		if b {
			n++
		}
	}
	return n
}

func (g *Game) starBricks(wi int) bool { return g.bricksFound(wi) >= BricksPerWorld }

func (g *Game) worldStars(wi int) int {
	n := 0
	for _, b := range []bool{g.starMonsters(wi), g.starBuild(wi), g.starBricks(wi)} {
		if b {
			n++
		}
	}
	return n
}

func (g *Game) totalStars() int {
	n := g.otherRealmStars()
	for wi := range Worlds {
		n += g.worldStars(wi)
	}
	return n
}

// realmStars: stelele lumii mari curente (fara celelalte lumi).
func (g *Game) realmStars() int {
	n := 0
	for wi := range Worlds {
		n += g.worldStars(wi)
	}
	return n
}

func (g *Game) gunUnlocked(k int) bool {
	return k >= 0 && k < NumGuns && (g.cheat() || g.totalStars() >= GunStars[k])
}

// ownOutfit: costumul e castigat (in cheat: toate).
func (g *Game) ownOutfit(i int) bool {
	return i >= 0 && i < len(g.Prog.OwnOutfit) && (g.cheat() || g.Prog.OwnOutfit[i])
}

// realmOpen: lumea mare se poate vizita (in cheat: toate).
func (g *Game) realmOpen(ri int) bool {
	return ri >= 0 && ri < len(Realms) && (g.cheat() || ri <= g.UnlockedRealm)
}

// checkStars: o stea noua poate aduce o arma noua.
func (g *Game) checkStars() {
	st := g.totalStars()
	pr := &g.Prog
	if st <= pr.LastStars {
		pr.LastStars = st
		return
	}
	old := pr.LastStars
	pr.LastStars = st
	if g.Debug == nil && g.Time < 2 {
		return // la incarcarea salvarii
	}
	for k := NumGuns - 1; k > 0; k-- {
		if GunStars[k] > old && GunStars[k] <= st {
			pr.Gun = k
			g.Player.Pose.Gun = k
			g.sayIcon(IconGun, "Armă nouă: "+GunNames[k]+"! O schimbi din cabina telefonică.")
			g.Play(g.Audio.Achievement)
			return
		}
	}
	g.sayIcon(IconStar, "Stea nouă!")
	g.Play(g.Audio.Coin)
}

// sectorul lumii in care se afla Aventurierul (-1 = pe mare / in tunel)
func (g *Game) currentWorld() int {
	return WorldOfSector(SectorAt(g.Player.Pos.X, g.Player.Pos.Z))
}

// markBeachPad: steaua 2 pe plaja = ai sarit pe ambele trambuline de acolo.
func (g *Game) markBeachPad(pi int) {
	w := g.City.W
	n := 0
	for i := range w.Pads {
		if SectorAt(w.Pads[i].Pos.X, w.Pads[i].Pos.Z) != SecWest {
			continue
		}
		if i == pi && n < len(g.Prog.BeachPads) {
			g.Prog.BeachPads[n] = true
		}
		n++
	}
}

// freeFogMeshes: norii lumii vechi raman in memorie (cache pe lume); MeshBuilder
// foloseste memorie Go, deci NU se apeleaza rl.UnloadMesh (ar cadea jocul).
var fogCache = map[int]struct {
	m [MaxZones]rl.Mesh
	c [MaxZones]rl.Vector3
}{}

func freeFogMeshes() {
	if fogReady {
		fogCache[fogRealm] = struct {
			m [MaxZones]rl.Mesh
			c [MaxZones]rl.Vector3
		}{fogMeshes, fogCenters}
	}
	fogReady = false
}

var fogRealm int
