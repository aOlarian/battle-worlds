package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// Realizari pentru Aventurierul: banner mare, cu trofeu desenat si sunet.
type Achievement struct {
	ID          string
	Title       string
	Description string
	Unlocked    bool
}

type AchievementManager struct {
	List     []*Achievement
	Queue    []*Achievement
	Active   *Achievement
	PopTimer float32
	OnUnlock func(a *Achievement)
}

func InitAchievements() *AchievementManager {
	return &AchievementManager{
		List: []*Achievement{
			{ID: "first_blood", Title: "Eroul Începător!", Description: "Ai învins primul tău inamic!"},
			{ID: "lego_builder", Title: "Marele Arhitect!", Description: "Ai pus prima ta piesă blocuri!"},
			{ID: "giant_upgrade", Title: "Super Mega Uriaș!", Description: "Ai mărit o construcție!"},
			{ID: "sharpshooter", Title: "Ochi de Vultur!", Description: "Ai nimerit de 5 ori la rând!"},
			{ID: "city_explorer", Title: "Exploratorul!", Description: "Ai vizitat toate cele 5 insule!"},
			{ID: "car_collector", Title: "Pilot de Curse!", Description: "Ai construit o mașină blocuri!"},
			{ID: "bluey_friend", Title: "Prietenul cățeilor!", Description: "Ai construit un cățel blocuri!"},
			{ID: "dino_master", Title: "Stăpânul Dinozaurilor!", Description: "Ai construit un dinozaur!"},
			{ID: "champion", Title: "Campionul Aventurierul!", Description: "Ai trecut de nivelul 5!"},
			{ID: "tunnel_rider", Title: "Prin Munte!", Description: "Ai trecut printr-un tunel!"},
			{ID: "bridge_master", Title: "Constructor de Poduri!", Description: "Ai traversat toate cele 4 poduri mari!"},
			{ID: "battle_name", Title: "D-A-V-I-D!", Description: "Ai construit toate literele numelui tău!"},
			{ID: "great_tunnel", Title: "Prin Marele Tunel!", Description: "Ai mers pe sub mare, de la un capăt la altul!"},
			{ID: "swimmer", Title: "Înotătorul!", Description: "Ai înotat în mare!"},
			{ID: "wheel_top", Title: "Pe acoperișul lumii!", Description: "Ai urcat în vârful Roții Uriașe!"},
			{ID: "coaster", Title: "Curajosul!", Description: "Ai mers cu montagne russe!"},
			{ID: "trampoline", Title: "Super Săritura!", Description: "Ai sărit 20 de metri pe trambulină!"},
			{ID: "roof_jumper", Title: "Pe acoperiș!", Description: "Ai aterizat pe acoperișul unui bloc!"},
			{ID: "driver", Title: "Șoferul Aventurierul!", Description: "Ai condus prima ta mașină!"},
			{ID: "tunnel_driver", Title: "Cu mașina pe sub mare!", Description: "Ai condus 600 m prin Marele Tunel!"},
			{ID: "car_boom", Title: "BUM!", Description: "Ai aruncat în aer o mașină (și ai luat aur)!"},
			{ID: "boat_sink", Title: "Pirat de blocuri!", Description: "Ai scufundat o barcă (și ai luat aur)!"},
			{ID: "plane_boom", Title: "Ochi de Pilot!", Description: "Ai nimerit avionul de pe cer! Pilotul a sărit cu parașuta."},
			{ID: "bird_hunter", Title: "Puf de pene!", Description: "Ai nimerit un pescăruș în zbor!"},
		},
	}
}

func (am *AchievementManager) find(id string) *Achievement {
	for _, a := range am.List {
		if a.ID == id {
			return a
		}
	}
	return nil
}

// MarkUnlocked: deblocare tacuta (la incarcarea salvarii).
func (am *AchievementManager) MarkUnlocked(id string) {
	if a := am.find(id); a != nil {
		a.Unlocked = true
	}
}

