package main

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Personaje in stil minifigurina: corp, cap rotund cu fata, par sau casca,
// picioare si brate animate separat. Toate piesele sunt mesh-uri rotunjite.

var (
	ColSkin = rl.NewColor(255, 214, 60, 255)
)

const (
	AccNone = iota
	AccHero
	AccHorns
	AccVisor
	AccHood
	AccRocks
	AccCrest
	AccLava
	AccDinoJaw
	AccIceCrown
	AccCap
	AccHair
	AccLongHair
)

type CharStyle struct {
	Shirt   rl.Color
	Pants   rl.Color
	Hair    rl.Color
	Eye     rl.Color
	Acc     int
	Print   rl.Color
	EyeGlo  bool
	Crown   bool
	Costume int
}

type CharMeshes struct {
	Tail         rl.Mesh
	HasTail      bool
	Body         rl.Mesh
	Leg          rl.Mesh
	Shin         rl.Mesh
	ForeL, ForeR rl.Mesh
	ArmL         rl.Mesh
	ArmR         rl.Mesh
	Cape         rl.Mesh
	HasCape      bool
	HeadTop      float32
}

var charCache = map[string]*CharMeshes{}

func styleKey(s CharStyle) string {
	return fmt.Sprintf("%v-%v-%v-%v-%d-%v-%v", s.Shirt, s.Pants, s.Hair, s.Eye, s.Acc, s.Print, s.EyeGlo) + fmt.Sprintf("-%v-%d", s.Crown, s.Costume)
}

// trunchi trapezoidal (mai lat jos, ca la minifigurine)
func torso(mb *MeshBuilder, y0, y1, wBot, wTop, d float32, col rl.Color) {
	hb, ht, hd := wBot/2, wTop/2, d/2
	p := [8]rl.Vector3{
		v3(-hb, y0, -hd), v3(hb, y0, -hd), v3(hb, y0, hd), v3(-hb, y0, hd),
		v3(-ht, y1, -hd), v3(ht, y1, -hd), v3(ht, y1, hd), v3(-ht, y1, hd),
	}
	mb.Quad(p[4], p[5], p[6], p[7], v3(0, 1, 0), col)
	mb.Quad(p[0], p[1], p[2], p[3], v3(0, -1, 0), col)
	mb.Quad(p[3], p[2], p[6], p[7], v3(0, 0, 1), col)
	mb.Quad(p[0], p[1], p[5], p[4], v3(0, 0, -1), col)
	mb.Quad(p[1], p[2], p[6], p[5], v3(1, 0.2, 0), shade(col, 0.95))
	mb.Quad(p[0], p[3], p[7], p[4], v3(-1, 0.2, 0), shade(col, 0.95))
}

