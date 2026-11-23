package main

import (
	"fmt"
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// MASINILE: se pot lovi (dau aur, explodeaza dupa cateva lovituri si
// reapar), iar Aventurierul se poate urca in ele (E) si le conduce oriunde:
// pe poduri, prin tunelurile din munti si prin Marele Tunel de sub Mare.
// ====================================================================

const (
	CarMaxHP      = float32(4) // 4 impuscaturi (sau 3 lovituri de sabie)
	CarRespawn    = float32(6) // dupa cat timp reapare in trafic
	CarHitGold    = 5
	CarBoomGold   = 50
	carHalfLen    = float32(2.25)
	carHalfWidth  = float32(1.1)
	driveMax      = float32(26)
	driveTurbo    = float32(34)
	driveReverse  = float32(9)
	carEnterReach = float32(4.2)
)

type DriveState struct {
	Active bool
	Car    int
	Speed  float32
	VelY   float32
	Steer  float32
	Bump   float32
	LookT  float32 // cat timp a trecut de la ultima miscare a mouse-ului
	Dist   float32 // metri condusi
	GTDist float32 // metri condusi prin Marele Tunel
}

// carFrame: punctul (x,z) in sistemul masinii: along = in fata, side = in dreapta.
func carFrame(c *Car, x, z float32) (along, side float32) {
	dx, dz := x-c.Pos.X, z-c.Pos.Z
	sy, cy := sinf(c.Yaw), cosf(c.Yaw)
	return dx*sy + dz*cy, dx*cy - dz*sy
}

func (c *Car) Hittable() bool { return c.Dead <= 0 && !c.Hidden }

// carAt: masina lovita de punctul p (proiectil), -1 = niciuna.
func (g *Game) carAt(p rl.Vector3, pad float32) int {
	for k := range g.City.Cars {
		c := &g.City.Cars[k]
		if !c.Hittable() || (g.Drive.Active && g.Drive.Car == k) {
			continue
		}
		if absf(p.X-c.Pos.X) > 4 || absf(p.Z-c.Pos.Z) > 4 {
			continue
		}
		dy := p.Y - c.Pos.Y
		if dy < -0.2 || dy > 1.9+pad {
			continue
		}
		a, s := carFrame(c, p.X, p.Z)
		if absf(a) < carHalfLen+pad && absf(s) < carHalfWidth+pad {
			return k
		}
	}
	return -1
}

func (g *Game) addGold(n int64, at rl.Vector3) {
	g.Player.Gold += n
	g.addFloat(at, fmt.Sprintf("+%d aur", n), LegoGoldBright, 28)
}

// carHit: masina primeste o lovitura (pistol sau sabie).
func (g *Game) carHit(k int, dmg float32, at rl.Vector3) {
	c := &g.City.Cars[k]
	if !c.Hittable() {
		return
	}
	c.HP -= dmg
	c.Flash = 0.14
	c.Speed *= 0.6
	g.addGold(CarHitGold, v3(c.Pos.X, c.Pos.Y+2.4, c.Pos.Z))
	g.PlayAt(g.Audio.ShotHit, c.Pos, 0.9, 60)
	g.PlayAt(g.Audio.Clack, c.Pos, 0.8, 40)
	for n := 0; n < 10; n++ {
		g.spawnParticle(at, v3((rand.Float32()-0.5)*6, rand.Float32()*4, (rand.Float32()-0.5)*6), []rl.Color{LegoYellow, LegoWhite, LegoOrange}[n%3], 0.16, 0.45)
	}
	if c.HP <= 0 {
		g.explodeCar(k)
	}
}

func (g *Game) explodeCar(k int) {
	c := &g.City.Cars[k]
	center := v3(c.Pos.X, c.Pos.Y+1.0, c.Pos.Z)
	c.Dead = CarRespawn
	c.Parked = false
	c.HP = 0
	g.addGold(CarBoomGold, v3(c.Pos.X, c.Pos.Y+3.4, c.Pos.Z))
	g.Achieve.Unlock("car_boom")
	g.PlayAt(g.Audio.Boom, center, 1, 120)
	if d := rl.Vector3Distance(center, g.Player.Pos); d < 30 {
		g.CamShake = maxf(g.CamShake, 0.9*(1-d/30))
	}
	// foc, scantei, fum si piese blocuri care zboara
	for n := 0; n < 46; n++ {
		col := []rl.Color{rl.NewColor(255, 200, 60, 255), rl.NewColor(255, 120, 30, 255), rl.NewColor(255, 240, 160, 255)}[n%3]
		g.spawnParticle(center, v3((rand.Float32()-0.5)*14, rand.Float32()*12+2, (rand.Float32()-0.5)*14), col, 0.45, 0.8)
	}
	for n := 0; n < 26; n++ {
		g.spawnParticle(center, v3((rand.Float32()-0.5)*5, rand.Float32()*6+3, (rand.Float32()-0.5)*5), rl.NewColor(70, 70, 76, 200), 0.8, 1.6)
	}
	for n := 0; n < 14; n++ {
		g.spawnParticle(center, v3((rand.Float32()-0.5)*16, rand.Float32()*10+4, (rand.Float32()-0.5)*16), carColors[c.Mesh%len(carColors)], 0.35, 1.4)
	}
	if g.Debug != nil {
		g.Debug.CarsBoomed++
	}
}

// respawnS: unde pe ruta reapare masina i: intre 45 si 140 m de Aventurierul, in spatele camerei
// (nu apare din senin in fata lui), la cel putin 14 m de alte masini de pe aceeasi ruta.
func (cw *CityWorld) respawnS(i int, playerPos, camPos rl.Vector3) float32 {
	car := &cw.Cars[i]
	r := car.Route
	look := rl.Vector3Subtract(playerPos, camPos)
	look.Y = 0
	if rl.Vector3Length(look) < 0.01 {
		look = v3(0, 0, 1)
	}
	look = rl.Vector3Normalize(look)
	free := func(s float32) bool {
		for j := range cw.Cars {
			o := &cw.Cars[j]
			if j == i || o.Route != r || o.Dead > 0 || o.Driven || o.Parked {
				continue
			}
			gap := absf(o.S - s)
			gap = minf(gap, r.Len-gap)
			if gap < 14 {
				return false
			}
		}
		return true
	}
	var good []float32
	var ok2 []float32
	for s := float32(0); s < r.Len; s += 5 {
		q, _ := r.Sample(s)
		d := rl.Vector3Distance(q, playerPos)
		if d < 45 || !free(s) {
			continue
		}
		to := rl.Vector3Subtract(q, camPos)
		to.Y = 0
		seen := rl.Vector3DotProduct(rl.Vector3Normalize(to), look) > 0.35 && d < 220
		switch {
		case d < 140 && !seen:
			good = append(good, s)
		case !seen:
			ok2 = append(ok2, s)
		}
	}
	if len(good) > 0 {
		return good[rand.Intn(len(good))]
	}
	if len(ok2) > 0 {
		return ok2[rand.Intn(len(ok2))]
	}
	return float32(math.Mod(float64(car.S+r.Len*0.5), float64(r.Len)))
}

var carColors = []rl.Color{LegoRed, LegoYellow, LegoBlue, LegoLime, LegoOrange, LegoCyan, LegoWhite, LegoPink}

// updateCarsGame: fum la masinile lovite, intoarcerea masinilor parcate in trafic.
func (g *Game) updateCarsGame(dt float32) {
	for k := range g.City.Cars {
		c := &g.City.Cars[k]
		c.Hidden = !c.Driven && g.lockedAt(c.Pos.X, c.Pos.Z, c.Pos.Y)
		if c.Flash > 0 {
			c.Flash -= dt
		}
		if c.Dead > 0 {
			continue
		}
		if c.HP <= CarMaxHP*0.5 && rand.Float32() < dt*10 {
			g.spawnParticle(v3(c.Pos.X+sinf(c.Yaw)*1.6, c.Pos.Y+1.2, c.Pos.Z+cosf(c.Yaw)*1.6), v3((rand.Float32()-0.5)*0.8, 2+rand.Float32(), (rand.Float32()-0.5)*0.8), rl.NewColor(90, 90, 96, 170), 0.5, 1.2)
		}
		if c.Parked && !(g.Drive.Active && g.Drive.Car == k) {
			// masina lasata departe se intoarce singura pe drumul ei
			if rl.Vector3Distance(c.Pos, g.Player.Pos) > 110 {
				c.Parked = false
			}
		}
	}
}

// nearestCarToEnter: masina in care poate urca Aventurierul acum.
func (g *Game) nearestCarToEnter() int {
	p := g.Player.Pos
	best, bd := -1, carEnterReach
	for k := range g.City.Cars {
		c := &g.City.Cars[k]
		if !c.Hittable() || absf(c.Pos.Y-p.Y) > 2.2 {
			continue
		}
		a, s := carFrame(c, p.X, p.Z)
		// distanta pana la marginea masinii
		da := maxf(absf(a)-carHalfLen, 0)
		ds := maxf(absf(s)-carHalfWidth, 0)
		if d := sqrtf(da*da + ds*ds); d < bd {
			best, bd = k, d
		}
	}
	return best
}

func (g *Game) enterCar(k int) {
	c := &g.City.Cars[k]
	c.Driven = true
	c.Parked = false
	c.Speed = 0
	g.Drive = DriveState{Active: true, Car: k}
	g.Ride = RideState{}
	g.Swimming = false
	g.CamYaw = c.Yaw + math.Pi
	g.CamPitch = 0.22
	g.CamDist = 8
	g.CamBoom = 8
	g.Player.Pos = c.Pos
	g.Player.OnGround = true
	g.Player.VelY = 0
	g.Achieve.Unlock("driver")
	g.Play(g.Audio.Click)
	g.sayIcon(IconCar, "Ești la volan! W = accelerezi, S = frânezi / înapoi, A/D = virezi, E = cobori.")
}

func (g *Game) exitCar() {
	w := g.City.W
	if !g.Drive.Active {
		return
	}
	c := &g.City.Cars[g.Drive.Car]
	c.Driven = false
	c.Parked = true
	c.Speed = 0
	g.Drive.Active = false
	p := &g.Player
	fx, fz := sinf(c.Yaw), cosf(c.Yaw)
	rx, rz := fz, -fx
	spots := [][2]float32{{-rx * 2.4, -rz * 2.4}, {rx * 2.4, rz * 2.4}, {-fx * 3.6, -fz * 3.6}, {fx * 3.6, fz * 3.6}}
	placed := false
	for _, s := range spots {
		x, z := c.Pos.X+s[0], c.Pos.Z+s[1]
		if w.CanStand(x, z, c.Pos.Y+0.3, 0.45) {
			gy, _ := w.GroundAt(x, z, c.Pos.Y+0.5)
			p.Pos = v3(x, gy, z)
			placed = true
			break
		}
	}
	if !placed {
		p.Pos = v3(c.Pos.X, c.Pos.Y+1.9, c.Pos.Z) // pe acoperisul masinii, sare jos
	}
	p.OnGround = placed
	p.VelY = 0
	p.Pose = Pose{} // se ridica de pe scaun (altfel ramanea asezat si cand inota)
	g.CamDist = 7
	g.Play(g.Audio.Click)
	g.Audio.EngineUpdate(false, 0)
}

// carCanGo: masina (centrul in x,z) poate sta acolo? intoarce si solul.
func (g *Game) carCanGo(x, z, y, yaw float32) (float32, bool) {
	w := g.City.W
	if g.lockedAt(x, z, y) {
		g.lockBump()
		return 0, false
	}
	fx, fz := sinf(yaw), cosf(yaw)
	rx, rz := fz, -fx
	ground := float32(-1000)
	// centrul, botul, spatele si colturile din fata
	pts := [][2]float32{{0, 0}, {carHalfLen, 0}, {-carHalfLen, 0}, {carHalfLen - 0.3, carHalfWidth - 0.2}, {carHalfLen - 0.3, -carHalfWidth + 0.2}, {-carHalfLen + 0.3, carHalfWidth - 0.2}, {-carHalfLen + 0.3, -carHalfWidth + 0.2}}
	for i, q := range pts {
		px := x + fx*q[0] + rx*q[1]
		pz := z + fz*q[0] + rz*q[1]
		// pe pante, botul e mai sus decat centrul: urcam rampe de pana la ~25 de grade
		lim := 0.5 + absf(q[0])*0.47
		gy, kind := w.GroundAt(px, pz, y+lim-0.6)
		if kind == SurfWater || gy > y+lim {
			return 0, false
		}
		if w.Solid(px, pz, maxf(y, gy)+0.7, 0.1) {
			return 0, false
		}
		if i == 0 {
			ground = gy
		}
	}
	return ground, true
}

// updateDriving: Aventurierul conduce masina (inlocuieste mersul pe jos).
func (g *Game) updateDriving(dt float32) {
	w := g.City.W
	d := &g.Drive
	c := &g.City.Cars[d.Car]
	p := &g.Player
	if !c.Hittable() {
		g.exitCar()
		return
	}
	throttle, steer := float32(0), float32(0)
	if rl.IsKeyDown(rl.KeyW) || rl.IsKeyDown(rl.KeyUp) {
		throttle = 1
	}
	if rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyDown) {
		throttle = -1
	}
	if rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyLeft) {
		steer = 1
	}
	if rl.IsKeyDown(rl.KeyD) || rl.IsKeyDown(rl.KeyRight) {
		steer = -1
	}
	brake := rl.IsKeyDown(rl.KeySpace)
	turbo := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
	if g.Debug != nil {
		if th, st, ok := g.Debug.DriveInput(g); ok {
			throttle, steer, brake = th, st, false
		}
	}
	top := driveMax
	if turbo {
		top = driveTurbo
	}
	switch {
	case brake:
		d.Speed = approach(d.Speed, 0, 30*dt)
	case throttle > 0:
		if d.Speed < 0 {
			d.Speed = approach(d.Speed, 0, 26*dt)
		} else {
			d.Speed = approach(d.Speed, top, (9+6*(1-d.Speed/top))*dt)
		}
	case throttle < 0:
		if d.Speed > 0 {
			d.Speed = approach(d.Speed, 0, 26*dt)
		} else {
			d.Speed = approach(d.Speed, -driveReverse, 8*dt)
		}
	default:
		d.Speed = approach(d.Speed, 0, 5*dt)
	}
	// virajul: cu cat merge mai repede, cu atat roata se intoarce mai putin
	d.Steer += (steer - d.Steer) * minf(dt*8, 1)
	spd := absf(d.Speed)
	if spd > 0.3 {
		rate := 1.9 * clampf(spd/5, 0, 1) * (1 - 0.45*clampf(spd/driveTurbo, 0, 1))
		sign := float32(1)
		if d.Speed < 0 {
			sign = -1
		}
		c.Yaw += d.Steer * rate * sign * dt
	}
	// mers pas cu pas (ca sa nu treaca prin pereti)
	dist := d.Speed * dt
	n := int(absf(dist)/0.35) + 1
	step := dist / float32(n)
	for i := 0; i < n; i++ {
		nx := c.Pos.X + sinf(c.Yaw)*step
		nz := c.Pos.Z + cosf(c.Yaw)*step
		gy, ok := g.carCanGo(nx, nz, c.Pos.Y, c.Yaw)
		if ok {
			// alte masini in cale
			for k := range g.City.Cars {
				o := &g.City.Cars[k]
				if k == d.Car || !o.Hittable() || absf(o.Pos.Y-c.Pos.Y) > 2 {
					continue
				}
				dx, dz := o.Pos.X-nx, o.Pos.Z-nz
				if dx*dx+dz*dz < 3.4*3.4 {
					a, _ := carFrame(c, o.Pos.X, o.Pos.Z)
					if a*sign32(step) > 0 {
						ok = false
						o.Speed = 0
						break
					}
				}
			}
		}
		if !ok {
			if absf(d.Speed) > 6 {
				g.PlayAt(g.Audio.Clack, c.Pos, 1, 40)
				g.CamShake = maxf(g.CamShake, clampf(absf(d.Speed)/40, 0.1, 0.5))
			}
			d.Speed = -d.Speed * 0.25
			break
		}
		c.Pos.X, c.Pos.Z = nx, nz
		d.Dist += absf(step)
		if w.InTunnel(nx, nz, c.Pos.Y+0.5) == 10 {
			d.GTDist += absf(step)
		}
		if gy < c.Pos.Y-0.9 {
			// a iesit de pe o margine: cade (fara pagube)
			c.Pos.Y += d.VelY * dt / float32(n)
		} else {
			c.Pos.Y += (gy - c.Pos.Y) * minf(dt*14, 1)
			if c.Pos.Y < gy {
				c.Pos.Y = maxf(c.Pos.Y, gy-0.05)
			}
			d.VelY = 0
		}
	}
	// gravitatie cand e in aer
	gy, _ := w.GroundAt(c.Pos.X, c.Pos.Z, c.Pos.Y+0.35)
	if c.Pos.Y > gy+0.05 {
		d.VelY -= Gravity * dt
		c.Pos.Y = maxf(c.Pos.Y+d.VelY*dt, gy)
		if c.Pos.Y <= gy {
			d.VelY = 0
		}
	} else {
		c.Pos.Y = gy
		d.VelY = 0
	}
	// inclinarea pe pante
	gf, _ := w.GroundAt(c.Pos.X+sinf(c.Yaw)*1.8, c.Pos.Z+cosf(c.Yaw)*1.8, c.Pos.Y+0.6)
	gb, _ := w.GroundAt(c.Pos.X-sinf(c.Yaw)*1.8, c.Pos.Z-cosf(c.Yaw)*1.8, c.Pos.Y+0.6)
	if absf(gf-gb) < 2 {
		c.Pitch += (-atan2f(gf-gb, 3.6) - c.Pitch) * minf(dt*10, 1)
	}
	c.Dir = v3(sinf(c.Yaw), 0, cosf(c.Yaw))
	c.Speed = absf(d.Speed)
	p.Pos = c.Pos
	p.Yaw = c.Yaw
	p.OnGround = true
	p.VelY = 0
	p.Pose = Pose{Sit: true}
	if rl.IsKeyPressed(rl.KeyQ) {
		g.PlayAt(g.Audio.CarHonk, c.Pos, 1, 60)
	}
	if d.GTDist > 600 {
		g.Achieve.Unlock("tunnel_driver")
	}
	g.Audio.EngineUpdate(true, d.Speed)
	if rl.IsKeyPressed(rl.KeyE) || (g.Debug != nil && g.Debug.PressE(g)) {
		g.exitCar()
		g.SkipE = true
	}
}

