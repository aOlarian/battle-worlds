package main

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type VendorMachine struct {
	Position  rl.Vector3
	SlotIndex int
}

type Citizen struct {
	Pos      rl.Vector3
	Yaw      float32
	Speed    float32
	Style    int
	Pose     Pose
	Pause    float32
	Greet    float32
	Home     rl.Vector2
	HomeR    float32
	TurnLeft float32
}

type Cloud struct {
	Pos  rl.Vector3
	Mesh int
	S    float32
}

type CityWorld struct {
	W      *World
	Chunks []Chunk
	Water  rl.Mesh
	Roads  []RoadSeg

	LegoSites  []ConstructionSite
	VendorPods []VendorMachine

	Citizens      []Citizen
	CitizenMeshes []*CharMeshes

	Routes     []*Route
	Cars       []Car
	CarMeshes  []rl.Mesh
	Boats      []Boat
	BoatMeshes []rl.Mesh

	PlaneMesh rl.Mesh
	Planes    []Plane
	Flocks    []Flock
	Birds     []Bird
	BirdBody  rl.Mesh
	BirdWingL rl.Mesh
	BirdWingR rl.Mesh
	ChuteMesh rl.Mesh
	Chute     Parachute

	Clouds      []Cloud
	CloudMeshes []rl.Mesh

	Booths  []Booth
	Wheels  []FerrisWheel
	Coaster *Coaster

	TurbineTower, TurbineRotor rl.Mesh
	Turbines                   []rl.Vector3
	TurbineAng                 float32

	RaceMeshes []rl.Mesh
	RaceCenter rl.Vector2
	RaceT      [2]float32

	Fountain rl.Vector3
	Rng      *rand.Rand

	// V 6.0
	Realm   int         // indexul lumii mari (0 = insulele)
	Zones   []RealmZone // zonele (lumile 1+)
	Station rl.Vector3  // gara trenului magic
	Nav     *NavGraph
}