func GetCharMeshes(s CharStyle) *CharMeshes {
	key := styleKey(s)
	if cm, ok := charCache[key]; ok {
		return cm
	}
	if f := realmChars[s.Costume]; f != nil {
		cm := f(s)
		if cm.HeadTop == 0 {
			cm.HeadTop = 2.75
		}
		charCache[key] = cm
		return cm
	}
	cm := &CharMeshes{HeadTop: 2.75}
	// ---- corp (sold, trunchi, gat, cap, fata, accesorii) ----
	mb := NewMB()
	mb.BeveledBox(v3(0, 1.0, 0), v3(0.98, 0.2, 0.52), 0.05, s.Pants)
	torso(mb, 1.08, 1.95, 1.0, 0.8, 0.5, s.Shirt)
	// imprimeu pe piept
	pc := s.Print
	if pc.A > 0 {
		mb.Quad(v3(-0.18, 1.35, 0.252), v3(0.18, 1.35, 0.252), v3(0.18, 1.7, 0.252), v3(-0.18, 1.7, 0.252), v3(0, 0, 1), pc)
	}
	mb.Cylinder(v3(0, 1.94, 0), v3(0, 2.06, 0), 0.19, 0.19, 12, ColSkin, false)
	// cap rotund
	skin := ColSkin
	if s.Costume == 2 {
		skin = LegoRed
	}
	mb.Cylinder(v3(0, 2.06, 0), v3(0, 2.58, 0), 0.33, 0.33, 18, skin, false)
	mb.EllipsoidF(v3(0, 2.58, 0), v3(0.33, 0.07, 0.33), 4, 18, func(n rl.Vector3) rl.Color { return skin })
	mb.EllipsoidF(v3(0, 2.06, 0), v3(0.33, 0.06, 0.33), 4, 18, func(n rl.Vector3) rl.Color { return skin })
	mb.Cylinder(v3(0, 2.6, 0), v3(0, 2.74, 0), 0.17, 0.17, 12, skin, true)
	// fata: ochi + zambet
	eye := s.Eye
	if s.EyeGlo {
		mb.Emis = 1
	}
	if s.Costume != 2 {
		for _, x := range []float32{-0.12, 0.12} {
			mb.Ellipsoid(v3(x, 2.4, 0.3), v3(0.055, 0.075, 0.04), 4, 6, eye)
		}
		mb.Emis = 0
		for k := 0; k < 7; k++ {
			a := float32(k)/6*2.2 + 0.47
			x := cosf(a) * 0.13
			y := 2.24 - sinf(a)*0.06
			mb.Box(v3(-x, y, 0.32), v3(0.05, 0.035, 0.03), rl.NewColor(60, 40, 30, 255))
		}
	}
	switch s.Acc {
	case AccHero:
		// casca aurie cu creasta rosie
		mb.EllipsoidF(v3(0, 2.55, 0), v3(0.38, 0.3, 0.38), 6, 16, func(n rl.Vector3) rl.Color { return LegoGold })
		mb.Box(v3(0, 2.52, 0.33), v3(0.7, 0.08, 0.1), LegoGoldBright)
		mb.BeveledBox(v3(0, 2.92, -0.05), v3(0.12, 0.3, 0.6), 0.05, LegoRed)
		mb.BeveledBox(v3(0, 1.6, 0.255), v3(0.38, 0.38, 0.03), 0.05, LegoGold)
	case AccHorns:
		mb.EllipsoidF(v3(0, 2.55, 0), v3(0.37, 0.25, 0.37), 5, 14, func(n rl.Vector3) rl.Color { return shade(s.Shirt, 0.6) })
		for _, sx := range []float32{-1, 1} {
			mb.CylinderC(v3(sx*0.3, 2.6, 0), v3(sx*0.62, 2.8, 0.05), 0.12, 0.09, 8, rl.NewColor(240, 230, 210, 255), rl.NewColor(240, 230, 210, 255), false)
			mb.CylinderC(v3(sx*0.62, 2.8, 0.05), v3(sx*0.66, 3.15, 0.1), 0.09, 0.01, 8, rl.NewColor(240, 230, 210, 255), LegoWhite, true)
		}
	case AccVisor:
		mb.EllipsoidF(v3(0, 2.52, 0), v3(0.37, 0.32, 0.37), 6, 14, func(n rl.Vector3) rl.Color { return LegoLightGrey })
		mb.Emis = 1
		mb.Box(v3(0, 2.4, 0.3), v3(0.6, 0.12, 0.12), LegoCyan)
		mb.Emis = 0
		for _, sx := range []float32{-1, 1} {
			mb.BeveledBox(v3(sx*0.5, 1.9, 0), v3(0.42, 0.22, 0.6), 0.08, LegoLightGrey)
		}
	case AccHood:
		mb.EllipsoidF(v3(0, 2.38, -0.04), v3(0.42, 0.48, 0.42), 7, 14, func(n rl.Vector3) rl.Color { return shade(s.Shirt, 0.55) })
		mb.Cylinder(v3(0, 1.9, -0.1), v3(0, 2.1, -0.1), 0.55, 0.45, 12, shade(s.Shirt, 0.5), false)
	case AccRocks:
		for _, sx := range []float32{-1, 1} {
			mb.CylinderC(v3(sx*0.45, 1.9, 0), v3(sx*0.62, 2.35, -0.05), 0.2, 0.02, 6, LegoDarkGrey, LegoLightGrey, true)
			mb.CylinderC(v3(sx*0.3, 1.95, -0.15), v3(sx*0.35, 2.3, -0.2), 0.14, 0.02, 6, LegoDarkGrey, LegoLightGrey, true)
		}
		mb.EllipsoidF(v3(0, 2.6, 0), v3(0.36, 0.18, 0.36), 4, 10, func(n rl.Vector3) rl.Color { return LegoDarkGrey })
	case AccCrest:
		for k := 0; k < 5; k++ {
			z := 0.25 - float32(k)*0.15
			mb.CylinderC(v3(0, 2.62, z), v3(0, 2.95-float32(k)*0.03, z-0.1), 0.08, 0.01, 6, LegoOrange, LegoYellow, true)
		}
		for _, sx := range []float32{-1, 1} {
			mb.CylinderC(v3(sx*0.3, 2.5, 0), v3(sx*0.5, 2.75, -0.2), 0.08, 0.01, 6, LegoOrange, LegoYellow, true)
		}
	case AccLava:
		mb.Emis = 1
		for k := 0; k < 4; k++ {
			y := 1.2 + float32(k)*0.2
			mb.Box(v3(0.05*float32(k%2)-0.05, y, 0.256), v3(0.6-float32(k)*0.08, 0.04, 0.02), LegoOrange)
		}
		mb.Emis = 0
		mb.EllipsoidF(v3(0, 2.6, 0), v3(0.36, 0.14, 0.36), 4, 10, func(n rl.Vector3) rl.Color { return LegoOrange })
	case AccDinoJaw:
		mb.BeveledBox(v3(0, 2.3, 0.35), v3(0.5, 0.25, 0.4), 0.08, s.Shirt)
		for k := 0; k < 4; k++ {
			mb.CylinderC(v3(-0.18+float32(k)*0.12, 2.18, 0.5), v3(-0.18+float32(k)*0.12, 2.08, 0.52), 0.035, 0.0, 4, LegoWhite, LegoWhite, false)
		}
		for k := 0; k < 4; k++ {
			mb.CylinderC(v3(0, 2.62, 0.15-float32(k)*0.15), v3(0, 2.85, 0.1-float32(k)*0.15), 0.1, 0.01, 4, s.Shirt, LegoLime, true)
		}
	case AccIceCrown:
		mb.Emis = 0.35
		for k := 0; k < 6; k++ {
			a := float32(k) / 6 * 2 * math.Pi
			x, z := cosf(a)*0.26, sinf(a)*0.26
			mb.CylinderC(v3(x, 2.6, z), v3(x*1.2, 3.0+float32(k%2)*0.15, z*1.2), 0.08, 0.0, 5, LegoCyan, LegoWhite, false)
		}
		mb.Emis = 0
	case AccCap:
		mb.EllipsoidF(v3(0, 2.6, 0), v3(0.36, 0.2, 0.36), 5, 14, func(n rl.Vector3) rl.Color { return s.Hair })
		mb.BeveledBox(v3(0, 2.6, 0.38), v3(0.5, 0.05, 0.3), 0.02, s.Hair)
	case AccHair:
		mb.EllipsoidF(v3(0, 2.58, -0.02), v3(0.37, 0.24, 0.37), 5, 14, func(n rl.Vector3) rl.Color { return s.Hair })
		mb.BeveledBox(v3(0, 2.35, -0.22), v3(0.66, 0.42, 0.22), 0.1, s.Hair)
	case AccLongHair:
		mb.EllipsoidF(v3(0, 2.58, -0.02), v3(0.38, 0.25, 0.38), 5, 14, func(n rl.Vector3) rl.Color { return s.Hair })
		mb.BeveledBox(v3(0, 2.15, -0.2), v3(0.72, 0.85, 0.24), 0.12, s.Hair)
	}
	if s.Crown {
		mb.Cylinder(v3(0, 2.72, 0), v3(0, 2.89, 0), 0.37, 0.37, 12, LegoGold, true)
		for k := 0; k < 8; k++ {
			a := float32(k) * math.Pi / 4
			x, z := cosf(a)*0.3, sinf(a)*0.3
			mb.Cylinder(v3(x, 2.85, z), v3(x, 3.06, z), 0.07, 0.01, 6, LegoGold, true)
		}
	}
	decorateCharacter(mb, s)
	decorateEnemy(mb, s)
	cm.Body = mb.Build()

	// ---- picior (pivot la sold) ----
	lb := NewMB()
	lb.BeveledBox(v3(0, -0.23, 0), v3(0.44, 0.48, 0.5), 0.07, s.Pants)
	cm.Leg = lb.Build()
	lb = NewMB()
	boot := s.Pants
	if s.Costume == 4 {
		boot = LegoRed
	}
	lb.BeveledBox(v3(0, -0.22, 0), v3(0.43, 0.46, 0.49), 0.07, boot)
	if s.Costume == 7 {
		for _, y := range []float32{-.12, -.30} {
			lb.Box(v3(0, y, 0), v3(.44, .09, .50), LegoWhite)
		}
	}
	if s.Costume == 6 {
		lb.Box(v3(0, -.34, .15), v3(.45, .09, .45), LegoYellow)
	}
	lb.BeveledBox(v3(0, -0.43, 0.07), v3(0.46, 0.14, 0.62), 0.05, shade(boot, 0.8))
	cm.Shin = lb.Build()
	for side := -1; side <= 1; side += 2 {
		ab := NewMB()
		ab.BeveledBox(v3(0, -0.18, 0), v3(0.27, 0.40, 0.30), 0.08, s.Shirt)
		upper := ab.Build()
		ab = NewMB()
		ab.BeveledBox(v3(0, -0.18, 0), v3(0.25, 0.39, 0.28), 0.08, s.Shirt)
		ab.Ellipsoid(v3(0, -0.39, 0.08), v3(0.14, 0.12, 0.12), 5, 8, ColSkin)
		fore := ab.Build()
		if side == -1 {
			cm.ArmL = upper
			cm.ForeL = fore
		} else {
			cm.ArmR = upper
			cm.ForeR = fore
		}
	}

	if s.Acc == AccHero || s.Costume == 4 || s.Costume == 5 {
		cb := NewMB()
		if s.Costume == 5 {
			cb.Box(v3(0, -.52, 0), v3(.92, 1.04, .05), capeColor(s))
			for k := 0; k < 4; k++ {
				x := -.46 + float32(k)*.23
				cb.Tri(v3(x, -1.03, .025), v3(x+.23, -1.03, .025), v3(x+.115, -1.38, .025), v3(0, 0, 1), capeColor(s))
				cb.Tri(v3(x+.115, -1.38, -.025), v3(x+.23, -1.03, -.025), v3(x, -1.03, -.025), v3(0, 0, -1), capeColor(s))
			}
		} else {
			cb.BeveledBox(v3(0, -0.65, 0), v3(0.92, 1.3, 0.05), 0.02, capeColor(s))
		}
		cm.Cape = cb.Build()
		cm.HasCape = true
	}
	if s.Costume == 3 {
		tb := NewMB()
		tb.Cylinder(v3(0, 0, 0), v3(0, -.25, -.95), .22, .03, 10, s.Shirt, true)
		cm.Tail = tb.Build()
		cm.HasTail = true
	}
	charCache[key] = cm
	return cm
}

