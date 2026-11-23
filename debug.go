package main

import (
	"fmt"
	"math"
	"os"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Mod de verificare automata (nu apare in jocul normal, nu salveaza nimic):
//   BATTLE_DEBUG=shots  -> capturi de ecran din locuri cheie
//   BATTLE_DEBUG=smoke  -> Aventurierul merge singur: piata, scarile, Marele Tunel (vest-est), parc, tunel de munte, poduri
//   BATTLE_DEBUG=fight  -> lupta in timp real (pistol de departe, sabie de aproape), 2 niveluri
//   BATTLE_DEBUG=fun    -> inot, trambulina, roata uriasa, montagne russe, lift, acoperis, constructie cu sabia

type DebugView struct {
	Name     string
	Pos      rl.Vector3
	CamYaw   float32
	CamPitch float32
	CamDist  float32
	Scene    GameScene
	Setup    func(g *Game)
	Wait     int
}

type DebugRun struct {
	v6        debug6
	NoLocks   bool // V 5.0: testele vechi merg pe toata harta (lumile inchise nu blocheaza)
	v5        debug5
	lvl0      int
	Mode      string
	Frame     int
	Views     []DebugView
	ViewIdx   int
	ViewF     int
	OutDir    string
	ShowFPS   bool
	FixedStep bool
	Way       []rl.Vector2
	WayIdx    int
	WayT      float32
	Log       []string
	Failures  int
	MaxTime   float32
	fpsSum    float32
	fpsN      int
	MaxY      float32
	MinY      float32

	Kills, Hits, Deaths, Levels int
	Swims                       int
	MaxApex                     float32
	MaxRideY                    float32
	CoasterLaps                 int
	waitT                       float32

	// controale scriptate
	moveDir    rl.Vector3
	moveOn     bool
	jumpOn     bool
	fireOn     bool
	eQueued    bool
	cQueued    bool
	CarsBoomed int
	ThingsHit  [4]int
	carK       int
	minY2      float32
	gold0      float32
	driveTh    float32
	driveSt    float32
	driveOn    bool
	step       int
	stepT      float32
	stepVal    float32
	stepInt    int
}

func ParseDebug() *DebugRun {
	m := os.Getenv("BATTLE_DEBUG")
	if m == "" {
		return nil
	}
	out := os.Getenv("BATTLE_DEBUG_OUT")
	if out == "" {
		out = "."
	}
	return &DebugRun{Mode: m, OutDir: out, ShowFPS: true, MaxTime: 900, FixedStep: m != "shots", NoLocks: os.Getenv("BATTLE_DEBUG_LOCKS") == ""}
}

// ---- legaturile cu jocul ----

func (d *DebugRun) Move(g *Game) (rl.Vector3, bool) {
	switch d.Mode {
	case "smoke":
		return d.smokeMove(g), true
	case "fight":
		mv, fire := d.fightControl(g)
		d.fireOn = fire
		return mv, true
	case "fun":
		return d.moveDir, d.moveOn
	}
	return rl.Vector3{}, false
}

func (d *DebugRun) Jump(g *Game) bool { return d.jumpOn }

func (d *DebugRun) Fire(g *Game) bool { return d.fireOn }

func (d *DebugRun) DriveInput(g *Game) (float32, float32, bool) {
	return d.driveTh, d.driveSt, d.driveOn
}

func (d *DebugRun) PressC(g *Game) bool {
	if d.cQueued {
		d.cQueued = false
		return true
	}
	return false
}

func (d *DebugRun) PressE(g *Game) bool {
	if d.eQueued {
		d.eQueued = false
		return true
	}
	return false
}

func (d *DebugRun) logf(format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	d.Log = append(d.Log, s)
	fmt.Println(s)
}

func (d *DebugRun) check(ok bool, format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	if ok {
		d.logf("OK    %s", s)
	} else {
		d.Failures++
		d.logf("GRESIT %s", s)
	}
}

func (d *DebugRun) Setup(g *Game) {
	g.SaveData = &SaveData{}
	if os.Getenv("BATTLE_DEBUG_SAVE") != "" {
		// salvarea reala (APPDATA ales de test): jucator ca din meniu
		g.SaveData.Players = []PlayerSave{CreatePlayer("Aventurierul")}
		g.PlayerIdx = 0
	}
	g.Player.Name = "Aventurierul"
	g.Difficulty = "medium"
	g.DifficultyIdx = 2
	g.applyDifficulty()
	g.spawnEnemies(1)
	for i := range g.City.LegoSites {
		s := &g.City.LegoSites[i]
		s.CurrentStage = (i*5 + 6) % (s.TotalStages + 1)
		if i == 8 || i == 9 {
			s.CurrentStage = s.TotalStages
		}
		if i == 0 {
			s.CurrentStage = s.TotalStages
			s.UpgradeLevel = 3
			s.TargetScale = 3
			s.CurrentScale = 3
		}
	}
	g.Scene = SceneMapExplore
	g.ShowHelp = 0
	d.MinY = 1e9
	if v := os.Getenv("BATTLE_DEBUG_PROBE"); v != "" {
		probeColliders(g, v)
	}
	if id := os.Getenv("BATTLE_REALM"); id != "" {
		// V 6.0: testele si capturile intr-o lume mare anume
		if ri := realmIndex(id); ri > 0 {
			g.switchRealm(ri)
			g.arrivedInRealm()
		}
	}
	if v := os.Getenv("BATTLE_DEBUG_LEVEL"); v != "" {
		fmt.Sscanf(v, "%d", &g.Level)
		for i := range g.Prog.OwnOutfit {
			g.Prog.OwnOutfit[i] = true
		}
		g.syncFog()
		g.spawnEnemies(g.Level)
	}
	switch d.Mode {
	case "shots":
		d.setupShots(g)
	case "smoke":
		w := g.City.W
		m0 := &w.Mountains[0]
		way := []rl.Vector2{}
		add := func(x, z float32) { way = append(way, rl.NewVector2(x, z)) }
		add(20, 23)   // gura scarilor din piata
		add(20, 7)    // jos, in Marele Tunel
		add(-250, 3)  // sub mare, langa liftul de vest
		add(-386, 2)  // sus pe rampa, pe Insula Plajei
		add(-300, -2) // inapoi in tunel
		add(250, -2)  // sub mare spre est
		add(390, -2)  // sus, in Parcul de Distractii
		add(330, -38) // la roata uriasa
		add(385, -2)  // coboara in santul Marelui Tunel
		add(300, -2)  // inapoi in tunel
		add(20, 3)    // sub piata
		add(20, 8)
		add(20, 23) // urca scarile
		add(26, 30)
		add(0, 42)
		x, z := m0.World(-40, 0)
		add(x, z)
		x, z = m0.World(0, 0)
		add(x, z)
		x, z = m0.World(PortalOut+14, 0)
		add(x, z)
		add(SecNW_X+12, SecNW_Z+12)
		add(SecNW_X+14, SecNW_Z+2)
		add(SecNE_X-14, SecNE_Z+2)
		add(SecNE_X+2, SecNE_Z+14)
		add(SecSE_X+2, SecSE_Z-14)
		add(SecSE_X-14, SecSE_Z+2)
		add(SecSW_X+14, SecSW_Z+2)
		add(SecSW_X+2, SecSW_Z-14)
		add(SecNW_X+2, SecNW_Z+14)
		d.Way = way
		g.Player.Pos = v3(0, 0.2, 30)
		g.Enemies = g.Enemies[:0]
	case "fight":
		g.Player.Pos = v3(0, 0.2, 30)
		g.Difficulty = "easy"
		g.applyDifficulty()
		g.spawnEnemies(1)
	case "fun":
		g.Enemies = g.Enemies[:0]
		g.Player.Gold = 5000
		if v := os.Getenv("BATTLE_DEBUG_FROM"); v != "" {
			fmt.Sscanf(v, "%d", &d.step)
		}
	}
}

// ---------------- traseul (smoke) ----------------

func (d *DebugRun) smokeMove(g *Game) rl.Vector3 {
	if d.WayIdx >= len(d.Way) {
		return rl.Vector3{}
	}
	p := g.Player.Pos
	t := d.Way[d.WayIdx]
	dx, dz := t.X-p.X, t.Y-p.Z
	dist := sqrtf(dx*dx + dz*dz)
	d.WayT += g.FrameTime
	if dist < 3 {
		d.logf("OK  punct %d (%.0f, %.0f) atins in %.1fs, y=%.2f", d.WayIdx, t.X, t.Y, d.WayT, p.Y)
		d.WayIdx++
		d.WayT = 0
		return rl.Vector3{}
	}
	if d.WayT > 90 {
		d.logf("BLOCAT la punctul %d (%.0f, %.0f): Aventurierul e la (%.1f, %.1f, %.1f)", d.WayIdx, t.X, t.Y, p.X, p.Y, p.Z)
		d.Failures++
		gy, _ := g.City.W.GroundAt(t.X, t.Y, p.Y+2)
		g.Player.Pos = v3(t.X, gy, t.Y)
		d.WayIdx++
		d.WayT = 0
		return rl.Vector3{}
	}
	g.CamYaw = atan2f(-dx, -dz)
	return rl.Vector3Normalize(v3(dx, 0, dz))
}

// ---------------- lupta (fight) ----------------

func (d *DebugRun) fightControl(g *Game) (rl.Vector3, bool) {
	p := g.Player.Pos
	best := -1
	bd := float32(1e9)
	for i := range g.Enemies {
		e := &g.Enemies[i]
		if !e.Alive {
			continue
		}
		dd := rl.Vector3Distance(e.Pos, p)
		if dd < bd {
			bd, best = dd, i
		}
	}
	if best < 0 {
		return rl.Vector3{}, false
	}
	e := &g.Enemies[best]
	dx, dz := e.Pos.X-p.X, e.Pos.Z-p.Z
	g.CamYaw = atan2f(-dx, -dz)
	g.CamPitch = 0.12
	d.WayT += g.FrameTime
	if d.WayT > 30 {
		g.Player.Pos = v3(e.Pos.X+6, e.Pos.Y, e.Pos.Z+6)
		d.WayT = 0
	}
	if bd > 12 {
		return rl.Vector3Normalize(v3(dx, 0, dz)), bd < 45 && g.AimEnemy == best
	}
	if bd > 3 && g.Weapon == WeaponSword {
		return rl.Vector3Normalize(v3(dx, 0, dz)), true
	}
	return rl.Vector3{}, g.AimEnemy >= 0 || g.Weapon == WeaponSword
}

// ---------------- capturi (shots) ----------------

func (d *DebugRun) setupShots(g *Game) {
	if curRealm > 0 {
		d.setupRealmShots(g)
		return
	}
	w := g.City.W
	m0 := &w.Mountains[0]
	pxIn, pzIn := m0.World(PortalOut+12, 0)
	pxT, pzT := m0.World(20, 0)
	b0 := &w.Bridges[0]
	bp := func(b *Bridge, t float32) rl.Vector3 { x, z := b.Point(t, 0); return v3(x, b.DeckY(t), z) }
	gt := &w.Corridors[0]
	enemyAt := func(off rl.Vector3) func(g *Game) {
		return func(g *Game) {
			e := &g.Enemies[0]
			e.Pos = rl.Vector3Add(g.Player.Pos, off)
			e.Alive = true
			e.Home = rl.NewVector2(e.Pos.X, e.Pos.Z)
			e.HomeR = 10
			g.Invuln = 99
		}
	}
	d.Views = []DebugView{
		{Name: "01_plaza", Pos: v3(0, 0, 30), CamYaw: 0.0, CamPitch: 0.3, CamDist: 9},
		{Name: "02_gun_crosshair", Pos: v3(0, 0, 34), CamYaw: 0, CamPitch: 0.12, CamDist: 6, Setup: enemyAt(v3(0, 0, -16)), Wait: 20},
		{Name: "03_sword_enemy", Pos: v3(0, 0, 34), CamYaw: 0, CamPitch: 0.2, CamDist: 6, Setup: func(g *Game) {
			enemyAt(v3(0.5, 0, -3))(g)
			g.SwingT = swingTime * 0.4
		}, Wait: 12},
		{Name: "04_sword_build", Pos: v3(-16, 0, -8), CamYaw: 0.2, CamPitch: 0.3, CamDist: 8, Wait: 20},
		{Name: "05_mountain_portal", Pos: v3(pxIn, 0, pzIn), CamYaw: atan2f(m0.Dir.X, m0.Dir.Y) + 0.35, CamPitch: 0.15, CamDist: 13},
		{Name: "06_mountain_tunnel", Pos: v3(pxT, 0, pzT), CamYaw: atan2f(m0.Dir.X, m0.Dir.Y), CamPitch: 0.08, CamDist: 6},
		{Name: "07_great_tunnel_portal", Pos: v3(-372, 0, 9), CamYaw: -1.25, CamPitch: 0.3, CamDist: 14},
		{Name: "08_great_tunnel", Pos: v3(-120, gt.FloorC(272), 2), CamYaw: -math32Pi / 2, CamPitch: 0.05, CamDist: 6},
		{Name: "09_undersea_fish", Pos: v3(-200, gt.FloorC(192), -3), CamYaw: -math32Pi/2 - 0.5, CamPitch: 0.05, CamDist: 5},
		{Name: "10_stairs_exit", Pos: v3(20, 0, 24), CamYaw: math32Pi, CamPitch: 0.7, CamDist: 9},
		{Name: "11_lift_islet", Pos: v3(256, 0, 9), CamYaw: 0.4, CamPitch: 0.25, CamDist: 18},
		{Name: "12_funpark", Pos: v3(300, 0, -10), CamYaw: -1.88, CamPitch: 0.3, CamDist: 30},
		{Name: "13_wheel_top", Setup: func(g *Game) {
			fw := &g.City.Wheels[len(g.City.Wheels)-1]
			best, by := 0, float32(-1)
			for k := 0; k < fw.N; k++ {
				if y := fw.CabinAxle(k).Y; y > by {
					best, by = k, y
				}
			}
			g.Ride = RideState{Kind: RideWheel, Idx: len(g.City.Wheels) - 1, Cabin: best}
			g.CamYaw = 2.6
			g.CamPitch = 0.55
			g.CamDist = 30
		}, Wait: 6},
		{Name: "14_coaster", Setup: func(g *Game) {
			c := g.City.Coaster
			c.S = c.nearestS(c.Ctrl[7])
			c.Running = true
			c.V = 8
			c.Traveled = 100
			c.Rider = true
			g.Ride = RideState{Kind: RideCoaster}
			g.CamPitch = 0.2
		}, Wait: 4},
		{Name: "15_trampolines", Pos: v3(384, 0, -36), CamYaw: 0.4, CamPitch: 0.3, CamDist: 18, Setup: func(g *Game) {
			pd := &g.City.W.Pads[len(g.City.W.Pads)-3]
			g.Player.Pos = v3(pd.Pos.X, pd.MatY()+14, pd.Pos.Z)
			g.Player.OnGround = false
			g.Player.VelY = 0
			g.Ride = RideState{Kind: RideTrampoline, Idx: len(g.City.W.Pads) - 3}
		}, Wait: 3},
		{Name: "16_swimming", Pos: v3(-350, -2.25, 76), CamYaw: 0.3, CamPitch: 0.3, CamDist: 9, Setup: func(g *Game) {
			g.Swimming = true
			g.Player.Pose.Swim = 1
			g.Player.Pose.Moving = 1
		}, Wait: 8},
		{Name: "17_suspension", Pos: bp(b0, 0.08), CamYaw: atan2f(-b0.Dir.X, -b0.Dir.Y) + 0.5, CamPitch: 0.25, CamDist: 20},
		{Name: "18_beach_island", Pos: v3(-330, 0, 10), CamYaw: -1.0, CamPitch: 0.4, CamDist: 30},
		{Name: "19_map", Pos: v3(0, 0, 30), CamYaw: 0, CamPitch: 0.3, CamDist: 9, Setup: func(g *Game) { g.ShowMap = true }},
		{Name: "20_menu", Scene: SceneMainMenu},
		{Name: "21_stairs_walkup", Pos: v3(20, 0, 27), CamYaw: 0, CamPitch: 0.12, CamDist: 6},
		{Name: "22_stairs_above", Pos: v3(26, 0, 22), CamYaw: 0.9, CamPitch: 0.75, CamDist: 12},
		{Name: "23_ramp_top", Pos: v3(-398, 0, 0), CamYaw: -math32Pi / 2, CamPitch: 0.15, CamDist: 6},
		{Name: "24_ramp_low", Pos: v3(-352, gt.FloorC(40), 1), CamYaw: -math32Pi / 2, CamPitch: 0.1, CamDist: 6},
		{Name: "25_park_ramp", Pos: v3(398, 0, 0), CamYaw: math32Pi / 2, CamPitch: 0.2, CamDist: 7},
		{Name: "26_tunnel_lookup", Pos: v3(-150, gt.FloorC(242), 0), CamYaw: -math32Pi / 2, CamPitch: -0.35, CamDist: 4},
		{Name: "27_drive_tunnel", Setup: func(g *Game) {
			c := &g.City.Cars[0]
			c.Parked = true
			c.Pos = v3(-150, gt.FloorC(242)+0.25, 2.2)
			c.Yaw = math32Pi / 2
			g.enterCar(0)
			g.CamYaw = c.Yaw + math32Pi + 0.3
		}, Wait: 10},
		{Name: "28_drive_bridge", Setup: func(g *Game) {
			c := &g.City.Cars[1]
			c.Parked = true
			x, z := b0.Point(0.3, 2.2)
			c.Pos = v3(x, b0.DeckY(0.3), z)
			c.Yaw = atan2f(b0.Dir.X, b0.Dir.Y)
			g.enterCar(1)
			g.CamYaw = c.Yaw + math32Pi - 0.4
			g.CamPitch = 0.3
		}, Wait: 10},
		{Name: "29_car_boom", Pos: v3(0, 0, 38), CamYaw: 0, CamPitch: 0.15, CamDist: 7, Setup: func(g *Game) {
			c := &g.City.Cars[2]
			c.Parked = true
			c.Dead = 0
			c.HP = 0.5
			c.Pos = v3(0, 0, 28)
			c.Yaw = math32Pi / 2
			g.carHit(2, 1, v3(0, 1, 28))
		}, Wait: 5},
		{Name: "30_coaster_3rd", Setup: func(g *Game) {
			c := g.City.Coaster
			c.S = c.nearestS(c.Ctrl[7])
			c.Running = true
			c.V = 8
			c.Traveled = 100
			c.Rider = true
			g.Ride = RideState{Kind: RideCoaster}
			g.CamYaw = 1.2
			g.CamPitch = 0.35
			g.CamDist = 11
		}, Wait: 6},
		{Name: "31_wheel_3rd", Setup: func(g *Game) {
			fw := &g.City.Wheels[len(g.City.Wheels)-1]
			g.Ride = RideState{Kind: RideWheel, Idx: len(g.City.Wheels) - 1, Cabin: (fw.LowestCabin() + fw.N/2 - 1) % fw.N}
			g.CamYaw = 2.2
			g.CamPitch = 0.2
			g.CamDist = 14
		}, Wait: 6},
		{Name: "32_site_upgrade", Pos: v3(-16, 0, -8), CamYaw: 0.2, CamPitch: 0.3, CamDist: 8, Setup: func(g *Game) {
			for i := range g.City.LegoSites {
				s := &g.City.LegoSites[i]
				if absf(s.Position.X+16) < 9 && absf(s.Position.Z+14) < 9 {
					for s.CurrentStage < s.TotalStages {
						g.placeBrick(s)
					}
				}
			}
			g.UpgradeHits = 2
			g.Floats = nil
			g.MessageT = 0
		}, Wait: 20},
		{Name: "33_prompt_car", Pos: v3(0, 0, 31.6), CamYaw: 0.4, CamPitch: 0.25, CamDist: 7, Setup: func(g *Game) {
			c := &g.City.Cars[3]
			c.Parked = true
			c.Dead = 0
			c.HP = CarMaxHP
			c.Pos = v3(0, 0, 28)
			c.Yaw = math32Pi / 2
		}, Wait: 8},
		{Name: "34_plane_aim", Pos: v3(0, 0, 40), CamDist: 7, Setup: func(g *Game) {
			// avionul de jos, adus la 70 m in fata (doar pentru captura)
			pl := &g.City.Planes[0]
			pl.Ang = math32Pi / 2
			pl.CZ = -30 - pl.RZ
			g.City.updateSkyTargets(0, g.Time, g.Player.Pos, g.Camera.Position, nil)
			d.aimThing(g, TgtPlane, 0)
			g.CamPitch += 0.08
		}, Wait: 3},
		{Name: "35_birds", Pos: v3(0, 0, 40), CamDist: 6, Setup: func(g *Game) {
			f := &g.City.Flocks[0]
			f.R = 4
			f.C = rl.NewVector2(0, 14)
			f.H = 12
			g.City.updateSkyTargets(0, g.Time, g.Player.Pos, g.Camera.Position, nil)
			d.aimThing(g, TgtBird, -1)
		}, Wait: 3},
		{Name: "36_boat_sink", Setup: func(g *Game) {
			// Aventurierul inoata langa barca 0, pe care o lovim pana se scufunda
			b := &g.City.Boats[0]
			b.Speed = 0
			b.Update(0, g.Time)
			d.teleport(g, b.Pos.X+14, b.Pos.Z+14)
			g.Swimming = true
			g.Player.Pos.Y = WaterY - 1.25
			g.CamDist = 10
			d.aimThing(g, TgtBoat, 0)
			g.CamPitch = 0.3
			b.HP = 1
			g.hitThing(TgtBoat, 0, b.Pos)
		}, Wait: 40},
		{Name: "37_roof_pad", Setup: func(g *Game) {
			pi, _ := d.cityPadWithRoof(g)
			pd := &g.City.W.Pads[pi]
			d.teleport(g, pd.Pos.X, pd.Pos.Z)
			g.mountPad(pi)
			g.Player.Pos.Y = pd.MatY() + 9
			g.Player.VelY = -2
			g.CamYaw = 0.8
			g.CamPitch = 0.35
			g.CamDist = 16
		}, Wait: 3},
		{Name: "38_parachute", Pos: v3(0, 0, 40), CamYaw: 0, CamPitch: -0.25, CamDist: 7, Setup: func(g *Game) {
			g.City.Chute = Parachute{Pos: v3(0, 14, 22), Yaw: 0.5, Active: true}
		}, Wait: 4},
		{Name: "39_banner_message", Pos: v3(0, 0, 30), CamYaw: 0, CamPitch: 0.3, CamDist: 9, Setup: func(g *Game) {
			g.LevelBanner = 3
			g.sayIcon(IconTunnel, "Marele Tunel de sub Mare - ai intrat în tunel!")
		}, Wait: 4},
		{Name: "40_lift_prompt", Pos: v3(256, 0, 9), CamYaw: 0.4, CamPitch: 0.25, CamDist: 9, Setup: func(g *Game) {
			l := g.City.W.Lifts[1]
			d.teleport(g, l.Top.X, l.Top.Z+1)
		}, Wait: 6},
	}
	if v := os.Getenv("BATTLE_DEBUG_VIEWS"); v != "" {
		// capturi ad-hoc: "nume:x,y,z,yaw,pitch,dist;nume2:..."
		d.Views = nil
		for _, part := range strings.Split(v, ";") {
			nv := strings.SplitN(part, ":", 2)
			if len(nv) != 2 {
				continue
			}
			var x, y, z, yaw, pitch, dist float32
			fmt.Sscanf(nv[1], "%f,%f,%f,%f,%f,%f", &x, &y, &z, &yaw, &pitch, &dist)
			d.Views = append(d.Views, DebugView{Name: nv[0], Pos: v3(x, y, z), CamYaw: yaw, CamPitch: pitch, CamDist: dist, Wait: 10})
		}
	}
	if v := os.Getenv("BATTLE_DEBUG_ONLY"); v != "" {
		keep := []DebugView{}
		for _, x := range d.Views {
			if strings.Contains(x.Name, v) {
				keep = append(keep, x)
			}
		}
		d.Views = keep
	}
	d.applyView(g)
}

const math32Pi = float32(3.14159265)

func (d *DebugRun) applyView(g *Game) {
	if d.ViewIdx >= len(d.Views) {
		return
	}
	if g.Drive.Active {
		g.exitCar()
	}
	v := d.Views[d.ViewIdx]
	g.Scene = SceneMapExplore
	g.ShowMap = false
	g.Ride = RideState{}
	g.Swimming = false
	g.Player.Pose = Pose{}
	g.SwingT = 0
	if v.Scene == SceneMainMenu {
		g.Scene = SceneMainMenu
		return
	}
	if len(g.Enemies) > 0 {
		g.Enemies[0].Pos = v3(0, -100, 0)
	}
	if v.Pos != (rl.Vector3{}) || v.Setup == nil {
		g.Player.Pos = v.Pos
		if v.Pos.Y > -2 || v.Pos.Y < -2.5 {
			gy, _ := g.City.W.GroundAt(v.Pos.X, v.Pos.Z, v.Pos.Y+0.5)
			g.Player.Pos.Y = gy
		}
		g.Player.OnGround = true
		g.CamYaw = v.CamYaw
		g.CamPitch = v.CamPitch
		g.CamDist = v.CamDist
		g.CamBoom = v.CamDist
		g.Player.Yaw = v.CamYaw + 3.14159
	}
	if v.Setup != nil {
		v.Setup(g)
	}
}

// ---------------- distractii (fun) ----------------

func (d *DebugRun) teleport(g *Game, x, z float32) {
	if g.Drive.Active {
		g.exitCar()
	}
	gy, _ := g.City.W.GroundAt(x, z, 60)
	g.Player.Pos = v3(x, gy, z)
	g.Player.OnGround = true
	g.Player.VelY = 0
	g.Swimming = false
	g.Ride = RideState{}
	d.moveOn, d.jumpOn, d.fireOn = false, false, false
}

func (d *DebugRun) next() {
	d.step++
	d.stepT = 0
	d.stepVal = 0
	d.stepInt = 0
}

func (d *DebugRun) funStep(g *Game) bool {
	dt := g.FrameTime
	d.stepT += dt
	p := &g.Player
	w := g.City.W
	switch d.step {
	case 0: // inot: de pe plaja in mare
		d.teleport(g, -356, 50)
		d.moveDir, d.moveOn = v3(0, 0, 1), true
		g.CamYaw = math32Pi
		d.next()
	case 1:
		if g.Swimming && p.Pos.Y < -1.8 {
			d.stepInt = 1
		}
		if d.stepT > 8 {
			sd := w.LandSD(p.Pos.X, p.Pos.Z)
			d.check(d.stepInt == 1 && g.Swimming, "inot: Aventurierul pluteste (y=%.2f, apa la %.1f)", p.Pos.Y, WaterY)
			d.check(sd > 25, "inot: a inotat liber in larg, la %.1f m de mal (fara limita de 10 m)", sd)
			d.moveDir = v3(0, 0, -1)
			d.next()
		}
	case 2:
		if d.stepT > 13 {
			d.check(!g.Swimming && p.Pos.Y > -0.6, "inot: a iesit singur din apa pe plaja (y=%.2f)", p.Pos.Y)
			d.next()
		}
	case 3: // trambulina gigant, cu SPATIU tinut apasat
		pi := len(w.Pads) - 2
		pd := &w.Pads[pi]
		d.teleport(g, pd.Pos.X+pd.R+1.8, pd.Pos.Z)
		d.eQueued = true
		d.jumpOn = true
		d.MaxApex = 0
		d.next()
	case 4:
		if d.stepT > 22 {
			pd := &w.Pads[len(w.Pads)-2]
			d.check(d.MaxApex >= pd.MaxApex*0.9, "trambulina: a sarit %.1f m (maxim %.0f)", d.MaxApex, pd.MaxApex)
			d.jumpOn = false
			d.next()
		}
	case 5: // fara SPATIU, sariturile scad pana la cele mici
		if d.stepT > 14 {
			d.check(g.PadApex < d.MaxApex*0.6, "trambulina: fara SPATIU sariturile scad (acum %.1f m)", g.PadApex)
			d.eQueued = true
			d.next()
		}
	case 6:
		if d.stepT > 1.5 {
			d.check(g.Ride.Kind == RideNone && p.Pos.Y < 1.5, "trambulina: E il coboara jos (y=%.2f)", p.Pos.Y)
			// SPATIU apasat doar aproape de panza (ca un copil), nu tinut
			pd := &w.Pads[len(w.Pads)-3]
			d.teleport(g, pd.Pos.X+pd.R+1.8, pd.Pos.Z)
			d.eQueued = true
			d.MaxApex = 0
			d.next()
		}
	case 7:
		pd := &w.Pads[len(w.Pads)-3]
		h := p.Pos.Y - pd.MatY()
		d.jumpOn = g.Ride.Kind == RideTrampoline && p.VelY < 0 && h < 1.2 && int(d.stepT*10)%2 == 0
		if d.stepT > 24 {
			d.check(d.MaxApex >= pd.MaxApex*0.9, "trambulina: apasand SPATIU doar jos, sare %.1f m (maxim %.0f)", d.MaxApex, pd.MaxApex)
			d.jumpOn = false
			d.eQueued = true
			d.next()
		}
	case 8: // roata uriasa
		if d.stepT < 1 {
			return false
		}
		fw := &g.City.Wheels[len(g.City.Wheels)-1]
		d.teleport(g, fw.Board.X, fw.Board.Z)
		d.eQueued = true
		d.MaxRideY = 0
		d.next()
	case 9:
		if d.stepT > 1 && d.stepInt == 0 {
			d.stepInt = 1
			d.check(g.Ride.Kind == RideWheel, "roata: E il urca in cabina")
		}
		fw := &g.City.Wheels[len(g.City.Wheels)-1]
		if d.stepT > 3.14159/fw.Speed+4 {
			d.check(d.MaxRideY > 58, "roata: a urcat pana la %.1f m", d.MaxRideY)
			d.eQueued = true
			d.next()
		}
	case 10:
		if d.stepT > 1.5 {
			d.check(g.Ride.Kind == RideNone && p.Pos.Y < 1, "roata: E il coboara pe platforma (y=%.2f)", p.Pos.Y)
			c := g.City.Coaster
			s0 := c.Ctrl[1]
			d.teleport(g, s0.X, s0.Z-2.6)
			d.MaxRideY = 0
			d.next()
		}
	case 11: // montagne russe: asteapta trenul in statie, urca
		c := g.City.Coaster
		if int(d.stepT*30)%150 == 0 {
			d.logf("      [coaster t=%.0fs] running=%v wait=%.1f S=%.0f/%.0f V=%.1f prompt=%d \"%s\" ride=%d", d.stepT, c.Running, c.Wait, c.S, c.Len, c.V, g.PromptKind, g.Prompt, g.Ride.Kind)
		}
		if g.PromptKind == PromptCoaster && g.Ride.Kind == RideNone && d.stepInt == 0 {
			d.eQueued = true
			d.stepInt = 1
		}
		if g.Ride.Kind == RideCoaster && d.stepInt == 1 {
			d.stepInt = 2
			d.logf("      montagne russe: urcat dupa %.1fs", d.stepT)
		}
		if d.stepInt == 2 && g.Ride.Kind == RideNone && d.stepT > 5 {
			d.check(d.CoasterLaps >= 1, "montagne russe: o tura completa (cel mai sus %.1f m, %.0fs)", d.MaxRideY, d.stepT)
			d.check(d.MaxRideY > 28, "montagne russe: a urcat dealul de %.0f m", d.MaxRideY)
			d.next()
		}
		if d.stepT > 160 {
			d.check(false, "montagne russe: nu a terminat tura (stare %d, S=%.0f/%.0f, V=%.1f)", d.stepInt, c.S, c.Len, c.V)
			d.next()
		}
	case 12: // liftul de pe insulita de vest
		if g.FadeDir != 0 || d.stepT < 1 {
			return false
		}
		l := w.Lifts[0]
		g.Player.Pos = l.Bottom
		g.Player.OnGround = true
		d.eQueued = true
		d.next()
	case 13:
		if d.stepT > 1.5 {
			l := w.Lifts[0]
			d.check(absf(p.Pos.Y-l.Top.Y) < 0.6 && absf(p.Pos.X-l.Top.X) < 3, "lift: a urcat pe insulita (y=%.2f)", p.Pos.Y)
			d.eQueued = true
			d.next()
		}
	case 14:
		if d.stepT > 1.5 {
			l := w.Lifts[0]
			d.check(absf(p.Pos.Y-l.Bottom.Y) < 0.8, "lift: a coborat in tunel (y=%.2f)", p.Pos.Y)
			d.next()
		}
	case 15: // acoperis: cade pe un bloc si ramane sus
		for i := range w.Boxes {
			b := &w.Boxes[i]
			if b.Roof == RoofFlat && b.Top > 8 {
				cx, cz := (b.MinX+b.MaxX)/2, (b.MinZ+b.MaxZ)/2
				d.teleport(g, cx, cz)
				g.Player.Pos = v3(cx, b.Top+6, cz)
				g.Player.OnGround = false
				d.next()
				d.stepVal = b.Top
				return false
			}
		}
		d.next()
	case 16:
		if d.stepT > 2 {
			d.check(absf(p.Pos.Y-d.stepVal) < 0.3 && p.OnGround, "acoperis: sta pe bloc la %.1f m (roof %.1f)", p.Pos.Y, d.stepVal)
			d.next()
		}
	case 17: // construieste lovind santierul cu sabia
		si := -1
		for i := range g.City.LegoSites {
			s := &g.City.LegoSites[i]
			if s.CurrentStage < s.TotalStages-3 {
				si = i
				break
			}
		}
		if si < 0 {
			d.check(false, "sabie: nu am gasit un santier neterminat")
			d.next()
			break
		}
		s := &g.City.LegoSites[si]
		d.teleport(g, s.Position.X+7, s.Position.Z)
		g.CamYaw = atan2f(7, 0)
		d.fireOn = true
		d.next()
		d.stepInt = s.CurrentStage
		d.stepVal = float32(si)
	case 18:
		if d.stepT > 3 {
			s := &g.City.LegoSites[int(d.stepVal)]
			d.check(g.Weapon == WeaponSword, "sabie: apare singura langa santier")
			d.check(s.CurrentStage > d.stepInt, "sabie: loviturile au pus %d piese", s.CurrentStage-d.stepInt)
			d.fireOn = false
			d.next()
		}
	case 19: // inamicii lovesc in timp real (fara ecran de lupta, camera libera)
		d.teleport(g, 0, 34)
		g.spawnEnemies(1)
		g.Enemies = g.Enemies[:2]
		for k := range g.Enemies {
			e := &g.Enemies[k]
			off := v3(0, 0, -2.6)
			if k == 1 {
				off = v3(6, 0, -13)
			}
			e.Pos = rl.Vector3Add(g.Player.Pos, off)
			e.Home = rl.NewVector2(e.Pos.X, e.Pos.Z)
			e.HomeR = 20
			e.AtkCD = 0.5
		}
		g.Enemies[0].Ranged = false
		g.Enemies[1].Ranged = true
		g.Invuln = 0
		d.Hits = 0
		d.next()
	case 20:
		if d.stepT > 8 {
			d.check(d.Hits >= 2, "inamici: l-au lovit de %d ori in 8 s (corp la corp si cu bile de foc)", d.Hits)
			d.check(g.Scene == SceneMapExplore || g.Scene == SceneGameOver, "inamici: nu apare niciun ecran de lupta (scena %d)", g.Scene)
			shots := 0
			for _, pr := range g.Projectiles {
				if pr.Enemy {
					shots++
				}
			}
			d.logf("      bile de foc in zbor acum: %d, viata aventurierului: %.0f", shots, g.Player.HP)
			d.next()
		}
	case 21: // trambulina: mergand spre ea se urca singur
		pd := &w.Pads[len(w.Pads)-4]
		d.teleport(g, pd.Pos.X+pd.R+3.2, pd.Pos.Z)
		d.moveDir, d.moveOn = v3(-1, 0, 0), true
		d.MaxApex = 0
		d.next()
	case 22:
		if g.Ride.Kind == RideTrampoline {
			d.moveOn = false
		}
		if d.stepT > 4 {
			d.check(g.Ride.Kind == RideTrampoline && d.MaxApex > 1.5, "trambulina: mergand spre ea s-a urcat singur (sare %.1f m)", d.MaxApex)
			// cade pe panza de sus, fara E: trebuie sa sara din nou
			pd := &w.Pads[len(w.Pads)-4]
			d.teleport(g, pd.Pos.X, pd.Pos.Z)
			p.Pos.Y = pd.MatY() + 7
			p.OnGround = false
			p.VelY = 0
			g.AirStartY = p.Pos.Y
			d.MaxApex = 0
			d.next()
		}
	case 23:
		if d.stepT > 3 {
			d.check(g.Ride.Kind == RideTrampoline && d.MaxApex > 1.5, "trambulina: cazand pe panza (fara E) sare din nou %.1f m", d.MaxApex)
			d.eQueued = true
			d.next()
		}
	case 24: // roata: camera libera din spate, C = din ochii aventurierului
		if d.stepT < 1 {
			return false
		}
		fw := &g.City.Wheels[len(g.City.Wheels)-1]
		d.teleport(g, fw.Board.X, fw.Board.Z)
		d.eQueued = true
		d.next()
	case 25:
		if d.stepT > 2.5 && d.stepInt == 0 {
			head := v3(p.Pos.X, p.Pos.Y+1.75, p.Pos.Z)
			dist := rl.Vector3Distance(g.Camera.Position, head)
			d.check(g.Ride.Kind == RideWheel && !g.Ride.First && dist > 6, "atractii: in roata camera e din spate (3rd person), la %.1f m", dist)
			g.CamYaw += 2.0
			g.CamPitch = -0.9
			d.stepInt = 1
		}
		if d.stepT > 3.2 && d.stepInt == 1 {
			head := v3(p.Pos.X, p.Pos.Y+1.75, p.Pos.Z)
			d.check(g.Camera.Position.Y < head.Y-1, "atractii: camera se roteste liber (si de jos in sus): cam y %.1f, cap %.1f", g.Camera.Position.Y, head.Y)
			d.cQueued = true
			d.stepInt = 2
		}
		if d.stepT > 4 && d.stepInt == 2 {
			head := v3(p.Pos.X, p.Pos.Y+1.75, p.Pos.Z)
			dist := rl.Vector3Distance(g.Camera.Position, head)
			d.check(g.Ride.First && dist < 0.5, "atractii: C = camera din ochii aventurierului (la %.2f m de cap)", dist)
			d.cQueued = true
			d.eQueued = true
			d.next()
		}
	case 26: // masina: urca, conduce prin Marele Tunel de la un capat la altul
		if d.stepT < 2 {
			return false
		}
		cw := g.City
		gtr := cw.Routes[len(cw.Routes)-1]
		k := -1
		for i := range cw.Cars {
			c := &cw.Cars[i]
			if c.Route == gtr {
				if k < 0 {
					k = i
				} else {
					c.Dead = 9999 // golim tunelul, ca masina aventurierului sa aiba drum liber
				}
			}
		}
		c := &cw.Cars[k]
		c.Parked = true
		c.Pos = v3(-396, 0, 2.2)
		c.Yaw = math32Pi / 2
		c.Dir = v3(1, 0, 0)
		c.Pitch = 0
		d.teleport(g, -396, -0.6)
		d.carK = k
		d.minY2 = 1e9
		d.eQueued = true
		d.next()
	case 27:
		if d.stepT > 0.5 && d.stepInt == 0 {
			d.check(g.Drive.Active, "masina: E il urca la volan")
			d.stepInt = 1
		}
		if g.Drive.Active {
			c := &g.City.Cars[g.Drive.Car]
			d.driveOn = true
			d.driveTh = 1
			d.driveSt = clampf((c.Pos.Z-2.2)*0.4+float32(math.Remainder(float64(c.Yaw-math32Pi/2), 2*math.Pi))*1.5, -1, 1)
			d.minY2 = minf(d.minY2, c.Pos.Y)
			if c.Pos.X > 386 || d.stepT > 60 {
				d.check(c.Pos.X > 380, "masina: a condus prin Marele Tunel pana la x=%.0f (in %.0fs, viteza %.0f km/h)", c.Pos.X, d.stepT, absf(g.Drive.Speed)*3.6)
				d.check(d.minY2 < -14, "masina: a coborat sub mare pana la %.1f m", d.minY2)
				d.check(g.Drive.GTDist > 600, "masina: %.0f m prin tunel", g.Drive.GTDist)
				d.driveOn = false
				d.eQueued = true
				d.next()
			}
		} else if d.stepT > 2 {
			d.check(false, "masina: nu e la volan")
			d.next()
		}
	case 28:
		if d.stepT > 1 {
			d.check(!g.Drive.Active && p.OnGround, "masina: E il coboara langa masina (y=%.2f)", p.Pos.Y)
			for i := range g.City.Cars {
				if g.City.Cars[i].Dead > 100 {
					g.City.Cars[i].Dead = 0.01
				}
			}
			g.Enemies = g.Enemies[:0] // V 5.0: monstrii nivelului 1 stau in piata si ar lua ei loviturile
			// masina de tinta in piata
			c := &g.City.Cars[d.carK]
			c.Parked = true
			c.HP = CarMaxHP
			c.Pos = v3(0, 0, 27.5)
			c.Yaw = math32Pi / 2
			c.Pitch = 0
			d.teleport(g, 0, 37)
			g.CamYaw = 0
			g.CamPitch = 0.08
			d.gold0 = float32(g.Player.Gold)
			d.CarsBoomed = 0
			d.next()
		}
	case 29:
		d.fireOn = d.stepT > 0.5 && d.CarsBoomed == 0
		if d.stepT > 5 {
			got := float32(g.Player.Gold) - d.gold0
			d.check(d.CarsBoomed == 1, "masina: a explodat dupa impuscaturi (%d explozii, aimcar %d)", d.CarsBoomed, g.AimCar)
			d.check(got >= float32(3*CarHitGold+CarBoomGold), "masina: a dat %.0f aur", got)
			d.next()
		}
	case 30:
		if d.stepT > CarRespawn+2 {
			c := &g.City.Cars[d.carK]
			d.check(c.Dead <= 0 && c.HP == CarMaxHP && !c.Parked, "masina: a reaparut in trafic dupa explozie (dead %.1f hp %.0f parked %v)", c.Dead, c.HP, c.Parked)
			d.next()
		}
	case 31: // pila de pod in mare: cade pe ea, sta pe ea, iese in apa, se urca inapoi
		k := d.findPier(g)
		if k < 0 {
			d.check(false, "pila: nu am gasit nicio pila de pod in mare")
			d.step = 34
			return false
		}
		c := w.Circles[k]
		d.carK = k
		d.teleport(g, c.X, c.Z)
		p.Pos.Y = c.Top + 4
		p.OnGround = false
		d.next()
	case 32:
		c := w.Circles[d.carK]
		if d.stepT > 2 && d.stepInt == 0 {
			d.check(p.OnGround && !g.Swimming && absf(p.Pos.Y-c.Top) < 0.25, "pila: a cazut pe pila si sta pe ea (y=%.2f, varful %.2f), nu inoata", p.Pos.Y, c.Top)
			// iese spre larg
			dx, dz := p.Pos.X-c.X, p.Pos.Z-c.Z
			if dx*dx+dz*dz < 0.01 {
				dx = 1
			}
			l := sqrtf(dx*dx + dz*dz)
			d.moveDir, d.moveOn = v3(dx/l, 0, dz/l), true
			d.stepInt = 1
			d.stepVal = d.stepT
		}
		if d.stepInt == 1 && d.stepT > d.stepVal+1.6 {
			d.check(g.Swimming, "pila: a sarit de pe pila in apa si inoata")
			d.moveDir = rl.Vector3Negate(d.moveDir)
			d.stepInt = 2
			d.stepVal = d.stepT
		}
		if d.stepInt == 2 && d.stepT > d.stepVal+4 {
			d.check(!g.Swimming && absf(p.Pos.Y-c.Top) < 0.3, "pila: inotand spre pila, s-a urcat pe ea (y=%.2f)", p.Pos.Y)
			d.moveOn = false
			d.next()
		}
	case 33: // blocat inauntrul pilei (cum a patit Augustin): trebuie sa poata iesi
		c := w.Circles[d.carK]
		d.teleport(g, c.X+c.R*0.3, c.Z)
		g.Swimming = true
		p.Pos.Y = WaterY - 1.25
		d.moveDir, d.moveOn = v3(1, 0, 0), true
		d.next()
	case 34:
		if d.stepT > 3 {
			c := w.Circles[d.carK]
			dx, dz := p.Pos.X-c.X, p.Pos.Z-c.Z
			d.check(sqrtf(dx*dx+dz*dz) > c.R+0.3, "pila: din interiorul pilei a iesit inotand (la %.1f m de centru, raza %.1f)", sqrtf(dx*dx+dz*dz), c.R)
			d.moveOn = false
			d.next()
		}
	case 35: // dupa masina: cand inoata, nu mai sta asezat
		k := g.nearestCarIdx(v3(0, 0, 40))
		c := &g.City.Cars[k]
		c.Parked = true
		d.teleport(g, c.Pos.X, c.Pos.Z)
		g.enterCar(k)
		d.next()
	case 36:
		if d.stepT > 1 && d.stepInt == 0 {
			d.eQueued = true
			d.stepInt = 1
		}
		if d.stepT > 2 && d.stepInt == 1 {
			d.teleport(g, -200, 40)
			g.enterWater()
			d.moveDir, d.moveOn = v3(1, 0, 0), true
			d.stepInt = 2
		}
		if d.stepT > 4 {
			d.check(g.Swimming && !p.Pose.Sit && !g.Drive.Active, "masina -> apa: Aventurierul inoata, nu mai sta asezat (sit=%v)", p.Pose.Sit)
			d.moveOn = false
			d.next()
		}
	case 37: // traficul nu se goleste: aruncam in aer 14 masini, dupa reaparitie trebuie sa fie toate
		d.teleport(g, 0, 40)
		n := 0
		for k := range g.City.Cars {
			if n < 14 && g.City.Cars[k].Hittable() {
				g.explodeCar(k)
				n++
			}
		}
		d.next()
	case 38:
		if d.stepT > CarRespawn+3 {
			alive, near := 0, 0
			for k := range g.City.Cars {
				c := &g.City.Cars[k]
				if c.Hittable() {
					alive++
					if rl.Vector3Distance(c.Pos, p.Pos) < 160 {
						near++
					}
				}
			}
			stacked := 0
			for i := range g.City.Cars {
				for j := i + 1; j < len(g.City.Cars); j++ {
					a, b := &g.City.Cars[i], &g.City.Cars[j]
					if a.Hittable() && b.Hittable() && rl.Vector3Distance(a.Pos, b.Pos) < 3 {
						stacked++
						d.logf("  suprapuse: %d (ruta %s, S %.0f, v %.1f) si %d (ruta %s, S %.0f, v %.1f)", i, a.Route.Name, a.S, a.Speed, j, b.Route.Name, b.S, b.Speed)
					}
				}
			}
			d.check(alive == len(g.City.Cars), "trafic: toate masinile au reaparut (%d / %d)", alive, len(g.City.Cars))
			d.check(near >= 6, "trafic: %d masini la mai putin de 160 m de Aventurierul", near)
			d.check(stacked == 0, "trafic: nicio masina suprapusa peste alta (%d)", stacked)
			d.next()
		}
	case 39: // avionul de jos: tinteste si trage pana il nimereste
		d.teleport(g, 0, 40)
		d.ThingsHit = [4]int{}
		d.gold0 = float32(g.Player.Gold)
		d.next()
	case 40:
		if d.aimThing(g, TgtPlane, 0) {
			d.fireOn = true
		}
		if d.ThingsHit[TgtPlane] > 0 || d.stepT > 40 {
			got := float32(g.Player.Gold) - d.gold0
			d.check(d.ThingsHit[TgtPlane] > 0, "avion: laserul a ajuns la avion si l-a doborat (%.0f s, +%.0f aur)", d.stepT, got)
			d.check(g.City.Chute.Active, "avion: pilotul coboara cu parasuta")
			d.fireOn = false
			d.next()
		}
	case 41: // un pescarus in zbor
		d.teleport(g, Sectors[0].Center.X, Sectors[0].Center.Y+30)
		d.next()
	case 42:
		if d.aimThing(g, TgtBird, -1) {
			d.fireOn = true
		}
		if d.ThingsHit[TgtBird] > 0 || d.stepT > 25 {
			d.check(d.ThingsHit[TgtBird] > 0, "pescarus: nimerit in zbor (%.0f s)", d.stepT)
			d.fireOn = false
			d.next()
		}
	case 43: // o barca: se scufunda dupa cateva lovituri (Aventurierul sta pe malul de langa ea)
		b := &g.City.Boats[0]
		found := false
		for r := b.RX + 4; r < b.RX+70 && !found; r += 3 {
			for k := 0; k < 24 && !found; k++ {
				a := float32(k) / 24 * 2 * math32Pi
				x, z := b.C.X+cosf(a)*r, b.C.Y+sinf(a)*r
				if w.NaturalLand(x, z) >= 0 && w.CanStand(x, z, 0.3, 0.45) {
					d.teleport(g, x, z)
					found = true
				}
			}
		}
		d.next()
	case 44:
		if d.aimThing(g, TgtBoat, -1) {
			d.fireOn = true
		}
		if d.ThingsHit[TgtBoat] > 0 || d.stepT > 30 {
			d.check(d.ThingsHit[TgtBoat] > 0, "barca: s-a scufundat dupa impuscaturi (%.0f s)", d.stepT)
			d.fireOn = false
			d.next()
		}
	case 45: // trambulina de langa bloc: sare pe acoperis
		pi, bi := d.cityPadWithRoof(g)
		if pi < 0 {
			d.check(false, "acoperis: nicio trambulina langa un bloc")
			d.step = 47
			return false
		}
		cnt := 0
		for k := range w.Pads {
			if _, ok := d.roofNear(g, k); ok {
				cnt++
			}
		}
		d.check(cnt >= 8, "acoperis: %d trambuline lipite de blocuri, fiecare sare mai sus de acoperis", cnt)
		pd := &w.Pads[pi]
		d.carK = bi
		d.stepVal = float32(pi)
		d.teleport(g, pd.Pos.X, pd.Pos.Z)
		g.mountPad(pi)
		d.jumpOn = true
		d.next()
	case 46:
		b := &w.Boxes[d.carK]
		top := b.SurfaceY((b.MinX+b.MaxX)/2, (b.MinZ+b.MaxZ)/2)
		cx, cz := (b.MinX+b.MaxX)/2, (b.MinZ+b.MaxZ)/2
		if g.Ride.Kind == RideTrampoline && p.Pos.Y > top+1.5 {
			dx, dz := cx-p.Pos.X, cz-p.Pos.Z
			l := sqrtf(dx*dx+dz*dz) + 0.001
			d.moveDir, d.moveOn = v3(dx/l, 0, dz/l), true
		}
		_, kind := w.GroundAt(p.Pos.X, p.Pos.Z, p.Pos.Y+0.2)
		if (p.OnGround && kind == SurfRoof && p.Pos.Y > top-1) || d.stepT > 30 {
			d.check(p.OnGround && kind == SurfRoof && p.Pos.Y > top-1, "acoperis: de pe trambulina a aterizat pe bloc (y=%.1f, acoperis %.1f, %.0f s)", p.Pos.Y, top, d.stepT)
			d.moveOn, d.jumpOn = false, false
			d.next()
		}
	default:
		return d.funStep5(g)
	}
	return false
}

// AfterFrame: true cand testul s-a terminat
func (d *DebugRun) AfterFrame(g *Game) bool {
	d.Frame++
	if d.Frame > 5 {
		d.fpsSum += float32(rl.GetFPS())
		d.fpsN++
	}
	switch d.Mode {
	case "shots":
		d.ViewF++
		wait := d.Views[d.ViewIdx].Wait
		if wait == 0 {
			wait = 8
		}
		if d.ViewF == wait {
			name := fmt.Sprintf("%s/%s.png", d.OutDir, d.Views[d.ViewIdx].Name)
			rl.TakeScreenshot(name)
			fmt.Println("captura:", name)
			d.ViewIdx++
			d.ViewF = 0
			if d.ViewIdx >= len(d.Views) {
				return true
			}
			d.applyView(g)
		}
	case "fight":
		if d.lvl0 == 0 {
			d.lvl0 = g.Level
		}
		if g.Level > d.lvl0 {
			d.lvl0 = g.Level
			d.Levels++
			fmt.Printf("NIVEL TERMINAT -> acum nivelul %d (aur %d, timp %.0fs)\n", g.Level, g.Player.Gold, g.Time)
		}
		if g.Scene == SceneGameOver {
			d.Deaths++
			g.Player.HP = g.Player.MaxHP
			g.ResetPlayerPosition()
			g.Scene = SceneMapExplore
		}
		if d.Levels >= 2 || g.Time > 900 {
			fmt.Printf("REZULTAT LUPTA: niveluri terminate %d, inamici invinsi %d, lovituri primite %d, infrangeri %d, nivel curent %d, timp %.0fs, FPS %.1f\n", d.Levels, d.Kills, d.Hits, d.Deaths, g.Level, g.Time, d.fpsSum/float32(maxInt(d.fpsN, 1)))
			return true
		}
	case "smoke":
		if os.Getenv("BATTLE_DEBUG_SNAP") != "" && d.Frame%450 == 0 {
			name := fmt.Sprintf("%s/snap_%04d_x%.0f_y%.0f.png", d.OutDir, d.Frame/450, g.Player.Pos.X, g.Player.Pos.Y)
			g.Draw()
			rl.TakeScreenshot(name)
		}
		if g.Player.Pos.Y > d.MaxY {
			d.MaxY = g.Player.Pos.Y
		}
		if g.Player.Pos.Y < d.MinY {
			d.MinY = g.Player.Pos.Y
		}
		if d.WayIdx >= len(d.Way) || g.Time > d.MaxTime {
			fmt.Printf("Inaltimea maxima atinsa: %.2f m, cea mai adanca: %.2f m (sub mare)\n", d.MaxY, d.MinY)
			fmt.Printf("REZULTAT: %d/%d puncte, %d blocaje, FPS mediu %.1f, timp %.0fs\n", d.WayIdx, len(d.Way), d.Failures, d.fpsSum/float32(maxInt(d.fpsN, 1)), g.Time)
			return true
		}
	case "fun":
		if g.Scene == SceneGameOver {
			g.Scene = SceneMapExplore
		}
		if d.funStep(g) || g.Time > d.MaxTime {
			fmt.Printf("REZULTAT DISTRACTII: %d verificari gresite, timp %.0fs, FPS %.1f\n", d.Failures, g.Time, d.fpsSum/float32(maxInt(d.fpsN, 1)))
			return true
		}
	}
	return false
}

// findPier: o pila de pod in mare (varful putin deasupra apei).
func (d *DebugRun) findPier(g *Game) int {
	w := g.City.W
	for k, c := range w.Circles {
		if c.R >= 1.5 && c.Top > WaterY && c.Top < WaterY+1.6 && w.NaturalLand(c.X, c.Z) < SeabedY+0.5 {
			return k
		}
	}
	return -1
}

func (g *Game) nearestCarIdx(p rl.Vector3) int {
	best, bd := 0, float32(1e9)
	for k := range g.City.Cars {
		c := &g.City.Cars[k]
		if !c.Hittable() || c.Route == g.City.Routes[len(g.City.Routes)-1] {
			continue
		}
		if dd := rl.Vector3Distance(c.Pos, p); dd < bd {
			best, bd = k, dd
		}
	}
	return best
}

// aimThing: indreapta camera spre tinta (idx < 0 = cea mai apropiata de felul asta).
// true cand tinta asistata a prins-o.
func (d *DebugRun) aimThing(g *Game, kind, idx int) bool {
	cw := g.City
	p := g.Player.Pos
	head := v3(p.X, p.Y+2.2, p.Z)
	if idx < 0 {
		bd := float32(1e9)
		n := map[int]int{TgtBoat: len(cw.Boats), TgtBird: len(cw.Birds), TgtPlane: len(cw.Planes)}[kind]
		for i := 0; i < n; i++ {
			if c, ok := g.thingCenter(kind, i); ok {
				if dd := rl.Vector3Distance(c, head); dd < bd && dd > 8 {
					bd, idx = dd, i
				}
			}
		}
		if idx < 0 {
			return false
		}
	}
	c, ok := g.thingCenter(kind, idx)
	if !ok {
		return false
	}
	dir := rl.Vector3Normalize(rl.Vector3Subtract(c, head))
	g.CamPitch = clampf(float32(math.Asin(float64(-dir.Y))), -0.75, 1.3)
	cp := cosf(g.CamPitch)
	if cp < 0.01 {
		cp = 0.01
	}
	g.CamYaw = atan2f(-dir.X, -dir.Z)
	return g.AimKind == kind
}

// roofNear: blocul lipit de trambulina k, daca trambulina il arunca mai sus de acoperis.
func (d *DebugRun) roofNear(g *Game, k int) (int, bool) {
	w := g.City.W
	pd := &w.Pads[k]
	for bi := range w.Boxes {
		b := &w.Boxes[bi]
		if b.Roof == RoofNone || b.Top < 5 {
			continue
		}
		dx := maxf(maxf(b.MinX-pd.Pos.X, pd.Pos.X-b.MaxX), 0)
		dz := maxf(maxf(b.MinZ-pd.Pos.Z, pd.Pos.Z-b.MaxZ), 0)
		gap := sqrtf(dx*dx+dz*dz) - pd.R
		top := b.SurfaceY((b.MinX+b.MaxX)/2, (b.MinZ+b.MaxZ)/2)
		if gap < 2.5 && pd.MaxApex+pd.MatY() > top+2 {
			return bi, true
		}
	}
	return -1, false
}

// cityPadWithRoof: trambulina langa cel mai inalt bloc.
func (d *DebugRun) cityPadWithRoof(g *Game) (int, int) {
	w := g.City.W
	bp, bb, bt := -1, -1, float32(0)
	for k := range w.Pads {
		if bi, ok := d.roofNear(g, k); ok {
			if t := w.Boxes[bi].Top; t > bt {
				bp, bb, bt = k, bi, t
			}
		}
	}
	return bp, bb
}