func InitCityWorld(progress func(string)) *CityWorld {
	if progress == nil {
		progress = func(string) {}
	}
	w := NewWorldLayout()
	cb := &CityBuilder{W: w, Rng: rand.New(rand.NewSource(2026))}
	cw := &CityWorld{W: w, Rng: rand.New(rand.NewSource(7))}
	buildHammerMesh()
	buildWeaponMeshes()
	buildPadMat()
	buildBoothMesh()

	// --- santierele blocuri si tonomatele ---
	progress("Pregătim șantierele blocuri...")
	blueprints := GetLegoBlueprints()
	letters := GetArenaLetterBlueprints()
	addSite := func(pos rl.Vector3, vend rl.Vector3, bp LegoItemDefinition, yaw float32) {
		idx := len(cw.LegoSites)
		cw.LegoSites = append(cw.LegoSites, ConstructionSite{
			Position: pos, Blueprint: bp, Meshes: BuildBlueprintMeshes(bp),
			TotalStages: len(bp.Bricks), UpgradeLevel: 1, TargetScale: 1, CurrentScale: 1, Yaw: yaw, ColliderIdx: -1,
		})
		cw.VendorPods = append(cw.VendorPods, VendorMachine{Position: vend, SlotIndex: idx})
		cb.Reserve = append(cb.Reserve, CircleCollider{X: pos.X, Z: pos.Z, R: 15}, CircleCollider{X: vend.X, Z: vend.Z, R: 3})
	}
	for si, s := range Sectors[:4] {
		in := s.Inner
		quads := [][2]float32{{in.X, -in.Y}, {-in.X, in.Y}}
		for k, q := range quads {
			pos := v3(s.Center.X+q[0]*30, 0, s.Center.Y+q[1]*30)
			vend := v3(s.Center.X+q[0]*19.5, 0, s.Center.Y+q[1]*19.5)
			yaw := atan2f(-q[0], -q[1])
			addSite(pos, vend, blueprints[si*2+k], yaw)
		}
	}
	for i, bp := range letters {
		x := -32 + 16*float32(i)
		addSite(v3(x, 0, -14), v3(x, 0, -3), bp, 0)
	}
	cb.reserveCityPads()

	progress("Asfaltăm șoselele...")
	cb.buildRoads()
	cb.placeBooths(cw)
	// roata din Cartoon City (cu cabine in care urci)
	lp := cb.landmarkPos(1)
	cw.Wheels = append(cw.Wheels, newWheel("Roata din Cartoon City", v3(lp.X, 14.1, lp.Y), 11, 8, 0.85))
	cb.buildLandmarks(cw)
	for _, p := range turbinePositions() {
		cb.Reserve = append(cb.Reserve, CircleCollider{X: p.X, Z: p.Z, R: 4})
		w.Circles = append(w.Circles, CircleCollider{X: p.X, Z: p.Z, R: 1.1, Top: 30})
	}
	progress("Construim podurile...")
	cb.buildBridges()
	progress("Săpăm tunelurile prin munți...")
	cb.buildMountainsAndTunnels()
	progress("Săpăm Marele Tunel de sub Mare...")
	cb.buildCorridorMeshes()
	progress("Ridicăm clădirile...")
	cb.buildBuildings()
	cb.placeRoofPads()
	cb.buildPlaza()
	progress("Construim Parcul de Distracții...")
	nPads := len(w.Pads)
	cb.buildBeachIsland()
	cb.buildFunPark(cw)
	{
		mb := NewMB()
		for i := 0; i < nPads; i++ {
			cb.padFrame(mb, &w.Pads[i])
		}
		for i := nPads; i < nPads+2 && i < len(w.Pads); i++ {
			cb.padFrame(mb, &w.Pads[i]) // trambulinele de pe plaja
		}
		cb.addChunk(mb, true)
	}
	{
		mb := NewMB()
		plates := []rl.Color{rl.NewColor(120, 200, 90, 255), rl.NewColor(110, 170, 230, 255), rl.NewColor(240, 200, 90, 255), rl.NewColor(200, 200, 205, 255)}
		for i := range cw.LegoSites {
			s := &cw.LegoSites[i]
			c := plates[3]
			if i < 8 {
				c = plates[i/2%3]
			}
			cb.baseplate(mb, s.Position.X, s.Position.Z, c)
			w.Boxes = append(w.Boxes, BoxCollider{MinX: s.Position.X - 5, MaxX: s.Position.X + 5, MinZ: s.Position.Z - 5, MaxZ: s.Position.Z + 5, Top: 0.4})
		}
		for _, v := range cw.VendorPods {
			cb.vendorKiosk(mb, v.Position.X, v.Position.Z, LegoRed)
		}
		cb.addChunk(mb, true)
	}
	progress("Plantăm copacii...")
	cb.buildVegetationAndProps()
	progress("Decoram lumile...")
	cb.buildThemeDecor()
	progress("Modelăm insulele...")
	cb.buildTerrain()
	progress("Umplem marea...")
	cw.Water = cb.buildWater()
	cw.Chunks = cb.Chunks
	cw.Roads = cb.Roads
	w.Roads = cb.Roads
	cw.finishBooths()
	w.RebuildGrid()
	for i := range cw.LegoSites {
		s := &cw.LegoSites[i]
		s.ColliderIdx = w.AddCircle(s.Position.X, s.Position.Z, 0, 0)
	}

	progress("Pornim traficul...")
	carCols := []rl.Color{LegoRed, LegoYellow, LegoBlue, LegoLime, LegoOrange, LegoCyan, LegoWhite, LegoPink}
	for _, c := range carCols {
		cw.CarMeshes = append(cw.CarMeshes, buildCarMesh(c, false))
	}
	cw.Routes = []*Route{w.BuildRingRoute(false), w.BuildRingRoute(true), w.BuildTunnelRoute(false), w.BuildTunnelRoute(true), w.BuildGreatTunnelRoute()}
	counts := []int{10, 10, 6, 6, 10}
	k := 0
	for ri, r := range cw.Routes {
		for i := 0; i < counts[ri]; i++ {
			spd := 10 + cw.Rng.Float32()*3
			if ri == 4 {
				spd = 14 + cw.Rng.Float32()*3
			}
			cw.Cars = append(cw.Cars, Car{Route: r, S: r.Len * float32(i) / float32(counts[ri]), Speed: spd, MaxSpd: spd, Mesh: k % len(cw.CarMeshes), HP: CarMaxHP})
			k++
		}
	}
	// --- barci ---
	boatCols := []rl.Color{LegoRed, LegoBlue, LegoYellow, LegoGreen, LegoOrange}
	for _, c := range boatCols {
		cw.BoatMeshes = append(cw.BoatMeshes, buildBoatMesh(c))
	}
	loops := []struct {
		c      rl.Vector2
		rx, rz float32
	}{
		{rl.NewVector2(0, -168), 34, 42}, {rl.NewVector2(0, 168), 34, 42},
		{rl.NewVector2(-168, 0), 42, 34}, {rl.NewVector2(168, 0), 42, 34},
		{rl.NewVector2(0, 0), 275, 275}, {rl.NewVector2(0, 0), 300, 290},
		{rl.NewVector2(-356, 0), 95, 110}, {rl.NewVector2(362, 0), 100, 120},
	}
	for i, l := range loops {
		n := 2
		if i == 4 || i == 5 {
			n = 3
		}
		for j := 0; j < n; j++ {
			sp := 5 + cw.Rng.Float32()*2
			if i%2 == 1 {
				sp = -sp
			}
			cw.Boats = append(cw.Boats, Boat{C: l.c, RX: l.rx, RZ: l.rz, Phase: float32(j) / float32(n) * 2 * math.Pi, Speed: sp, Mesh: (i + j) % len(cw.BoatMeshes), HP: BoatMaxHP})
		}
	}
	// --- cetateni (si pe insulele noi) ---
	for i := 0; i < 12; i++ {
		st := CitizenStyle(i)
		if rs := RealmCharStyle(0, RoleNPC1+i%3); realmChars[rs.Costume] != nil {
			st = rs // V 6.0: locuitorii Lumii Marii (turisti, salvamari, inghetata)
		}
		cw.CitizenMeshes = append(cw.CitizenMeshes, GetCharMeshes(st))
	}
	for i := 0; i < 62; i++ {
		si := i % 7
		var home rl.Vector2
		var hr float32
		switch si {
		case 4:
			home, hr = rl.NewVector2(0, 8), 34
		case SecWest:
			home, hr = rl.NewVector2(WestX+10, 0), 48
		case SecPark:
			home, hr = rl.NewVector2(EastX-5, -10), 55
		default:
			home, hr = Sectors[si].Center, 52
		}
		var pos rl.Vector3
		for tries := 0; tries < 80; tries++ {
			x := home.X + (cw.Rng.Float32()*2-1)*hr
			z := home.Y + (cw.Rng.Float32()*2-1)*hr
			if w.CanStand(x, z, 0.3, 0.4) {
				g, _ := w.GroundAt(x, z, 0.5)
				if absf(g) < 0.6 {
					pos = v3(x, g, z)
					break
				}
			}
		}
		cw.Citizens = append(cw.Citizens, Citizen{Pos: pos, Yaw: cw.Rng.Float32() * 6.28, Speed: 1.6 + cw.Rng.Float32()*0.6,
			Style: i % len(cw.CitizenMeshes), Home: home, HomeR: hr, Pose: Pose{Walk: cw.Rng.Float32() * 6}})
	}
	// --- cer: avion, nori ---
	cw.PlaneMesh = buildPlaneMesh()
	cw.buildSkyTargets()
	for i := 0; i < 6; i++ {
		cw.CloudMeshes = append(cw.CloudMeshes, buildCloudMesh(i*13+5))
	}
	for i := 0; i < 22; i++ {
		cw.Clouds = append(cw.Clouds, Cloud{Pos: v3((cw.Rng.Float32()*2-1)*560, 92+cw.Rng.Float32()*38, (cw.Rng.Float32()*2-1)*480), Mesh: i % 6, S: 0.9 + cw.Rng.Float32()*0.8})
	}
	cw.TurbineTower, cw.TurbineRotor = buildTurbine()
	cw.Turbines = turbinePositions()
	cw.RaceMeshes = []rl.Mesh{buildCarMesh(LegoRed, true), buildCarMesh(LegoBlue, true)}
	cw.RaceCenter = cb.landmarkPos(2)
	cw.RaceT = [2]float32{0, 3}
	cw.Fountain = v3(0, 0, 20)
	// V 6.0: gara trenului magic (lumea 0)
	cw.Station = islandStationSpot(w)
	{
		mb := NewMB()
		buildStationMesh(mb, cw.Station, &Realms[0])
		stationColliders(w, cw.Station)
		cb.Chunks = cw.Chunks
		cb.addChunk(mb, true)
		cw.Chunks = cb.Chunks
		w.RebuildGrid()
	}
	return cw
}

