package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// MeshBuilder strange geometrie in memorie (pozitii, normale, culori, date extra)
// si o urca pe placa video ca un singur rl.Mesh. Toata lumea statica e construita
// asa, ca jocul sa deseneze cateva mesh-uri mari in loc de mii de cuburi pe cadru.
//
// Texcoord-ul fiecarui varf transporta date de iluminare, nu UV:
//
//	x = AO (cat cer "vede" suprafata: 1 afara, ~0.25 in tunel)
//	y = emisie (0 normal, 1 = lampa/neon care straluceste singur)
type MeshBuilder struct {
	V    []float32
	N    []float32
	C    []uint8
	T    []float32
	AO   float32
	Emis float32
	M    rl.Matrix
	hasM bool
}

var meshKeepAlive [][]float32
var colorKeepAlive [][]uint8

func NewMB() *MeshBuilder {
	return &MeshBuilder{AO: 1, Emis: 0, M: rl.MatrixIdentity()}
}

// SetMatrix aplica o transformare tuturor primitivelor urmatoare.
func (b *MeshBuilder) SetMatrix(m rl.Matrix) { b.M = m; b.hasM = true }
func (b *MeshBuilder) ClearMatrix()          { b.M = rl.MatrixIdentity(); b.hasM = false }

func (b *MeshBuilder) VertexCount() int { return len(b.V) / 3 }

func (b *MeshBuilder) tp(p rl.Vector3) rl.Vector3 {
	if !b.hasM {
		return p
	}
	return rl.Vector3Transform(p, b.M)
}

func (b *MeshBuilder) tn(n rl.Vector3) rl.Vector3 {
	if !b.hasM {
		return n
	}
	m := b.M
	r := rl.NewVector3(
		m.M0*n.X+m.M4*n.Y+m.M8*n.Z,
		m.M1*n.X+m.M5*n.Y+m.M9*n.Z,
		m.M2*n.X+m.M6*n.Y+m.M10*n.Z,
	)
	return rl.Vector3Normalize(r)
}

func (b *MeshBuilder) raw(p, n rl.Vector3, c rl.Color) {
	b.V = append(b.V, p.X, p.Y, p.Z)
	b.N = append(b.N, n.X, n.Y, n.Z)
	b.C = append(b.C, c.R, c.G, c.B, c.A)
	b.T = append(b.T, b.AO, b.Emis)
}

// TriS adauga un triunghi cu normale netede; ordinea varfurilor e corectata
// automat ca fata sa fie spre normala medie (fara probleme de "winding").
func (b *MeshBuilder) TriS(p1, n1, p2, n2, p3, n3 rl.Vector3, c1, c2, c3 rl.Color) {
	p1, p2, p3 = b.tp(p1), b.tp(p2), b.tp(p3)
	n1, n2, n3 = b.tn(n1), b.tn(n2), b.tn(n3)
	g := rl.Vector3CrossProduct(rl.Vector3Subtract(p2, p1), rl.Vector3Subtract(p3, p1))
	avg := rl.Vector3Add(rl.Vector3Add(n1, n2), n3)
	if rl.Vector3DotProduct(g, avg) < 0 {
		p2, p3 = p3, p2
		n2, n3 = n3, n2
		c2, c3 = c3, c2
	}
	b.raw(p1, n1, c1)
	b.raw(p2, n2, c2)
	b.raw(p3, n3, c3)
}

// Tri adauga un triunghi plat, orientat spre "hint".
func (b *MeshBuilder) Tri(p1, p2, p3 rl.Vector3, hint rl.Vector3, c rl.Color) {
	g := rl.Vector3CrossProduct(rl.Vector3Subtract(p2, p1), rl.Vector3Subtract(p3, p1))
	if rl.Vector3Length(g) < 1e-9 {
		return
	}
	if rl.Vector3DotProduct(g, hint) < 0 {
		p2, p3 = p3, p2
		g = rl.Vector3Negate(g)
	}
	n := rl.Vector3Normalize(g)
	b.TriS(p1, n, p2, n, p3, n, c, c, c)
}

// Quad plat (a,b,c,d in ordine pe contur), orientat spre hint.
func (b *MeshBuilder) Quad(p1, p2, p3, p4 rl.Vector3, hint rl.Vector3, c rl.Color) {
	b.Tri(p1, p2, p3, hint, c)
	b.Tri(p1, p3, p4, hint, c)
}