// Pose: starea animatiei unui personaj
type Pose struct {
	Walk      float32 // faza mersului
	Moving    float32 // 0..1 (amestec intre stat pe loc si mers)
	Attack    float32 // >0 cand bratul drept e ridicat (trage / loveste)
	Air       bool
	Hurt      float32
	Wave      float32 // salut (cetateni)
	Breathe   float32
	Weapon    int     // 0 nimic, 1 pistol laser, 2 sabie
	Raise     float32 // 0..1 arma apare in mana
	Swing     float32 // 0..1 progresul loviturii cu sabia
	SwingSide float32
	AimPitch  float32
	Swim      float32 // 0..1 inoata (corpul se apleaca in fata)
	Sit       bool    // in vagonul de montagne russe
	HandsUp   bool
	// V 5.0 (setate de joc, folosite de animatie)
	Gun         int     // modelul armei de foc cand Weapon == WeaponGun: 0..NumGuns-1 (vezi GunNames)
	Speed       float32 // viteza orizontala reala, m/s (mers ~5.5, alergare ~9)
	Land        float32 // 1 -> 0 dupa aterizare (amortizare in genunchi)
	Turn        float32 // viteza de rotire, rad/s (aplecare in curba)
	AirVelocity float32 // viteza verticala, pozitiva la urcare
}

