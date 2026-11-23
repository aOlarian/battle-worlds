package main

import (
	"fmt"
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// LUPTA IN TIMP REAL (fara ecran separat de lupta, camera nu se blocheaza)
//   - departe: pistol laser, tinta din mijlocul ecranului (ajutor la tintire);
//   - aproape de un inamic sau de un santier blocuri: sabia apare singura;
//   - inamicii vin si lovesc (cu pregatire vizibila) sau trag cu bile de foc.
// ====================================================================

const (
	WeaponNone  = 0
	WeaponGun   = 1
	WeaponSword = 2
)

const swingTime = float32(0.34)

func (g *Game) nearestEnemy() (int, float32) {
	p := g.Player.Pos
	best, bd := -1, float32(1e9)
	for i := range g.Enemies {
		e := &g.Enemies[i]
		if !e.Alive || absf(e.Pos.Y-p.Y) > 4 {
			continue
		}
		dx, dz := e.Pos.X-p.X, e.Pos.Z-p.Z
		if d := sqrtf(dx*dx + dz*dz); d < bd {
			best, bd = i, d
		}
	}
	return best, bd
}

// nearestSite: santierul la care Aventurierul poate construi lovind cu sabia
func (g *Game) nearestSite() int {
	p := g.Player.Pos
	best, bd := -1, float32(1e9)
	for i := range g.City.LegoSites {
		s := &g.City.LegoSites[i]
		if s.CurrentStage >= s.TotalStages && s.UpgradeLevel >= 5 {
			continue
		}
		dx, dz := p.X-s.Position.X, p.Z-s.Position.Z
		d := sqrtf(dx*dx + dz*dz)
		reach := maxf(s.Meshes.Radius*s.Scale(), 3) + 4.5
		if d < reach && d-reach < bd && absf(p.Y-s.Position.Y) < 3 {
			best, bd = i, d-reach
		}
	}
	for _, v := range g.City.VendorPods {
		dx, dz := p.X-v.Position.X, p.Z-v.Position.Z
		if dx*dx+dz*dz < 4.5*4.5 {
			s := &g.City.LegoSites[v.SlotIndex]
			if s.CurrentStage < s.TotalStages || s.UpgradeLevel < 5 {
				return v.SlotIndex
			}
		}
	}
	return best
}

// computeAim: ce e in mijlocul ecranului (inamic sau perete), cu ajutor la tintire.
func (g *Game) computeAim() {
	w := g.City.W
	cam := g.Camera
	dir := rl.Vector3Normalize(rl.Vector3Subtract(cam.Target, cam.Position))
	maxD := float32(150)
	hitD := maxD
	for s := float32(2); s < maxD; s += 0.8 {
		q := rl.Vector3Add(cam.Position, rl.Vector3Scale(dir, s))
		gy, _ := w.GroundAt(q.X, q.Z, q.Y)
		if q.Y < gy || w.Solid(q.X, q.Z, q.Y, 0) || w.tunnelCeilingHit(q) {
			hitD = s
			break
		}
	}
	g.AimEnemy = -1
	bestScore := float32(1e9)
	for i := range g.Enemies {
		e := &g.Enemies[i]
		if !e.Alive {
			continue
		}
		chest := v3(e.Pos.X, e.Pos.Y+1.4*e.Scale, e.Pos.Z)
		to := rl.Vector3Subtract(chest, cam.Position)
		along := rl.Vector3DotProduct(to, dir)
		if along < 1 || along > hitD+1.5 || along > 90 {
			continue
		}
		perp := rl.Vector3Length(rl.Vector3Subtract(to, rl.Vector3Scale(dir, along)))
		// tinta "lipicioasa": raza efectiva creste cu distanta (ajutor pentru copii)
		allow := 1.3*e.Scale + along*0.045
		if perp < allow && perp/allow+along*0.002 < bestScore {
			bestScore = perp/allow + along*0.002
			g.AimEnemy = i
		}
	}
	// masinile se pot tinti si ele (dau aur)
	g.AimCar = -1
	if g.AimEnemy < 0 {
		bestCar := float32(1e9)
		for k := range g.City.Cars {
			c := &g.City.Cars[k]
			if !c.Hittable() || (g.Drive.Active && g.Drive.Car == k) {
				continue
			}
			ctr := v3(c.Pos.X, c.Pos.Y+0.9, c.Pos.Z)
			to := rl.Vector3Subtract(ctr, cam.Position)
			along := rl.Vector3DotProduct(to, dir)
			if along < 1 || along > hitD+2.5 || along > 90 {
				continue
			}
			perp := rl.Vector3Length(rl.Vector3Subtract(to, rl.Vector3Scale(dir, along)))
			allow := 1.7 + along*0.03
			if perp < allow && along < bestCar {
				bestCar = along
				g.AimCar = k
			}
		}
	}
	// barcile, avioanele si pescarusii (cand nu e niciun inamic sau masina in tinta)
	g.AimKind, g.AimIdx = TgtNone, -1
	if g.AimEnemy < 0 && g.AimCar < 0 {
		g.aimThings(cam.Position, dir, hitD, maxD)
	}
	if g.AimEnemy >= len(g.Enemies) {
		g.AimEnemy = -1
	}
	if g.AimEnemy >= 0 {
		e := &g.Enemies[g.AimEnemy]
		g.AimPoint = v3(e.Pos.X, e.Pos.Y+1.4*e.Scale, e.Pos.Z)
	} else if g.AimCar >= 0 {
		c := &g.City.Cars[g.AimCar]
		g.AimPoint = v3(c.Pos.X, c.Pos.Y+0.9, c.Pos.Z)
	} else if c, ok := g.thingCenter(g.AimKind, g.AimIdx); ok {
		g.AimPoint = c
	} else {
		g.AimPoint = rl.Vector3Add(cam.Position, rl.Vector3Scale(dir, hitD))
	}
}

func (g *Game) updateWeapons(dt float32) {
	p := &g.Player
	if g.ShootCD > 0 {
		g.ShootCD -= dt
	}
	if g.WeaponT > 0 {
		g.WeaponT -= dt
	}
	g.NearSite = g.nearestSite()
	g.NearVendor = -1
	for i, v := range g.City.VendorPods {
		dx, dz := p.Pos.X-v.Position.X, p.Pos.Z-v.Position.Z
		if dx*dx+dz*dz < 81 && v.SlotIndex == g.NearSite {
			g.NearVendor = i
		}
	}
	if g.NearVendor < 0 && g.NearSite >= 0 {
		for i, v := range g.City.VendorPods {
			if v.SlotIndex == g.NearSite {
				g.NearVendor = i
			}
		}
	}
	if g.NearSite != g.UpgradeSite {
		g.UpgradeHits = 0
	}
	// ce arma are in mana
	_, nd := g.nearestEnemy()
	want := WeaponGun
	if g.NearSite >= 0 || nd < 5.5 || (g.Weapon == WeaponSword && nd < 7.5) {
		want = WeaponSword
	}
	if g.Swimming || g.Ride.Kind == RideWheel || g.Ride.Kind == RideCoaster || g.Drive.Active {
		want = WeaponNone
	}
	if want != g.Weapon {
		g.Weapon = want
		g.WeaponT = 0.25
		g.SwingT = 0
		switch want {
		case WeaponSword:
			g.Play(g.Audio.DrawSword)
		case WeaponGun:
			g.Play(g.Audio.DrawGun)
		}
	}
	p.Pose.Weapon = g.Weapon
	p.Pose.Raise = 1 - clampf(g.WeaponT/0.25, 0, 1)
	p.Pose.AimPitch = g.CamPitch
	g.computeAim()

	if g.SwingT > 0 {
		prev := g.SwingT
		g.SwingT -= dt
		// lovitura "intra" la mijlocul miscarii
		if prev > swingTime*0.5 && g.SwingT <= swingTime*0.5 {
			g.swordStrike()
		}
		g.swordTrail()
	}
	p.Pose.Swing = 0
	if g.SwingT > 0 {
		p.Pose.Swing = 1 - g.SwingT/swingTime
		p.Pose.SwingSide = g.SwingSide
	}
	if g.Weapon == WeaponNone || g.WeaponT > 0.05 || g.Fade > 0 {
		return
	}
	fire := rl.IsMouseButtonDown(rl.MouseButtonLeft)
	if g.Debug != nil {
		fire = g.Debug.Fire(g)
	}
	if !fire {
		return
	}
	switch g.Weapon {
	case WeaponGun:
		if g.ShootCD <= 0 {
			g.fireGun()
		}
	case WeaponSword:
		if g.SwingT <= 0 {
			g.SwingT = swingTime
			g.SwingSide = -g.SwingSide
			if g.SwingSide == 0 {
				g.SwingSide = 1
			}
			// se intoarce spre inamicul cel mai apropiat (ajutor)
			if i, d := g.nearestEnemy(); i >= 0 && d < 4.5 {
				e := &g.Enemies[i]
				g.CamYaw = turnTo(g.CamYaw, atan2f(-(e.Pos.X-p.Pos.X), -(e.Pos.Z-p.Pos.Z)), 0.35)
			}
			g.Play(g.Audio.Swing)
		}
	}
}

func (g *Game) muzzle() rl.Vector3 {
	p := &g.Player
	_, fwd, right := g.camVectors()
	f := rl.Vector3Normalize(v3(fwd.X, 0, fwd.Z))
	m := v3(p.Pos.X, p.Pos.Y+1.95, p.Pos.Z)
	m = rl.Vector3Add(m, rl.Vector3Scale(right, 0.5))
	return rl.Vector3Add(m, rl.Vector3Scale(f, 1.0))
}

func (g *Game) shootProjectile() {
	p := &g.Player
	g.ShootCD = 0.2
	origin := g.muzzle()
	dir := rl.Vector3Normalize(rl.Vector3Subtract(g.AimPoint, origin))
	p.Pose.Attack = 0.15
	speed, life := float32(62), float32(2.4)
	if g.AimKind != TgtNone {
		// spre avion / pescarus / barca: laser rapid, care ajunge si la 600 m
		speed = laserFast
		life = rl.Vector3Distance(g.AimPoint, origin)/speed + 0.8
	}
	g.Projectiles = append(g.Projectiles, Projectile3D{
		Position: origin, Velocity: rl.Vector3Scale(dir, speed), Color: LegoCyan, Life: life,
		Damage: (p.Attack + g.BonusDamage) * 0.6, Target: g.AimEnemy,
		CarTgt: g.AimCar + 1, TgtKind: g.AimKind, TgtIdx: g.AimIdx,
	})
	for n := 0; n < 5; n++ {
		g.spawnParticle(origin, v3((rand.Float32()-0.5)*2+dir.X*3, (rand.Float32()-0.5)*2+dir.Y*3, (rand.Float32()-0.5)*2+dir.Z*3), rl.NewColor(200, 250, 255, 255), 0.12, 0.15)
	}
	g.Play(g.Audio.Shoot)
}

// swordStrike: loveste tot ce e in fata, aproape; daca nu e niciun inamic, construieste.
func (g *Game) swordStrike() {
	p := &g.Player
	fx, fz := sinf(p.Yaw), cosf(p.Yaw)
	hit := false
	for i := range g.Enemies {
		e := &g.Enemies[i]
		if !e.Alive || absf(e.Pos.Y-p.Pos.Y) > 3 {
			continue
		}
		dx, dz := e.Pos.X-p.Pos.X, e.Pos.Z-p.Pos.Z
		d := sqrtf(dx*dx + dz*dz)
		reach := 2.6 + 0.7*e.Scale
		if d > reach {
			continue
		}
		if d > 0.8 && (dx*fx+dz*fz)/d < 0.15 {
			continue
		}
		hit = true
		dmg := (p.Attack + g.BonusDamage) * 1.4
		e.HP -= dmg
		e.Flash = 0.18
		e.Stun = 0.4
		e.Windup = 0
		if d > 0.01 {
			e.DodgeDir = v3(dx/d, 0, dz/d)
			e.DodgeT = 0.16
		}
		g.CamShake = 0.45
		g.addFloat(v3(e.Pos.X, e.Pos.Y+3*e.Scale, e.Pos.Z), fmt.Sprintf("-%.0f", dmg), LegoOrange, 32)
		g.PlayAt(g.Audio.SwordHit, e.Pos, 1, 40)
		chest := v3(e.Pos.X, e.Pos.Y+1.4*e.Scale, e.Pos.Z)
		for n := 0; n < 16; n++ {
			g.spawnParticle(chest, v3((rand.Float32()-0.5)*7, rand.Float32()*5, (rand.Float32()-0.5)*7), []rl.Color{LegoCyan, LegoWhite, LegoYellow}[n%3], 0.16, 0.45)
		}
		if e.HP <= 0 {
			g.defeatEnemy(i, 150)
		}
	}
	// si masinile din fata (dau aur)
	for k := range g.City.Cars {
		c := &g.City.Cars[k]
		if !c.Hittable() || absf(c.Pos.Y-p.Pos.Y) > 2.5 {
			continue
		}
		a, s := carFrame(c, p.Pos.X, p.Pos.Z)
		da, ds := maxf(absf(a)-carHalfLen, 0), maxf(absf(s)-carHalfWidth, 0)
		if sqrtf(da*da+ds*ds) > 2.2 {
			continue
		}
		dx, dz := c.Pos.X-p.Pos.X, c.Pos.Z-p.Pos.Z
		if dl := sqrtf(dx*dx + dz*dz); dl > 0.1 && (dx*fx+dz*fz)/dl < -0.2 {
			continue
		}
		hit = true
		g.CamShake = 0.35
		g.carHit(k, 1.4, v3(c.Pos.X, c.Pos.Y+1.1, c.Pos.Z))
	}
	if hit || g.NearSite < 0 {
		return
	}
	g.hitSite(g.NearSite)
}

func (g *Game) hitSite(idx int) {
	site := &g.City.LegoSites[idx]
	g.PlayAt(g.Audio.Build, site.Position, 1, 40)
	top := v3(site.Position.X, site.Position.Y+1.5, site.Position.Z)
	for n := 0; n < 10; n++ {
		g.spawnParticle(top, v3((rand.Float32()-0.5)*5, rand.Float32()*5+2, (rand.Float32()-0.5)*5), []rl.Color{LegoGold, LegoWhite, LegoRed}[n%3], 0.2, 0.6)
	}
	if site.CurrentStage < site.TotalStages {
		g.placeBrick(site)
		return
	}
	if site.UpgradeLevel < 5 {
		if g.UpgradeSite != idx {
			g.UpgradeSite = idx
			g.UpgradeHits = 0
		}
		g.UpgradeHits++
		if g.UpgradeHits >= 5 {
			g.UpgradeHits = 0
			g.upgradeSite(site)
		} else {
			g.sayIcon(IconGrow, fmt.Sprintf("Mărire la %dx (%d aur): mai lovește de %d ori!", site.UpgradeLevel+1, 250*site.UpgradeLevel, 5-g.UpgradeHits))
		}
	}
}

// urma luminoasa a sabiei
func (g *Game) swordTrail() {
	p := &g.Player
	t := 1 - g.SwingT/swingTime
	a := (t - 0.5) * 2.4 * g.SwingSide
	yaw := p.Yaw + a
	tip := v3(p.Pos.X+sinf(yaw)*2.2, p.Pos.Y+1.6+0.9*(0.5-t), p.Pos.Z+cosf(yaw)*2.2)
	for k := 0; k < 2; k++ {
		g.spawnParticle(tip, v3((rand.Float32()-0.5)*0.4, (rand.Float32()-0.5)*0.4, (rand.Float32()-0.5)*0.4), rl.NewColor(150, 240, 255, 200), 0.16, 0.18)
	}
}

func (g *Game) updateProjectiles(dt float32) {
	_, _, dodgeChance := g.enemyMultipliers()
	w := g.City.W
	p := &g.Player
	for i := len(g.Projectiles) - 1; i >= 0; i-- {
		pr := &g.Projectiles[i]
		prev := pr.Position
		// putina "cautare" a tintei asistate, ca sa nu rateze la limita
		if !pr.Enemy && pr.Target >= 0 && pr.Target < len(g.Enemies) && g.Enemies[pr.Target].Alive {
			e := &g.Enemies[pr.Target]
			to := rl.Vector3Normalize(rl.Vector3Subtract(v3(e.Pos.X, e.Pos.Y+1.4*e.Scale, e.Pos.Z), pr.Position))
			sp := rl.Vector3Length(pr.Velocity)
			pr.Velocity = rl.Vector3Scale(rl.Vector3Normalize(rl.Vector3Lerp(rl.Vector3Scale(pr.Velocity, 1/sp), to, minf(dt*6, 1))), sp)
		}
		if !pr.Enemy && pr.CarTgt > 0 && pr.CarTgt <= len(g.City.Cars) && g.City.Cars[pr.CarTgt-1].Hittable() {
			c := &g.City.Cars[pr.CarTgt-1]
			to := rl.Vector3Normalize(rl.Vector3Subtract(v3(c.Pos.X, c.Pos.Y+0.9, c.Pos.Z), pr.Position))
			sp := rl.Vector3Length(pr.Velocity)
			pr.Velocity = rl.Vector3Scale(rl.Vector3Normalize(rl.Vector3Lerp(rl.Vector3Scale(pr.Velocity, 1/sp), to, minf(dt*8, 1))), sp)
		}
		if !pr.Enemy && pr.TgtKind != TgtNone {
			if c, ok := g.thingCenter(pr.TgtKind, pr.TgtIdx); ok {
				to := rl.Vector3Normalize(rl.Vector3Subtract(c, pr.Position))
				sp := rl.Vector3Length(pr.Velocity)
				pr.Velocity = rl.Vector3Scale(rl.Vector3Normalize(rl.Vector3Lerp(rl.Vector3Scale(pr.Velocity, 1/sp), to, minf(dt*12, 1))), sp)
			}
		}
		pr.Position = rl.Vector3Add(pr.Position, rl.Vector3Scale(pr.Velocity, dt))
		pr.Life -= dt
		hit := false
		if pr.Enemy {
			if rand.Float32() < 0.8 {
				g.spawnParticle(prev, v3((rand.Float32()-0.5)*0.8, rand.Float32()*0.8, (rand.Float32()-0.5)*0.8), rl.NewColor(255, 150, 40, 255), 0.16, 0.3)
			}
			chest := v3(p.Pos.X, p.Pos.Y+1.4, p.Pos.Z)
			if rl.Vector3Distance(pr.Position, chest) < 1.05 {
				hit = true
				g.hurtPlayer(pr.Damage, prev)
			}
		} else {
			if rand.Float32() < 0.7 {
				g.spawnParticle(prev, v3((rand.Float32()-0.5)*0.6, (rand.Float32()-0.5)*0.6, (rand.Float32()-0.5)*0.6), gunSpecs[clampInt(pr.Gun, 0, NumGuns-1)].Glow, 0.12, 0.25)
				if pr.Gun == GunRocket {
					g.spawnParticle(prev, v3((rand.Float32()-0.5)*0.5, 0.6+rand.Float32()*0.5, (rand.Float32()-0.5)*0.5), rl.NewColor(225, 225, 230, 200), 0.32, 0.7)
				}
			}
			for k := range g.Enemies {
				e := &g.Enemies[k]
				if !e.Alive {
					continue
				}
				chest := v3(e.Pos.X, e.Pos.Y+1.4*e.Scale, e.Pos.Z)
				d := rl.Vector3Distance(pr.Position, chest)
				// eschiva: inamicul vede proiectilul venind si sare in lateral
				if d < 9 && d > 3 && e.DodgeCD <= 0 && e.DodgeT <= 0 && pr.Target != k {
					e.DodgeCD = 3
					if rand.Float64() < dodgeChance {
						side := rl.Vector3Normalize(v3(-pr.Velocity.Z, 0, pr.Velocity.X))
						if rand.Float32() < 0.5 {
							side = rl.Vector3Negate(side)
						}
						e.DodgeDir = side
						e.DodgeT = 0.35
						g.addFloat(v3(e.Pos.X, e.Pos.Y+3.2*e.Scale, e.Pos.Z), "Eschivă!", LegoYellow, 26)
						g.HitStreak = 0
					}
				}
				if d < 1.3*e.Scale+pr.Size*0.6 {
					hit = true
					for n := 0; n < 10; n++ {
						g.spawnParticle(pr.Position, v3((rand.Float32()-0.5)*6, rand.Float32()*5, (rand.Float32()-0.5)*6), gunSpecs[clampInt(pr.Gun, 0, NumGuns-1)].Col, 0.18, 0.5)
					}
					if pr.Splash > 0 {
						g.explodeAt(pr.Position, pr.Splash, pr.Damage)
						pr.Splash = 0
					} else {
						g.damageEnemy(k, pr.Damage, pr.Position, pr.Gun == GunBlaster)
						if pr.Kind == ProjWeb && k < len(g.Enemies) && g.Enemies[k].Alive {
							g.Enemies[k].Stun = maxf(g.Enemies[k].Stun, 2.5) // prins in panza
						}
					}
					break
				}
			}
			if !hit {
				// barci, avioane, pescarusi: verificam tot drumul facut in cadrul asta (laserul e rapid)
				for k := 1; k <= 6 && !hit; k++ {
					q := rl.Vector3Lerp(prev, pr.Position, float32(k)/6)
					if kind, idx := g.thingAt(q); kind != TgtNone {
						hit = true
						g.hitThing(kind, idx, q)
					}
				}
			}
			if !hit {
				mid := rl.Vector3Lerp(prev, pr.Position, 0.5)
				for _, q := range []rl.Vector3{mid, pr.Position} {
					if k := g.carAt(q, 0.15); k >= 0 {
						hit = true
						g.carHit(k, 1, q)
						break
					}
				}
			}
		}
		if !hit {
			gy, _ := w.GroundAt(pr.Position.X, pr.Position.Z, pr.Position.Y)
			if pr.Position.Y < gy || w.Solid(pr.Position.X, pr.Position.Z, pr.Position.Y, 0) || w.tunnelCeilingHit(pr.Position) {
				hit = true
				c := LegoYellow
				if pr.Enemy {
					c = LegoOrange
				}
				for n := 0; n < 8; n++ {
					g.spawnParticle(prev, v3((rand.Float32()-0.5)*4, rand.Float32()*4, (rand.Float32()-0.5)*4), c, 0.15, 0.4)
				}
			}
		}
		if hit && !pr.Enemy && pr.Splash > 0 {
			g.explodeAt(pr.Position, pr.Splash, pr.Damage)
		}
		if hit || pr.Life <= 0 {
			if !hit && pr.Life <= 0 && !pr.Enemy {
				g.HitStreak = 0
			}
			g.Projectiles = append(g.Projectiles[:i], g.Projectiles[i+1:]...)
		}
	}
}

func (g *Game) hurtPlayer(dmg float64, from rl.Vector3) {
	p := &g.Player
	if g.Invuln > 0 || g.Scene != SceneMapExplore || g.Drive.Active {
		return
	}
	dmg = math.Max(dmg-p.Defense/3-g.BonusShield, 4)
	p.HP -= dmg
	g.Invuln = 0.6
	g.HurtFlash = 0.45
	g.SinceHurt = 0
	g.CamShake = 0.8
	g.Play(g.Audio.Hurt)
	g.addFloat(v3(p.Pos.X, p.Pos.Y+3, p.Pos.Z), fmt.Sprintf("-%.0f", dmg), LegoRed, 32)
	dx, dz := p.Pos.X-from.X, p.Pos.Z-from.Z
	if d := sqrtf(dx*dx + dz*dz); d > 0.01 && g.Ride.Kind == RideNone {
		for k := 0; k < 4; k++ {
			g.City.W.MovePlayer(&p.Pos, dx/d*0.25, dz/d*0.25, 0.45, p.Pos.Y)
		}
	}
	if g.Debug != nil {
		g.Debug.Hits++
	}
	if p.HP <= 0 {
		p.HP = 0
		g.Play(g.Audio.Death)
		g.Scene = SceneGameOver
		g.unlockCursor()
	}
}

func (g *Game) defeatEnemy(k int, score int64) {
	e := &g.Enemies[k]
	if !e.Alive {
		return
	}
	e.HP = 0
	e.Alive = false
	g.Player.Gold += e.Gold
	g.Player.Silver += e.Silver
	g.Score += score
	g.Achieve.Unlock("first_blood")
	g.Play(g.Audio.Victory)
	g.Play(g.Audio.Coin)
	g.addFloat(v3(e.Pos.X, e.Pos.Y+3.5*e.Scale, e.Pos.Z), fmt.Sprintf("+%d aur", e.Gold), LegoGoldBright, 34)
	for n := 0; n < 40; n++ {
		c := []rl.Color{LegoGold, LegoGoldBright, LegoWhite, enemyColors[e.Kind%len(enemyColors)]}[n%4]
		g.spawnParticle(v3(e.Pos.X, e.Pos.Y+1.5, e.Pos.Z), v3((rand.Float32()-0.5)*9, rand.Float32()*9+2, (rand.Float32()-0.5)*9), c, 0.3, 1.4)
	}
	if g.Debug != nil {
		g.Debug.Kills++
	}
	g.afterDefeat()
}

// levelComplete: se face la sfarsitul cadrului (nu in mijlocul buclelor peste monstri/proiectile)
func (g *Game) levelComplete() { g.pendingLevelUp = true }

func (g *Game) doLevelComplete() {
	g.pendingLevelUp = false
	g.Level++
	if g.Level > 5 {
		g.Achieve.Unlock("champion")
	}
	g.Play(g.Audio.LevelUp)
	g.Play(g.Audio.Victory)
	g.Enemies = g.Enemies[:0]
	g.AimEnemy = -1 // altfel tinta (crucea) citeste Seful sters si jocul cade
	if wi := g.Level - 1; wi < NumWorlds {
		g.startReveal(wi)
	} else if wi == NumWorlds {
		// V 6.0: ultima insula a lumii: vine trenul magic spre lumea urmatoare
		g.startReveal(-1)
		g.onRealmComplete()
	} else {
		// toate lumile sunt deschise: monstri noi peste tot, fara harta
		g.startReveal(-1)
	}
	g.autosave()
}

// ---------------------------------------------------------------
// INAMICII: plimbare, urmarire, lovituri si bile de foc
// ---------------------------------------------------------------

func (g *Game) updateEnemies(dt float32) {
	w := g.City.W
	p := &g.Player
	speedM := float32(1)
	switch g.Difficulty {
	case "easy", "cheat":
		speedM = 0.8
	case "hard":
		speedM = 1.25
	}
	_, atkM, _ := g.enemyMultipliers()
	busy := g.Ride.Kind == RideWheel || g.Ride.Kind == RideCoaster || g.Swimming || g.InTunnel >= 10
	g.pickAttacker(dt, busy)
	for i := range g.Enemies {
		e := &g.Enemies[i]
		if !e.Alive {
			continue
		}
		e.Pose.Breathe += dt * 2
		if e.Flash > 0 {
			e.Flash -= dt
		}
		if e.DodgeCD > 0 {
			e.DodgeCD -= dt
		}
		if e.Stun > 0 {
			e.Stun -= dt
		}
		if e.Pose.Attack > 0 {
			e.Pose.Attack -= dt
		}
		var mx, mz float32
		moving := false
		dx := p.Pos.X - e.Pos.X
		dz := p.Pos.Z - e.Pos.Z
		dist := sqrtf(dx*dx + dz*dz)
		hx := p.Pos.X - e.Home.X
		hz := p.Pos.Z - e.Home.Y
		playerInZone := hx*hx+hz*hz < (e.HomeR+14)*(e.HomeR+14)
		engage := dist < 30 && playerInZone && !busy && absf(p.Pos.Y-e.Pos.Y) < 5 && g.Scene == SceneMapExplore
		if e.DodgeT > 0 {
			e.DodgeT -= dt
			mx = e.DodgeDir.X * 10 * dt
			mz = e.DodgeDir.Z * 10 * dt
			moving = true
		} else if engage && e.Stun <= 0 && i != g.Attacker {
			// V 6.0: isi asteapta randul: sta la 10-13 m, se uita la Aventurierul si se plimba usor in cerc
			e.Alert = true
			e.Windup = 0
			e.Yaw = turnTo(e.Yaw, atan2f(dx, dz), dt*5)
			switch {
			case dist < 10:
				// nu fuge de Aventurierul (il poate lovi cu sabia); doar nu vine mai aproape singur
			case dist > 14:
				s := 3.6 * speedM * dt
				mx, mz = dx/dist*s, dz/dist*s
				moving = true
			default:
				side := float32(1)
				if i%2 == 1 {
					side = -1
				}
				s := 1.4 * dt * side
				mx, mz = -dz/dist*s, dx/dist*s
				moving = true
			}
			if e.AtkCD < 1.2 {
				e.AtkCD = 1.2 // cand ii vine randul, nu loveste imediat
			}
		} else if engage && e.Stun <= 0 {
			if !e.Alert {
				e.Alert = true
				g.PlayAt(g.Audio.Engage, e.Pos, 0.8, 40)
			}
			e.Yaw = turnTo(e.Yaw, atan2f(dx, dz), dt*7)
			e.AtkCD -= dt
			if e.Ranged {
				// pastreaza distanta si trage bile de foc (se pot evita miscandu-te)
				switch {
				case dist > 15:
					s := 4.2 * speedM * dt
					mx, mz = dx/dist*s, dz/dist*s
					moving = true
				case dist < 8:
					s := 2.6 * dt
					mx, mz = -dx/dist*s, -dz/dist*s
					moving = true
				}
				if e.Windup > 0 {
					e.Windup -= dt
					e.Pose.Attack = 0.3
					if e.Windup <= 0 {
						g.enemyFire(e, float64(atkM))
						e.AtkCD = 2.6 + rand.Float32()*0.8
					}
				} else if e.AtkCD <= 0 && dist < 26 && g.Invuln <= 0.3 {
					e.Windup = 0.6
				}
			} else {
				if dist > 2.3 {
					sp := float32(4.6)
					if dist > 10 {
						sp = 5.6
					}
					s := sp * speedM * dt
					mx, mz = dx/dist*s, dz/dist*s
					moving = true
				}
				if e.Windup > 0 {
					e.Windup -= dt
					e.Pose.Attack = 0.3
					if e.Windup <= 0 {
						if dist < 3.4 {
							g.hurtPlayer(e.Attack*0.9, e.Pos)
						}
						e.Pose.Attack = 0.25
						e.AtkCD = 1.4 + rand.Float32()*0.5
					}
				} else if e.AtkCD <= 0 && dist < 2.9 {
					e.Windup = 0.55
					g.PlayAt(g.Audio.Windup, e.Pos, 0.9, 25)
				}
			}
		} else {
			e.Windup = 0
			if !engage {
				e.Alert = false
			}
			e.WanderT -= dt
			if e.WanderT <= 0 {
				e.WanderT = 2 + rand.Float32()*3
				e.WanderY = rand.Float32() * 2 * math.Pi
				if rand.Float32() < 0.3 {
					e.WanderY = -999 // sta pe loc
				}
			}
			ex := e.Pos.X - e.Home.X
			ez := e.Pos.Z - e.Home.Y
			if ex*ex+ez*ez > e.HomeR*e.HomeR {
				e.WanderY = atan2f(-ex, -ez)
			}
			if e.WanderY > -900 && e.Stun <= 0 {
				e.Yaw = turnTo(e.Yaw, e.WanderY, dt*3)
				s := 2.2 * dt
				mx, mz = sinf(e.Yaw)*s, cosf(e.Yaw)*s
				moving = true
			}
		}
		if moving {
			if w.Move(&e.Pos, mx, mz, 0.6*e.Scale) {
				e.Pose.Walk += dt * 8
				e.Pose.Moving = minf(e.Pose.Moving+dt*5, 1)
			} else {
				e.WanderY = e.Yaw + math.Pi
				e.Pose.Moving = maxf(e.Pose.Moving-dt*5, 0)
			}
		} else {
			e.Pose.Moving = maxf(e.Pose.Moving-dt*5, 0)
		}
		// nu iese de pe insula lui
		ex := e.Pos.X - e.Home.X
		ez := e.Pos.Z - e.Home.Y
		if d := sqrtf(ex*ex + ez*ez); d > e.HomeR+6 {
			e.Pos.X = e.Home.X + ex/d*(e.HomeR+6)
			e.Pos.Z = e.Home.Y + ez/d*(e.HomeR+6)
		}
		gy, _ := w.GroundAt(e.Pos.X, e.Pos.Z, e.Pos.Y)
		e.Pos.Y += (gy - e.Pos.Y) * minf(dt*12, 1)
	}
}

func (g *Game) enemyFire(e *Enemy, atkM float64) {
	p := &g.Player
	from := v3(e.Pos.X+sinf(e.Yaw)*0.8, e.Pos.Y+2.0*e.Scale, e.Pos.Z+cosf(e.Yaw)*0.8)
	to := v3(p.Pos.X, p.Pos.Y+1.3, p.Pos.Z)
	dir := rl.Vector3Normalize(rl.Vector3Subtract(to, from))
	g.Projectiles = append(g.Projectiles, Projectile3D{Position: from, Velocity: rl.Vector3Scale(dir, 15), Color: LegoOrange, Life: 3, Damage: e.Attack * 0.75, Target: -1, Enemy: true})
	g.PlayAt(g.Audio.EnemyShot, from, 1, 50)
	_ = atkM
}

func turnTo(cur, target, k float32) float32 {
	d := float32(math.Remainder(float64(target-cur), 2*math.Pi))
	return cur + d*minf(k, 1)
}

// ---------------------------------------------------------------
// PARTICULE + TEXTE PLUTITOARE
// ---------------------------------------------------------------

func (g *Game) spawnParticle(pos, vel rl.Vector3, col rl.Color, size, life float32) {
	if len(g.Particles) > 1800 {
		return
	}
	g.Particles = append(g.Particles, Particle3D{Position: pos, Velocity: vel, Color: col, Size: size, Life: life, MaxLife: life, Gravity: 9})
}

func (g *Game) addFloat(pos rl.Vector3, text string, col rl.Color, size float32) {
	g.Floats = append(g.Floats, FloatText{Pos: pos, Text: text, Color: col, Life: 1.4, Size: size})
}

func (g *Game) updateParticles() {
	dt := g.FrameTime
	for i := len(g.Particles) - 1; i >= 0; i-- {
		p := &g.Particles[i]
		p.Position = rl.Vector3Add(p.Position, rl.Vector3Scale(p.Velocity, dt))
		p.Velocity.Y -= p.Gravity * dt
		p.Life -= dt
		if p.Life <= 0 {
			g.Particles[i] = g.Particles[len(g.Particles)-1]
			g.Particles = g.Particles[:len(g.Particles)-1]
		}
	}
	for i := len(g.Floats) - 1; i >= 0; i-- {
		f := &g.Floats[i]
		f.Pos.Y += dt * 1.5
		f.Life -= dt
		if f.Life <= 0 {
			g.Floats = append(g.Floats[:i], g.Floats[i+1:]...)
		}
	}
	// apa din fantana
	if rand.Float32() < 0.8 {
		f := g.City.Fountain
		a := rand.Float32() * 2 * math.Pi
		g.Particles = append(g.Particles, Particle3D{Position: v3(f.X, 3.9, f.Z), Velocity: v3(cosf(a)*1.6, 4.5+rand.Float32(), sinf(a)*1.6), Color: rl.NewColor(190, 235, 255, 220), Size: 0.14, Life: 1.1, MaxLife: 1.1, Gravity: 9})
	}
}

// pickAttacker: un singur monstru ataca odata (pentru un copil de 5 ani).
// Ataca cel cu care se lupta Aventurierul: cand Aventurierul se apropie la 5 m de alt monstru,
// acela devine atacatorul. Ceilalti asteapta (nu vin mai aproape de 10 m, dar
// nici nu fug, ca Aventurierul sa-i poata lovi). Daca atacatorul cade sau Aventurierul pleaca,
// randul trece la cel mai apropiat.
func (g *Game) pickAttacker(dt float32, busy bool) {
	p := &g.Player
	ok := func(i int) bool {
		if i < 0 || i >= len(g.Enemies) {
			return false
		}
		e := &g.Enemies[i]
		if !e.Alive || busy || g.Scene != SceneMapExplore {
			return false
		}
		dx, dz := p.Pos.X-e.Pos.X, p.Pos.Z-e.Pos.Z
		hx, hz := p.Pos.X-e.Home.X, p.Pos.Z-e.Home.Y
		return dx*dx+dz*dz < 32*32 && hx*hx+hz*hz < (e.HomeR+14)*(e.HomeR+14) && absf(p.Pos.Y-e.Pos.Y) < 5
	}
	d2 := func(i int) float32 {
		e := &g.Enemies[i]
		return sq(p.Pos.X-e.Pos.X) + sq(p.Pos.Z-e.Pos.Z)
	}
	cur := g.Attacker
	best, bd := -1, float32(1e9)
	for i := range g.Enemies {
		if ok(i) && d2(i) < bd {
			best, bd = i, d2(i)
		}
	}
	if ok(cur) {
		// pastreaza atacatorul, afara de cazul in care Aventurierul s-a apucat de altul (la 5 m)
		if best >= 0 && best != cur && bd < 5*5 && d2(cur) > bd+9 && g.Enemies[cur].Windup <= 0 {
			g.Attacker = best
		}
		return
	}
	g.Attacker = best
	_ = dt
}