// QuadC: quad plat cu culori pe varfuri.
func (b *MeshBuilder) QuadC(p1, p2, p3, p4 rl.Vector3, hint rl.Vector3, c1, c2, c3, c4 rl.Color) {
	g := rl.Vector3CrossProduct(rl.Vector3Subtract(p2, p1), rl.Vector3Subtract(p3, p1))
	if rl.Vector3Length(g) < 1e-9 {
		g = rl.Vector3CrossProduct(rl.Vector3Subtract(p3, p1), rl.Vector3Subtract(p4, p1))
	}
	if rl.Vector3DotProduct(g, hint) < 0 {
		g = rl.Vector3Negate(g)
	}
	n := rl.Vector3Normalize(g)
	b.TriS(p1, n, p2, n, p3, n, c1, c2, c3)
	b.TriS(p1, n, p3, n, p4, n, c1, c3, c4)
}

// Box: cutie aliniata pe axe (in spatiul matricii curente).
func (b *MeshBuilder) Box(c rl.Vector3, s rl.Vector3, col rl.Color) {
	hx, hy, hz := s.X/2, s.Y/2, s.Z/2
	v := func(x, y, z float32) rl.Vector3 { return rl.NewVector3(c.X+x, c.Y+y, c.Z+z) }
	top := shade(col, 1.0)
	b.Quad(v(-hx, hy, -hz), v(hx, hy, -hz), v(hx, hy, hz), v(-hx, hy, hz), rl.NewVector3(0, 1, 0), top)
	b.Quad(v(-hx, -hy, -hz), v(hx, -hy, -hz), v(hx, -hy, hz), v(-hx, -hy, hz), rl.NewVector3(0, -1, 0), col)
	b.Quad(v(-hx, -hy, hz), v(hx, -hy, hz), v(hx, hy, hz), v(-hx, hy, hz), rl.NewVector3(0, 0, 1), col)
	b.Quad(v(-hx, -hy, -hz), v(hx, -hy, -hz), v(hx, hy, -hz), v(-hx, hy, -hz), rl.NewVector3(0, 0, -1), col)
	b.Quad(v(hx, -hy, -hz), v(hx, hy, -hz), v(hx, hy, hz), v(hx, -hy, hz), rl.NewVector3(1, 0, 0), col)
	b.Quad(v(-hx, -hy, -hz), v(-hx, hy, -hz), v(-hx, hy, hz), v(-hx, -hy, hz), rl.NewVector3(-1, 0, 0), col)
}

// BoxNoBottom e la fel ca Box, fara fata de jos (pentru ce sta pe sol).
func (b *MeshBuilder) BoxNoBottom(c rl.Vector3, s rl.Vector3, col rl.Color) {
	hx, hy, hz := s.X/2, s.Y/2, s.Z/2
	v := func(x, y, z float32) rl.Vector3 { return rl.NewVector3(c.X+x, c.Y+y, c.Z+z) }
	b.Quad(v(-hx, hy, -hz), v(hx, hy, -hz), v(hx, hy, hz), v(-hx, hy, hz), rl.NewVector3(0, 1, 0), col)
	b.Quad(v(-hx, -hy, hz), v(hx, -hy, hz), v(hx, hy, hz), v(-hx, hy, hz), rl.NewVector3(0, 0, 1), col)
	b.Quad(v(-hx, -hy, -hz), v(hx, -hy, -hz), v(hx, hy, -hz), v(-hx, hy, -hz), rl.NewVector3(0, 0, -1), col)
	b.Quad(v(hx, -hy, -hz), v(hx, hy, -hz), v(hx, hy, hz), v(hx, -hy, hz), rl.NewVector3(1, 0, 0), col)
	b.Quad(v(-hx, -hy, -hz), v(-hx, hy, -hz), v(-hx, hy, hz), v(-hx, -hy, hz), rl.NewVector3(-1, 0, 0), col)
}