// Armele aventurierului: pistol laser de jucarie si sabie laser.
const NumGuns = 6

// GunNames: numele armelor (pentru parinte; copilul vede desenul). Ordinea e fixa (salvare).
var GunNames = [NumGuns]string{"Pistol laser", "Laser rosu", "Blaster mare", "Lansator de rachete", "Raza continua", "Tun cu bule"}

var gunMesh, swordMesh rl.Mesh
var gunMeshes [NumGuns]rl.Mesh

// GunMuzzle: varful tevii in spatiul local al modelului armei (inainte de rotirea din mana).
var GunMuzzle [NumGuns]rl.Vector3
var weaponsReady bool

func buildWeaponMeshes() {
	// pistolul: teava pe +Z, prinderea in origine
	gb := NewMB()
	gb.BeveledBox(v3(0, 0.08, 0.25), v3(0.22, 0.26, 0.62), 0.06, LegoBlue)
	gb.BeveledBox(v3(0, 0.25, 0.18), v3(0.12, 0.1, 0.36), 0.03, LegoOrange)
	gb.BeveledBox(v3(0, -0.12, 0.02), v3(0.16, 0.32, 0.16), 0.04, LegoDarkGrey)
	gb.Cylinder(v3(0, 0.1, 0.55), v3(0, 0.1, 0.95), 0.07, 0.06, 10, LegoLightGrey, true)
	gb.Emis = 1
	gb.Cylinder(v3(0, 0.1, 0.95), v3(0, 0.1, 1.0), 0.075, 0.075, 10, LegoCyan, true)
	gb.Box(v3(0, 0.08, 0.1), v3(0.24, 0.05, 0.2), LegoCyan)
	gb.Emis = 0
	gunMesh = gb.Build()
	gunMeshes[0] = gunMesh
	GunMuzzle[0] = v3(0, 0.1, 1.0)
	buildExtraGuns()
	// sabia: lama pe +Z
	sb := NewMB()
	sb.Cylinder(v3(0, 0, -0.18), v3(0, 0, 0.28), 0.06, 0.06, 8, LegoDarkGrey, true)
	sb.BeveledBox(v3(0, 0, 0.3), v3(0.42, 0.09, 0.09), 0.03, LegoGold)
	sb.Cylinder(v3(0, 0, .27), v3(0, 0, .40), .10, .08, 12, LegoLightGrey, true)
	sb.Emis = 1
	sb.CylinderC(v3(0, 0, 0.34), v3(0, 0, 1.75), 0.075, 0.04, 10, LegoCyan, rl.NewColor(200, 250, 255, 255), true)
	sb.Emis = 0
	swordMesh = sb.Build()
	weaponsReady = true
}

