package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// BATTLE_GALLERY=1 BATTLE_GALLERY_REALM=<id> BATTLE_DEBUG_OUT=<folder>:
// capturi cu personajele si vehiculul unei lumi (pentru agentii de continut).
func runRealmGallery(g *Game, id string, out string) {
	ri := realmIndex(id)
	if ri < 0 {
		fmt.Println("lume necunoscuta:", id)
		return
	}
	os.MkdirAll(out, 0755)
	R := &Realms[ri]
	sky := R.SkyHorizon
	floor := NewMB()
	gc := R.Grass
	if gc.A == 0 {
		gc = LegoGreen
	}
	floor.Box(v3(0, -.2, 0), v3(60, .4, 20), gc)
	ground := floor.Build()
	shot := func(name string, cam rl.Camera3D, draw func(d DrawFn)) {
		g.R.BeginFrame(cam, 0)
		g.R.ShadowRange = 25
		g.R.ShadowPass(v3(0, 0, 0), func(mat rl.Material) { draw(func(m rl.Mesh, t rl.Matrix, _ rl.Color) { rl.DrawMesh(m, mat, t) }) })
		rl.BeginDrawing()
		rl.ClearBackground(sky)
		rl.BeginMode3D(cam)
		g.R.DrawSky(cam)
		g.R.BindShadow()
		g.R.DrawLit(ground, rl.MatrixIdentity(), rl.White)
		draw(g.R.DrawLit)
		rl.EndMode3D()
		img := rl.LoadImageFromScreen()
		rl.EndDrawing()
		p := filepath.Join(out, id+"_"+name+".png")
		if !rl.ExportImage(*img, p) {
			fmt.Println("Export esuat:", p)
		}
		rl.UnloadImage(img)
	}
	type actor struct {
		role  int
		pose  Pose
		yaw   float32
		scale float32
		crown bool
	}
	row := func(name string, as []actor, camZ, camY, tgtY float32) {
		sp := float32(2.4)
		width := float32(len(as)-1) * sp
		for _, a := range as {
			if a.scale > 1.5 {
				sp = 4
				width = float32(len(as)-1) * sp
			}
		}
		cam := rl.Camera3D{Position: v3(0, camY, float32(math.Max(float64(camZ), float64(width)*1.2))), Target: v3(0, tgtY, 0), Up: v3(0, 1, 0), Fovy: 45, Projection: rl.CameraPerspective}
		shot(name, cam, func(d DrawFn) {
			for i, a := range as {
				s := RealmCharStyle(ri, a.role)
				s.Crown = a.crown
				sc := a.scale
				if sc == 0 {
					sc = 1
				}
				DrawCharacter(d, GetCharMeshes(s), v3(float32(i)*sp-width/2, 0, 0), a.yaw, sc, a.pose, rl.White)
			}
		})
	}
	pi := float32(math.Pi)
	row("hero_turn", []actor{{yaw: 0}, {yaw: pi / 2}, {yaw: pi}, {yaw: -pi / 2}}, 9, 4, 1.4)
	row("hero_motion", []actor{
		{pose: Pose{Walk: 1, Moving: 1, Speed: 5.5}, yaw: pi / 2},
		{pose: Pose{Walk: 2.5, Moving: 1, Speed: 9}, yaw: pi / 2},
		{pose: Pose{Weapon: WeaponGun, Raise: 1}, yaw: pi / 4},
		{pose: Pose{Weapon: WeaponSword, Raise: 1, Swing: .3, SwingSide: 1}},
		{pose: Pose{Air: true, AirVelocity: 5}, yaw: pi / 2},
	}, 10, 4, 1.4)
	row("enemies", []actor{{role: RoleEnemy1}, {role: RoleEnemy2}, {role: RoleEnemy3}, {role: RoleEnemy4}}, 9, 4, 1.4)
	row("enemies_back", []actor{{role: RoleEnemy1, yaw: pi}, {role: RoleEnemy2, yaw: pi}, {role: RoleEnemy3, yaw: pi}, {role: RoleEnemy4, yaw: pi}}, 9, 4, 1.4)
	row("boss", []actor{{role: RoleHero}, {role: RoleBoss, scale: 2.4, crown: true}}, 14, 6, 3)
	row("npcs", []actor{{role: RoleNPC1, pose: Pose{Wave: .3}}, {role: RoleNPC2}, {role: RoleNPC3, pose: Pose{Walk: 1, Moving: 1, Speed: 3}}, {role: RoleHero}}, 9, 4, 1.4)
	if f := realmVehicles[R.Vehicle]; f != nil {
		cols := []rl.Color{R.Main, R.Accent, LegoWhite}
		ms := []rl.Mesh{}
		for _, c := range cols {
			ms = append(ms, f(c))
		}
		cam := rl.Camera3D{Position: v3(9, 6, 11), Target: v3(0, 1, 0), Up: v3(0, 1, 0), Fovy: 45, Projection: rl.CameraPerspective}
		shot("vehicles", cam, func(d DrawFn) {
			for i, m := range ms {
				t := rl.MatrixMultiply(rl.MatrixRotateY(0.6), rl.MatrixTranslate(float32(i-1)*5.5, 0, 0))
				d(m, t, rl.White)
			}
			// Aventurierul langa vehicul, pentru scara
			DrawCharacter(d, GetCharMeshes(RealmCharStyle(ri, RoleHero)), v3(0, 0, 4), 0.6, 1, Pose{}, rl.White)
		})
	}
	fmt.Println("galerie gata:", id)
}