// onRoad: pe asfalt (pentru sunetul pasilor)
func (w *World) onRoad(x, z float32) bool {
	for _, r := range w.Roads {
		if r.HalfW < 0 {
			if d := sqrtf(x*x + z*z); absf(d-48) < 4.5 {
				return true
			}
			continue
		}
		if distToSeg(x, z, r.A, r.B) < r.HalfW {
			return true
		}
	}
	return false
}

// PadAt: trambulina de sub (x,z)
func (w *World) PadAt(x, z float32) int {
	for i := range w.Pads {
		p := &w.Pads[i]
		dx, dz := x-p.Pos.X, z-p.Pos.Z
		if dx*dx+dz*dz < p.R*p.R {
			return i
		}
	}
	return -1
}

func (cw *CityWorld) Update(dt, t float32, playerPos, listener rl.Vector3, audio *GameAudio) {
	w := cw.W
	// santiere
	for i := range cw.LegoSites {
		s := &cw.LegoSites[i]
		s.Update(dt)
		if s.HammerActive && s.HammerStrikesLeft > 0 && s.HammerTime > 0.22 {
			s.HammerTime = 0
			s.HammerStrikesLeft--
			if audio != nil {
				audio.PlayAt(audio.Build, s.Position, listener, 0.8, 40)
			}
			if s.HammerStrikesLeft <= 0 {
				s.HammerActive = false
			}
		}
		if s.ColliderIdx >= 0 {
			c := &w.Circles[s.ColliderIdx]
			if s.CurrentStage > 0 {
				c.R = s.Meshes.Radius * s.Scale() * 0.75
				c.Top = 50
			} else {
				c.R = 0
				c.Top = 0
			}
		}
	}
	// cetateni
	for i := range cw.Citizens {
		c := &cw.Citizens[i]
		c.Pose.Breathe += dt * 2
		if c.Greet > 0 {
			c.Greet -= dt
		}
		if c.Pose.Wave > 0 {
			c.Pose.Wave -= dt
			if c.Pose.Wave < 0 {
				c.Pose.Wave = 0
			}
		}
		dx := playerPos.X - c.Pos.X
		dz := playerPos.Z - c.Pos.Z
		d2 := dx*dx + dz*dz
		if d2 < 25 && c.Greet <= 0 && absf(playerPos.Y-c.Pos.Y) < 2 {
			c.Greet = 12
			c.Pose.Wave = 1.6
			c.Pause = 1.6
			c.Yaw = atan2f(dx, dz)
			if audio != nil {
				audio.PlayAt(audio.Greet, c.Pos, listener, 0.8, 20)
			}
		}
		if c.Pause > 0 {
			c.Pause -= dt
			c.Pose.Moving = maxf(c.Pose.Moving-dt*4, 0)
			continue
		}
		if cw.Rng.Float32() < 0.004 {
			c.Pause = 1 + cw.Rng.Float32()*3
		}
		if cw.Rng.Float32() < 0.01 {
			c.TurnLeft = (cw.Rng.Float32()*2 - 1) * 1.2
		}
		turn := clampf(c.TurnLeft, -1, 1) * dt
		c.Yaw += turn
		c.TurnLeft -= turn
		hx := c.Home.X - c.Pos.X
		hz := c.Home.Y - c.Pos.Z
		if hx*hx+hz*hz > c.HomeR*c.HomeR {
			c.Yaw = atan2f(hx, hz)
		}
		step := c.Speed * dt
		mx := sinf(c.Yaw) * step
		mz := cosf(c.Yaw) * step
		if w.Move(&c.Pos, mx, mz, 0.4) {
			c.Pose.Moving = minf(c.Pose.Moving+dt*4, 1)
			c.Pose.Walk += dt * 7
		} else {
			c.Yaw += math.Pi/2 + cw.Rng.Float32()*math.Pi
		}
		g, _ := w.GroundAt(c.Pos.X, c.Pos.Z, c.Pos.Y)
		c.Pos.Y += (g - c.Pos.Y) * minf(dt*12, 1)
	}
	// masini (urmaresc ruta, opresc pentru Aventurierul si pentru masina din fata)
	for i := range cw.Cars {
		car := &cw.Cars[i]
		if car.Driven || car.Parked {
			continue // o conduce Aventurierul sau sta parcata unde a lasat-o
		}
		if car.Dead > 0 {
			car.Dead -= dt
			if car.Dead <= 0 {
				// reapare pe drumul ei, nu prea departe de Aventurierul (ca traficul din jurul lui sa
				// nu se goleasca), dar in afara privirii lui si nu peste alta masina
				car.S = cw.respawnS(i, playerPos, listener)
				car.HP = CarMaxHP
				car.Speed = 0
				car.Pos, car.Dir = car.Route.Sample(car.S)
			}
			continue
		}
		target := car.MaxSpd
		ahead := rl.Vector3Add(car.Pos, rl.Vector3Scale(car.Dir, 6))
		if rl.Vector3Distance(ahead, playerPos) < 4.2 && absf(playerPos.Y-car.Pos.Y) < 3 {
			target = 0
			if car.Honk <= 0 && audio != nil {
				audio.PlayAt(audio.CarHonk, car.Pos, listener, 1, 60)
				car.Honk = 4
			}
		}
		for j := range cw.Cars {
			o := &cw.Cars[j]
			if j == i || o.Route != car.Route || o.Dead > 0 || o.Driven || o.Parked {
				continue
			}
			gap := o.S - car.S
			if gap < 0 {
				gap += car.Route.Len
			}
			if gap < 11 {
				target = minf(target, maxf(gap-6, 0)*1.5)
			}
		}
		car.Speed += (target - car.Speed) * minf(dt*2.5, 1)
		car.S += car.Speed * dt
		if car.S > car.Route.Len {
			car.S -= car.Route.Len
		}
		p, d := car.Route.Sample(car.S)
		pf, _ := car.Route.Sample(car.S + 2.5)
		pb, _ := car.Route.Sample(car.S - 2.5)
		yawT := atan2f(pf.X-pb.X, pf.Z-pb.Z)
		dyaw := float32(math.Remainder(float64(yawT-car.Yaw), 2*math.Pi))
		car.Yaw += dyaw * minf(dt*8, 1)
		car.Pitch = -atan2f(pf.Y-pb.Y, 5)
		// sunetul masinii care trece pe langa Aventurierul
		if audio != nil && car.Speed > 5 {
			dNow := rl.Vector3Distance(p, playerPos)
			if dNow < 9 && car.Whoosh <= 0 {
				audio.PlayAt(audio.CarWhoosh, p, listener, 0.9, 30)
				car.Whoosh = 3
			}
		}
		if car.Whoosh > 0 {
			car.Whoosh -= dt
		}
		car.Pos = p
		car.Dir = d
		if car.Honk > 0 {
			car.Honk -= dt
		}
	}
	for i := range cw.Boats {
		cw.Boats[i].Update(dt, t)
	}
	cw.updateSkyTargets(dt, t, playerPos, listener, audio)
	for i := range cw.Clouds {
		c := &cw.Clouds[i]
		c.Pos.X += dt * 2.2
		if c.Pos.X > 600 {
			c.Pos.X = -600
		}
	}
	for i := range cw.Wheels {
		fw := &cw.Wheels[i]
		fw.Ang += dt * fw.Speed
	}
	if cw.Coaster != nil {
		cw.Coaster.Update(dt, audio, listener)
	}
	for i := range w.Pads {
		p := &w.Pads[i]
		if p.Dip > 0 {
			p.Dip = maxf(p.Dip-dt*2.5, 0)
		}
	}
	cw.TurbineAng += dt * 1.6
	cw.RaceT[0] += dt * 1.15
	cw.RaceT[1] += dt * 1.05
}