func approach(v, target, step float32) float32 {
	if v < target {
		return minf(v+step, target)
	}
	return maxf(v-step, target)
}

func sign32(v float32) float32 {
	if v < 0 {
		return -1
	}
	return 1
}

// driveCamera: camera in spatele masinii; mouse-ul o roteste, apoi revine singura in spate.
func (g *Game) driveCamera(dt float32) {
	w := g.City.W
	c := &g.City.Cars[g.Drive.Car]
	md := rl.GetMouseDelta()
	if absf(md.X)+absf(md.Y) > 0.5 {
		g.Drive.LookT = 0
	} else {
		g.Drive.LookT += dt
	}
	if g.Drive.LookT > 1.0 && absf(g.Drive.Speed) > 2 {
		want := c.Yaw + math.Pi
		if g.Drive.Speed < 0 {
			want = c.Yaw
		}
		dy := float32(math.Remainder(float64(want-g.CamYaw), 2*math.Pi))
		g.CamYaw += dy * minf(dt*2.5, 1)
		g.CamPitch += (0.22 - g.CamPitch) * minf(dt*1.5, 1)
	}
	back, fwd, _ := g.camVectors()
	pivot := v3(c.Pos.X, c.Pos.Y+2.3, c.Pos.Z)
	dist := clampf(g.CamDist, 5, 16)
	desired := rl.Vector3Add(pivot, rl.Vector3Scale(back, dist))
	clear := w.CameraClear(pivot, desired)
	boom := rl.Vector3Distance(pivot, clear)
	if boom < g.CamBoom {
		g.CamBoom = boom
	} else {
		g.CamBoom += (boom - g.CamBoom) * minf(dt*4, 1)
	}
	pos := rl.Vector3Add(pivot, rl.Vector3Scale(back, g.CamBoom))
	target := rl.Vector3Add(pivot, rl.Vector3Scale(fwd, 10))
	if g.CamShake > 0 {
		s := g.CamShake * 0.12
		j := v3((rand.Float32()-0.5)*s, (rand.Float32()-0.5)*s, (rand.Float32()-0.5)*s)
		pos = rl.Vector3Add(pos, j)
		target = rl.Vector3Add(target, j)
	}
	g.Camera.Position = pos
	g.Camera.Target = target
}
