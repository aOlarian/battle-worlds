package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Tuneluri reale: fiecare trece pe sub o creasta de munte. Muntele are iarba,
// stanca si zapada pe varfuri; tunelul are portaluri de piatra cu arc, ziduri
// de sprijin in santul de acces, sosea, trotuare, faianta, lampi si ventilatoare.

var (
	ColRock     = rl.NewColor(150, 138, 124, 255)
	ColRockDark = rl.NewColor(112, 104, 96, 255)
	ColMeadow   = rl.NewColor(116, 176, 82, 255)
	ColSnow     = rl.NewColor(244, 248, 252, 255)
	ColTile     = rl.NewColor(236, 230, 214, 255)
	ColVault    = rl.NewColor(200, 198, 192, 255)
	ColPortal   = rl.NewColor(182, 172, 158, 255)
)

func (cb *CityBuilder) mountainColor(h, ny, noise float32) rl.Color {
	rock := mixColor(ColRockDark, ColRock, 0.4+0.6*noise)
	if h < WaterY+0.2 {
		return shade(ColRockDark, 0.8)
	}
	grassAmt := smoothstep(0.5, 0.72, ny) * (1 - smoothstep(15, 21, h+noise*4))
	c := mixColor(rock, shade(ColMeadow, 0.9+0.2*noise), grassAmt)
	snowAmt := smoothstep(21, 25, h+noise*3) * smoothstep(0.45, 0.65, ny)
	c = mixColor(c, ColSnow, snowAmt)
	return c
}

func (cb *CityBuilder) buildMountain(m *Mountain) {
	mb := NewMB()
	const step = float32(1.0)
	const uLo = float32(-36)
	const uHi = float32(66)
	const vMax = float32(30)
	nu := int((uHi - uLo) / step)
	nv := int(2 * vMax / step)
	H := make([]float32, (nu+1)*(nv+1))
	R := make([]float32, (nu+1)*(nv+1))
	idx := func(i, j int) int { return j*(nu+1) + i }
	for j := 0; j <= nv; j++ {
		for i := 0; i <= nu; i++ {
			u := uLo + float32(i)*step
			v := -vMax + float32(j)*step
			x, z := m.World(u, v)
			r := m.Raw(u, v)
			R[idx(i, j)] = r
			H[idx(i, j)] = cb.W.LandHeight(x, z) + r - 0.08
		}
	}
	nrm := func(i, j int) rl.Vector3 {
		i0, i1 := maxInt(i-1, 0), minInt(i+1, nu)
		j0, j1 := maxInt(j-1, 0), minInt(j+1, nv)
		du := (H[idx(i1, j)] - H[idx(i0, j)]) / (float32(i1-i0) * step)
		dv := (H[idx(i, j1)] - H[idx(i, j0)]) / (float32(j1-j0) * step)
		// gradient in coordonate lume
		gx := du*m.Dir.X + dv*m.Perp.X
		gz := du*m.Dir.Y + dv*m.Perp.Y
		return rl.Vector3Normalize(v3(-gx, 1, -gz))
	}
	pt := func(i, j int) rl.Vector3 {
		u := uLo + float32(i)*step
		v := -vMax + float32(j)*step
		x, z := m.World(u, v)
		return v3(x, H[idx(i, j)], z)
	}
	col := func(i, j int, n rl.Vector3) rl.Color {
		p := pt(i, j)
		noise := fbm(p.X*0.15, p.Z*0.15, 3)
		return cb.mountainColor(p.Y, n.Y, noise)
	}
	for j := 0; j < nv; j++ {
		for i := 0; i < nu; i++ {
			if R[idx(i, j)] <= 0.05 && R[idx(i+1, j)] <= 0.05 && R[idx(i, j+1)] <= 0.05 && R[idx(i+1, j+1)] <= 0.05 {
				continue
			}
			uc := uLo + (float32(i)+0.5)*step
			vc := -vMax + (float32(j)+0.5)*step
			if m.InTrench(uc, vc) {
				continue
			}
			n00, n10, n01, n11 := nrm(i, j), nrm(i+1, j), nrm(i, j+1), nrm(i+1, j+1)
			mb.TriS(pt(i, j), n00, pt(i+1, j), n10, pt(i+1, j+1), n11, col(i, j, n00), col(i+1, j, n10), col(i+1, j+1, n11))
			mb.TriS(pt(i, j), n00, pt(i+1, j+1), n11, pt(i, j+1), n01, col(i, j, n00), col(i+1, j+1, n11), col(i, j+1, n01))
		}
	}
	// brazi si bolovani pe versanti
	placed := 0
	for tries := 0; tries < 900 && placed < 46; tries++ {
		u := -32 + cb.Rng.Float32()*94
		v := (cb.Rng.Float32()*2 - 1) * 26
		if absf(v) < 11 && (u > PortalOut-4 || u < -PortalIn+4) {
			continue
		}
		r := m.Raw(u, v)
		if r < 2.5 || r > 19 {
			continue
		}
		x, z := m.World(u, v)
		y := cb.W.LandHeight(x, z) + r
		if y < WaterY+1 {
			continue
		}
		// panta
		dr := absf(m.Raw(u+1, v)-r) + absf(m.Raw(u, v+1)-r)
		if dr > 1.6 {
			continue
		}
		cb.pineTree(mb, x, y, z, 0.9+cb.Rng.Float32()*0.5)
		placed++
	}
	for k := 0; k < 8; k++ {
		u := -28 + cb.Rng.Float32()*84
		v := (cb.Rng.Float32()*2 - 1) * 24
		if absf(v) < 10 {
			continue
		}
		x, z := m.World(u, v)
		y := cb.W.LandHeight(x, z) + m.Raw(u, v)
		s := 0.8 + cb.Rng.Float32()*1.4
		mb.EllipsoidF(v3(x, y, z), v3(s*1.3, s*0.8, s), 5, 7, func(n rl.Vector3) rl.Color { return shade(ColRock, 0.85+0.25*n.Y) })
	}
	cb.addChunk(mb, true)
}