func (cw *CityWorld) racePose(i int) (rl.Vector3, float32) {
	t := cw.RaceT[i]
	ax, az := float32(16)+float32(i)*1.6-0.8, float32(10)+float32(i)*1.6-0.8
	x := cw.RaceCenter.X + cosf(t)*ax
	z := cw.RaceCenter.Y + sinf(t)*az
	dx := -sinf(t) * ax
	dz := cosf(t) * az
	return v3(x, 0.08, z), atan2f(dx, dz)
}

// DrawOpaque deseneaza tot ce e opac. shadow = pasul de umbre (fara obiecte mici departe).
func (cw *CityWorld) DrawOpaque(r *Renderer, draw DrawFn, t float32, shadow bool, focus rl.Vector3) {
	for i := range cw.Chunks {
		ch := &cw.Chunks[i]
		if shadow {
			if !ch.Shadows || !r.InShadowRange(ch.Center, ch.Radius, focus) {
				continue
			}
		} else if !r.SphereVisible(ch.Center, ch.Radius) {
			continue
		}
		draw(ch.Mesh, rl.MatrixIdentity(), rl.White)
	}
	near := func(p rl.Vector3, rad float32) bool {
		if shadow {
			dx, dz := p.X-focus.X, p.Z-focus.Z
			lim := r.ShadowRange*1.2 + rad
			return dx*dx+dz*dz < lim*lim
		}
		if !r.SphereVisible(p, rad) {
			return false
		}
		dx, dz := p.X-r.camPos.X, p.Z-r.camPos.Z
		return dx*dx+dz*dz < 300*300
	}
	for i := range cw.LegoSites {
		s := &cw.LegoSites[i]
		if near(s.Position, s.Meshes.Radius*s.Scale()+s.Meshes.Height*s.Scale()) {
			s.Draw(draw, false, t)
			s.DrawHammer(draw)
		}
	}
	for i := range cw.Citizens {
		c := &cw.Citizens[i]
		if near(c.Pos, 2) {
			DrawCharacter(draw, cw.CitizenMeshes[c.Style], c.Pos, c.Yaw, 0.85, c.Pose, rl.White)
		}
	}
	for i := range cw.Cars {
		c := &cw.Cars[i]
		if c.Dead > 0 || c.Hidden {
			continue
		}
		if near(c.Pos, 3) {
			m := rl.MatrixMultiply(rl.MatrixRotateX(c.Pitch), rl.MatrixRotateY(c.Yaw))
			m = rl.MatrixMultiply(m, rl.MatrixTranslate(c.Pos.X, c.Pos.Y, c.Pos.Z))
			tint := rl.White
			if c.HP <= CarMaxHP*0.5 {
				tint = rl.NewColor(150, 140, 135, 255) // lovita: se innegreste
			}
			if !shadow && c.Flash > 0 {
				r.SetFlash(0.5)
			}
			draw(cw.CarMeshes[c.Mesh], m, tint)
			if !shadow && c.Flash > 0 {
				r.SetFlash(0)
			}
		}
	}
	for i := 0; i < len(cw.RaceMeshes); i++ {
		p, yaw := cw.racePose(i)
		if near(p, 3) {
			draw(cw.RaceMeshes[i], rl.MatrixMultiply(rl.MatrixRotateY(yaw), rl.MatrixTranslate(p.X, p.Y, p.Z)), rl.White)
		}
	}
	for i := range cw.Boats {
		b := &cw.Boats[i]
		sink := float32(0)
		if b.Dead > 0 {
			// se scufunda incet (cu prova in sus), apoi dispare pana reapare
			sink = BoatRespawn - b.Dead
			if sink > boatSinkTime {
				continue
			}
		}
		if near(b.Pos, 5) {
			m := rl.MatrixMultiply(rl.MatrixRotateX(-sink*0.25), rl.MatrixRotateZ(b.Roll+sink*0.12))
			m = rl.MatrixMultiply(m, rl.MatrixRotateY(b.Yaw))
			m = rl.MatrixMultiply(m, rl.MatrixTranslate(b.Pos.X, b.Pos.Y-sink*1.4, b.Pos.Z))
			tint := rl.White
			if b.HP <= BoatMaxHP*0.5 && b.Dead <= 0 {
				tint = rl.NewColor(160, 150, 145, 255)
			}
			if !shadow && b.Flash > 0 {
				r.SetFlash(0.5)
			}
			draw(cw.BoatMeshes[b.Mesh], m, tint)
			if !shadow && b.Flash > 0 {
				r.SetFlash(0)
			}
		}
	}
	for i := range cw.Wheels {
		fw := &cw.Wheels[i]
		if near(fw.Pos, fw.R+6) {
			fw.Draw(draw)
		}
	}
	if cw.Coaster != nil {
		p, _ := cw.Coaster.At(cw.Coaster.S)
		if near(p, 12) {
			cw.Coaster.Draw(draw)
		}
	}
	for i := range cw.W.Pads {
		pd := &cw.W.Pads[i]
		if near(pd.Pos, pd.R+1) {
			pd.DrawMat(draw)
		}
	}
	for i, p := range cw.Turbines {
		if !near(p, 14) {
			continue
		}
		yaw := float32(0.6)
		draw(cw.TurbineTower, rl.MatrixMultiply(rl.MatrixRotateY(yaw), rl.MatrixTranslate(p.X, p.Y, p.Z)), rl.White)
		m := rl.MatrixMultiply(rl.MatrixRotateZ(cw.TurbineAng+float32(i)), rl.MatrixTranslate(0, 26.6, 1.4))
		m = rl.MatrixMultiply(m, rl.MatrixRotateY(yaw))
		m = rl.MatrixMultiply(m, rl.MatrixTranslate(p.X, p.Y, p.Z))
		draw(cw.TurbineRotor, m, rl.White)
	}
	cw.drawSkyTargets(r, draw, t, shadow, near)
	cw.drawBooths(draw, near)
	cw.DrawThemeAnimated(draw, t, shadow, near)
	if !shadow {
		for i := range cw.Clouds {
			c := &cw.Clouds[i]
			m := rl.MatrixMultiply(rl.MatrixScale(c.S, c.S*0.8, c.S), rl.MatrixTranslate(c.Pos.X, c.Pos.Y, c.Pos.Z))
			draw(cw.CloudMeshes[c.Mesh], m, rl.White)
		}
	}
}

