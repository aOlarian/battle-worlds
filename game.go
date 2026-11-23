package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const WindowTitle = "BATTLE WORLDS 6.0 - Lumile"

var (
	ScreenWidth  int32 = 1280
	ScreenHeight int32 = 720
)

type GameScene int

const (
	ScenePlayerSelect GameScene = iota
	ScenePlayerCreate
	SceneMainMenu
	SceneDifficulty
	SceneMapExplore
	SceneWinnerGlory
	SceneGameOver
	ScenePauseMenu
	SceneHelp
	SceneBooth // V 5.0: in cabina telefonica
)

const Gravity = float32(28)

type Player struct {
	Name     string
	Pos      rl.Vector3
	Yaw      float32
	VelY     float32
	OnGround bool
	Pose     Pose
	Meshes   *CharMeshes

	Attack  float64
	Defense float64
	HP      float64
	MaxHP   float64
	Gold    int64
	Silver  int64
}

type Enemy struct {
	Name   string
	Kind   int
	Meshes *CharMeshes
	Pos    rl.Vector3
	Yaw    float32
	Scale  float32
	Pose   Pose

	HP, MaxHP       float64
	Attack, Defense float64
	Gold, Silver    int64
	Alive           bool
	Home            rl.Vector2
	HomeR           float32
	Ranged          bool

	WanderT  float32
	WanderY  float32
	DodgeT   float32
	DodgeCD  float32
	DodgeDir rl.Vector3
	Stun     float32
	Flash    float32
	AtkCD    float32 // pauza pana la urmatorul atac
	Windup   float32 // >0: isi pregateste lovitura (se vede)
	Alert    bool
	Boss     bool // V 5.0: Seful insulei (apare dupa ceilalti)
}

type Particle3D struct {
	Position rl.Vector3
	Velocity rl.Vector3
	Color    rl.Color
	Size     float32
	Life     float32
	MaxLife  float32
	Gravity  float32
}

type Projectile3D struct {
	Position rl.Vector3
	Velocity rl.Vector3
	Color    rl.Color
	Life     float32
	Damage   float64
	Target   int
	CarTgt   int // masina tintita + 1 (0 = niciuna)
	TgtKind  int // barca / avion / pescarus tintit (TgtBoat...)
	TgtIdx   int
	Enemy    bool // proiectil tras de un inamic
	Gun      int  // V 5.0: arma care l-a tras
	Size     float32
	Splash   float32 // raza exploziei (rachete)
	Kind     int     // V 6.0: forma glontului (ProjWeb, ProjBat...)
}

type FloatText struct {
	Pos   rl.Vector3
	Text  string
	Color rl.Color
	Life  float32
	Size  float32
}

// Plimbari / atractii
const (
	RideNone = iota
	RideWheel
	RideCoaster
	RideTrampoline
)

type RideState struct {
	Kind      int
	Idx       int
	Cabin     int
	T         float32
	MaxY      float32
	StartLaps int
	Look      float32
	First     bool // camera din ochii aventurierului (tasta C); altfel din spate (3rd person)
}

var difficultyOptions = []struct {
	Name  string
	Key   string
	Desc  string
	Color rl.Color
}{
	{"CHEAT", "cheat", "Pentru părinte: totul deblocat (lumi, insule, arme, personaje), 99999 aur, 1000 viață.", rl.NewColor(0, 220, 255, 255)},
	{"UȘOR", "easy", "Inamicii au 70% viață și lovesc mai încet. Perfect pentru Aventurierul!", rl.NewColor(110, 230, 110, 255)},
	{"NORMAL", "medium", "Dificultate echilibrată, pe toate cele 5 insule.", rl.NewColor(245, 210, 80, 255)},
	{"GREU", "hard", "Inamicii au 160% viață și lovesc puternic!", rl.NewColor(255, 90, 90, 255)},
}

var mouseSensLevels = []float32{0.0014, 0.0019, 0.0025, 0.0032, 0.0041}

type Game struct {
	Scene     GameScene
	PrevScene GameScene

	R        *Renderer
	City     *CityWorld
	Camera   rl.Camera3D
	CamYaw   float32
	CamPitch float32
	CamDist  float32
	CamBoom  float32
	CamShake float32
	Player   Player
	Enemies  []Enemy
	Achieve  *AchievementManager
	Audio    *GameAudio

	Particles   []Particle3D
	Projectiles []Projectile3D
	Floats      []FloatText

	// arme: pistol laser de departe, sabie de aproape (se schimba singure)
	Weapon        int
	WeaponT       float32
	SwingT        float32
	SwingSide     float32
	ShootCD       float32
	AimPoint      rl.Vector3
	AimKind       int // barca / avion / pescarus din tinta (TgtNone = niciuna)
	AimIdx        int
	AimEnemy      int
	Attacker      int // V 6.0: singurul monstru care ataca acum (ceilalti asteapta pe rand)
	Train         TrainState
	Zip           ZipState
	arriveFirst   bool // prima sosire in lumea curenta (primeste eroul)
	RealmStates   []RealmState
	UnlockedRealm int // cea mai mare lume mare deschisa
	cityCache     map[int]*CityWorld
	islandLook    *[6][3]float32
	islandFogD    float32

	AimCar      int // masina din tinta (-1 = niciuna)
	NearSite    int
	UpgradeHits int
	UpgradeSite int
	HurtFlash   float32
	SinceHurt   float32

	// miscare
	Swimming   bool
	SwimStroke float32
	AirStartY  float32

	// plimbari
	Ride        RideState
	Fade        float32
	FadeDir     float32
	FadeAction  func()
	Prompt      string
	PromptIcon  IconKind
	PromptKind  int
	PromptIdx   int
	PadApex     float32
	PadPress    float32
	PadBoost    bool
	PadBoostMsg float32

	// condus masina
	Drive DriveState

	// progres / profil
	SaveData       *SaveData
	PlayerList     []string
	PlayerIdx      int
	CreateName     string
	SelectIdx      int
	MenuIdx        int
	DifficultyIdx  int
	PauseIdx       int
	Difficulty     string
	Level          int
	Score          int64
	TotalBuilt     int
	BonusDamage    float64
	BonusHP        float64
	BonusShield    float64
	HitStreak      int
	ZonesVisited   [5]bool
	BridgesCrossed [4]bool
	MouseSens      int

	// stare
	FrameTime    float32
	Time         float32
	CursorLocked bool
	ShouldQuit   bool
	NearVendor   int
	StepTimer    float32
	Invuln       float32
	Message      string
	MessageT     float32
	MessageIcon  IconKind
	ShowHelp     float32
	ShowMap      bool
	InTunnel     int
	OnBridge     int
	BridgeEntryT float32
	BridgeLastT  float32
	Zone         int
	ZoneBanner   float32
	LevelBanner  float32
	SkipMouse    int
	GTEnterX     float32
	SkipE        bool

	Font     rl.Font
	FontBold rl.Font
	HasFont  bool

	Debug *DebugRun

	// V 5.0
	Prog           ProgState
	Booth          BoothUI
	BeamT          float32
	BeamEnd        rl.Vector3
	beamCarAcc     float32
	beamThingAcc   float32
	floatCD        float32
	brickChimeT    float32
	prevYaw        float32
	prevPos        rl.Vector3
	landT          float32
	pendingLevelUp bool
	prevAir        bool
}

