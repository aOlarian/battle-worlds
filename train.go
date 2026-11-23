package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// V 6.0: TRENUL MAGIC intre lumile mari.
// Fiecare lume are o gara (peron + casuta + sine scurte). Cand lumea e
// terminata (toate insulele), trenul coboara din cer si asteapta la peron.
// Aventurierul urca (E sau doar intra in vagon), trenul zboara spre cer, ecranul
// se albeste, se incarca lumea noua, iar trenul coboara la gara ei si il lasa
// pe Aventurierul pe peron. Tot asa se calatoreste si din cabina telefonica.
// ====================================================================

const (
	TrainOff = iota
	TrainArrive
	TrainWait
	TrainDepart
	TrainFadeOut
	TrainArriveNew
	TrainUnload
)

const (
	trainArriveT = 5.0
	trainDepartT = 3.6
	trainFadeT   = 1.0
)

type TrainState struct {
	Phase  int
	T      float32
	Target int // lumea spre care pleaca
	Carry  bool
}

var trainMesh rl.Mesh
var trainReady bool

// realmStationSpot: loc liber langa piateta din centru (departe de drumuri).
func realmStationSpot(cb *CityBuilder, hub rl.Vector2) rl.Vector3 {
	best := rl.NewVector3(hub.X+36, 0, hub.Y)
	bs := float32(1e9)
	for ri := 0; ri < 8; ri++ {
		r := 30 + float32(ri)*3
		for ai := 0; ai < 48; ai++ {
			a := float32(ai) / 48 * 2 * math.Pi
			x, z := hub.X+cosf(a)*r, hub.Y+sinf(a)*r
			if cb.nearRoad(x, z, 12) || cb.reserved(x, z, 15) {
				continue
			}
			if s := r; s < bs {
				bs = s
				best = v3(x, 0, z)
			}
		}
	}
	return best
}

// directia sinelor: perpendicular pe directia spre centru (peronul priveste spre piateta)
func stationAxes(st rl.Vector3) (along, toHub rl.Vector2) {
	toHub = rl.Vector2Normalize(rl.NewVector2(-st.X, -st.Z))
	if toHub.X == 0 && toHub.Y == 0 {
		toHub = rl.NewVector2(0, -1)
	}
	along = rl.NewVector2(-toHub.Y, toHub.X)
	return
}

func buildStationMesh(mb *MeshBuilder, st rl.Vector3, D *RealmDef) {
	along, toHub := stationAxes(st)
	P := func(u, v, y float32) rl.Vector3 {
		return v3(st.X+along.X*u+toHub.X*v, st.Y+y, st.Z+along.Y*u+toHub.Y*v)
	}
	yaw := atan2f(along.X, along.Y)
	mb.SetMatrix(rl.MatrixMultiply(rl.MatrixRotateY(yaw), rl.MatrixTranslate(st.X, st.Y, st.Z)))
	// peronul (spre piateta), sinele (in spate)
	stone := lerpColor(D.Path, LegoWhite, 0.25)
	mb.BeveledBox(v3(0, 0.25, 0), v3(5, 0.5, 26), 0.1, stone) // local X = toHub? (rotit: z = along)
	mb.Box(v3(-2.4, 0.52, 0), v3(0.3, 0.04, 26), LegoYellow)  // linia galbena
	for _, x := range []float32{-4.6, -6.0} {
		mb.Box(v3(x, 0.12, 0), v3(0.18, 0.18, 34), rl.NewColor(150, 150, 160, 255)) // sine
	}
	for k := -16; k <= 16; k += 2 {
		mb.Box(v3(-5.3, 0.05, float32(k)), v3(2.4, 0.1, 0.45), rl.NewColor(110, 80, 55, 255)) // traverse
	}
	// casuta garii cu acoperis si ceas
	mb.BeveledBox(v3(1.4, 2.0, 0), v3(2.2, 3.0, 8), 0.15, lerpColor(D.Main, LegoWhite, 0.35))
	mb.Quad(v3(0, 3.5, -4.6), v3(0, 3.5, 4.6), v3(2.9, 4.6, 4.6), v3(2.9, 4.6, -4.6), v3(-0.4, 1, 0), D.Accent)
	mb.Quad(v3(2.9, 4.6, -4.6), v3(2.9, 4.6, 4.6), v3(2.9, 3.5, 4.6), v3(2.9, 3.5, -4.6), v3(1, 0, 0), shade(D.Accent, 0.8))
	mb.Cylinder(v3(0.28, 3.0, 0), v3(0.2, 3.0, 0), 0.75, 0.75, 16, LegoWhite, true)
	mb.Box(v3(0.16, 3.15, 0), v3(0.05, 0.4, 0.07), LegoBlack)
	mb.Box(v3(0.16, 3.0, 0.15), v3(0.05, 0.07, 0.35), LegoBlack)
	// stalpi cu felinare la capete
	for _, z := range []float32{-11, 11} {
		mb.Cylinder(v3(-1.5, 0.5, z), v3(-1.5, 4.2, z), 0.12, 0.1, 8, LegoBlack, false)
		mb.Emis = 1
		mb.Ellipsoid(v3(-1.5, 4.4, z), v3(0.35, 0.45, 0.35), 5, 8, rl.NewColor(255, 230, 140, 255))
		mb.Emis = 0
	}
	mb.ClearMatrix()
	_ = P
}