// BeveledBox: cutie cu muchii rotunjite (aspect de jucarie / Pixar).
func (b *MeshBuilder) BeveledBox(c rl.Vector3, s rl.Vector3, bevel float32, col rl.Color) {
	hx, hy, hz := s.X/2, s.Y/2, s.Z/2
	mn := hx
	if hy < mn {
		mn = hy
	}
	if hz < mn {
		mn = hz
	}
	if bevel > mn*0.95 {
		bevel = mn * 0.95
	}
	if bevel <= 0.001 {
		b.Box(c, s, col)
		return
	}
	ix, iy, iz := hx-bevel, hy-bevel, hz-bevel
	coords := func(h float32) [4]float32 { return [4]float32{-h, -h + bevel, h - bevel, h} }
	cx, cy, cz := coords(hx), coords(hy), coords(hz)
	proj := func(x, y, z float32) (rl.Vector3, rl.Vector3) {
		qx := clampf(x, -ix, ix)
		qy := clampf(y, -iy, iy)
		qz := clampf(z, -iz, iz)
		d := rl.NewVector3(x-qx, y-qy, z-qz)
		l := rl.Vector3Length(d)
		var n rl.Vector3
		if l < 1e-6 {
			n = rl.NewVector3(0, 1, 0)
		} else {
			n = rl.Vector3Scale(d, 1/l)
		}
		p := rl.NewVector3(c.X+qx+n.X*bevel, c.Y+qy+n.Y*bevel, c.Z+qz+n.Z*bevel)
		return p, n
	}
	face := func(axis int, sign float32) {
		var ua, va [4]float32
		switch axis {
		case 0:
			ua, va = cy, cz
		case 1:
			ua, va = cx, cz
		default:
			ua, va = cx, cy
		}
		pt := func(i, j int) (rl.Vector3, rl.Vector3) {
			u, v := ua[i], va[j]
			switch axis {
			case 0:
				return proj(sign*hx, u, v)
			case 1:
				return proj(u, sign*hy, v)
			default:
				return proj(u, v, sign*hz)
			}
		}
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				p1, n1 := pt(i, j)
				p2, n2 := pt(i+1, j)
				p3, n3 := pt(i+1, j+1)
				p4, n4 := pt(i, j+1)
				b.TriS(p1, n1, p2, n2, p3, n3, col, col, col)
				b.TriS(p1, n1, p3, n3, p4, n4, col, col, col)
			}
		}
	}
	for axis := 0; axis < 3; axis++ {
		face(axis, 1)
		face(axis, -1)
	}
}

// Cylinder intre doua puncte, cu raze diferite (con trunchiat), netezit.
func (b *MeshBuilder) Cylinder(p0, p1 rl.Vector3, r0, r1 float32, sides int, col rl.Color, caps bool) {
	b.CylinderC(p0, p1, r0, r1, sides, col, col, caps)
}

func (b *MeshBuilder) CylinderC(p0, p1 rl.Vector3, r0, r1 float32, sides int, col0, col1 rl.Color, caps bool) {
	axis := rl.Vector3Subtract(p1, p0)
	L := rl.Vector3Length(axis)
	if L < 1e-6 {
		return
	}
	ax := rl.Vector3Scale(axis, 1/L)
	ref := rl.NewVector3(0, 1, 0)
	if absf(ax.Y) > 0.9 {
		ref = rl.NewVector3(1, 0, 0)
	}
	u := rl.Vector3Normalize(rl.Vector3CrossProduct(ax, ref))
	w := rl.Vector3CrossProduct(ax, u)
	slope := (r0 - r1) / L
	for i := 0; i < sides; i++ {
		a0 := float64(i) / float64(sides) * 2 * math.Pi
		a1 := float64(i+1) / float64(sides) * 2 * math.Pi
		d0 := rl.Vector3Add(rl.Vector3Scale(u, float32(math.Cos(a0))), rl.Vector3Scale(w, float32(math.Sin(a0))))
		d1 := rl.Vector3Add(rl.Vector3Scale(u, float32(math.Cos(a1))), rl.Vector3Scale(w, float32(math.Sin(a1))))
		n0 := rl.Vector3Normalize(rl.Vector3Add(d0, rl.Vector3Scale(ax, slope)))
		n1 := rl.Vector3Normalize(rl.Vector3Add(d1, rl.Vector3Scale(ax, slope)))
		q00 := rl.Vector3Add(p0, rl.Vector3Scale(d0, r0))
		q01 := rl.Vector3Add(p0, rl.Vector3Scale(d1, r0))
		q10 := rl.Vector3Add(p1, rl.Vector3Scale(d0, r1))
		q11 := rl.Vector3Add(p1, rl.Vector3Scale(d1, r1))
		b.TriS(q00, n0, q01, n1, q11, n1, col0, col0, col1)
		b.TriS(q00, n0, q11, n1, q10, n0, col0, col1, col1)
		if caps {
			if r1 > 0.0001 {
				b.Tri(p1, q10, q11, ax, col1)
			}
			if r0 > 0.0001 {
				b.Tri(p0, q00, q01, rl.Vector3Negate(ax), col0)
			}
		}
	}
}

