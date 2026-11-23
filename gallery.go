package main

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
	"os"
	"path/filepath"
)

type galleryActor struct {
	style CharStyle
	pose  Pose
	yaw   float32
	scale float32
}

func RunGallery(g *Game) {
	out := os.Getenv("BATTLE_DEBUG_OUT")
	if out == "" {
		out = "."
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		fmt.Println(err)
		return
	}
	floor := NewMB()
	floor.Box(v3(0, -.2, 0), v3(45, .4, 12), LegoGreen)
	ground := floor.Build()
	capture := func(name string, actors []galleryActor) {
		spacing := float32(2.1)
		width := float32(len(actors)-1) * spacing
		cam := rl.Camera3D{Position: v3(0, 5, float32(math.Max(10, float64(width)*1.25))), Target: v3(0, 1.4, 0), Up: v3(0, 1, 0), Fovy: 45, Projection: rl.CameraPerspective}
		if name == "boss" {
			cam = rl.Camera3D{Position: v3(0, 6, 16), Target: v3(0, 3, 0), Up: v3(0, 1, 0), Fovy: 45, Projection: rl.CameraPerspective}
		}
		if len(actors) == 1 && len(name) >= 4 && name[:4] == "aim_" {
			cam = rl.Camera3D{Position: v3(1.7, 2.55, -2.9), Target: v3(.25, 1.65, .3), Up: v3(0, 1, 0), Fovy: 38, Projection: rl.CameraPerspective}
		}
		meshes := make([]*CharMeshes, len(actors))
		for i, a := range actors {
			meshes[i] = GetCharMeshes(a.style)
		}
		drawActors := func(draw DrawFn) {
			for i, a := range actors {
				sc := a.scale
				if sc == 0 {
					sc = 1
				}
				DrawCharacter(draw, meshes[i], v3(float32(i)*spacing-width/2, 0, 0), a.yaw, sc, a.pose, rl.White)
			}
		}
		g.R.BeginFrame(cam, 0)
		g.R.ShadowRange = 25
		g.R.ShadowPass(v3(0, 0, 0), func(mat rl.Material) { drawActors(func(m rl.Mesh, t rl.Matrix, _ rl.Color) { rl.DrawMesh(m, mat, t) }) })
		rl.BeginDrawing()
		rl.ClearBackground(rl.NewColor(190, 225, 250, 255))
		rl.BeginMode3D(cam)
		g.R.DrawSky(cam)
		g.R.BindShadow()
		g.R.DrawLit(ground, rl.MatrixIdentity(), rl.White)
		drawActors(g.R.DrawLit)
		rl.EndMode3D()
		img := rl.LoadImageFromScreen()
		rl.EndDrawing()
		path := filepath.Join(out, name+".png")
		if !rl.ExportImage(*img, path) {
			fmt.Println("Export esuat:", path)
		}
		rl.UnloadImage(img)
	}
	if id := os.Getenv("BATTLE_GALLERY_REALM"); id != "" {
		runRealmGallery(g, id, out)
		return
	}
	if os.Getenv("BATTLE_GALLERY_ENEMIES") == "1" {
		for wi := 0; wi < 7; wi++ {
			a := []galleryActor{}
			for i := 0; i < 4; i++ {
				a = append(a, galleryActor{style: ThemedEnemyStyle(wi, i)})
			}
			capture(fmt.Sprintf("enemies_%d", wi), a)
		}
		boss := ThemedEnemyStyle(0, 0)
		boss.Crown = true
		capture("boss", []galleryActor{{style: Outfits[0].Style}, {style: boss, scale: 2.4}})
		return
	}
	for i, o := range Outfits {
		capture(fmt.Sprintf("outfit_%d_%s", i, o.ID), []galleryActor{{style: o.Style}, {style: o.Style, yaw: math.Pi}, {style: o.Style, yaw: math.Pi / 2}, {style: o.Style, yaw: math.Pi, pose: Pose{Walk: 1, Moving: 1, Speed: 5.5}}})
	}
	group := []galleryActor{}
	for _, o := range Outfits {
		group = append(group, galleryActor{style: o.Style})
	}
	capture("outfits_group", group)
	for _, speed := range []float32{5.5, 9} {
		a := []galleryActor{}
		for i := 0; i < 6; i++ {
			a = append(a, galleryActor{style: Outfits[0].Style, pose: Pose{Walk: float32(i) * math.Pi / 3, Moving: 1, Speed: speed}, yaw: math.Pi / 2})
		}
		capture(fmt.Sprintf("motion_%.0f_profile", speed), a)
		for i := range a {
			a[i].yaw = math.Pi
		}
		capture(fmt.Sprintf("motion_%.0f_back", speed), a)
	}
	for i := 0; i < 6; i++ {
		capture(fmt.Sprintf("aim_%d", i), []galleryActor{{style: Outfits[0].Style, pose: Pose{Weapon: WeaponGun, Raise: 1, Gun: i}}})
		captureGun(g, out, i)
	}
	capture("idle_jump", []galleryActor{{style: Outfits[0].Style, pose: Pose{Breathe: 1}, yaw: math.Pi / 2}, {style: Outfits[0].Style, pose: Pose{Air: true, AirVelocity: 5}, yaw: math.Pi / 2}, {style: Outfits[0].Style, pose: Pose{Air: true, AirVelocity: -5}, yaw: math.Pi / 2}, {style: Outfits[0].Style, pose: Pose{Land: 1}, yaw: math.Pi / 2}})
	capture("poses", []galleryActor{{style: HeroStyle(), pose: Pose{Air: true}}, {style: HeroStyle(), pose: Pose{Weapon: WeaponSword, Raise: 1, Swing: .5}}, {style: HeroStyle(), pose: Pose{Swim: 1, Walk: 1}}})
	special := []struct {
		name string
		pose Pose
	}{
		{"aim_up", Pose{Weapon: WeaponGun, Raise: 1, AimPitch: .35}},
		{"recoil", Pose{Weapon: WeaponGun, Raise: 1, Attack: 1}},
		{"walk_aim", Pose{Weapon: WeaponGun, Raise: 1, Moving: 1, Speed: 5.5, Walk: 1}},
		{"sword", Pose{Weapon: WeaponSword, Raise: 1, Swing: .3, SwingSide: 1}},
		{"swim", Pose{Swim: 1, Walk: 1}},
		{"sit", Pose{Sit: true}},
		{"hands_up", Pose{Sit: true, HandsUp: true}},
		{"wave", Pose{Wave: .3}},
		{"hurt", Pose{Hurt: 1}},
	}
	for _, item := range special {
		capture("special_"+item.name, []galleryActor{{style: HeroStyle(), pose: item.pose}, {style: HeroStyle(), pose: item.pose, yaw: math.Pi / 2}, {style: HeroStyle(), pose: item.pose, yaw: math.Pi}})
	}
	for wi := 0; wi < 7; wi++ {
		a := []galleryActor{}
		for i := 0; i < 4; i++ {
			a = append(a, galleryActor{style: ThemedEnemyStyle(wi, i)})
		}
		capture(fmt.Sprintf("enemies_%d", wi), a)
	}
	boss := ThemedEnemyStyle(0, 0)
	boss.Crown = true
	capture("boss", []galleryActor{{style: Outfits[0].Style}, {style: boss, scale: 2.4}})
}

func captureGun(g *Game, out string, i int) {
	cam := rl.Camera3D{Position: v3(1.35, .95, 2.35), Target: v3(0, .1, .55), Up: v3(0, 1, 0), Fovy: 27, Projection: rl.CameraPerspective}
	g.R.BeginFrame(cam, 0)
	rl.BeginDrawing()
	rl.ClearBackground(rl.NewColor(190, 225, 250, 255))
	rl.BeginMode3D(cam)
	g.R.DrawSky(cam)
	g.R.DrawLit(gunMeshes[i], rl.MatrixIdentity(), rl.White)
	rl.EndMode3D()
	img := rl.LoadImageFromScreen()
	rl.EndDrawing()
	path := filepath.Join(out, fmt.Sprintf("gun_close_%d.png", i))
	if !rl.ExportImage(*img, path) {
		fmt.Println("Export esuat:", path)
	}
	rl.UnloadImage(img)
}