func (g *Game) Play(s *Sfx) {
	if g.Audio != nil {
		g.Audio.Play(s)
	}
}

func (g *Game) PlayAt(s *Sfx, pos rl.Vector3, vol, maxDist float32) {
	if g.Audio != nil {
		g.Audio.PlayAt(s, pos, g.Camera.Position, vol, maxDist)
	}
}

func NewGame() *Game {
	g := &Game{
		Scene:         ScenePlayerSelect,
		CamDist:       7,
		CamBoom:       7,
		CamPitch:      0.28,
		Level:         1,
		Difficulty:    "medium",
		DifficultyIdx: 2,
		PlayerIdx:     -1,
		NearVendor:    -1,
		NearSite:      -1,
		AimEnemy:      -1,
		UpgradeSite:   -1,
		InTunnel:      -1,
		OnBridge:      -1,
		Zone:          -1,
		Weapon:        1,
		MouseSens:     2,
		Achieve:       InitAchievements(),
	}
	g.Player = Player{Name: "Aventurierul", Meshes: GetCharMeshes(HeroStyle()), Attack: 45, Defense: 30, HP: 250, MaxHP: 250, Gold: 1000, Silver: 500}
	g.Prog = newProgState()
	g.Prog.GPS.World = -1
	g.Camera = rl.Camera3D{Position: v3(0, 8, 12), Target: v3(0, 1, 0), Up: v3(0, 1, 0), Fovy: 62, Projection: rl.CameraPerspective}
	return g
}

func (g *Game) ResetPlayerPosition() {
	if g.Drive.Active {
		g.exitCar()
	}
	g.Player.Pos = v3(0, 0.3, 30)
	g.Player.Yaw = math.Pi
	if curRealm > 0 && g.City != nil {
		g.Player.Pos = g.City.StationDoor()
	}
	g.CamYaw = 0
	g.Player.VelY = 0
	g.Player.OnGround = true
	g.Swimming = false
	g.Ride = RideState{}
	g.Invuln = 3
}

// ---------------------------------------------------------------
// DIFICULTATE
// ---------------------------------------------------------------

func (g *Game) baseStats() (hp, atk, def float64) {
	switch g.Difficulty {
	case "cheat":
		return 1000, 40, 100
	case "easy":
		return 350, 55, 40
	case "hard":
		return 200, 38, 20
	}
	return 250, 45, 30
}

// applyDifficulty seteaza viata/atacul de baza dupa dificultate; aurul nu se pierde.
func (g *Game) applyDifficulty() {
	hp, atk, def := g.baseStats()
	g.Player.MaxHP = hp + g.BonusHP
	g.Player.HP = g.Player.MaxHP
	g.Player.Attack = atk
	g.Player.Defense = def
	if g.Difficulty == "cheat" {
		if g.Player.Gold < 99999 {
			g.Player.Gold = 99999
		}
		if g.Player.Silver < 99999 {
			g.Player.Silver = 99999
		}
	}
}

func (g *Game) enemyMultipliers() (hpM, atkM, dodge float64) {
	switch g.Difficulty {
	case "cheat":
		return 0.5, 0.2, 0.05
	case "easy":
		return 0.7, 0.55, 0.08
	case "hard":
		return 1.6, 1.4, 0.22
	}
	return 1.0, 1.0, 0.14
}

var enemyNames = []string{"Mega Minotaur", "Cyber Gladiator", "Shadow Phantom", "Titan Golem", "Fire Dragon Knight", "Lava Beast", "Cyber Dino Bot", "Ice Giant"}
var enemyColors = []rl.Color{LegoRed, LegoPurple, rl.NewColor(60, 70, 90, 255), LegoLightGrey, LegoOrange, rl.NewColor(60, 50, 50, 255), LegoTeal, rl.NewColor(120, 180, 255, 255)}

// inamicii care trag de la distanta (ceilalti vin la lupta corp la corp)
var enemyRanged = []bool{false, true, true, false, true, false, false, true}

func (g *Game) aliveEnemies() int {
	n := 0
	for i := range g.Enemies {
		if g.Enemies[i].Alive {
			n++
		}
	}
	return n
}

// ---------------------------------------------------------------
// CURSOR
// ---------------------------------------------------------------

func (g *Game) lockCursor() {
	if !g.CursorLocked {
		rl.DisableCursor()
		g.CursorLocked = true
		g.SkipMouse = 2 // primul delta dupa blocare e un salt mare - il ignoram
	}
}

func (g *Game) unlockCursor() {
	if g.CursorLocked {
		rl.EnableCursor()
		g.CursorLocked = false
	}
}

func (g *Game) say(msg string) {
	g.Message = msg
	g.MessageT = 3
	g.MessageIcon = IconNone
}

// sayIcon: mesaj desenat (pictograma mare) cu textul mic dedesubt, pentru parinte.
func (g *Game) sayIcon(icon IconKind, msg string) {
	g.say(msg)
	g.MessageIcon = icon
}

// ---------------------------------------------------------------
// BUCLA PRINCIPALA
// ---------------------------------------------------------------

func (g *Game) Update() {
	g.FrameTime = minf(rl.GetFrameTime(), 0.05)
	if g.Debug != nil && g.Debug.FixedStep {
		g.FrameTime = 1.0 / 30
	}
	g.Time += g.FrameTime
	g.Achieve.Update(g.FrameTime)
	if g.MessageT > 0 {
		g.MessageT -= g.FrameTime
	}
	if g.CamShake > 0 {
		g.CamShake = maxf(g.CamShake-g.FrameTime*6, 0)
	}
	g.updateFade(g.FrameTime)
	switch g.Scene {
	case ScenePlayerSelect:
		g.updatePlayerSelect()
	case ScenePlayerCreate:
		g.updatePlayerCreate()
	case SceneMainMenu:
		g.updateMainMenu()
	case SceneDifficulty:
		g.updateDifficulty()
	case SceneMapExplore:
		g.updateExplore()
	case SceneWinnerGlory:
		g.updateLevelComplete()
	case SceneGameOver:
		g.updateGameOver()
	case ScenePauseMenu:
		g.updatePause()
	case SceneHelp:
		g.updateHelp()
	case SceneBooth:
		g.updateBooth()
	}
	if g.Scene == SceneMapExplore {
		g.updateParticles()
		fwd := rl.Vector3Normalize(rl.Vector3Subtract(g.Camera.Target, g.Camera.Position))
		right := rl.Vector3Normalize(rl.Vector3CrossProduct(fwd, v3(0, 1, 0)))
		g.Audio.SetListener(g.Camera.Position, right)
		p := g.Player.Pos
		onLand := g.City.W.NaturalLand(p.X, p.Z) > WaterY || p.Y > 2
		shore := absf(g.City.W.LandSD(p.X, p.Z))
		g.Audio.UpdateAmbience(g.FrameTime, onLand, g.InTunnel >= 0, shore, g.Camera.Position)
	} else {
		g.Audio.Silence(g.FrameTime)
	}
}

