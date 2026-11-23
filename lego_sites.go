package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// DrawFn deseneaza un mesh cu o transformare; aceeasi functie serveste pasul de umbre.
type DrawFn func(mesh rl.Mesh, m rl.Matrix, tint rl.Color)

// Mesh-urile caramizilor, construite o singura data pentru fiecare model.
type BlueprintMeshes struct {
	Bricks []rl.Mesh // cate unul pe caramida, in coordonatele modelului
	Full   rl.Mesh   // tot modelul (pentru holograma)
	Radius float32
	Height float32
}

func addLegoBrick(mb *MeshBuilder, pos, size rl.Vector3, col rl.Color) {
	mb.BeveledBox(pos, size, minf(0.06, minf(size.X, minf(size.Y, size.Z))*0.2), col)
	nx := int(math.Round(float64(size.X / 0.8)))
	nz := int(math.Round(float64(size.Z / 0.8)))
	if nx < 1 {
		nx = 1
	}
	if nz < 1 {
		nz = 1
	}
	stepX := size.X / float32(nx)
	stepZ := size.Z / float32(nz)
	top := pos.Y + size.Y/2
	r := minf(0.17, minf(stepX, stepZ)*0.3)
	for i := 0; i < nx; i++ {
		for j := 0; j < nz; j++ {
			x := pos.X - size.X/2 + stepX*(float32(i)+0.5)
			z := pos.Z - size.Z/2 + stepZ*(float32(j)+0.5)
			mb.Cylinder(v3(x, top-0.01, z), v3(x, top+0.12, z), r, r, 10, shade(col, 1.04), true)
		}
	}
}

func BuildBlueprintMeshes(bp LegoItemDefinition) BlueprintMeshes {
	out := BlueprintMeshes{}
	full := NewMB()
	for _, b := range bp.Bricks {
		mb := NewMB()
		addLegoBrick(mb, b.RelPos, b.Size, b.Color)
		addLegoBrick(full, b.RelPos, b.Size, b.Color)
		out.Bricks = append(out.Bricks, mb.Build())
		r := sqrtf(b.RelPos.X*b.RelPos.X+b.RelPos.Z*b.RelPos.Z) + maxf(b.Size.X, b.Size.Z)/2
		if r > out.Radius {
			out.Radius = r
		}
		if h := b.RelPos.Y + b.Size.Y/2; h > out.Height {
			out.Height = h
		}
	}
	out.Full = full.Build()
	return out
}

type ConstructionSite struct {
	Position     rl.Vector3
	Blueprint    LegoItemDefinition
	Meshes       BlueprintMeshes
	CurrentStage int
	TotalStages  int
	UpgradeLevel int
	TargetScale  float32
	CurrentScale float32
	Yaw          float32

	HammerStrikesLeft int
	HammerTime        float32
	HammerActive      bool
	PopTime           float32 // animatia ultimei caramizi puse
	ColliderIdx       int
}

func (cs *ConstructionSite) Scale() float32 {
	return cs.CurrentScale * cs.Blueprint.BaseScale
}

func (cs *ConstructionSite) Update(dt float32) {
	if absf(cs.CurrentScale-cs.TargetScale) > 0.001 {
		cs.CurrentScale += (cs.TargetScale - cs.CurrentScale) * minf(dt*3.0, 1)
		if absf(cs.CurrentScale-cs.TargetScale) < 0.01 {
			cs.CurrentScale = cs.TargetScale
		}
	}
	if cs.HammerActive {
		cs.HammerTime += dt
	}
	if cs.PopTime > 0 {
		cs.PopTime -= dt
	}
}

func (cs *ConstructionSite) transform(extraScale float32, lift float32) rl.Matrix {
	s := cs.Scale() * extraScale
	m := rl.MatrixScale(s, s, s)
	m = rl.MatrixMultiply(m, rl.MatrixRotateY(cs.Yaw))
	return rl.MatrixMultiply(m, rl.MatrixTranslate(cs.Position.X, cs.Position.Y+0.4+lift, cs.Position.Z))
}

func (cs *ConstructionSite) Draw(draw DrawFn, ghosts bool, t float32) {
	base := cs.transform(1, 0)
	for i := 0; i < cs.CurrentStage && i < len(cs.Meshes.Bricks); i++ {
		m := base
		if i == cs.CurrentStage-1 && cs.PopTime > 0 {
			// caramida noua "cade" la locul ei si sare putin
			p := cs.PopTime / 0.45
			drop := p * p * 2.5 * cs.Scale()
			brick := cs.Blueprint.Bricks[i]
			c := brick.RelPos
			s := cs.Scale()
			mm := rl.MatrixTranslate(-c.X, -c.Y, -c.Z)
			k := 1 + 0.25*sinf(p*math.Pi)
			mm = rl.MatrixMultiply(mm, rl.MatrixScale(k, k, k))
			mm = rl.MatrixMultiply(mm, rl.MatrixTranslate(c.X, c.Y+drop/s, c.Z))
			m = rl.MatrixMultiply(mm, base)
		}
		draw(cs.Meshes.Bricks[i], m, rl.White)
	}
	if ghosts && cs.CurrentStage < len(cs.Meshes.Bricks) {
		a := uint8(90 + 60*sinf(t*4))
		draw(cs.Meshes.Bricks[cs.CurrentStage], base, rl.NewColor(255, 255, 255, a))
	}
}

// Ciocanul animat care bate caramida
var hammerMesh rl.Mesh

func buildHammerMesh() {
	mb := NewMB()
	mb.Cylinder(v3(0, -1.6, 0), v3(0, 0.1, 0), 0.12, 0.12, 8, LegoBrown, true)
	mb.Cylinder(v3(0, -1.65, 0), v3(0, -1.2, 0), 0.15, 0.15, 8, LegoRed, true)
	mb.BeveledBox(v3(0, 0.25, 0), v3(1.2, 0.6, 0.6), 0.1, LegoGold)
	mb.BeveledBox(v3(0.68, 0.25, 0), v3(0.2, 0.5, 0.5), 0.05, LegoDarkGrey)
	hammerMesh = mb.Build()
}

func (cs *ConstructionSite) DrawHammer(draw DrawFn) {
	if !cs.HammerActive {
		return
	}
	swing := sinf(cs.HammerTime*28) * 50
	idx := cs.CurrentStage - 1
	target := cs.Position
	if idx >= 0 && idx < len(cs.Blueprint.Bricks) {
		c := cs.Blueprint.Bricks[idx].RelPos
		p := rl.Vector3Transform(c, cs.transform(1, 0))
		target = p
	}
	m := rl.MatrixMultiply(rl.MatrixTranslate(-0.7, 1.2, 0), rl.MatrixRotateZ(swing*rl.Deg2rad))
	m = rl.MatrixMultiply(m, rl.MatrixTranslate(target.X+1.0, target.Y+0.8, target.Z))
	draw(hammerMesh, m, rl.White)
}

func DrawVendorHologram(draw DrawFn, vendorPos rl.Vector3, meshes BlueprintMeshes, t float32) {
	fit := float32(1.2) / maxf(meshes.Height, 1)
	if r := 1.0 / maxf(meshes.Radius, 0.5); r < fit {
		fit = r
	}
	y := vendorPos.Y + 3.5 + sinf(t*2.5)*0.15
	m := rl.MatrixScale(fit, fit, fit)
	m = rl.MatrixMultiply(m, rl.MatrixRotateY(t*1.2))
	m = rl.MatrixMultiply(m, rl.MatrixTranslate(vendorPos.X, y, vendorPos.Z))
	draw(meshes.Full, m, rl.NewColor(170, 240, 255, 200))
}