func (am *AchievementManager) Unlock(id string) {
	a := am.find(id)
	if a == nil || a.Unlocked {
		return
	}
	a.Unlocked = true
	am.Queue = append(am.Queue, a)
	if am.OnUnlock != nil {
		am.OnUnlock(a)
	}
}

func (am *AchievementManager) Count() (int, int) {
	n := 0
	for _, a := range am.List {
		if a.Unlocked {
			n++
		}
	}
	return n, len(am.List)
}

func (am *AchievementManager) Update(dt float32) {
	if am.Active == nil && len(am.Queue) > 0 {
		am.Active = am.Queue[0]
		am.Queue = am.Queue[1:]
		am.PopTimer = 4
	}
	if am.Active != nil {
		am.PopTimer -= dt
		if am.PopTimer <= 0 {
			am.Active = nil
		}
	}
}

func DrawTrophy(cx, cy, s float32) {
	gold := LegoGold
	dark := rl.NewColor(200, 150, 20, 255)
	rl.DrawRectangleRounded(rl.NewRectangle(cx-s*0.45, cy-s*0.5, s*0.9, s*0.55), 0.6, 8, gold)
	rl.DrawRing(rl.NewVector2(cx-s*0.45, cy-s*0.3), s*0.12, s*0.2, 90, 270, 12, gold)
	rl.DrawRing(rl.NewVector2(cx+s*0.45, cy-s*0.3), s*0.12, s*0.2, -90, 90, 12, gold)
	rl.DrawRectangle(int32(cx-s*0.08), int32(cy+s*0.05), int32(s*0.16), int32(s*0.22), dark)
	rl.DrawRectangleRounded(rl.NewRectangle(cx-s*0.3, cy+s*0.25, s*0.6, s*0.16), 0.4, 6, dark)
	DrawStar(cx, cy-s*0.25, s*0.16, LegoWhite)
}

func DrawStar(cx, cy, r float32, col rl.Color) {
	pts := make([]rl.Vector2, 10)
	for i := 0; i < 10; i++ {
		rr := r
		if i%2 == 1 {
			rr = r * 0.45
		}
		a := float32(i)*0.6283185 - 1.5707963
		pts[i] = rl.NewVector2(cx+cosf(a)*rr, cy+sinf(a)*rr)
	}
	c := rl.NewVector2(cx, cy)
	for i := 0; i < 10; i++ {
		j := (i + 1) % 10
		rl.DrawTriangle(c, pts[j], pts[i], col)
	}
}

func (am *AchievementManager) DrawPopup(g *Game) {
	a := am.Active
	if a == nil {
		return
	}
	w := float32(560)
	h := float32(104)
	x := (float32(ScreenWidth) - w) / 2
	y := float32(24)
	if am.PopTimer > 3.6 {
		t := (4 - am.PopTimer) / 0.4
		y = -h + (h+24)*easeOutBack(t)
	} else if am.PopTimer < 0.4 {
		y = -h + (h+24)*(am.PopTimer/0.4)
	}
	rl.DrawRectangleRounded(rl.NewRectangle(x-4, y-4, w+8, h+8), 0.3, 10, LegoGold)
	rl.DrawRectangleRounded(rl.NewRectangle(x, y, w, h), 0.3, 10, rl.NewColor(30, 34, 48, 250))
	rl.DrawRectangleRounded(rl.NewRectangle(x+12, y+12, 80, 80), 0.3, 8, rl.NewColor(255, 236, 140, 255))
	DrawTrophy(x+52, y+50, 60)
	g.text("REALIZARE NOUĂ!", x+108, y+12, 20, LegoGold, true)
	g.text(a.Title, x+108, y+36, 30, LegoWhite, true)
	g.text(a.Description, x+108, y+72, 20, LegoCyan, false)
}

func easeOutBack(t float32) float32 {
	t = clampf(t, 0, 1)
	c1 := float32(1.70158)
	c3 := c1 + 1
	u := t - 1
	return 1 + c3*u*u*u + c1*u*u
}