// startFade: ecranul se intuneca, se face actiunea (urcare / coborare), apoi revine.
func (g *Game) startFade(action func()) {
	if g.FadeDir != 0 {
		return
	}
	g.FadeDir = 1
	g.FadeAction = action
}

func (g *Game) updateFade(dt float32) {
	if g.FadeDir == 0 {
		return
	}
	g.Fade += g.FadeDir * dt / 0.3
	if g.FadeDir > 0 && g.Fade >= 1 {
		g.Fade = 1
		if g.FadeAction != nil {
			g.FadeAction()
			g.FadeAction = nil
		}
		g.CamBoom = 1.5
		g.FadeDir = -1
	} else if g.FadeDir < 0 && g.Fade <= 0 {
		g.Fade = 0
		g.FadeDir = 0
	}
}

func (g *Game) updateExplore() {
	dt := g.FrameTime
	p := &g.Player
	g.City.Update(dt, g.Time, p.Pos, g.Camera.Position, g.Audio)
	g.lockCursor()
	g.mouseLook()
	if g.Prog.Reveal.Active {
		g.updateReveal(dt)
	}
	g.updateTrain(dt)
	// cat timp e harta de dezvaluire pe ecran, Aventurierul sta pe loc
	frozen := g.revealMapAlpha() > 0.02
	if g.trainCutscene() {
		frozen = true
	}

	switch {
	case frozen:
	case g.Drive.Active:
		g.updateDriving(dt)
	case g.Ride.Kind == RideWheel || g.Ride.Kind == RideCoaster:
		g.updateRide(dt)
	default:
		g.updateMovement(dt)
	}
	p.Pose.Breathe += dt * 2
	g.updatePoseExtras(dt)

	g.trackPlaces()
	if g.Drive.Active {
		g.driveCamera(dt)
	} else {
		g.updateCamera(dt)
	}
	g.trainCamera()
	if !frozen {
		g.updateInteractions()
		g.updateWeapons(dt)
		g.updateEnemies(dt)
	}
	g.updateCarsGame(dt)
	g.updateProjectiles(dt)
	if g.pendingLevelUp {
		g.doLevelComplete()
	}
	g.updateProgress(dt)
	g.brickChime(dt)
	g.updateBeam(dt)
	if g.Scene != SceneMapExplore {
		return
	}

	if g.Invuln > 0 {
		g.Invuln -= dt
	}
	if g.HurtFlash > 0 {
		g.HurtFlash -= dt
	}
	g.SinceHurt += dt
	if g.SinceHurt > 6 && p.HP < p.MaxHP {
		p.HP = math.Min(p.MaxHP, p.HP+p.MaxHP*0.04*float64(dt))
	}
	if g.PadBoostMsg > 0 {
		g.PadBoostMsg -= dt
	}
	if g.ShowHelp > 0 {
		g.ShowHelp -= dt
	}
	if rl.IsKeyPressed(rl.KeyH) {
		if g.ShowHelp > 0 {
			g.ShowHelp = 0
		} else {
			g.ShowHelp = 30
		}
	}
	if rl.IsKeyPressed(rl.KeyM) || rl.IsKeyPressed(rl.KeyTab) {
		g.ShowMap = !g.ShowMap
	}
	if rl.IsKeyPressed(rl.KeyEscape) || rl.IsKeyPressed(rl.KeyP) {
		g.PrevScene = SceneMapExplore
		g.Scene = ScenePauseMenu
		g.PauseIdx = 0
		g.unlockCursor()
	}
}

// ---- mouse: camera se roteste imediat, fara netezire (fara intarziere) ----
func (g *Game) mouseLook() {
	md := rl.GetMouseDelta()
	if g.SkipMouse > 0 {
		g.SkipMouse--
		md = rl.Vector2{}
	}
	sens := mouseSensLevels[clampInt(g.MouseSens, 0, len(mouseSensLevels)-1)]
	if g.Ride.Kind == RideCoaster && g.Ride.First {
		g.Ride.Look = float32(math.Remainder(float64(g.Ride.Look-md.X*sens), 2*math.Pi))
	} else {
		g.CamYaw -= md.X * sens
	}
	if g.Ride.Kind == RideWheel || g.Ride.Kind == RideCoaster {
		// in atractii camera se misca liber in toate directiile
		g.CamPitch = clampf(g.CamPitch+md.Y*sens, -1.25, 1.45)
	} else {
		g.CamPitch = clampf(g.CamPitch+md.Y*sens, -0.75, 1.3)
	}
	maxD := float32(16)
	if g.Ride.Kind == RideWheel {
		maxD = 48
	}
	if wh := rl.GetMouseWheelMove(); wh != 0 {
		g.CamDist = clampf(g.CamDist-wh*1.2, 2.5, maxD)
	}
	if g.CamDist > maxD {
		g.CamDist += (maxD - g.CamDist) * 0.1
	}
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (g *Game) moveInput() rl.Vector3 {
	fwd := v3(-sinf(g.CamYaw), 0, -cosf(g.CamYaw))
	right := v3(cosf(g.CamYaw), 0, -sinf(g.CamYaw))
	var mv rl.Vector3
	if rl.IsKeyDown(rl.KeyW) || rl.IsKeyDown(rl.KeyUp) {
		mv = rl.Vector3Add(mv, fwd)
	}
	if rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyDown) {
		mv = rl.Vector3Subtract(mv, fwd)
	}
	if rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyLeft) {
		mv = rl.Vector3Subtract(mv, right)
	}
	if rl.IsKeyDown(rl.KeyD) || rl.IsKeyDown(rl.KeyRight) {
		mv = rl.Vector3Add(mv, right)
	}
	if g.Debug != nil {
		if dm, ok := g.Debug.Move(g); ok {
			mv = dm
		}
	}
	return mv
}

func (g *Game) jumpPressed() bool {
	if g.Debug != nil && g.Debug.Jump(g) {
		return true
	}
	return rl.IsKeyPressed(rl.KeySpace) || rl.IsMouseButtonPressed(rl.MouseButtonRight)
}

func (g *Game) jumpHeld() bool {
	if g.Debug != nil && g.Debug.Jump(g) {
		return true
	}
	return rl.IsKeyDown(rl.KeySpace) || rl.IsMouseButtonDown(rl.MouseButtonRight)
}