// DrawTransparent: holograme, caramizi-fantoma si stalpii de lumina ai santierelor
func (cw *CityWorld) DrawTransparent(r *Renderer, draw DrawFn, t float32, near int) {
	for i := range cw.VendorPods {
		v := &cw.VendorPods[i]
		if !r.SphereVisible(v.Position, 6) {
			continue
		}
		s := &cw.LegoSites[v.SlotIndex]
		DrawVendorHologram(draw, v.Position, s.Meshes, t)
	}
	for i := range cw.LegoSites {
		s := &cw.LegoSites[i]
		if s.CurrentStage < s.TotalStages && r.SphereVisible(s.Position, 10*s.Scale()) {
			if s.CurrentStage < len(s.Meshes.Bricks) {
				a := uint8(80 + 60*sinf(t*4))
				draw(s.Meshes.Bricks[s.CurrentStage], s.transform(1, 0), rl.NewColor(255, 255, 255, a))
			}
		}
	}
}

// DrawBeacons: stalp de lumina aurie deasupra santierelor neterminate (se vad de departe)
func (cw *CityWorld) DrawBeacons(r *Renderer, t float32, near int) {
	rl.DisableDepthMask()
	for i := range cw.LegoSites {
		s := &cw.LegoSites[i]
		if s.CurrentStage >= s.TotalStages {
			continue
		}
		dx, dz := s.Position.X-r.camPos.X, s.Position.Z-r.camPos.Z
		if dx*dx+dz*dz > 320*320 || !r.SphereVisible(v3(s.Position.X, 20, s.Position.Z), 22) {
			continue
		}
		pulse := 0.5 + 0.5*sinf(t*3+float32(i))
		top := s.Position.Y + maxf(s.Meshes.Height*s.Scale(), 2) + 2
		rl.DrawCylinder(v3(s.Position.X, top, s.Position.Z), 0.7, 0.25, 38, 10, rl.NewColor(255, 215, 60, uint8(55+40*pulse)))
		// caramida care se roteste deasupra
		y := top + 1.6 + sinf(t*2.5+float32(i))*0.4
		rl.DrawCube(v3(s.Position.X, y, s.Position.Z), 1.3, 0.6, 0.8, rl.NewColor(255, 200, 30, 230))
		for _, ox := range []float32{-0.32, 0.32} {
			rl.DrawCylinder(v3(s.Position.X+ox, y+0.3, s.Position.Z), 0.17, 0.17, 0.16, 8, rl.NewColor(255, 225, 80, 240))
		}
		// inel care pulseaza pe placa, mai puternic cand Aventurierul e aproape
		rr := float32(5.6) + pulse*0.6
		a := uint8(70 + 60*pulse)
		if i == near {
			a = 200
		}
		rl.DrawCylinderWires(v3(s.Position.X, 0.45, s.Position.Z), rr, rr, 0.1, 32, rl.NewColor(255, 220, 60, a))
	}
	rl.EnableDepthMask()
}
