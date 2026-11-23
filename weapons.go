package main

import (
	"fmt"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// V 5.0: CELE 6 ARME DE FOC (se aleg din cabina telefonica)
//   0 pistol laser   - bila albastra, ritm normal
//   1 laser rosu     - raze rosii subtiri, foarte rapide
//   2 blaster mare   - bila portocalie mare, lenta, loveste tare si impinge
//   3 rachete        - racheta cu fum, explodeaza si loveste tot in jur
//   4 raza continua  - cat tii apasat, raza mov loveste tinta (nu rateaza)
//   5 tun cu bule    - 5 bule in evantai, bune de aproape
// ====================================================================

const (
	GunPistol = iota
	GunRed
	GunBlaster
	GunRocket
	GunBeam
	GunBubble
)

type gunSpec struct {
	CD     float32
	Speed  float32
	Dmg    float64 // inmultit cu atacul aventurierului
	Size   float32
	Col    rl.Color
	Glow   rl.Color
	Splash float32
}

var gunSpecs = [NumGuns]gunSpec{
	{CD: 0.20, Speed: 62, Dmg: 0.60, Size: 0.22, Col: rl.NewColor(220, 250, 255, 255), Glow: rl.NewColor(80, 210, 255, 110)},
	{CD: 0.11, Speed: 95, Dmg: 0.40, Size: 0.14, Col: rl.NewColor(255, 220, 220, 255), Glow: rl.NewColor(255, 40, 40, 140)},
	{CD: 0.55, Speed: 46, Dmg: 1.75, Size: 0.5, Col: rl.NewColor(255, 240, 160, 255), Glow: rl.NewColor(255, 140, 30, 130)},
	{CD: 0.85, Speed: 40, Dmg: 2.4, Size: 0.3, Col: rl.NewColor(240, 60, 50, 255), Glow: rl.NewColor(255, 200, 80, 120), Splash: 6.5},
	{CD: 0.08, Speed: 0, Dmg: 0.27, Size: 0, Col: rl.NewColor(240, 200, 255, 255), Glow: rl.NewColor(190, 90, 255, 120)},
	{CD: 0.45, Speed: 34, Dmg: 0.26, Size: 0.32, Col: rl.NewColor(210, 240, 255, 200), Glow: rl.NewColor(255, 150, 230, 90)},
}

func (g *Game) gunDamage(k int) float64 {
	p := &g.Player
	return (p.Attack + g.BonusDamage) * gunSpecs[k].Dmg * (1 + 0.06*float64(k))
}

// fireGun: trage cu arma aleasa (inlocuieste shootProjectile din V 4).
func (g *Game) fireGun() {
	k := clampInt(g.Prog.Gun, 0, NumGuns-1)
	sp := gunSpecs[k]
	g.ShootCD = sp.CD
	p := &g.Player
	p.Pose.Attack = 0.15
	if k == GunBeam {
		g.beamTick()
		return
	}
	origin := g.muzzle()
	dir := rl.Vector3Normalize(rl.Vector3Subtract(g.AimPoint, origin))
	speed := sp.Speed
	life := float32(2.4)
	if g.AimKind != TgtNone {
		speed = maxf(speed, laserFast)
		life = rl.Vector3Distance(g.AimPoint, origin)/speed + 0.8
	}
	shotKind := ProjLaser
	if k == 0 {
		shotKind = g.projKind()
	}
	add := func(d rl.Vector3, dmg float64, target, car int) {
		g.Projectiles = append(g.Projectiles, Projectile3D{
			Position: origin, Velocity: rl.Vector3Scale(d, speed), Color: sp.Col, Life: life,
			Damage: dmg, Target: target, CarTgt: car + 1, TgtKind: g.AimKind, TgtIdx: g.AimIdx,
			Gun: k, Size: sp.Size, Splash: sp.Splash,
			Kind: shotKind,
		})
	}
	if k == GunBubble {
		right := rl.Vector3Normalize(rl.Vector3CrossProduct(dir, v3(0, 1, 0)))
		for i := -2; i <= 2; i++ {
			d := rl.Vector3Normalize(rl.Vector3Add(dir, rl.Vector3Add(rl.Vector3Scale(right, float32(i)*0.07), v3(0, (rand.Float32()-0.5)*0.05, 0))))
			tgt := -1
			if i == 0 {
				tgt = g.AimEnemy
			}
			add(d, g.gunDamage(k), tgt, -1)
		}
	} else {
		add(dir, g.gunDamage(k), g.AimEnemy, g.AimCar)
	}
	for n := 0; n < 6; n++ {
		g.spawnParticle(origin, v3((rand.Float32()-0.5)*2+dir.X*3, (rand.Float32()-0.5)*2+dir.Y*3, (rand.Float32()-0.5)*2+dir.Z*3), sp.Col, 0.12, 0.15)
	}
	switch k {
	case GunBlaster:
		g.Audio.PlayV(g.Audio.Shoot, 1, 0.7)
		g.CamShake = maxf(g.CamShake, 0.15)
	case GunRocket:
		g.Audio.PlayV(g.Audio.CarWhoosh, 0.9, 1.3)
	case GunBubble:
		g.Audio.PlayV(g.Audio.Spring, 0.7, 1.6)
	case GunRed:
		g.Audio.PlayV(g.Audio.Shoot, 0.8, 1.35)
	default:
		g.Play(g.Audio.Shoot)
	}
}

// beamTick: raza continua loveste ce e in tinta (cu ajutor la tintire).
func (g *Game) beamTick() {
	g.BeamT = 0.14
	origin := g.muzzle()
	g.BeamEnd = g.AimPoint
	dmg := g.gunDamage(GunBeam)
	if rand.Float32() < 0.35 {
		g.Audio.PlayV(g.Audio.ShotHit, 0.35, 1.8)
	}
	switch {
	case g.AimEnemy >= 0 && g.AimEnemy < len(g.Enemies) && g.Enemies[g.AimEnemy].Alive:
		e := &g.Enemies[g.AimEnemy]
		if rl.Vector3Distance(origin, e.Pos) > 80 {
			return
		}
		g.BeamEnd = v3(e.Pos.X, e.Pos.Y+1.4*e.Scale, e.Pos.Z)
		g.damageEnemy(g.AimEnemy, dmg, g.BeamEnd, false)
	case g.AimCar >= 0 && g.AimCar < len(g.City.Cars) && g.City.Cars[g.AimCar].Hittable():
		c := &g.City.Cars[g.AimCar]
		g.BeamEnd = v3(c.Pos.X, c.Pos.Y+0.9, c.Pos.Z)
		g.beamCarAcc += 0.34
		if g.beamCarAcc >= 1 {
			g.beamCarAcc = 0
			g.carHit(g.AimCar, 1, g.BeamEnd)
		}
	case g.AimKind != TgtNone:
		if c, ok := g.thingCenter(g.AimKind, g.AimIdx); ok {
			g.BeamEnd = c
			g.beamThingAcc += 0.34
			if g.beamThingAcc >= 1 {
				g.beamThingAcc = 0
				g.hitThing(g.AimKind, g.AimIdx, c)
			}
		}
	}
	for n := 0; n < 3; n++ {
		g.spawnParticle(g.BeamEnd, v3((rand.Float32()-0.5)*4, rand.Float32()*3, (rand.Float32()-0.5)*4), gunSpecs[GunBeam].Col, 0.14, 0.35)
	}
}

// damageEnemy: o lovitura de arma de foc (comuna pentru proiectile, raza, explozii).
func (g *Game) damageEnemy(k int, dmg float64, at rl.Vector3, push bool) {
	e := &g.Enemies[k]
	if !e.Alive {
		return
	}
	e.HP -= dmg
	e.Flash = 0.15
	e.Alert = true
	g.CamShake = maxf(g.CamShake, 0.12)
	g.HitStreak++
	if g.HitStreak >= 5 {
		g.Achieve.Unlock("sharpshooter")
	}
	if g.floatCD <= 0 {
		g.addFloat(v3(e.Pos.X, e.Pos.Y+3*e.Scale, e.Pos.Z), fmt.Sprintf("-%.0f", dmg), LegoOrange, 30)
		g.floatCD = 0.12
	}
	g.PlayAt(g.Audio.ShotHit, e.Pos, 1, 60)
	if push && !e.Boss {
		d := rl.Vector3Normalize(v3(e.Pos.X-g.Player.Pos.X, 0, e.Pos.Z-g.Player.Pos.Z))
		e.DodgeDir = d
		e.DodgeT = 0.25
		e.Stun = maxf(e.Stun, 0.4)
	}
	if e.HP <= 0 {
		g.defeatEnemy(k, 100)
	}
}

// explodeAt: explozia rachetei (loveste tot ce e in raza).
func (g *Game) explodeAt(c rl.Vector3, r float32, dmg float64) {
	g.PlayAt(g.Audio.Boom, c, 0.9, 120)
	g.CamShake = maxf(g.CamShake, 0.35)
	for n := 0; n < 46; n++ {
		col := []rl.Color{LegoOrange, LegoYellow, LegoRed, rl.NewColor(90, 90, 95, 255)}[n%4]
		g.spawnParticle(c, v3((rand.Float32()-0.5)*14, rand.Float32()*10, (rand.Float32()-0.5)*14), col, 0.35, 0.9)
	}
	for k := range g.Enemies {
		e := &g.Enemies[k]
		if !e.Alive {
			continue
		}
		d := rl.Vector3Distance(c, v3(e.Pos.X, e.Pos.Y+1.2*e.Scale, e.Pos.Z))
		if d < r+0.8*e.Scale {
			f := 1 - 0.5*clampf(d/r, 0, 1)
			g.damageEnemy(k, dmg*float64(f), c, true)
		}
	}
	for k := range g.City.Cars {
		cc := &g.City.Cars[k]
		if cc.Hittable() && rl.Vector3Distance(c, cc.Pos) < r+1.5 {
			g.carHit(k, 2, c)
		}
	}
	if kind, idx := g.thingAt(c); kind != TgtNone {
		g.hitThing(kind, idx, c)
	}
}

func (g *Game) updateBeam(dt float32) {
	if g.BeamT > 0 {
		g.BeamT -= dt
	}
	if g.floatCD > 0 {
		g.floatCD -= dt
	}
}

func (g *Game) drawBeam() {
	if g.BeamT <= 0 || g.Weapon != WeaponGun {
		return
	}
	a := g.muzzle()
	b := g.BeamEnd
	k := clampf(g.BeamT/0.14, 0, 1)
	w := 0.09 + 0.03*sinf(g.Time*60)
	rl.DrawCylinderEx(a, b, w, w, 6, rl.NewColor(250, 230, 255, uint8(230*k)))
	rl.DrawCylinderEx(a, b, w*3.2, w*3.2, 8, rl.NewColor(180, 80, 255, uint8(90*k)))
	rl.DrawSphereEx(b, 0.45+0.15*sinf(g.Time*40), 8, 8, rl.NewColor(220, 160, 255, uint8(160*k)))
}

// drawPlayerShot: cum arata fiecare proiectil al aventurierului.
func drawPlayerShot(pr *Projectile3D) {
	if drawThemedShot(pr) {
		return
	}
	sp := gunSpecs[clampInt(pr.Gun, 0, NumGuns-1)]
	switch pr.Gun {
	case GunRed:
		d := rl.Vector3Normalize(pr.Velocity)
		tail := rl.Vector3Subtract(pr.Position, rl.Vector3Scale(d, 1.6))
		rl.DrawCylinderEx(tail, pr.Position, 0.06, 0.06, 6, sp.Col)
		rl.DrawCylinderEx(tail, pr.Position, 0.16, 0.16, 6, sp.Glow)
	case GunRocket:
		d := rl.Vector3Normalize(pr.Velocity)
		back := rl.Vector3Subtract(pr.Position, rl.Vector3Scale(d, 0.9))
		rl.DrawCylinderEx(back, pr.Position, 0.16, 0.16, 8, rl.NewColor(80, 170, 70, 255))
		rl.DrawCylinderEx(pr.Position, rl.Vector3Add(pr.Position, rl.Vector3Scale(d, 0.35)), 0.16, 0.02, 8, sp.Col)
		rl.DrawSphereEx(back, 0.22, 6, 6, rl.NewColor(255, 200, 80, 220))
	case GunBubble:
		rl.DrawSphereEx(pr.Position, pr.Size, 10, 10, sp.Col)
		rl.DrawSphereEx(v3(pr.Position.X-pr.Size*0.35, pr.Position.Y+pr.Size*0.35, pr.Position.Z), pr.Size*0.25, 6, 6, rl.NewColor(255, 255, 255, 230))
	default:
		rl.DrawSphereEx(pr.Position, pr.Size, 8, 8, sp.Col)
		rl.DrawSphereEx(pr.Position, pr.Size*1.9, 8, 8, sp.Glow)
	}
}