// ---- mers, inot, sarituri, trambuline ----
func (g *Game) updateMovement(dt float32) {
	w := g.City.W
	p := &g.Player
	// pe jos nu sta niciodata asezat (dupa masina, roata, montagne russe)
	p.Pose.Sit, p.Pose.HandsUp = false, false
	if g.updateZip(dt) {
		return
	}
	mv := g.moveInput()
	moving := rl.Vector3Length(mv) > 0.01
	blocked := false
	sprint := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
	speed := float32(11)
	if sprint {
		speed = 16
	}
	ground, kind := w.GroundAt(p.Pos.X, p.Pos.Z, p.Pos.Y+0.2)
	wading := kind == SurfWater && !g.Swimming && p.OnGround
	switch {
	case g.Swimming:
		speed = 6.5
		if sprint {
			speed = 9.5
		}
	case wading:
		speed *= 0.65
	case g.Ride.Kind == RideTrampoline:
		speed = 7.5
	}
	speed *= g.abilitySpeed(p.OnGround)
	if moving {
		mv = rl.Vector3Normalize(mv)
		step := speed * dt
		n := int(step/0.25) + 1
		testY := p.Pos.Y
		if g.Swimming {
			testY = p.Pos.Y + 1.0
		}
		movedAny := false
		for k := 0; k < n; k++ {
			before := p.Pos
			if w.MovePlayer(&p.Pos, mv.X*step/float32(n), mv.Z*step/float32(n), 0.45, testY) {
				movedAny = true
			}
			if g.lockedAt(p.Pos.X, p.Pos.Z, p.Pos.Y) && !g.lockedAt(before.X, before.Z, before.Y) {
				p.Pos = before
				g.lockBump()
				break
			}
		}
		if movedAny {
			p.Pose.Walk += dt * speed * 0.75
			p.Pose.Moving = minf(p.Pose.Moving+dt*6, 1)
			if p.OnGround && !g.Swimming {
				g.StepTimer += dt
				if g.StepTimer > 0.33*11/speed {
					g.StepTimer = 0
					g.footstep(kind)
				}
			}
		} else {
			p.Pose.Moving = maxf(p.Pose.Moving-dt*6, 0)
			blocked = true
		}
		if g.Swimming {
			p.Yaw = turnTo(p.Yaw, atan2f(mv.X, mv.Z), dt*8)
		}
	} else {
		p.Pose.Moving = maxf(p.Pose.Moving-dt*6, 0)
	}
	// a mers pana in marginea unei trambuline: se urca singur pe ea
	if moving && p.OnGround && !g.Swimming && g.Ride.Kind == RideNone {
		for pi := range w.Pads {
			pd := &w.Pads[pi]
			dx, dz := pd.Pos.X-p.Pos.X, pd.Pos.Z-p.Pos.Z
			dd := sqrtf(dx*dx + dz*dz)
			// lipit de marginea trambulinei si mergand spre ea
			if dd < pd.R+1.25 && dd > 0.01 && (dx*mv.X+dz*mv.Z)/dd > 0.4 && p.Pos.Y < pd.MatY()-0.2 && absf(p.Pos.Y-pd.Pos.Y) < 1 {
				g.mountPad(pi)
				return
			}
		}
	}
	if !g.Swimming {
		// Aventurierul priveste mereu unde tinteste camera (ca la shootere)
		p.Yaw = turnTo(p.Yaw, g.CamYaw+math.Pi, dt*18)
	}

	ground, kind = w.GroundAt(p.Pos.X, p.Pos.Z, p.Pos.Y+0.2)
	depth := WaterY - ground
	waterHere := kind == SurfWater

	// ---- inot ----
	if g.Swimming {
		g.SwimStroke += dt * (1.2 + 1.8*p.Pose.Moving)
		target := WaterY - 1.25 + sinf(g.Time*2.2)*0.06
		if g.surfing() {
			target = WaterY + 0.2 + sinf(g.Time*2.2)*0.06
		}
		p.Pos.Y += (target - p.Pos.Y) * minf(dt*6, 1)
		p.VelY = 0
		p.OnGround = true
		p.Pose.Swim = minf(p.Pose.Swim+dt*3, 1)
		if g.surfing() {
			p.Pose.Swim = 0 // sta in picioare pe placa
		}
		if moving && int(g.SwimStroke*1.0)%2 == 0 && rand.Float32() < dt*3 {
			g.PlayAt(g.Audio.Swim, p.Pos, 0.8, 30)
		}
		if rand.Float32() < dt*8*p.Pose.Moving {
			g.splash(v3(p.Pos.X, WaterY, p.Pos.Z), 2, 0.6)
		}
		if moving && blocked {
			// lovit de ceva iesit putin din apa (pila de pod, soclu, chei): se urca pe el
			mv = rl.Vector3Normalize(mv)
			ax, az := p.Pos.X+mv.X*1.0, p.Pos.Z+mv.Z*1.0
			gy, k := w.GroundAt(ax, az, WaterY+1.3)
			if k != SurfWater && gy > WaterY-0.5 && gy < WaterY+1.95 && !w.Solid(ax, az, gy+StepUp, 0.4) {
				g.Swimming = false
				p.Pos = v3(ax, gy, az)
				p.OnGround = true
				p.VelY = 0
				p.Pose.Swim = 0
				g.splash(v3(p.Pos.X, WaterY, p.Pos.Z), 8, 1.5)
				g.Play(g.Audio.Splash)
				return
			}
		}
		if !waterHere || depth < 1.0 {
			// a ajuns la mal: se ridica in picioare
			g.Swimming = false
			p.Pos.Y = maxf(ground, p.Pos.Y)
		} else if g.jumpPressed() {
			p.VelY = 7
			p.OnGround = false
			g.Swimming = false
			g.splash(v3(p.Pos.X, WaterY, p.Pos.Z), 10, 2)
			g.Play(g.Audio.Splash)
		}
		p.Pose.Air = false
		return
	}
	p.Pose.Swim = maxf(p.Pose.Swim-dt*4, 0)

	// ---- trambulina: apasarea pe panza ----
	if g.PadPress > 0 {
		g.PadPress -= dt
		pad := &w.Pads[g.Ride.Idx]
		if g.jumpPressed() || g.jumpHeld() {
			g.PadBoost = true
		}
		pad.Dip = 0.25 + 0.5*g.PadApex/pad.MaxApex
		p.Pos.Y = pad.MatY() - pad.Dip*0.8
		if g.PadPress <= 0 {
			g.launchFromPad(pad)
		}
		return
	}

	// ---- saritura + gravitatie ----
	if g.Ride.Kind != RideTrampoline && g.flyTakeoff() {
		g.Play(g.Audio.Jump)
	}
	if p.OnGround && g.jumpPressed() && g.Ride.Kind != RideTrampoline {
		p.VelY = g.jumpVel()
		p.OnGround = false
		g.AirStartY = p.Pos.Y
		g.Play(g.Audio.Jump)
	}
	if p.OnGround && p.VelY <= 0 {
		if ground < p.Pos.Y-0.7 {
			p.OnGround = false // a calcat in gol
			g.AirStartY = p.Pos.Y
		} else {
			p.Pos.Y += (ground - p.Pos.Y) * minf(dt*18, 1)
			if waterHere && depth >= 1.1 {
				g.enterWater()
			}
		}
	}
	if !p.OnGround {
		p.VelY -= Gravity * dt
		if g.Ride.Kind == RideNone {
			g.abilityAir(dt)
		}
		p.Pos.Y += p.VelY * dt
		if g.Ride.Kind == RideTrampoline && p.VelY < 0 {
			// toata coborarea e fereastra in care SPATIU face saritura mai mare
			if g.jumpPressed() {
				g.PadBoost = true
			}
		}
		if p.Pos.Y <= ground && p.VelY <= 0 {
			fall := g.AirStartY - ground
			if kind == SurfPad {
				if pi := w.PadAt(p.Pos.X, p.Pos.Z); pi >= 0 {
					// orice aterizare pe panza (din saritura sau de pe acoperis) il arunca din nou
					g.landOnPad(pi)
					return
				}
			}
			p.Pos.Y = ground
			p.VelY = 0
			p.OnGround = true
			if g.Ride.Kind == RideTrampoline {
				g.Ride = RideState{}
			}
			if fall > 1.5 {
				g.Play(g.Audio.Land)
				g.CamShake = minf(fall*0.04, 0.6)
			}
			if kind == SurfRoof && ground > 3 {
				g.Achieve.Unlock("roof_jumper")
			}
		}
		// in apa adanca: inoata
		if waterHere && depth >= 1.1 && p.Pos.Y <= WaterY-0.5 && p.VelY <= 0 {
			g.enterWater()
		}
		// tavanul tunelurilor
		if t := w.InTunnel(p.Pos.X, p.Pos.Z, p.Pos.Y); t >= 0 {
			ceil := float32(5.5)
			if t >= 10 {
				c := &w.Corridors[t-10]
				s, lat := c.Local(p.Pos.X, p.Pos.Z)
				ceil = c.CeilY(s, lat) - 2.9
			}
			if p.Pos.Y > ceil {
				p.Pos.Y = ceil
				p.VelY = minf(p.VelY, 0)
			}
		}
	}
	p.Pose.Air = !p.OnGround
	// siguranta: daca a cazut cumva sub lume (nu in tuneluri), il readucem in piata
	if p.Pos.Y < ground-3 && w.corridorVolume(p.Pos.X, p.Pos.Z, p.Pos.Y) < 0 {
		g.ResetPlayerPosition()
	}
}