// arcul golului de tunel (in plan v,y), parcurs de la dreapta-jos la stanga-jos
func tunnelHoleOutline(n int) [][2]float32 {
	pts := [][2]float32{}
	wallH := TubeWallH
	r := TubeHalfW
	total := wallH*2 + float32(math.Pi)*r
	for k := 0; k <= n; k++ {
		s := total * float32(k) / float32(n)
		switch {
		case s <= wallH:
			pts = append(pts, [2]float32{r, s})
		case s <= wallH+float32(math.Pi)*r:
			a := (s - wallH) / r
			pts = append(pts, [2]float32{r * cosf(a), wallH + r*sinf(a)})
		default:
			pts = append(pts, [2]float32{-r, wallH - (s - wallH - float32(math.Pi)*r)})
		}
	}
	return pts
}

func (cb *CityBuilder) buildTunnel(m *Mountain) {
	mb := NewMB()
	neon := m.Neon
	hw := TubeHalfW
	uA := -(PortalIn + 1)
	uB := PortalOut + 1
	dEnd := func(u float32) float32 { return minf(u-uA, uB-u) }
	P := func(u, v, y float32) rl.Vector3 {
		x, z := m.World(u, v)
		return v3(x, y, z)
	}
	dir3 := v3(m.Dir.X, 0, m.Dir.Y)
	per3 := v3(m.Perp.X, 0, m.Perp.Y)
	up := v3(0, 1, 0)
	ao := func(u float32) float32 { return lerpf(0.2, 1.0, 1-smoothstep(-0.5, 5, dEnd(u))) }
	glow := func(u float32) float32 {
		return (0.2 + 0.14*cosf(2*math.Pi*u/4)) * smoothstep(0, 4, dEnd(u))
	}
	vaultSeg := 18
	for u := uA; u < uB-0.001; u += 1 {
		u1 := u + 1
		um := u + 0.5
		mb.AO = ao(um)
		mb.Emis = glow(um) * 0.6
		// sosea + marcaj galben
		mb.Quad(P(u, -4.5, 0.05), P(u1, -4.5, 0.05), P(u1, 4.5, 0.05), P(u, 4.5, 0.05), up, ColAsphalt)
		if int(math.Floor(float64(u)))%3 == 0 {
			mb.Quad(P(u, -0.12, 0.06), P(u1, -0.12, 0.06), P(u1, 0.12, 0.06), P(u, 0.12, 0.06), up, ColYellowLn)
		}
		for _, s := range []float32{-1, 1} {
			mb.Quad(P(u, s*4.5, 0.25), P(u1, s*4.5, 0.25), P(u1, s*hw, 0.25), P(u, s*hw, 0.25), up, ColSidewalk)
			mb.Quad(P(u, s*4.5, 0.05), P(u1, s*4.5, 0.05), P(u1, s*4.5, 0.25), P(u, s*4.5, 0.25), rl.Vector3Scale(per3, -s), ColCurb)
			// perete cu faianta + dunga colorata
			in := rl.Vector3Scale(per3, -s)
			mb.Emis = glow(um)
			mb.Quad(P(u, s*hw, 0.25), P(u1, s*hw, 0.25), P(u1, s*hw, 1.0), P(u, s*hw, 1.0), in, ColTile)
			mb.Quad(P(u, s*hw, 1.0), P(u1, s*hw, 1.0), P(u1, s*hw, 1.45), P(u, s*hw, 1.45), in, neon)
			mb.Quad(P(u, s*hw, 1.45), P(u1, s*hw, 1.45), P(u1, s*hw, TubeWallH), P(u, s*hw, TubeWallH), in, ColTile)
		}
		// bolta
		for k := 0; k < vaultSeg; k++ {
			a0 := float32(k) / float32(vaultSeg) * math.Pi
			a1 := float32(k+1) / float32(vaultSeg) * math.Pi
			v0, y0 := hw*cosf(a0), TubeWallH+hw*sinf(a0)
			v1, y1 := hw*cosf(a1), TubeWallH+hw*sinf(a1)
			am := (a0 + a1) / 2
			inward := v3(-per3.X*cosf(am), -sinf(am), -per3.Z*cosf(am))
			c := ColVault
			if k%2 == 0 {
				c = shade(ColVault, 0.96)
			}
			mb.Quad(P(u, v0, y0), P(u1, v0, y0), P(u1, v1, y1), P(u, v1, y1), inward, c)
		}
		mb.Emis = 0
	}
	// lampi pe bolta, iesiri de urgenta, ventilatoare
	mb.AO = 0.5
	for u := uA + 2; u < uB-1; u += 4 {
		for _, s := range []float32{-1, 1} {
			a := float32(math.Pi/2) - s*0.75
			v := (hw - 0.15) * cosf(a)
			y := TubeWallH + (hw-0.15)*sinf(a)
			p := P(u, v, y)
			mb.Emis = 1
			mb.SetMatrix(matTRS(p, atan2f(m.Dir.X, m.Dir.Y), 1))
			mb.Box(v3(0, 0, 0), v3(0.35, 0.18, 1.6), rl.NewColor(255, 246, 210, 255))
			mb.ClearMatrix()
			mb.Emis = 0
		}
	}
	for u := uA + 8; u < uB-4; u += 16 {
		for _, s := range []float32{-1, 1} {
			mb.Emis = 1
			p := P(u, s*(hw-0.06), 2.2)
			mb.SetMatrix(matTRS(p, atan2f(m.Dir.X, m.Dir.Y), 1))
			mb.Box(v3(0, 0, 0), v3(0.08, 0.35, 0.7), rl.NewColor(60, 220, 110, 255))
			mb.ClearMatrix()
			mb.Emis = 0
		}
	}
	for _, u := range []float32{-6, 14, 34} {
		mb.AO = 0.4
		for _, s := range []float32{-1, 1} {
			c := P(u, s*1.6, TubeWallH+hw-1.3)
			a := rl.Vector3Add(c, rl.Vector3Scale(dir3, -1.4))
			b := rl.Vector3Add(c, rl.Vector3Scale(dir3, 1.4))
			mb.Cylinder(a, b, 0.55, 0.55, 12, LegoDarkGrey, true)
			mb.Cylinder(rl.Vector3Add(c, v3(0, 0.5, 0)), rl.Vector3Add(c, v3(0, 1.2, 0)), 0.08, 0.08, 4, LegoDarkGrey, false)
		}
	}
	mb.AO = 1
	mb.Emis = 0

	// portalurile: zid de piatra care urmareste profilul muntelui + arc de bolti
	for _, side := range []float32{-1, 1} {
		uf, ub, auEnd := uB, PortalOut, MountAuOut // fata portalului, spatele, capatul muntelui
		if side < 0 {
			uf, ub, auEnd = uA, -PortalIn, MountAuIn
		}
		outN := rl.Vector3Scale(dir3, side)
		topAt := func(v float32) float32 {
			x, z := m.World(ub, v)
			return cb.W.LandHeight(x, z) + m.Raw(ub, v) + 0.7
		}
		// contur exterior: dreapta-jos -> sus -> stanga -> stanga-jos
		outer := [][2]float32{}
		tw := TrenchHalfW
		nTop := 24
		total := float32(0)
		rawPts := [][2]float32{{tw, 0}}
		for k := 0; k <= nTop; k++ {
			v := tw - 2*tw*float32(k)/float32(nTop)
			rawPts = append(rawPts, [2]float32{v, topAt(v)})
		}
		rawPts = append(rawPts, [2]float32{-tw, 0})
		// reesantionare dupa lungime
		lens := []float32{0}
		for k := 1; k < len(rawPts); k++ {
			dx := rawPts[k][0] - rawPts[k-1][0]
			dy := rawPts[k][1] - rawPts[k-1][1]
			total += sqrtf(dx*dx + dy*dy)
			lens = append(lens, total)
		}
		N := 64
		for k := 0; k <= N; k++ {
			s := total * float32(k) / float32(N)
			j := 1
			for j < len(lens)-1 && lens[j] < s {
				j++
			}
			t := (s - lens[j-1]) / maxf(lens[j]-lens[j-1], 1e-4)
			outer = append(outer, [2]float32{lerpf(rawPts[j-1][0], rawPts[j][0], t), lerpf(rawPts[j-1][1], rawPts[j][1], t)})
		}
		hole := tunnelHoleOutline(N)
		stone := func(v, y float32) rl.Color {
			iy := int32(math.Floor(float64(y / 0.9)))
			ix := int32(math.Floor(float64((v + float32(iy&1)*0.6) / 1.2)))
			return shade(ColPortal, 0.88+0.2*hash2(ix+int32(m.Index*50), iy))
		}
		for k := 0; k < N; k++ {
			a0, a1 := hole[k], hole[k+1]
			b0, b1 := outer[k], outer[k+1]
			pa0, pa1 := P(uf, a0[0], a0[1]), P(uf, a1[0], a1[1])
			pb0, pb1 := P(uf, b0[0], b0[1]), P(uf, b1[0], b1[1])
			c := stone((a0[0]+b0[0])/2, (a0[1]+b0[1])/2)
			mb.Quad(pa0, pa1, pb1, pb0, outN, c)
		}
		// grosimea portalului (intradosul de 1 m) si capacul de sus
		for k := 0; k < N; k++ {
			a0, a1 := hole[k], hole[k+1]
			mid := v3((a0[0]+a1[0])/2, (a0[1]+a1[1])/2-TubeWallH, 0)
			inward := rl.Vector3Normalize(v3(-per3.X*mid.X, -maxf(mid.Y, 0), -per3.Z*mid.X))
			mb.Quad(P(uf, a0[0], a0[1]), P(uf, a1[0], a1[1]), P(ub, a1[0], a1[1]), P(ub, a0[0], a0[1]), inward, shade(ColPortal, 0.8))
		}
		for k := 1; k < nTop+1; k++ {
			p0 := rawPts[k]
			p1 := rawPts[k+1]
			if k+1 >= len(rawPts)-1 {
				break
			}
			mb.Quad(P(uf, p0[0], p0[1]), P(uf, p1[0], p1[1]), P(ub, p1[0], p1[1]), P(ub, p0[0], p0[1]), up, shade(ColPortal, 0.95))
		}
		// inelul de bolti (voussoirs) in relief + cheia de bolta aurie + neon
		ringN := 15
		for k := 0; k < ringN; k++ {
			a0 := float32(k) / float32(ringN) * math.Pi
			a1 := float32(k+1) / float32(ringN) * math.Pi
			r0, r1 := hw, hw+1.1
			o := uf + side*0.18
			q := func(a, r, uu float32) rl.Vector3 { return P(uu, r*cosf(a), TubeWallH+r*sinf(a)) }
			c := ColPortal
			if k%2 == 0 {
				c = shade(ColPortal, 1.12)
			}
			if k == ringN/2 {
				c = LegoGold
			}
			mb.Quad(q(a0, r0, o), q(a1, r0, o), q(a1, r1, o), q(a0, r1, o), outN, c)
			am := (a0 + a1) / 2
			mb.Quad(q(a0, r1, o), q(a1, r1, o), q(a1, r1, uf), q(a0, r1, uf), v3(per3.X*cosf(am), sinf(am), per3.Z*cosf(am)), shade(c, 0.85))
			mb.Quad(q(a0, r0, o), q(a1, r0, o), q(a1, r0, uf), q(a0, r0, uf), v3(-per3.X*cosf(am), -sinf(am), -per3.Z*cosf(am)), shade(c, 0.85))
			mb.Emis = 1
			mb.Quad(q(a0, r1+0.05, o+side*0.02), q(a1, r1+0.05, o+side*0.02), q(a1, r1+0.32, o+side*0.02), q(a0, r1+0.32, o+side*0.02), outN, neon)
			mb.Emis = 0
		}
		// borne cu dungi galben-negru la intrare
		for _, s := range []float32{-1, 1} {
			for k := 0; k < 4; k++ {
				c := LegoYellow
				if k%2 == 1 {
					c = LegoBlack
				}
				p := P(uf+side*0.6, s*(hw-0.3), 0.25+float32(k)*0.25+0.125)
				mb.Cylinder(v3(p.X, p.Y-0.125, p.Z), v3(p.X, p.Y+0.125, p.Z), 0.25, 0.25, 10, c, k == 3)
			}
		}
		// zidurile de sprijin din sant
		uStart := absf(uf)
		for uu := uStart; uu < auEnd+1; uu += 1 {
			u0 := side * uu
			u1 := side * (uu + 1)
			for _, s := range []float32{-1, 1} {
				v := s * TrenchHalfW
				t0 := topAtWall(cb, m, u0, v)
				t1 := topAtWall(cb, m, u1, v)
				if t0 < 0.3 && t1 < 0.3 {
					continue
				}
				in := rl.Vector3Scale(per3, -s)
				for y := float32(0); y < maxf(t0, t1); y += 1.0 {
					ya0, yb0 := minf(y, t0), minf(y+1, t0)
					ya1, yb1 := minf(y, t1), minf(y+1, t1)
					if yb0-ya0 < 0.01 && yb1-ya1 < 0.01 {
						continue
					}
					c := shade(ColConcrete, 0.9+0.1*float32(int(y)%2))
					mb.Quad(P(u0, v, ya0), P(u1, v, ya1), P(u1, v, yb1), P(u0, v, yb0), in, c)
				}
				// coronament
				mb.Quad(P(u0, v, t0), P(u1, v, t1), P(u1, v+s*0.5, t1), P(u0, v+s*0.5, t0), up, ColConcrete2)
			}
		}
		// drumul din sant (pana la pod / insula)
		var endU float32 = MountAuOut
		if side < 0 {
			endU = 28
		}
		ax, az := m.World(uf, 0)
		bx, bz := m.World(side*endU, 0)
		cb.roadArm(mb, rl.NewVector2(ax, az), rl.NewVector2(bx, bz), 4.5, true, true)
	}
	cb.addChunk(mb, true)
}

func topAtWall(cb *CityBuilder, m *Mountain, u, v float32) float32 {
	x, z := m.World(u, v)
	return maxf(cb.W.LandHeight(x, z), 0) + m.Raw(u, v)
}

func (cb *CityBuilder) buildMountainsAndTunnels() {
	for i := range cb.W.Mountains {
		cb.buildMountain(&cb.W.Mountains[i])
		cb.buildTunnel(&cb.W.Mountains[i])
	}
}