// Sphere neteda (elipsoid daca scale e neuniform).
func (b *MeshBuilder) Ellipsoid(c rl.Vector3, r rl.Vector3, rings, slices int, col rl.Color) {
	b.EllipsoidF(c, r, rings, slices, func(n rl.Vector3) rl.Color { return col })
}

func (b *MeshBuilder) EllipsoidF(c rl.Vector3, r rl.Vector3, rings, slices int, colf func(n rl.Vector3) rl.Color) {
	pt := func(i, j int) (rl.Vector3, rl.Vector3) {
		th := float64(i) / float64(rings) * math.Pi
		ph := float64(j) / float64(slices) * 2 * math.Pi
		n := rl.NewVector3(float32(math.Sin(th)*math.Cos(ph)), float32(math.Cos(th)), float32(math.Sin(th)*math.Sin(ph)))
		p := rl.NewVector3(c.X+n.X*r.X, c.Y+n.Y*r.Y, c.Z+n.Z*r.Z)
		nn := rl.Vector3Normalize(rl.NewVector3(n.X/r.X, n.Y/r.Y, n.Z/r.Z))
		return p, nn
	}
	for i := 0; i < rings; i++ {
		for j := 0; j < slices; j++ {
			p1, n1 := pt(i, j)
			p2, n2 := pt(i+1, j)
			p3, n3 := pt(i+1, j+1)
			p4, n4 := pt(i, j+1)
			c1, c2, c3, c4 := colf(n1), colf(n2), colf(n3), colf(n4)
			if i != 0 {
				b.TriS(p1, n1, p2, n2, p3, n3, c1, c2, c3)
			}
			if i != rings-1 {
				b.TriS(p1, n1, p3, n3, p4, n4, c1, c3, c4)
			}
		}
	}
}

func (b *MeshBuilder) Sphere(c rl.Vector3, r float32, col rl.Color) {
	b.Ellipsoid(c, rl.NewVector3(r, r, r), 8, 12, col)
}

// Ring: inel plat (de ex. borduri rotunde, fantani).
func (b *MeshBuilder) Disc(c rl.Vector3, rIn, rOut float32, sides int, col rl.Color) {
	for i := 0; i < sides; i++ {
		a0 := float64(i) / float64(sides) * 2 * math.Pi
		a1 := float64(i+1) / float64(sides) * 2 * math.Pi
		o0 := rl.NewVector3(c.X+rOut*float32(math.Cos(a0)), c.Y, c.Z+rOut*float32(math.Sin(a0)))
		o1 := rl.NewVector3(c.X+rOut*float32(math.Cos(a1)), c.Y, c.Z+rOut*float32(math.Sin(a1)))
		if rIn <= 0.0001 {
			b.Tri(c, o0, o1, rl.NewVector3(0, 1, 0), col)
			continue
		}
		i0 := rl.NewVector3(c.X+rIn*float32(math.Cos(a0)), c.Y, c.Z+rIn*float32(math.Sin(a0)))
		i1 := rl.NewVector3(c.X+rIn*float32(math.Cos(a1)), c.Y, c.Z+rIn*float32(math.Sin(a1)))
		b.Quad(i0, o0, o1, i1, rl.NewVector3(0, 1, 0), col)
	}
}

// Tube: cilindru gol vertical (perete cu grosime) - pentru fantana, turnuri.
func (b *MeshBuilder) Tube(c rl.Vector3, rIn, rOut, h float32, sides int, col rl.Color) {
	b.Cylinder(c, rl.NewVector3(c.X, c.Y+h, c.Z), rOut, rOut, sides, col, false)
	b.Disc(rl.NewVector3(c.X, c.Y+h, c.Z), rIn, rOut, sides, col)
	// perete interior (normale spre interior)
	for i := 0; i < sides; i++ {
		a0 := float64(i) / float64(sides) * 2 * math.Pi
		a1 := float64(i+1) / float64(sides) * 2 * math.Pi
		p0 := rl.NewVector3(c.X+rIn*float32(math.Cos(a0)), c.Y, c.Z+rIn*float32(math.Sin(a0)))
		p1 := rl.NewVector3(c.X+rIn*float32(math.Cos(a1)), c.Y, c.Z+rIn*float32(math.Sin(a1)))
		mid := rl.NewVector3((p0.X+p1.X)/2, 0, (p0.Z+p1.Z)/2)
		hint := rl.NewVector3(c.X-mid.X, 0, c.Z-mid.Z)
		b.Quad(p0, p1, rl.NewVector3(p1.X, c.Y+h, p1.Z), rl.NewVector3(p0.X, c.Y+h, p0.Z), hint, col)
	}
}