func (g *Game) enterWater() {
	p := &g.Player
	if !g.Swimming {
		g.Swimming = true
		g.Ride = RideState{}
		g.Play(g.Audio.Splash)
		g.splash(v3(p.Pos.X, WaterY, p.Pos.Z), 18, 3)
		g.Achieve.Unlock("swimmer")
		if g.Debug != nil {
			g.Debug.Swims++
		}
	}
	p.VelY = 0
	p.OnGround = true
}

func (g *Game) splash(pos rl.Vector3, n int, spread float32) {
	for k := 0; k < n; k++ {
		g.spawnParticle(pos, v3((rand.Float32()-0.5)*spread*2, rand.Float32()*spread*2+1, (rand.Float32()-0.5)*spread*2), rl.NewColor(220, 245, 255, 230), 0.16, 0.7)
	}
}

func (g *Game) footstep(kind SurfaceKind) {
	a := g.Audio
	if a == nil {
		return
	}
	p := g.Player.Pos
	w := g.City.W
	var s *Sfx
	switch kind {
	case SurfTunnel, SurfBridge, SurfRoof:
		s = a.StepConcrete
	case SurfPad:
		s = a.Spring
	case SurfWater:
		g.PlayAt(a.Swim, p, 0.45, 30)
		g.splash(v3(p.X, WaterY, p.Z), 3, 0.8)
		return
	case SurfMountain:
		s = a.StepGrass
		if p.Y > 18 {
			s = a.StepSand
		}
	default:
		sd := w.LandSD(p.X, p.Z)
		switch {
		case sd > -GrassInset+0.5:
			s = a.StepSand
		case p.X*p.X+p.Z*p.Z < 43*43 || w.onRoad(p.X, p.Z):
			s = a.StepConcrete
		default:
			s = a.StepGrass
		}
	}
	a.PlayV(s, 1, 1)
}

// ---- trambuline ----
func (g *Game) landOnPad(pi int) {
	w := g.City.W
	p := &g.Player
	if g.Ride.Kind != RideTrampoline {
		g.Ride = RideState{Kind: RideTrampoline, Idx: pi}
		// cazut de sus (de pe acoperis, din saritura): sare inapoi aproape la fel de sus
		g.PadApex = clampf((g.AirStartY-w.Pads[pi].MatY())*0.85, 2.5, w.Pads[pi].MaxApex)
	}
	g.Ride.Idx = pi
	g.markBeachPad(pi)
	pad := &w.Pads[pi]
	p.Pos.Y = pad.MatY()
	p.VelY = 0
	p.OnGround = true
	g.PadPress = 0.16
	if g.jumpHeld() {
		g.PadBoost = true
	}
	g.PlayAt(g.Audio.Spring, pad.Pos, 0.7, 30)
}

func (g *Game) launchFromPad(pad *Trampoline) {
	p := &g.Player
	if g.PadBoost {
		g.PadApex = minf(g.PadApex*1.7+4, pad.MaxApex)
		g.PadBoostMsg = 0.8
	} else {
		g.PadApex = maxf(g.PadApex*0.8, 2.5)
	}
	g.PadBoost = false
	p.VelY = sqrtf(2 * Gravity * g.PadApex)
	p.OnGround = false
	g.AirStartY = pad.MatY() + g.PadApex
	pad.Dip = 0
	g.Audio.PlayV(g.Audio.Boing, 1, clampf(1.25-g.PadApex/50, 0.7, 1.25))
	if g.PadApex >= 20 {
		g.Achieve.Unlock("trampoline")
	}
	if g.Debug != nil && g.PadApex > g.Debug.MaxApex {
		g.Debug.MaxApex = g.PadApex
	}
}

// ---------------------------------------------------------------
// CAMERA: peste umarul drept, ca tinta din mijlocul ecranului sa nu fie acoperita de Aventurierul
// ---------------------------------------------------------------

func (g *Game) camVectors() (back, fwd, right rl.Vector3) {
	cp := cosf(g.CamPitch)
	back = v3(sinf(g.CamYaw)*cp, sinf(g.CamPitch), cosf(g.CamYaw)*cp)
	fwd = rl.Vector3Negate(back)
	right = v3(cosf(g.CamYaw), 0, -sinf(g.CamYaw))
	return
}

func (g *Game) updateCamera(dt float32) {
	w := g.City.W
	p := &g.Player
	if g.Ride.Kind == RideCoaster || g.Ride.Kind == RideWheel {
		g.rideCamera(dt)
		return
	}
	back, fwd, right := g.camVectors()
	shoulder := float32(1.25)
	height := float32(2.05)
	if g.Ride.Kind == RideWheel {
		shoulder = 0
		height = 2.2
	}
	if g.Swimming {
		height = 2.3
	}
	head := v3(p.Pos.X, p.Pos.Y+height, p.Pos.Z)
	pivot := rl.Vector3Add(head, rl.Vector3Scale(right, shoulder))
	pivot = w.CameraClear(head, pivot)
	desired := rl.Vector3Add(pivot, rl.Vector3Scale(back, g.CamDist))
	clear := w.CameraClear(pivot, desired)
	boom := rl.Vector3Distance(pivot, clear)
	if boom < g.CamBoom {
		g.CamBoom = boom
	} else {
		g.CamBoom += (boom - g.CamBoom) * minf(dt*5, 1)
	}
	pos := rl.Vector3Add(pivot, rl.Vector3Scale(back, g.CamBoom))
	target := rl.Vector3Add(pivot, rl.Vector3Scale(fwd, 12))
	if g.CamShake > 0 {
		s := g.CamShake * 0.12
		j := v3((rand.Float32()-0.5)*s, (rand.Float32()-0.5)*s, (rand.Float32()-0.5)*s)
		pos = rl.Vector3Add(pos, j)
		target = rl.Vector3Add(target, j)
	}
	g.Camera.Position = pos
	g.Camera.Target = target
}