func stationColliders(w *World, st rl.Vector3) {
	along, toHub := stationAxes(st)
	// casuta: cerc de 2.4 m in spatele peronului (spre piateta)
	c := v3(st.X+toHub.X*1.4, 0, st.Z+toHub.Y*1.4)
	for _, u := range []float32{-2.6, 0, 2.6} {
		w.Circles = append(w.Circles, CircleCollider{X: c.X + along.X*u, Z: c.Z + along.Y*u, R: 1.4, Top: st.Y + 4.6})
	}
}

// StationDoor: unde sta Aventurierul cand coboara din tren (pe peron, langa usa vagonului).
func (cw *CityWorld) StationDoor() rl.Vector3 {
	_, toHub := stationAxes(cw.Station)
	return v3(cw.Station.X+toHub.X*(-1.6), cw.Station.Y+0.5, cw.Station.Z+toHub.Y*(-1.6))
}

// pozitia de oprire a trenului (pe sine)
func (cw *CityWorld) trainRest() (rl.Vector3, rl.Vector2) {
	along, toHub := stationAxes(cw.Station)
	return v3(cw.Station.X-toHub.X*5.3, cw.Station.Y+0.2, cw.Station.Z-toHub.Y*5.3), along
}

func buildTrainMesh() {
	mb := NewMB()
	red := rl.NewColor(220, 40, 50, 255)
	blue := rl.NewColor(40, 90, 210, 255)
	gold := LegoGoldBright
	// locomotiva (fata spre +Z), lungime ~7
	mb.BeveledBox(v3(0, 1.0, 0), v3(2.4, 0.6, 7.2), 0.15, LegoBlack)
	mb.Cylinder(v3(0, 1.95, -0.6), v3(0, 1.95, 3.4), 1.05, 1.05, 18, red, true)
	mb.Cylinder(v3(0, 1.95, 3.4), v3(0, 1.95, 3.6), 1.1, 1.1, 18, gold, true)
	mb.Cylinder(v3(0, 2.8, 2.5), v3(0, 4.0, 2.5), 0.32, 0.55, 12, LegoBlack, true)
	mb.Cylinder(v3(0, 4.0, 2.5), v3(0, 4.15, 2.5), 0.6, 0.6, 12, gold, true)
	mb.Ellipsoid(v3(0, 3.05, 1.0), v3(0.4, 0.35, 0.4), 6, 10, gold)
	mb.BeveledBox(v3(0, 2.3, -2.2), v3(2.4, 2.6, 2.6), 0.15, blue)
	mb.BeveledBox(v3(0, 3.75, -2.2), v3(2.7, 0.3, 3.0), 0.1, red)
	mb.Box(v3(1.21, 2.6, -2.2), v3(0.04, 1.0, 1.4), rl.NewColor(150, 210, 250, 255))
	mb.Box(v3(-1.21, 2.6, -2.2), v3(0.04, 1.0, 1.4), rl.NewColor(150, 210, 250, 255))
	mb.Emis = 1
	mb.Ellipsoid(v3(0, 2.0, 3.7), v3(0.4, 0.4, 0.15), 6, 10, rl.NewColor(255, 240, 170, 255))
	mb.Emis = 0
	mb.Box(v3(0, 0.6, 4.0), v3(2.2, 0.5, 0.6), gold) // aparatoarea
	for _, z := range []float32{-2.4, 0, 2.2} {
		for _, x := range []float32{-1.25, 1.25} {
			mb.Cylinder(v3(x, 0.75, z), v3(x*1.08, 0.75, z), 0.7, 0.7, 14, red, true)
			mb.Cylinder(v3(x*1.08, 0.75, z), v3(x*1.1, 0.75, z), 0.25, 0.25, 8, gold, true)
		}
	}
	// 2 vagoane colorate cu ferestre
	cols := []rl.Color{rl.NewColor(255, 200, 40, 255), rl.NewColor(80, 190, 90, 255)}
	for k, c := range cols {
		z0 := -7.6 - float32(k)*7.4
		mb.BeveledBox(v3(0, 0.9, z0), v3(2.3, 0.5, 6.6), 0.12, LegoBlack)
		mb.BeveledBox(v3(0, 2.3, z0), v3(2.5, 2.4, 6.6), 0.2, c)
		mb.BeveledBox(v3(0, 3.6, z0), v3(2.7, 0.3, 6.9), 0.12, shade(c, 0.75))
		for i := -2; i <= 2; i++ {
			for _, x := range []float32{-1.26, 1.26} {
				mb.Box(v3(x, 2.6, z0+float32(i)*1.25), v3(0.04, 0.9, 0.8), rl.NewColor(150, 210, 250, 255))
			}
		}
		for _, z := range []float32{-2.4, 2.4} {
			for _, x := range []float32{-1.2, 1.2} {
				mb.Cylinder(v3(x, 0.55, z0+z), v3(x*1.08, 0.55, z0+z), 0.5, 0.5, 12, LegoBlack, true)
			}
		}
		// stelute aurii pe laterale (tren magic)
		for _, x := range []float32{-1.27, 1.27} {
			mb.Box(v3(x, 1.5, z0), v3(0.04, 0.3, 4.8), gold)
		}
	}
	trainMesh = mb.Build()
	trainReady = true
}

