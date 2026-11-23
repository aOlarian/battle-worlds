package main

import (
	"embed"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Sunete reale (inregistrari CC0 de la Kenney si OpenGameArt, vezi assets/sfx/CREDITS.txt),
// incluse in exe. Fiecare sunet are variante si copii ("alias") ca sa se poata
// suprapune (pasi, focuri). Sunetele din lume se aud mai incet cu distanta si
// din stanga/dreapta. Ambianta: pasari pe uscat, valuri la mal, vuiet in tunel.

//go:embed assets/sfx/*.ogg
var sfxFS embed.FS

type Sfx struct {
	Snds        [][]rl.Sound // [varianta][copie]
	Vol         float32
	PitchJitter float32
}

type GameAudio struct {
	Shoot, Swing, SwordHit, ShotHit, DrawSword, DrawGun       *Sfx
	StepGrass, StepConcrete, StepWood, StepSand               *Sfx
	Build, LegoClick, Coin, Jump, Land, Boing, Spring, Hurt   *Sfx
	Victory, LevelUp, BuildDone, Achievement, Death, Engage   *Sfx
	Click, Menu, Error, Ding, Greet, Windup, EnemyShot, Clack *Sfx
	Splash, Swim, Wave, CarHonk, CarWhoosh, Airplane          *Sfx
	Boom, Engine                                              *Sfx
	engineOn                                                  bool

	Birds, TunnelHum rl.Music
	// raylib citeste ogg-ul ambiantei direct din aceasta memorie pe toata durata jocului:
	// o tinem referita si fixata, altfel GC-ul Go o elibereaza si jocul se blocheaza
	// la primul cadru pe PC-urile cu placa de sunet (bug V4.1 "Pornim traficul").
	musicData        [][]byte
	musicPin         runtime.Pinner
	hasMusic         bool
	birdsVol, humVol float32
	waveT            float32

	ready    bool
	all      []rl.Sound
	aliases  []rl.Sound
	lisPos   rl.Vector3
	lisRight rl.Vector3
}

func (a *GameAudio) PlayV(s *Sfx, vol, pitch float32) {
	if a == nil || !a.ready || s == nil || len(s.Snds) == 0 {
		return
	}
	v := s.Snds[rand.Intn(len(s.Snds))]
	snd := v[0]
	for _, c := range v {
		if !rl.IsSoundPlaying(c) {
			snd = c
			break
		}
	}
	if s.PitchJitter > 0 {
		pitch *= 1 + (rand.Float32()*2-1)*s.PitchJitter
	}
	rl.SetSoundVolume(snd, clampf(vol*s.Vol, 0, 1))
	rl.SetSoundPitch(snd, pitch)
	rl.SetSoundPan(snd, 0.5)
	rl.PlaySound(snd)
}

func (a *GameAudio) Play(s *Sfx) { a.PlayV(s, 1, 1) }

// SetListener: unde e "urechea" (camera) si directia spre dreapta ei.
func (a *GameAudio) SetListener(pos, right rl.Vector3) {
	if a == nil {
		return
	}
	a.lisPos = pos
	a.lisRight = right
}

// PlayAt: sunet in lume, mai incet cu distanta si usor din stanga / dreapta.
func (a *GameAudio) PlayAt(s *Sfx, pos, listener rl.Vector3, vol, maxDist float32) {
	if a == nil || !a.ready || s == nil || len(s.Snds) == 0 {
		return
	}
	d := rl.Vector3Distance(pos, listener)
	if d > maxDist {
		return
	}
	k := 1 - d/maxDist
	att := k * k
	v := s.Snds[rand.Intn(len(s.Snds))]
	snd := v[0]
	for _, c := range v {
		if !rl.IsSoundPlaying(c) {
			snd = c
			break
		}
	}
	pan := float32(0.5)
	if d > 0.5 && rl.Vector3Length(a.lisRight) > 0.5 {
		side := rl.Vector3DotProduct(rl.Vector3Scale(rl.Vector3Subtract(pos, listener), 1/d), a.lisRight)
		pan = 0.5 - side*0.3 // in raylib 1.0 = stanga
	}
	pitch := float32(1)
	if s.PitchJitter > 0 {
		pitch = 1 + (rand.Float32()*2-1)*s.PitchJitter
	}
	rl.SetSoundVolume(snd, clampf(vol*att*s.Vol, 0, 1))
	rl.SetSoundPitch(snd, pitch)
	rl.SetSoundPan(snd, pan)
	rl.PlaySound(snd)
}

func InitAudio() *GameAudio {
	a := &GameAudio{}
	if os.Getenv("BATTLE_NOAUDIO") != "" {
		return a
	}
	rl.InitAudioDevice()
	a.ready = rl.IsAudioDeviceReady()
	if !a.ready {
		return a
	}
	rl.SetMasterVolume(0.8)
	if os.Getenv("BATTLE_DEBUG") != "" {
		rl.SetMasterVolume(0) // testele automate merg in liniste, dar cu sunetele incarcate
	}
	L := func(prefix string, vol, jitter float32, copies int) *Sfx { return a.load(prefix, vol, jitter, copies) }
	a.Shoot = L("shoot", 0.55, 0.06, 3)
	a.Swing = L("swing", 0.8, 0.08, 2)
	a.SwordHit = L("swordhit", 0.9, 0.08, 2)
	a.ShotHit = L("shothit", 0.7, 0.1, 2)
	a.DrawSword = L("draw_sword", 0.6, 0.05, 1)
	a.DrawGun = L("draw_gun", 0.7, 0.05, 1)
	a.StepGrass = L("step_grass", 0.45, 0.08, 2)
	a.StepConcrete = L("step_concrete", 0.45, 0.08, 2)
	a.StepWood = L("step_wood", 0.45, 0.08, 2)
	a.StepSand = L("step_snow", 0.4, 0.1, 2)
	a.Build = L("build", 0.8, 0.08, 2)
	a.LegoClick = L("lego_click", 0.8, 0.1, 2)
	a.Coin = L("coin", 0.7, 0.05, 2)
	a.Jump = L("jump", 0.5, 0.1, 2)
	a.Land = L("land", 0.5, 0.1, 2)
	a.Boing = L("boing", 0.75, 0, 2)
	a.Spring = L("spring", 0.5, 0.1, 2)
	a.Hurt = L("hurt", 0.9, 0.1, 2)
	a.Victory = L("victory", 0.7, 0, 1)
	a.LevelUp = L("levelup", 0.8, 0, 1)
	a.BuildDone = L("builddone", 0.8, 0, 1)
	a.Achievement = L("achievement", 0.8, 0, 1)
	a.Death = L("death", 0.8, 0, 1)
	a.Engage = L("engage", 0.5, 0, 1)
	a.Click = L("click", 0.6, 0, 2)
	a.Menu = L("menu", 0.5, 0, 2)
	a.Error = L("error", 0.5, 0, 1)
	a.Ding = L("ding", 0.7, 0, 1)
	a.Greet = L("greet", 0.35, 0.15, 2)
	a.Windup = L("windup", 0.6, 0.1, 3)
	a.EnemyShot = L("enemy_shot", 0.5, 0.08, 3)
	a.Clack = L("clack", 0.45, 0.1, 3)
	a.Splash = L("splash", 0.7, 0.08, 2)
	a.Swim = L("swim", 0.5, 0.12, 2)
	a.Wave = L("wave", 0.6, 0.05, 1)
	a.Airplane = L("plane", 0.5, 0, 1)
	a.CarHonk = a.fromSamples(genHonk(), 0.35)
	a.CarWhoosh = a.fromSamples(genWhoosh(), 0.4)
	a.Boom = a.fromSamples(genBoom(), 0.95)
	a.Engine = a.fromSamples(genEngine(), 0.4)
	if os.Getenv("BATTLE_DEBUG") != "" {
		fmt.Println("sunete incarcate:", len(a.all), "copii:", len(a.aliases))
	}
	// ambianta in bucla
	if data, err := sfxFS.ReadFile("assets/sfx/amb_birds.ogg"); err == nil {
		a.keepMusicData(data)
		a.Birds = rl.LoadMusicStreamFromMemory(".ogg", data, int32(len(data)))
		if data2, err2 := sfxFS.ReadFile("assets/sfx/amb_tunnel.ogg"); err2 == nil {
			a.keepMusicData(data2)
			a.TunnelHum = rl.LoadMusicStreamFromMemory(".ogg", data2, int32(len(data2)))
			a.hasMusic = rl.IsMusicValid(a.Birds) && rl.IsMusicValid(a.TunnelHum)
		}
		if a.hasMusic {
			a.Birds.Looping = true
			a.TunnelHum.Looping = true
			rl.SetMusicVolume(a.Birds, 0)
			rl.SetMusicVolume(a.TunnelHum, 0)
			rl.PlayMusicStream(a.Birds)
			rl.PlayMusicStream(a.TunnelHum)
		}
	}
	return a
}

func (a *GameAudio) load(prefix string, vol, jitter float32, copies int) *Sfx {
	s := &Sfx{Vol: vol, PitchJitter: jitter}
	entries, err := sfxFS.ReadDir("assets/sfx")
	if err != nil {
		return s
	}
	names := []string{}
	for _, e := range entries {
		n := e.Name()
		if n == prefix+".ogg" || (strings.HasPrefix(n, prefix+"_") && strings.HasSuffix(n, ".ogg") && !strings.Contains(strings.TrimPrefix(n, prefix+"_"), "_")) {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	for _, n := range names {
		data, err := sfxFS.ReadFile("assets/sfx/" + n)
		if err != nil {
			continue
		}
		wave := rl.LoadWaveFromMemory(".ogg", data, int32(len(data)))
		if !rl.IsWaveValid(wave) {
			continue
		}
		base := rl.LoadSoundFromWave(wave)
		rl.UnloadWave(wave)
		a.all = append(a.all, base)
		v := []rl.Sound{base}
		for k := 1; k < copies; k++ {
			al := rl.LoadSoundAlias(base)
			a.aliases = append(a.aliases, al)
			v = append(v, al)
		}
		s.Snds = append(s.Snds, v)
	}
	return s
}

func (a *GameAudio) fromSamples(samples []float64, vol float32) *Sfx {
	data := makeWAV(samples)
	wave := rl.LoadWaveFromMemory(".wav", data, int32(len(data)))
	snd := rl.LoadSoundFromWave(wave)
	rl.UnloadWave(wave)
	a.all = append(a.all, snd)
	return &Sfx{Snds: [][]rl.Sound{{snd}}, Vol: vol}
}

// UpdateAmbience: volumul pasarilor (pe uscat), al vuietului din tunel si valurile de la mal.
func (a *GameAudio) UpdateAmbience(dt float32, onLand, inTunnel bool, shoreDist float32, listener rl.Vector3) {
	if a == nil || !a.ready {
		return
	}
	if a.hasMusic {
		tb := float32(0)
		if onLand && !inTunnel {
			tb = 0.32
		}
		th := float32(0)
		if inTunnel {
			th = 0.55
		}
		a.birdsVol += (tb - a.birdsVol) * minf(dt*1.5, 1)
		a.humVol += (th - a.humVol) * minf(dt*2, 1)
		rl.SetMusicVolume(a.Birds, a.birdsVol)
		rl.SetMusicVolume(a.TunnelHum, a.humVol)
		rl.UpdateMusicStream(a.Birds)
		rl.UpdateMusicStream(a.TunnelHum)
	}
	a.waveT -= dt
	if a.waveT <= 0 {
		a.waveT = 2.2 + rand.Float32()*2.6
		if !inTunnel && shoreDist < 40 {
			a.PlayV(a.Wave, 0.25+0.75*(1-shoreDist/40), 0.9+rand.Float32()*0.2)
		}
	}
}

// Muzica de ambianta se opreste cand jocul e in meniu.
func (a *GameAudio) Silence(dt float32) {
	a.UpdateAmbience(dt, false, false, 1e9, rl.Vector3{})
}

func (a *GameAudio) CloseAudio() {
	if a == nil || !a.ready {
		return
	}
	for _, s := range a.aliases {
		rl.UnloadSoundAlias(s)
	}
	for _, snd := range a.all {
		rl.UnloadSound(snd)
	}
	if a.hasMusic {
		rl.UnloadMusicStream(a.Birds)
		rl.UnloadMusicStream(a.TunnelHum)
	}
	a.musicPin.Unpin()
	a.musicData = nil
	rl.CloseAudioDevice()
}

// ---------------- sunete generate (claxon, masina care trece) ----------------

const sampleRate = 44100

func makeWAV(samples []float64) []byte {
	numSamples := len(samples)
	dataSize := numSamples * 2
	buf := make([]byte, 44+dataSize)
	copy(buf[0:4], "RIFF")
	binary.LittleEndian.PutUint32(buf[4:8], uint32(36+dataSize))
	copy(buf[8:12], "WAVE")
	copy(buf[12:16], "fmt ")
	binary.LittleEndian.PutUint32(buf[16:20], 16)
	binary.LittleEndian.PutUint16(buf[20:22], 1)
	binary.LittleEndian.PutUint16(buf[22:24], 1)
	binary.LittleEndian.PutUint32(buf[24:28], sampleRate)
	binary.LittleEndian.PutUint32(buf[28:32], sampleRate*2)
	binary.LittleEndian.PutUint16(buf[32:34], 2)
	binary.LittleEndian.PutUint16(buf[34:36], 16)
	copy(buf[36:40], "data")
	binary.LittleEndian.PutUint32(buf[40:44], uint32(dataSize))
	for i, s := range samples {
		s = math.Max(-1, math.Min(1, s))
		binary.LittleEndian.PutUint16(buf[44+i*2:46+i*2], uint16(int16(s*32767.0)))
	}
	return buf
}

// claxon dublu, cu armonice si un filtru simplu (mai aproape de o masina reala)
func genHonk() []float64 {
	n := int(sampleRate * 0.55)
	out := make([]float64, n)
	lp := 0.0
	for i := 0; i < n; i++ {
		t := float64(i) / sampleRate
		on := (t < 0.18) || (t > 0.26 && t < 0.5)
		if !on {
			lp *= 0.995
			out[i] = lp
			continue
		}
		v := 0.0
		for _, f := range []float64{392, 494} {
			ph := math.Mod(t*f, 1)
			v += (ph*2 - 1) * 0.5 // dinte de fierastrau
		}
		lp += (v - lp) * 0.18
		env := math.Min(1, math.Min(t*80, 1))
		out[i] = lp * 0.5 * env
	}
	return out
}

func genWhoosh() []float64 {
	n := int(sampleRate * 0.9)
	out := make([]float64, n)
	lp, lp2 := 0.0, 0.0
	for i := 0; i < n; i++ {
		t := float64(i) / sampleRate
		env := math.Sin(math.Pi * t / 0.9)
		noise := rand.Float64()*2 - 1
		k := 0.04 + 0.12*env
		lp += (noise - lp) * k
		lp2 += (lp - lp2) * k
		hum := math.Sin(2*math.Pi*(90+40*env)*t) * 0.3
		out[i] = (lp2*2.2 + hum) * env * 0.5
	}
	return out
}

// keepMusicData: memoria unui stream ramane valida cat traieste jocul.
func (a *GameAudio) keepMusicData(data []byte) {
	if len(data) == 0 {
		return
	}
	a.musicData = append(a.musicData, data)
	a.musicPin.Pin(&data[0])
}

// explozie: bubuitura joasa + zgomot care se stinge (fara fisier, generata)
func genBoom() []float64 {
	n := int(sampleRate * 1.6)
	out := make([]float64, n)
	lp, lp2 := 0.0, 0.0
	for i := 0; i < n; i++ {
		t := float64(i) / sampleRate
		env := math.Exp(-t*3.2) * math.Min(1, t*400)
		noise := rand.Float64()*2 - 1
		k := 0.05 + 0.25*math.Exp(-t*6)
		lp += (noise - lp) * k
		lp2 += (lp - lp2) * 0.12
		thump := math.Sin(2*math.Pi*(70*math.Exp(-t*2.5)+28)*t) * math.Exp(-t*4.5)
		out[i] = (lp2*2.6 + lp*0.6 + thump*0.9) * env * 0.75
	}
	return out
}

// motorul masinii: 0,5 s care se repeta fara cusatura (56 Hz = 28 de perioade intregi)
func genEngine() []float64 {
	n := int(sampleRate * 0.5)
	out := make([]float64, n)
	lp := 0.0
	for i := 0; i < n; i++ {
		t := float64(i) / sampleRate
		f := 56.0
		v := math.Sin(2*math.Pi*f*t)*0.55 + math.Sin(2*math.Pi*2*f*t)*0.3 + math.Sin(2*math.Pi*3*f*t)*0.12
		// "pufaitul" cilindrilor
		v *= 0.75 + 0.25*math.Sin(2*math.Pi*4*f*t/2)
		lp += ((rand.Float64()*2 - 1) - lp) * 0.08
		out[i] = (v + lp*0.25) * 0.55
	}
	return out
}

// EngineUpdate: zgomotul motorului cat timp Aventurierul conduce (tonul creste cu viteza).
func (a *GameAudio) EngineUpdate(on bool, speed float32) {
	if a == nil || !a.ready || a.Engine == nil || len(a.Engine.Snds) == 0 {
		return
	}
	snd := a.Engine.Snds[0][0]
	if !on {
		if a.engineOn {
			rl.StopSound(snd)
			a.engineOn = false
		}
		return
	}
	a.engineOn = true
	s := float32(math.Abs(float64(speed)))
	rl.SetSoundPitch(snd, 0.75+s/26)
	rl.SetSoundVolume(snd, 0.22+0.25*minf(s/26, 1))
	if !rl.IsSoundPlaying(snd) {
		rl.PlaySound(snd)
	}
}