func (g *Game) coasterCamera() {
	c := g.City.Coaster
	seat := c.SeatPos()
	_, t := c.At(c.S + 2)
	yaw := atan2f(t.X, t.Z) + g.Ride.Look
	horiz := sqrtf(t.X*t.X + t.Z*t.Z)
	pitch := clampf(atan2f(t.Y, horiz)-(g.CamPitch-0.28)*0.9, -1.2, 1.0)
	dir := v3(sinf(yaw)*cosf(pitch), sinf(pitch), cosf(yaw)*cosf(pitch))
	// ochii aventurierului, deasupra scaunului (in sistemul vagonului, ca sa nu intre in el la coborari)
	m := c.carMatrix(0)
	o := rl.Vector3Transform(v3(0, 0, 0), m)
	carUp := rl.Vector3Normalize(rl.Vector3Subtract(rl.Vector3Transform(v3(0, 1, 0), m), o))
	eye := rl.Vector3Add(seat, rl.Vector3Scale(carUp, 2.0))
	g.Camera.Position = eye
	g.Camera.Target = rl.Vector3Add(eye, rl.Vector3Scale(dir, 10))
}

// ---------------------------------------------------------------
// ZONE, PODURI, TUNELURI
// ---------------------------------------------------------------

func (g *Game) trackPlaces() {
	p := &g.Player
	w := g.City.W
	z := SectorAt(p.Pos.X, p.Pos.Z)
	if z >= 0 && z != g.Zone {
		g.Zone = z
		g.ZoneBanner = 3
		if z < 5 && !g.ZonesVisited[z] {
			g.ZonesVisited[z] = true
			all := true
			for _, v := range g.ZonesVisited {
				all = all && v
			}
			if all {
				g.Achieve.Unlock("city_explorer")
			}
		}
	}
	if g.ZoneBanner > 0 {
		g.ZoneBanner -= g.FrameTime
	}
	t := w.InTunnel(p.Pos.X, p.Pos.Z, p.Pos.Y)
	if t >= 0 && t != g.InTunnel {
		g.Achieve.Unlock("tunnel_rider")
		if t == 10 {
			g.GTEnterX = p.Pos.X
			g.sayIcon(IconTunnel, "Marele Tunel de sub Mare - 780 de metri, pe uscat, pe sub mare!")
		} else {
			g.sayIcon(IconTunnel, w.TunnelName(t)+" - ai intrat în tunel!")
		}
		g.Play(g.Audio.Engage)
	}
	if t == 10 && g.GTEnterX != 0 && absf(p.Pos.X-g.GTEnterX) > 560 {
		g.Achieve.Unlock("great_tunnel")
		g.GTEnterX = 0
	}
	if t < 0 && g.InTunnel == 10 {
		g.GTEnterX = 0
	}
	g.InTunnel = t
	bi, _ := w.BridgeAt(p.Pos.X, p.Pos.Z)
	if bi >= 0 && bi < 4 && p.OnGround && p.Pos.Y > -0.5 {
		b := &w.Bridges[bi]
		bt, _ := b.Local(p.Pos.X, p.Pos.Z)
		if g.OnBridge != bi {
			g.OnBridge = bi
			g.BridgeEntryT = bt
			g.sayIcon(IconBridge, b.Name)
		}
		g.BridgeLastT = bt
	} else if bi < 0 && g.OnBridge >= 0 {
		if absf(g.BridgeLastT-g.BridgeEntryT) > 0.85 && !g.BridgesCrossed[g.OnBridge] {
			g.BridgesCrossed[g.OnBridge] = true
			all := true
			for _, c := range g.BridgesCrossed {
				all = all && c
			}
			if all {
				g.Achieve.Unlock("bridge_master")
			}
		}
		g.OnBridge = -1
	}
}

// ---------------------------------------------------------------
// TASTA E: tonomat, lift, roata, montagne russe, trambulina
// ---------------------------------------------------------------

const (
	PromptNone = iota
	PromptVendor
	PromptLiftUp
	PromptLiftDown
	PromptWheel
	PromptCoaster
	PromptPad
	PromptExitPad
	PromptCar
	PromptTrain
	PromptBooth
)