// pozitia trenului: la oprire (s=0); s<0 = in aer inainte de gara, s>0 = pleaca spre cer
func (cw *CityWorld) trainPose(s float32) (rl.Vector3, float32, float32) {
	rest, along := cw.trainRest()
	dir := v3(along.X, 0, along.Y)
	yaw := atan2f(along.X, along.Y)
	// traiectorie: pe orizontala de-a lungul sinelor, pe verticala o parabola
	x := s * 60
	h := s * s * 55
	pitch := float32(0)
	if s != 0 {
		pitch = -atanf(2*s*55/60) * 0.6
	}
	p := rl.Vector3Add(rest, rl.Vector3Scale(dir, x))
	p.Y += h
	return p, yaw, pitch
}

func atanf(x float32) float32 { return float32(math.Atan(float64(x))) }

func (g *Game) drawTrain(draw DrawFn) {
	tr := &g.Train
	if tr.Phase == TrainOff || tr.Phase == TrainFadeOut || !trainReady || g.City == nil {
		return
	}
	s := g.trainS()
	p, yaw, pitch := g.City.trainPose(s)
	m := rl.MatrixMultiply(rl.MatrixRotateX(pitch), rl.MatrixRotateY(yaw))
	m = rl.MatrixMultiply(m, rl.MatrixTranslate(p.X, p.Y, p.Z))
	draw(trainMesh, m, rl.White)
}

// trainS: parametrul traiectoriei pentru faza curenta
func (g *Game) trainS() float32 {
	tr := &g.Train
	switch tr.Phase {
	case TrainArrive, TrainArriveNew:
		k := 1 - clampf(tr.T/trainArriveT, 0, 1)
		return -k * k * 2.2 // incetineste la sosire
	case TrainDepart:
		k := clampf(tr.T/trainDepartT, 0, 1)
		return k * k * 2.6
	case TrainUnload:
		if tr.T < 1.2 {
			return 0
		}
		k := clampf((tr.T-1.2)/trainDepartT, 0, 1)
		return k * k * 2.6
	}
	return 0
}

// trainCutscene: Aventurierul nu se poate misca (e in tren sau priveste plecarea)
func (g *Game) trainCutscene() bool {
	ph := g.Train.Phase
	return ph == TrainDepart || ph == TrainFadeOut || ph == TrainArriveNew || (ph == TrainUnload && g.Train.T < 1.2)
}

