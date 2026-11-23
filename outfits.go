package main

import rl "github.com/gen2brain/raylib-go/raylib"

// ====================================================================
// V 5.0: COSTUMELE aventurierului (cate unul pentru fiecare lume).
// Indexul 0 = costumul de baza (mereu disponibil). Ordinea e fixa:
// se salveaza dupa ID. Style se da lui GetCharMeshes.
// ====================================================================

type Outfit struct {
	ID      string
	Name    string // pentru parinte
	Style   CharStyle
	Ability int // V 6.0: abilitatea eroului (AbilSurf, AbilGlide...)
	Realm   int // lumea mare din care e eroul (-1 = Aventurierul)
}

var Outfits = []Outfit{
	{ID: "hero", Name: "Aventurierul", Style: RealmCharStyle(9, RoleHero), Realm: -1}, // V 6.0: Aventurierul de desen animat (Terra), nu figurina blocuri
	{ID: "r_mare", Name: "Surferul", Style: RealmCharStyle(0, RoleHero), Ability: AbilSurf, Realm: 0},
	{ID: "r_monstri", Name: "Micul Vampir", Style: RealmCharStyle(1, RoleHero), Ability: AbilGlide, Realm: 1},
	{ID: "r_panze", Name: "Băiatul-Păianjen", Style: RealmCharStyle(2, RoleHero), Ability: AbilWeb, Realm: 2},
	{ID: "r_eroi", Name: "Eroul Zburător", Style: RealmCharStyle(3, RoleHero), Ability: AbilFly, Realm: 3},
	{ID: "r_craciun", Name: "Spiridușul", Style: RealmCharStyle(4, RoleHero), Ability: AbilSled, Realm: 4},
	{ID: "r_paste", Name: "Iepurașul", Style: RealmCharStyle(5, RoleHero), Ability: AbilHop, Realm: 5},
	{ID: "r_halloween", Name: "Micul Vrăjitor", Style: RealmCharStyle(6, RoleHero), Ability: AbilFly, Realm: 6},
}

// outfitOfRealm: costumul eroului lumii ri.
func outfitOfRealm(ri int) int {
	for i, o := range Outfits {
		if o.Realm == ri {
			return i
		}
	}
	return -1
}

// ThemedEnemyStyle: cum arata monstrul i din lumea wi (0..NumWorlds-1).
// Implicit: monstrii vechi; Terra ii poate imbraca tematic.
func ThemedEnemyStyle(wi, i int) CharStyle {
	if role := RoleEnemy1 + ((wi+i)%4+4)%4; realmChars[RealmCharCode(curRealm, role)] != nil {
		return RealmCharStyle(curRealm, role)
	}
	wi = clampInt(wi, 0, 6)
	colors := [7][5]rl.Color{
		{LegoRed, LegoOrange, LegoYellow, LegoBlue, LegoGreen},
		{LegoPurple, LegoBlue, LegoCyan, LegoRed, LegoOrange},
		{LegoGreen, LegoLime, LegoOrange, LegoYellow, LegoCyan},
		{LegoCyan, LegoBlue, LegoOrange, LegoPurple, LegoGreen},
		{LegoPurple, LegoPink, LegoBlue, LegoRed, LegoOrange},
		{LegoOrange, LegoRed, LegoPink, LegoYellow, LegoCyan},
		{LegoGreen, LegoLime, LegoCyan, LegoYellow, LegoGreen},
	}
	idx := ((i % 5) + 5) % 5
	c := colors[wi][idx]
	return CharStyle{Shirt: c, Pants: shade(c, .7), Hair: c, Eye: rl.NewColor(35, 30, 30, 255), Costume: 10 + wi}
}
