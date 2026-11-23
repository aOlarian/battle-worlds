package main

import "math"

// V 5.0: viteza reala, rotirea si aterizarea aventurierului, pentru animatia naturala.
func (g *Game) updatePoseExtras(dt float32) {
	p := &g.Player
	if dt <= 0 {
		return
	}
	dx, dz := p.Pos.X-g.prevPos.X, p.Pos.Z-g.prevPos.Z
	sp := sqrtf(dx*dx+dz*dz) / dt
	if sp > 30 || g.Drive.Active || g.Ride.Kind == RideWheel || g.Ride.Kind == RideCoaster {
		sp = 0
	}
	p.Pose.Speed = lerpf(p.Pose.Speed, sp, minf(dt*10, 1))
	dy := float32(math.Remainder(float64(p.Yaw-g.prevYaw), 2*math.Pi))
	p.Pose.Turn = clampf(lerpf(p.Pose.Turn, dy/dt, minf(dt*8, 1)), -6, 6)
	if p.OnGround && g.prevAir {
		g.landT = 1
	}
	g.prevAir = !p.OnGround && !g.Swimming
	if g.landT > 0 {
		g.landT = maxf(0, g.landT-dt/0.3)
	}
	p.Pose.Land = g.landT
	p.Pose.Gun = g.Prog.Gun
	g.prevPos = p.Pos
	g.prevYaw = p.Yaw
}