func (g *Game) updateInteractions() {
	w := g.City.W
	cw := g.City
	p := &g.Player
	g.Prompt = ""
	g.PromptKind = PromptNone
	g.PromptIcon = IconNone
	set := func(k, idx int, icon IconKind, msg string) {
		if g.PromptKind == PromptNone && g.Prompt == "" {
			g.PromptKind, g.PromptIdx, g.Prompt, g.PromptIcon = k, idx, msg, icon
		}
	}
	dist2 := func(a rl.Vector3, x, z float32) float32 { return (a.X-x)*(a.X-x) + (a.Z-z)*(a.Z-z) }
	if g.Drive.Active || g.SkipE {
		// la volan E inseamna "cobori" (tratat in updateDriving)
		g.SkipE = false
		return
	}
	if g.Ride.Kind == RideTrampoline {
		set(PromptExitPad, g.Ride.Idx, IconExit, "E = cobori de pe trambulină")
	} else if g.Ride.Kind == RideNone && !g.Swimming {
		for i := range w.Lifts {
			l := &w.Lifts[i]
			if dist2(p.Pos, l.Bottom.X, l.Bottom.Z) < 3.5*3.5 && absf(p.Pos.Y-l.Bottom.Y) < 2 {
				set(PromptLiftUp, i, IconLiftUp, "E = liftul spre suprafață (insulița din mare)")
			}
			if dist2(p.Pos, l.Top.X, l.Top.Z) < 3.5*3.5 && absf(p.Pos.Y-l.Top.Y) < 2 {
				set(PromptLiftDown, i, IconLiftDown, "E = liftul în jos, în Marele Tunel")
			}
		}
		for i := range cw.Booths {
			b := &cw.Booths[i]
			d := b.DoorPos(1.4)
			if dist2(p.Pos, d.X, d.Z) < 2.6*2.6 && absf(p.Pos.Y-b.Pos.Y) < 1.6 {
				set(PromptBooth, i, IconPhone, "E = intri în cabina telefonică (hartă, costume, arme)")
			}
		}
		if g.Train.Phase == TrainWait {
			door := cw.StationDoor()
			if dist2(p.Pos, door.X, door.Z) < 8*8 {
				set(PromptTrain, 0, IconTrain, "E = urci în trenul magic (lumea următoare)")
			}
		}
		if k := g.nearestCarToEnter(); k >= 0 {
			set(PromptCar, k, IconCar, "E = urci în mașină și o conduci!")
		}
		for i := range cw.Wheels {
			fw := &cw.Wheels[i]
			if dist2(p.Pos, fw.Board.X, fw.Board.Z) < 7*7 && p.Pos.Y < 2 {
				set(PromptWheel, i, IconWheel, fmt.Sprintf("E = urcă în %s", fw.Name))
			}
		}
		if c := cw.Coaster; c != nil {
			s0 := c.Ctrl[1]
			if dist2(p.Pos, s0.X, s0.Z-2.3) < 9*9 && p.Pos.Y < 3 {
				if !c.Running && c.Wait > 0.6 {
					set(PromptCoaster, 0, IconCoaster, "E = urcă în montagne russe!")
				} else {
					set(PromptNone, 0, IconWait, "Montagne russe: așteaptă să vină trenul în stație...")
				}
			}
		}
		for i := range w.Pads {
			pd := &w.Pads[i]
			r := pd.R + 2.6
			if dist2(p.Pos, pd.Pos.X, pd.Pos.Z) < r*r {
				set(PromptPad, i, IconTrampoline, "E = sari pe "+lowerFirst(pd.Name))
			}
		}
	}
	if !rl.IsKeyPressed(rl.KeyE) && !(g.Debug != nil && g.Debug.PressE(g)) {
		// tonomatul se foloseste in continuare cu E (sau lovind santierul cu sabia)
		return
	}
	switch g.PromptKind {
	case PromptLiftUp:
		l := w.Lifts[g.PromptIdx]
		g.Play(g.Audio.Ding)
		g.startFade(func() {
			p.Pos = l.Top
			p.Pos.Z += 1.6
			p.OnGround = true
			g.CamYaw = math.Pi
			g.sayIcon(IconLiftUp, "Ai ieșit pe insulița din mijlocul mării!")
		})
		return
	case PromptLiftDown:
		l := w.Lifts[g.PromptIdx]
		g.Play(g.Audio.Ding)
		g.startFade(func() {
			p.Pos = l.Bottom
			p.Pos.Z -= 1.2
			p.OnGround = true
			g.CamYaw = 0
			g.sayIcon(IconLiftDown, "Ai coborât în Marele Tunel de sub Mare!")
		})
		return
	case PromptWheel:
		idx := g.PromptIdx
		g.startFade(func() {
			fw := &g.City.Wheels[idx]
			if SectorAt(fw.Pos.X, fw.Pos.Z) == SecPark {
				g.Prog.RodeWheel = true
			}
			g.Ride = RideState{Kind: RideWheel, Idx: idx, Cabin: fw.LowestCabin()}
			g.CamDist = 12
			g.sayIcon(IconWheel, "Ești în roată! Mouse = privești în jur, rotița = mai departe, C = camera, E = cobori.")
		})
		return
	case PromptCoaster:
		g.startFade(func() {
			c := g.City.Coaster
			g.Ride = RideState{Kind: RideCoaster, StartLaps: c.Laps}
			g.CamDist = 10
			g.CamPitch = 0.3
			c.Rider = true
			g.Prog.RodeCoast = true
			c.Wait = 1.2
			g.sayIcon(IconCoaster, "Ține-te bine! Mouse = privești în jur, C = din ochii aventurierului.")
		})
		return
	case PromptBooth:
		g.openBooth(g.PromptIdx)
		return
	case PromptPad:
		g.mountPad(g.PromptIdx)
		return
	case PromptCar:
		g.enterCar(g.PromptIdx)
		return
	case PromptTrain:
		g.boardTrain()
		return
	case PromptExitPad:
		pd := &w.Pads[g.Ride.Idx]
		g.Ride = RideState{}
		g.PadPress = 0
		_, fwd, _ := g.camVectors()
		dir := rl.Vector2Normalize(rl.NewVector2(fwd.X, fwd.Z))
		x, z := pd.Pos.X+dir.X*(pd.R+1.4), pd.Pos.Z+dir.Y*(pd.R+1.4)
		gy, _ := w.GroundAt(x, z, 50)
		p.Pos = v3(x, gy, z)
		p.VelY = 0
		p.OnGround = true
		return
	}
	if g.NearVendor >= 0 {
		g.interactWithLegoVendor(g.NearVendor)
	}
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	if r[0] >= 'A' && r[0] <= 'Z' {
		r[0] += 'a' - 'A'
	}
	return string(r)
}

// ---- roata si montagne russe ----
func (g *Game) updateRide(dt float32) {
	p := &g.Player
	g.Ride.T += dt
	p.Pose.Moving = 0
	p.Pose.Air = false
	p.Pose.Swim = 0
	switch g.Ride.Kind {
	case RideWheel:
		fw := &g.City.Wheels[g.Ride.Idx]
		p.Pos = fw.CabinFloor(g.Ride.Cabin)
		p.OnGround = true
		p.Yaw = turnTo(p.Yaw, g.CamYaw+math.Pi, dt*8)
		p.Pose.Sit = false
		if p.Pos.Y > g.Ride.MaxY {
			g.Ride.MaxY = p.Pos.Y
		}
		if fw.R > 20 && p.Pos.Y > 55 {
			g.Achieve.Unlock("wheel_top")
		}
		if g.Debug != nil && p.Pos.Y > g.Debug.MaxRideY {
			g.Debug.MaxRideY = p.Pos.Y
		}
		if rl.IsKeyPressed(rl.KeyE) || (g.Debug != nil && g.Debug.PressE(g)) {
			idx := g.Ride.Idx
			g.startFade(func() {
				fw := &g.City.Wheels[idx]
				p.Pos = fw.Board
				p.OnGround = true
				g.Ride = RideState{}
				g.CamDist = minf(g.CamDist, 9)
			})
		}
	case RideCoaster:
		c := g.City.Coaster
		p.Pos = c.SeatPos()
		_, yaw, _ := c.CarPose(0)
		p.Yaw = yaw
		p.Pose.Sit = true
		_, t := c.At(c.S)
		p.Pose.HandsUp = c.Running && (t.Y < -0.25 || c.V > 14)
		if g.Debug != nil && p.Pos.Y > g.Debug.MaxRideY {
			g.Debug.MaxRideY = p.Pos.Y
		}
		if c.Laps > g.Ride.StartLaps && !c.Running {
			g.Achieve.Unlock("coaster")
			if g.Debug != nil {
				g.Debug.CoasterLaps++
			}
			g.startFade(func() {
				c.Rider = false
				g.Ride = RideState{}
				p.Pose.Sit, p.Pose.HandsUp = false, false
				s0 := c.Ctrl[1]
				p.Pos = v3(s0.X, 0.5, s0.Z-2.6)
				p.OnGround = true
				g.sayIcon(IconCoaster, "Ce tură! Mai mergi o dată? Apasă E în stație.")
			})
		}
	}
}

// ---------------------------------------------------------------
// CONSTRUCTIE LEGO
// ---------------------------------------------------------------

func (g *Game) burstAt(site *ConstructionSite, col rl.Color, n int, spread float32) {
	top := v3(site.Position.X, site.Position.Y+site.Meshes.Height*site.Scale()*0.6+1, site.Position.Z)
	for k := 0; k < n; k++ {
		g.spawnParticle(top, v3((rand.Float32()-0.5)*spread, rand.Float32()*6+3, (rand.Float32()-0.5)*spread), col, 0.25, 1.2)
	}
}