// Build urca geometria pe GPU.
func (b *MeshBuilder) Build() rl.Mesh {
	if len(b.V) == 0 {
		return rl.Mesh{}
	}
	mesh := rl.Mesh{
		VertexCount:   int32(len(b.V) / 3),
		TriangleCount: int32(len(b.V) / 9),
		Vertices:      &b.V[0],
		Normals:       &b.N[0],
		Colors:        &b.C[0],
		Texcoords:     &b.T[0],
	}
	rl.UploadMesh(&mesh, false)
	meshKeepAlive = append(meshKeepAlive, b.V, b.N, b.T)
	colorKeepAlive = append(colorKeepAlive, b.C)
	return mesh
}

// ---------- utilitare mici ----------

func clampf(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func absf(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

func lerpf(a, b, t float32) float32 { return a + (b-a)*t }

func smoothstep(e0, e1, x float32) float32 {
	t := clampf((x-e0)/(e1-e0), 0, 1)
	return t * t * (3 - 2*t)
}

func sinf(x float32) float32  { return float32(math.Sin(float64(x))) }
func cosf(x float32) float32  { return float32(math.Cos(float64(x))) }
func sqrtf(x float32) float32 { return float32(math.Sqrt(float64(x))) }
func atan2f(y, x float32) float32 {
	return float32(math.Atan2(float64(y), float64(x)))
}

func v3(x, y, z float32) rl.Vector3 { return rl.NewVector3(x, y, z) }

// shade inmulteste culoarea (lumineaza / intuneca), pastrand alpha.
func shade(c rl.Color, f float32) rl.Color {
	cl := func(v float32) uint8 {
		if v > 255 {
			return 255
		}
		if v < 0 {
			return 0
		}
		return uint8(v)
	}
	return rl.NewColor(cl(float32(c.R)*f), cl(float32(c.G)*f), cl(float32(c.B)*f), c.A)
}

func mixColor(a, b rl.Color, t float32) rl.Color {
	t = clampf(t, 0, 1)
	return rl.NewColor(
		uint8(float32(a.R)+(float32(b.R)-float32(a.R))*t),
		uint8(float32(a.G)+(float32(b.G)-float32(a.G))*t),
		uint8(float32(a.B)+(float32(b.B)-float32(a.B))*t),
		uint8(float32(a.A)+(float32(b.A)-float32(a.A))*t),
	)
}

func withAlpha(c rl.Color, a uint8) rl.Color { return rl.NewColor(c.R, c.G, c.B, a) }

// ---------- zgomot determinist (pentru teren, munti, variatii de culoare) ----------

func hash2(x, y int32) float32 {
	h := uint32(x)*374761393 + uint32(y)*668265263
	h = (h ^ (h >> 13)) * 1274126177
	h ^= h >> 16
	return float32(h&0xffffff) / float32(0xffffff)
}

func valueNoise(x, y float32) float32 {
	xi := int32(math.Floor(float64(x)))
	yi := int32(math.Floor(float64(y)))
	fx := x - float32(xi)
	fy := y - float32(yi)
	ux := fx * fx * (3 - 2*fx)
	uy := fy * fy * (3 - 2*fy)
	a := hash2(xi, yi)
	b := hash2(xi+1, yi)
	c := hash2(xi, yi+1)
	d := hash2(xi+1, yi+1)
	return lerpf(lerpf(a, b, ux), lerpf(c, d, ux), uy)
}

func fbm(x, y float32, oct int) float32 {
	sum, amp, freq, norm := float32(0), float32(1), float32(1), float32(0)
	for i := 0; i < oct; i++ {
		sum += valueNoise(x*freq, y*freq) * amp
		norm += amp
		amp *= 0.5
		freq *= 2.03
	}
	return sum / norm
}

// matrice utile
func matTRS(pos rl.Vector3, yaw float32, scale float32) rl.Matrix {
	m := rl.MatrixScale(scale, scale, scale)
	m = rl.MatrixMultiply(m, rl.MatrixRotateY(yaw))
	m = rl.MatrixMultiply(m, rl.MatrixTranslate(pos.X, pos.Y, pos.Z))
	return m
}