// callTrain: trenul vine la gara (lumea e terminata sau calatorie ceruta).
func (g *Game) callTrain(target int) {
	if !trainReady {
		buildTrainMesh()
	}
	g.Train = TrainState{Phase: TrainArrive, Target: target}
	g.Play(g.Audio.Ding)
}

// travelNow: plecare imediata (din cabina): ecranul se albeste si se calatoreste.
func (g *Game) travelNow(target int) {
	if !trainReady {
		buildTrainMesh()
	}
	g.Train = TrainState{Phase: TrainFadeOut, Target: target}
}

func (g *Game) updateTrain(dt float32) {
	tr := &g.Train
	if tr.Phase == TrainOff {
		return
	}
	tr.T += dt
	p := &g.Player
	switch tr.Phase {
	case TrainArrive:
		if tr.T >= trainArriveT {
			tr.Phase, tr.T = TrainWait, 0
			g.Play(g.Audio.Ding)
		}
	case TrainWait:
		door := g.City.StationDoor()
		d := sqrtf(sq(p.Pos.X-door.X) + sq(p.Pos.Z-door.Z))
		if d < 2.6 && absf(p.Pos.Y-door.Y) < 2 {
			g.boardTrain()
		}
	case TrainDepart:
		if tr.T >= trainDepartT {
			tr.Phase, tr.T = TrainFadeOut, 0
		}
	case TrainFadeOut:
		if tr.T >= trainFadeT {
			g.switchRealm(tr.Target)
			tr.Phase, tr.T = TrainArriveNew, 0
		}
	case TrainArriveNew:
		if tr.T >= trainArriveT {
			tr.Phase, tr.T = TrainUnload, 0
			door := g.City.StationDoor()
			p.Pos = door
			p.VelY = 0
			g.arrivedInRealm()
		}
	case TrainUnload:
		if tr.T >= 1.2+trainDepartT {
			tr.Phase = TrainOff
		}
	}
}

// trainCamera: camera de film cat Aventurierul e in tren.
func (g *Game) trainCamera() {
	if !g.trainCutscene() {
		return
	}
	s := g.trainS()
	p, _, _ := g.City.trainPose(s)
	rest, along := g.City.trainRest()
	_, toHub := stationAxes(g.City.Station)
	eye := v3(rest.X+toHub.X*22-along.X*10, rest.Y+7, rest.Z+toHub.Y*22-along.Y*10)
	g.Camera.Position = eye
	g.Camera.Target = v3(p.X, p.Y+2, p.Z)
}

// trainFade: cat de alb e ecranul (0..1)
func (g *Game) trainFade() float32 {
	tr := &g.Train
	switch tr.Phase {
	case TrainDepart:
		return clampf((tr.T-(trainDepartT-1))/1, 0, 1)
	case TrainFadeOut:
		return 1
	case TrainArriveNew:
		return clampf(1-tr.T/1.2, 0, 1)
	}
	return 0
}

func (g *Game) drawTrainOverlay() {
	tr := &g.Train
	if f := g.trainFade(); f > 0 {
		rl.DrawRectangle(0, 0, ScreenWidth, ScreenHeight, rl.NewColor(255, 252, 240, uint8(255*f)))
		if f > 0.6 && tr.Target >= 0 && tr.Target < len(Realms) {
			R := &Realms[tr.Target]
			a := uint8(255 * clampf((f-0.6)/0.4, 0, 1))
			for k := 0; k < 9; k++ {
				ang := g.Time*0.8 + float32(k)*0.7
				drawStar(sw()/2+cosf(ang)*220, sh()/2+sinf(ang)*120, 18, fade(R.Main, a))
			}
			drawIcon(IconTrain, sw()/2, sh()/2-10, 120, fade(R.Main, a))
			g.hint(R.Name, sw()/2, sh()/2+70, a)
		}
	}
}

// boardTrain: Aventurierul urca in tren (E langa usa sau intra direct).
func (g *Game) boardTrain() {
	tr := &g.Train
	if tr.Phase != TrainWait {
		return
	}
	tr.Phase, tr.T = TrainDepart, 0
	tr.Carry = true
	g.stopGPS()
	g.Play(g.Audio.LevelUp)
}