func DrawCharacter(draw DrawFn, cm *CharMeshes, pos rl.Vector3, yaw, scale float32, p Pose, tint rl.Color) {
	run := clampf((p.Speed-5.5)/3.5, 0, 1)
	swing := sinf(p.Walk) * (0.55 + run*0.22) * p.Moving
	bob := absf(sinf(p.Walk))*0.06*p.Moving + sinf(p.Breathe)*0.012 - clampf(p.Land, 0, 1)*0.12
	if p.Air {
		swing = 0.5
	}
	// la inot corpul se apleaca in fata, cu pieptul ca pivot
	var body rl.Matrix
	if p.Swim > 0.01 {
		body = rl.MatrixMultiply(rl.MatrixTranslate(0, -1.4, 0), rl.MatrixRotateX(1.25*p.Swim))
		body = rl.MatrixMultiply(body, rl.MatrixTranslate(0, 1.4, 0))
	} else {
		lean := clampf(p.Speed/9, 0, 1) * p.Moving * 0.12
		roll := clampf(p.Turn, -3, 3)*p.Moving*0.035 + sinf(p.Breathe*0.35)*0.018*(1-p.Moving)
		body = rl.MatrixMultiply(rl.MatrixTranslate(0, -1, 0), rl.MatrixRotateX(lean))
		body = rl.MatrixMultiply(body, rl.MatrixRotateZ(roll))
		body = rl.MatrixMultiply(body, rl.MatrixTranslate(0, 1, 0))
	}
	// Lovitura si tresarirea antreneaza si trunchiul.
	if p.Weapon == WeaponSword && p.Swing > 0 {
		t := smoothstep(0, 1, p.Swing)
		body = rl.MatrixMultiply(body, rl.MatrixRotateY(p.SwingSide*(0.45-0.9*t)))
	}
	if p.Hurt > 0 {
		h := clampf(p.Hurt, 0, 1)
		body = rl.MatrixMultiply(body, rl.MatrixRotateX(-0.18*h))
		body = rl.MatrixMultiply(body, rl.MatrixRotateZ(0.13*h))
	}
	root := func(local rl.Matrix) rl.Matrix {
		m := rl.MatrixMultiply(local, body)
		m = rl.MatrixMultiply(m, rl.MatrixScale(scale, scale, scale))
		m = rl.MatrixMultiply(m, rl.MatrixRotateY(yaw))
		return rl.MatrixMultiply(m, rl.MatrixTranslate(pos.X, pos.Y, pos.Z))
	}
	draw(cm.Body, root(rl.MatrixTranslate(0, bob, 0)), tint)
	for _, side := range []float32{-1, 1} {
		a := swing * side
		switch {
		case p.Sit:
			a = -1.45
		case p.Swim > 0.01:
			a = sinf(p.Walk*2.2+side*1.6+p.Breathe*2) * 0.35
		case p.Air:
			a = -0.24 - 0.15*side
		}
		m := rl.MatrixMultiply(rl.MatrixRotateX(a), rl.MatrixTranslate(side*0.24, 0.95+bob, 0))
		draw(cm.Leg, root(m), tint)
		knee := (0.10+clampf(-sinf(p.Walk)*side, 0, 1)*(0.75+run*0.45))*p.Moving + clampf(p.Land, 0, 1)*0.65
		if p.Air {
			knee = 0.4 + 0.55*clampf(p.AirVelocity/5, 0, 1)
		}
		if p.Sit {
			knee = 1.3
		}
		shin := rl.MatrixMultiply(rl.MatrixRotateX(knee), rl.MatrixTranslate(0, -0.46, 0))
		draw(cm.Shin, root(rl.MatrixMultiply(shin, m)), tint)
	}
	if cm.HasTail {
		tm := rl.MatrixMultiply(rl.MatrixRotateY(sinf(p.Walk)*.25*p.Moving+sinf(p.Breathe)*.08), rl.MatrixTranslate(0, 1.1+bob, -.25))
		draw(cm.Tail, root(tm), tint)
	}
	// brate
	armA := -swing * 0.9
	ra := swing * 0.9
	var rz, lz float32
	switch {
	case p.Swim > 0.01:
		// craul: bratele se rotesc pe rand
		ph := p.Walk*1.1 + p.Breathe*0.6
		armA = -float32(math.Mod(float64(ph), 2*math.Pi))
		ra = -float32(math.Mod(float64(ph+math.Pi), 2*math.Pi))
	case p.Sit && p.HandsUp:
		armA, ra = -2.9, -2.9
		lz, rz = -0.25, 0.25
	case p.Sit:
		armA, ra = -0.9, -0.9
	case p.Weapon == WeaponGun:
		k := p.Raise
		ra = lerpf(ra, -1.57+p.AimPitch-clampf(p.Attack, 0, 1)*0.20, k)
		armA = lerpf(armA, -1.25+p.AimPitch*0.8, k)
		lz = 1.05 * k
	case p.Weapon == WeaponSword:
		if p.Swing > 0 {
			t := smoothstep(0, 1, p.Swing)
			ra = lerpf(-2.7, -0.3, t)
			rz = p.SwingSide * lerpf(-0.8, 0.9, t)
		} else {
			// sabia tinuta in sus, in fata, ca sa se vada din spate
			ra = lerpf(ra, -0.95+swing*0.2, p.Raise)
			rz = 0.25 * p.Raise
		}
	}
	if p.Attack > 0 && p.Weapon == WeaponNone {
		ra = -1.45 * smoothstep(0, 0.12, p.Attack)
	}
	if p.Wave > 0 {
		ra = -2.6 + sinf(p.Wave*14)*0.35
		rz = 0.4
	}
	twist := sinf(p.Walk) * p.Moving * 0.08
	mL := rl.MatrixMultiply(rl.MatrixRotateX(armA), rl.MatrixRotateZ(lz))
	mL = rl.MatrixMultiply(mL, rl.MatrixTranslate(-0.56, 1.86+bob, twist))
	draw(cm.ArmL, root(mL), tint)
	leftElbow := float32(-0.3)
	if p.Weapon == WeaponGun {
		leftElbow = lerpf(leftElbow, -0.35, p.Raise)
	}
	fl := rl.MatrixMultiply(rl.MatrixRotateX(leftElbow), rl.MatrixTranslate(0, -0.38, 0))
	draw(cm.ForeL, root(rl.MatrixMultiply(fl, mL)), tint)
	mR := rl.MatrixMultiply(rl.MatrixRotateX(ra), rl.MatrixRotateZ(rz))
	mR = rl.MatrixMultiply(mR, rl.MatrixTranslate(0.56, 1.86+bob, -twist))
	draw(cm.ArmR, root(mR), tint)
	rightElbow := float32(-0.18)
	if p.Weapon == WeaponGun {
		rightElbow = lerpf(rightElbow, 0, p.Raise)
	}
	fr := rl.MatrixMultiply(rl.MatrixRotateX(rightElbow), rl.MatrixTranslate(0, -0.38, 0))
	draw(cm.ForeR, root(rl.MatrixMultiply(fr, mR)), tint)
	if weaponsReady && p.Weapon != WeaponNone && p.Raise > 0.05 {
		s := p.Raise
		var wl rl.Matrix
		var mesh rl.Mesh
		if p.Weapon == WeaponGun {
			wl = rl.MatrixMultiply(rl.MatrixScale(s, s, s), rl.MatrixRotateX(math.Pi/2))
			mesh = gunMeshes[clampInt(p.Gun, 0, NumGuns-1)]
		} else {
			wl = rl.MatrixScale(s, s, s)
			mesh = swordMesh
		}
		wl = rl.MatrixMultiply(wl, rl.MatrixTranslate(0, -0.39, 0.08))
		draw(mesh, root(rl.MatrixMultiply(rl.MatrixMultiply(wl, fr), mR)), rl.White)
	}
	if cm.HasCape {
		capeA := 0.12 + 0.35*p.Moving + sinf(p.Walk*1.5)*0.08*p.Moving
		if p.Air {
			capeA = 0.9
		}
		if p.Sit {
			capeA = 0.05
		}
		m := rl.MatrixMultiply(rl.MatrixRotateX(capeA), rl.MatrixTranslate(0, 1.93+bob, -0.27))
		draw(cm.Cape, root(m), tint)
	}
}