// placeBrick: pune urmatoarea piesa (cu E la tonomat sau lovind santierul cu sabia)
func (g *Game) placeBrick(site *ConstructionSite) bool {
	if site.CurrentStage >= site.TotalStages {
		return false
	}
	cost := int64(100)
	if g.Player.Gold < cost {
		g.sayIcon(IconNoCoin, "Nu ai destul aur! Învinge inamici ca să câștigi aur.")
		g.Play(g.Audio.Error)
		return false
	}
	g.Player.Gold -= cost
	site.CurrentStage++
	site.PopTime = 0.45
	g.Achieve.Unlock("lego_builder")
	g.burstAt(site, LegoGold, 14, 4)
	g.PlayAt(g.Audio.LegoClick, site.Position, 1, 40)
	if site.CurrentStage == site.TotalStages {
		g.TotalBuilt++
		g.Play(g.Audio.BuildDone)
		g.burstAt(site, LegoCyan, 40, 10)
		g.sayIcon(IconTrophy, fmt.Sprintf("Bravo! Ai terminat: %s!", cleanName(site.Blueprint.Name)))
		switch site.Blueprint.Theme {
		case ThemeCar:
			g.Achieve.Unlock("car_collector")
		case ThemeCartoon:
			g.Achieve.Unlock("bluey_friend")
		case ThemeDino:
			g.Achieve.Unlock("dino_master")
		}
		if g.lettersDone() {
			g.Achieve.Unlock("battle_name")
		}
	}
	g.autosave()
	return true
}

func (g *Game) upgradeSite(site *ConstructionSite) bool {
	if site.UpgradeLevel >= 5 {
		g.sayIcon(IconCheck, "Construcția e deja la mărimea maximă (5x)!")
		return false
	}
	cost := int64(250 * site.UpgradeLevel)
	if g.Player.Gold < cost {
		g.sayIcon(IconNoCoin, fmt.Sprintf("Mărirea costă %d aur. Mai adună aur!", cost))
		g.Play(g.Audio.Error)
		return false
	}
	g.Player.Gold -= cost
	site.UpgradeLevel++
	site.TargetScale = float32(site.UpgradeLevel)
	site.HammerActive = true
	site.HammerStrikesLeft = 5
	site.HammerTime = 0.25
	g.Achieve.Unlock("giant_upgrade")
	g.burstAt(site, LegoCyan, 30, 10)
	g.Play(g.Audio.LevelUp)
	g.sayIcon(IconGrow, fmt.Sprintf("%s crește la %dx!", cleanName(site.Blueprint.Name), site.UpgradeLevel))
	g.autosave()
	return true
}

func (g *Game) interactWithLegoVendor(vendorIdx int) {
	pod := &g.City.VendorPods[vendorIdx]
	site := &g.City.LegoSites[pod.SlotIndex]
	if site.HammerActive && site.HammerStrikesLeft > 0 {
		return
	}
	if site.CurrentStage < site.TotalStages {
		if g.placeBrick(site) {
			site.HammerActive = true
			site.HammerStrikesLeft = 3
			site.HammerTime = 0.25
		}
		return
	}
	g.upgradeSite(site)
}

func (g *Game) lettersDone() bool {
	for i := 8; i < len(g.City.LegoSites); i++ {
		s := &g.City.LegoSites[i]
		if s.CurrentStage < s.TotalStages {
			return false
		}
	}
	return true
}

func cleanName(s string) string {
	out := []rune{}
	for _, r := range s {
		if r == '[' || r == ']' {
			continue
		}
		out = append(out, r)
	}
	res := string(out)
	for len(res) > 0 && res[len(res)-1] == ' ' {
		res = res[:len(res)-1]
	}
	return res
}

// ---------------------------------------------------------------
// SALVARE
// ---------------------------------------------------------------

func (g *Game) autosave() {
	if g.Debug != nil && os.Getenv("BATTLE_DEBUG_SAVE") == "" {
		return
	}
	if g.PlayerIdx >= 0 && g.PlayerIdx < len(g.SaveData.Players) {
		UpdatePlayerData(&g.SaveData.Players[g.PlayerIdx], g)
		g.SaveData.MouseSens = g.MouseSens + 1
		if err := SavePlayers(g.SaveData); err != nil {
			g.say("Nu am putut salva jocul: " + err.Error())
		}
	}
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// mountPad: Aventurierul se urca pe trambulina (cu E sau mergand spre ea) si incepe sa sara.
func (g *Game) mountPad(i int) {
	w := g.City.W
	p := &g.Player
	pd := &w.Pads[i]
	p.Pos = v3(pd.Pos.X, pd.MatY()+0.5, pd.Pos.Z)
	g.Ride = RideState{Kind: RideTrampoline, Idx: i}
	g.PadApex = 2.5
	g.PadPress = 0
	p.VelY = 5
	p.OnGround = false
	g.AirStartY = p.Pos.Y + 2
	g.sayIcon(IconTrampoline, "Apasă SPAȚIU (sau ține-l apăsat) ca să sari tot mai sus! E = cobori.")
}

// rideCamera: in roata si in montagne russe camera e libera (mouse in toate directiile).
// Implicit din spate (3rd person); C = din ochii aventurierului.
func (g *Game) rideCamera(dt float32) {
	w := g.City.W
	p := &g.Player
	if rl.IsKeyPressed(rl.KeyC) || (g.Debug != nil && g.Debug.PressC(g)) {
		g.Ride.First = !g.Ride.First
		if g.Ride.First {
			g.Ride.Look = 0
			g.sayIcon(IconEye, "Camera: din ochii aventurierului (C = din spate)")
		} else {
			g.sayIcon(IconCamera, "Camera: din spate (C = din ochii aventurierului)")
		}
	}
	if g.Ride.First && g.Ride.Kind == RideCoaster {
		g.coasterCamera()
		return
	}
	back, fwd, _ := g.camVectors()
	var head rl.Vector3
	if g.Ride.Kind == RideCoaster {
		c := g.City.Coaster
		m := c.carMatrix(0)
		o := rl.Vector3Transform(v3(0, 0, 0), m)
		carUp := rl.Vector3Normalize(rl.Vector3Subtract(rl.Vector3Transform(v3(0, 1, 0), m), o))
		head = rl.Vector3Add(c.SeatPos(), rl.Vector3Scale(carUp, 1.9))
	} else {
		head = v3(p.Pos.X, p.Pos.Y+1.75, p.Pos.Z)
	}
	if g.Ride.First {
		// in cabina roții: privesti liber de unde stai
		g.Camera.Position = head
		g.Camera.Target = rl.Vector3Add(head, rl.Vector3Scale(fwd, 10))
		return
	}
	dist := clampf(g.CamDist, 4, 48)
	g.CamBoom += (dist - g.CamBoom) * minf(dt*6, 1)
	pos := rl.Vector3Add(head, rl.Vector3Scale(back, g.CamBoom))
	if gy, _ := w.GroundAt(pos.X, pos.Z, pos.Y+50); pos.Y < gy+0.6 {
		pos.Y = gy + 0.6
	}
	if pos.Y < WaterY+0.6 {
		pos.Y = WaterY + 0.6
	}
	g.Camera.Position = pos
	g.Camera.Target = head
}