// ---------------- stiluri ----------------

func HeroStyle() CharStyle {
	// V 6.0: Aventurierul de desen animat (Terra); vechea figurina doar daca lipseste
	if s := RealmCharStyle(9, RoleHero); realmChars[s.Costume] != nil {
		return s
	}
	return CharStyle{Shirt: LegoBlue, Pants: LegoRed, Eye: rl.NewColor(30, 30, 40, 255), Acc: AccHero}
}

var enemyAcc = []int{AccHorns, AccVisor, AccHood, AccRocks, AccCrest, AccLava, AccDinoJaw, AccIceCrown}
var enemyEye = []rl.Color{LegoRed, LegoCyan, LegoLime, LegoOrange, LegoYellow, LegoOrange, LegoRed, LegoCyan}

func EnemyStyle(kind int, col rl.Color) CharStyle {
	k := kind % len(enemyAcc)
	return CharStyle{Shirt: col, Pants: shade(col, 0.45), Eye: enemyEye[k], Acc: enemyAcc[k], EyeGlo: true}
}

func CitizenStyle(i int) CharStyle {
	shirts := []rl.Color{LegoRed, LegoBlue, LegoYellow, LegoGreen, LegoPink, LegoPurple, LegoCyan, LegoOrange, LegoWhite, LegoLime}
	pants := []rl.Color{rl.NewColor(40, 70, 140, 255), rl.NewColor(60, 60, 70, 255), rl.NewColor(120, 80, 50, 255), rl.NewColor(30, 110, 60, 255)}
	hairs := []rl.Color{rl.NewColor(90, 56, 30, 255), rl.NewColor(40, 34, 30, 255), rl.NewColor(240, 200, 90, 255), rl.NewColor(200, 90, 40, 255)}
	accs := []int{AccHair, AccLongHair, AccCap, AccHair}
	return CharStyle{
		Shirt: shirts[i%len(shirts)],
		Pants: pants[(i/2)%len(pants)],
		Hair:  hairs[(i/3)%len(hairs)],
		Eye:   rl.NewColor(30, 30, 40, 255),
		Acc:   accs[(i*7/5)%len(accs)],
		Print: []rl.Color{{}, LegoWhite, LegoYellow, {}}[(i)%4],
	}
}
