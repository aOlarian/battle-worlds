package main

import rl "github.com/gen2brain/raylib-go/raylib"

// ====================================================================
// V 6.0: LUMILE MARI (Realms). O lume mare = o harta intreaga, cu tema ei
// completa (personaje, decor, lumina, transport, arme, efecte) si cu mai
// multe insule / zone (nivelurile ei). Lumea 0 = harta cu insule din V 5.0
// (tema: marea, plaja, soarele). Celelalte lumi au harti de uscat
// (padure, oras, zapada, pajiste...), construite de RealmBuilder.
//
// CONTRACT pentru agentii de continut (Terra):
//   - personajele unei lumi: registerRealmChar(cod, functie) in fisierul
//     realm_<id>_chars.go; codul = RealmCharCode(lume, rol)
//   - vehiculul lumii:      registerRealmVehicle(id, functie)
//   - decorul lumii:        registerRealmDecor(id, functie) (realm_<id>_decor.go)
// Nimic altceva nu se modifica in fisierele comune.
// ====================================================================

// Rolurile personajelor dintr-o lume (al doilea parametru din RealmCharCode).
const (
	RoleHero   = 0 // eroul lumii (costumul pe care il primeste Aventurierul)
	RoleEnemy1 = 1 // 4 feluri de monstri
	RoleEnemy2 = 2
	RoleEnemy3 = 3
	RoleEnemy4 = 4
	RoleBoss   = 5 // Seful (se deseneaza cu scale 2.4 si coroana desenata de joc)
	RoleNPC1   = 6 // 3 locuitori prietenosi (se plimba, saluta)
	RoleNPC2   = 7
	RoleNPC3   = 8
)

// RealmCharCode: codul pus in CharStyle.Costume pentru personajul unei lumi.
func RealmCharCode(realm, role int) int { return 100 + realm*10 + role }

// RealmCharStyle: stilul (cheia) unui personaj de lume.
func RealmCharStyle(realm, role int) CharStyle {
	return CharStyle{Costume: RealmCharCode(realm, role), Eye: rl.NewColor(30, 30, 40, 255)}
}

var realmChars = map[int]func(s CharStyle) *CharMeshes{}

// registerRealmChar: functia construieste TOATE piesele personajului
// (Body, Leg, Shin, ArmL/ArmR, ForeL/ForeR, optional Cape/Tail) pe scheletul
// din DrawCharacter (characters.go):
//   - Body: in coordonatele personajului (talpi la y=0, fata spre +Z):
//     sold ~y 1.0, umeri ~y 1.86, cap pana la ~y 2.75 (HeadTop);
//   - Leg: pivot la sold (0.24 lateral, y 0.95), coboara 0.46; Shin: pivot la genunchi, coboara ~0.48;
//   - ArmL/ArmR: pivot la umar (±0.56, y 1.86), coboara 0.38; ForeL/ForeR: pivot la cot, coboara ~0.40,
//     mana la y -0.39 (arma se prinde acolo);
//   - Cape: pivot la (0, 1.93, -0.27), atarna in jos; Tail: pivot la (0, 1.1, -0.25).
//
// Formele pot fi oricum (rotunde, organice, nu cuburi blocuri), dar pivotii raman aceiasi.
func registerRealmChar(code int, f func(s CharStyle) *CharMeshes) { realmChars[code] = f }

var realmVehicles = map[string]func(col rl.Color) rl.Mesh{}

// registerRealmVehicle: vehiculul care circula pe drumurile lumii (si cu care
// conduce Aventurierul). Ca buildCarMesh (vehicles.go): fata spre +Z, rotile pe y=0,
// lungime ~4.4, latime ~2.1, inaltime <= 2.8.
func registerRealmVehicle(id string, f func(col rl.Color) rl.Mesh) { realmVehicles[id] = f }

// Abilitatile speciale ale eroilor (le are Aventurierul cand poarta costumul eroului).
const (
	AbilNone  = iota
	AbilSurf  // pe apa sta pe o placa de surf si merge repede
	AbilGlide // in aer, SPATIU tinut = planeaza ca un liliac
	AbilWeb   // Q / click dreapta = se trage cu panza spre acoperisul tintit; pistolul trage panze care opresc monstrii
	AbilFly   // SPATIU tinut = zboara in sus; in aer se misca liber
	AbilSled  // pe zapada aluneca repede; arunca bulgari
	AbilHop   // sare foarte sus (salt de iepuras)
)

// Felul proiectilului (cum arata glontul armei in lumea respectiva).
const (
	ProjLaser = iota
	ProjBat
	ProjWeb
	ProjStar
	ProjSnow
	ProjEgg
	ProjBubble
	ProjCandy
)

// RealmDef: o lume mare.
type RealmDef struct {
	ID     string
	Name   string // pentru parinte
	Main   rl.Color
	Accent rl.Color
	Island bool // true doar pentru lumea 0 (harta cu insule)
	Night  bool // noapte cu luna mare pe cer (lampile si ferestrele lumineaza)

	// lumina si cerul
	SkyTop, SkyHorizon rl.Color
	FogColor           rl.Color
	SunColor           rl.Color
	Ambient            float32 // 0.3..0.7
	FogDensity         float32

	// pamantul
	Grass, Dirt, Path rl.Color // iarba (sau nisip/zapada), pamant, poteci/drumuri

	Vehicle  string // id pentru registerRealmVehicle ("" = masina obisnuita)
	Ability  int    // abilitatea eroului
	Proj     int    // felul proiectilului
	ProjCol  rl.Color
	Zones    []ZoneDef
	HeroName string    // numele costumului (pentru parinte)
	Enemies  [4]string // numele monstrilor (rolurile 1-4), pentru parinte
	Boss     string
}

// ZoneDef: o insula / zona = un nivel din lume.
type ZoneDef struct {
	Name   string
	Short  string
	Main   rl.Color
	Accent rl.Color
}

// Realms: ordinea in care se deschid lumile.
var Realms = []RealmDef{
	{ID: "mare", Name: "Lumea Mării", Island: true, Main: rl.NewColor(40, 170, 220, 255), Accent: rl.NewColor(255, 200, 60, 255),
		SkyTop: rl.NewColor(110, 185, 245, 255), SkyHorizon: rl.NewColor(205, 235, 250, 255), FogColor: rl.NewColor(200, 228, 245, 255), Path: rl.NewColor(205, 195, 175, 255), Grass: rl.NewColor(120, 196, 90, 255), Dirt: rl.NewColor(190, 170, 130, 255),
		SunColor: rl.NewColor(255, 246, 225, 255), Ambient: 0.5, FogDensity: 1,
		Ability: AbilSurf, Proj: ProjBubble, ProjCol: rl.NewColor(120, 220, 255, 255), HeroName: "Surferul",
		Zones: []ZoneDef{
			{"Portul Soarelui", "Port", rl.NewColor(240, 160, 40, 255), rl.NewColor(40, 120, 200, 255)},
			{"Golful Piraților", "Pirati", rl.NewColor(150, 90, 50, 255), rl.NewColor(230, 40, 40, 255)},
			{"Jungla Papagalilor", "Jungla", rl.NewColor(60, 170, 70, 255), rl.NewColor(255, 120, 40, 255)},
			{"Insula Farului", "Far", rl.NewColor(230, 60, 60, 255), rl.NewColor(250, 250, 250, 255)},
			{"Recifele de Corali", "Corali", rl.NewColor(255, 120, 150, 255), rl.NewColor(60, 200, 220, 255)},
			{"Plaja Soarelui", "Plaja", rl.NewColor(40, 190, 210, 255), rl.NewColor(255, 150, 60, 255)},
			{"Parcul Acvatic", "Aqua", rl.NewColor(60, 140, 240, 255), rl.NewColor(255, 220, 50, 255)},
		}},
	{ID: "monstri", Name: "Hotelul Monștrilor", Night: true, Main: rl.NewColor(120, 70, 170, 255), Accent: rl.NewColor(120, 230, 120, 255),
		SkyTop: rl.NewColor(20, 26, 70, 255), SkyHorizon: rl.NewColor(95, 75, 140, 255), FogColor: rl.NewColor(80, 70, 125, 255),
		SunColor: rl.NewColor(215, 222, 255, 255), Ambient: 0.62, FogDensity: 0.7,
		Grass: rl.NewColor(70, 110, 80, 255), Dirt: rl.NewColor(95, 80, 75, 255), Path: rl.NewColor(130, 120, 125, 255),
		Vehicle: "monstri", Ability: AbilGlide, Proj: ProjBat, ProjCol: rl.NewColor(190, 120, 255, 255), HeroName: "Micul Vampir",
		Zones: []ZoneDef{
			{"Gara din Pădure", "Gara", rl.NewColor(150, 110, 70, 255), rl.NewColor(255, 200, 80, 255)},
			{"Castelul-Hotel", "Castel", rl.NewColor(120, 70, 170, 255), rl.NewColor(255, 210, 90, 255)},
			{"Pădurea cu Licurici", "Licurici", rl.NewColor(50, 120, 70, 255), rl.NewColor(220, 255, 120, 255)},
			{"Satul Vârcolacilor", "Sat", rl.NewColor(140, 100, 70, 255), rl.NewColor(240, 160, 60, 255)},
			{"Lacul cu Ceață", "Lac", rl.NewColor(80, 130, 170, 255), rl.NewColor(200, 220, 255, 255)},
		}},
	{ID: "panze", Name: "Orașul Pânzelor", Main: rl.NewColor(210, 40, 50, 255), Accent: rl.NewColor(40, 90, 210, 255),
		SkyTop: rl.NewColor(100, 170, 240, 255), SkyHorizon: rl.NewColor(220, 235, 250, 255), FogColor: rl.NewColor(205, 220, 240, 255),
		SunColor: rl.NewColor(255, 245, 230, 255), Ambient: 0.5, FogDensity: 1,
		Grass: rl.NewColor(110, 170, 80, 255), Dirt: rl.NewColor(150, 140, 130, 255), Path: rl.NewColor(80, 82, 90, 255),
		Vehicle: "panze", Ability: AbilWeb, Proj: ProjWeb, ProjCol: rl.NewColor(245, 245, 255, 255), HeroName: "Băiatul-Păianjen",
		Zones: []ZoneDef{
			{"Piața Zgârie-Norilor", "Centru", rl.NewColor(210, 40, 50, 255), rl.NewColor(40, 90, 210, 255)},
			{"Cartierul Podurilor", "Poduri", rl.NewColor(90, 120, 200, 255), rl.NewColor(240, 200, 60, 255)},
			{"Parcul Central", "Parc", rl.NewColor(80, 170, 80, 255), rl.NewColor(240, 240, 240, 255)},
			{"Portul Containerelor", "Port", rl.NewColor(240, 140, 40, 255), rl.NewColor(60, 160, 200, 255)},
			{"Turnul Laboratorului", "Turn", rl.NewColor(120, 200, 220, 255), rl.NewColor(150, 60, 200, 255)},
		}},
	{ID: "eroi", Name: "Orașul Eroilor Zburători", Main: rl.NewColor(40, 110, 220, 255), Accent: rl.NewColor(255, 200, 40, 255),
		SkyTop: rl.NewColor(80, 160, 250, 255), SkyHorizon: rl.NewColor(255, 235, 200, 255), FogColor: rl.NewColor(235, 230, 220, 255),
		SunColor: rl.NewColor(255, 240, 200, 255), Ambient: 0.52, FogDensity: 0.8,
		Grass: rl.NewColor(100, 180, 90, 255), Dirt: rl.NewColor(170, 160, 140, 255), Path: rl.NewColor(90, 92, 100, 255),
		Vehicle: "eroi", Ability: AbilFly, Proj: ProjStar, ProjCol: rl.NewColor(255, 220, 60, 255), HeroName: "Eroul Zburător",
		Zones: []ZoneDef{
			{"Piața Statuii", "Statuie", rl.NewColor(40, 110, 220, 255), rl.NewColor(255, 200, 40, 255)},
			{"Ziarul și Gara", "Ziar", rl.NewColor(200, 200, 210, 255), rl.NewColor(230, 50, 50, 255)},
			{"Fortăreața de Cristal", "Cristal", rl.NewColor(150, 220, 255, 255), rl.NewColor(255, 255, 255, 255)},
			{"Docurile", "Docuri", rl.NewColor(200, 120, 60, 255), rl.NewColor(60, 120, 180, 255)},
			{"Muntele cu Antene", "Antene", rl.NewColor(130, 130, 140, 255), rl.NewColor(255, 60, 60, 255)},
		}},
	{ID: "craciun", Name: "Satul de Crăciun", Main: rl.NewColor(200, 30, 40, 255), Accent: rl.NewColor(40, 150, 70, 255),
		SkyTop: rl.NewColor(150, 190, 235, 255), SkyHorizon: rl.NewColor(240, 240, 250, 255), FogColor: rl.NewColor(230, 236, 248, 255),
		SunColor: rl.NewColor(255, 240, 225, 255), Ambient: 0.6, FogDensity: 1.2,
		Grass: rl.NewColor(240, 246, 252, 255), Dirt: rl.NewColor(200, 210, 225, 255), Path: rl.NewColor(190, 160, 130, 255),
		Vehicle: "craciun", Ability: AbilSled, Proj: ProjSnow, ProjCol: rl.NewColor(250, 252, 255, 255), HeroName: "Spiridușul",
		Zones: []ZoneDef{
			{"Piața cu Bradul", "Brad", rl.NewColor(40, 150, 70, 255), rl.NewColor(255, 210, 60, 255)},
			{"Atelierul Jucăriilor", "Atelier", rl.NewColor(200, 30, 40, 255), rl.NewColor(255, 255, 255, 255)},
			{"Satul de Turtă Dulce", "Turta", rl.NewColor(190, 120, 60, 255), rl.NewColor(255, 240, 240, 255)},
			{"Lacul Înghețat", "Patinoar", rl.NewColor(150, 210, 240, 255), rl.NewColor(255, 255, 255, 255)},
			{"Peștera Morocănoșilor", "Pestera", rl.NewColor(100, 170, 70, 255), rl.NewColor(200, 30, 40, 255)},
		}},
	{ID: "paste", Name: "Pajiștea de Paște", Main: rl.NewColor(250, 180, 220, 255), Accent: rl.NewColor(150, 220, 120, 255),
		SkyTop: rl.NewColor(120, 195, 250, 255), SkyHorizon: rl.NewColor(240, 245, 220, 255), FogColor: rl.NewColor(230, 240, 230, 255),
		SunColor: rl.NewColor(255, 248, 225, 255), Ambient: 0.55, FogDensity: 0.8,
		Grass: rl.NewColor(120, 200, 90, 255), Dirt: rl.NewColor(160, 120, 80, 255), Path: rl.NewColor(225, 200, 150, 255),
		Vehicle: "paste", Ability: AbilHop, Proj: ProjEgg, ProjCol: rl.NewColor(255, 150, 200, 255), HeroName: "Iepurașul",
		Zones: []ZoneDef{
			{"Poiana Iepurașilor", "Poiana", rl.NewColor(150, 220, 120, 255), rl.NewColor(250, 180, 220, 255)},
			{"Ferma cu Puișori", "Ferma", rl.NewColor(255, 220, 60, 255), rl.NewColor(200, 70, 50, 255)},
			{"Grădina cu Lalele", "Lalele", rl.NewColor(240, 80, 120, 255), rl.NewColor(255, 230, 80, 255)},
			{"Moara de Vânt", "Moara", rl.NewColor(230, 200, 150, 255), rl.NewColor(120, 170, 230, 255)},
			{"Dealul Ouălor Uriașe", "Oua", rl.NewColor(180, 140, 240, 255), rl.NewColor(255, 200, 80, 255)},
		}},
	{ID: "halloween", Name: "Satul Dovlecilor", Night: true, Main: rl.NewColor(255, 130, 20, 255), Accent: rl.NewColor(140, 60, 200, 255),
		SkyTop: rl.NewColor(28, 20, 62, 255), SkyHorizon: rl.NewColor(140, 75, 85, 255), FogColor: rl.NewColor(95, 65, 100, 255),
		SunColor: rl.NewColor(255, 238, 210, 255), Ambient: 0.62, FogDensity: 0.7,
		Grass: rl.NewColor(80, 100, 60, 255), Dirt: rl.NewColor(110, 80, 60, 255), Path: rl.NewColor(150, 120, 100, 255),
		Vehicle: "halloween", Ability: AbilFly, Proj: ProjCandy, ProjCol: rl.NewColor(255, 150, 40, 255), HeroName: "Micul Vrăjitor",
		Zones: []ZoneDef{
			{"Strada cu Dovleci", "Strada", rl.NewColor(255, 130, 20, 255), rl.NewColor(140, 60, 200, 255)},
			{"Câmpul de Dovleci", "Camp", rl.NewColor(230, 110, 30, 255), rl.NewColor(90, 160, 60, 255)},
			{"Casa Vrăjitoarei Bune", "Vrajitoare", rl.NewColor(120, 60, 180, 255), rl.NewColor(120, 230, 120, 255)},
			{"Târgul de Bomboane", "Bomboane", rl.NewColor(255, 100, 170, 255), rl.NewColor(255, 220, 60, 255)},
			{"Conacul Fantomelor Vesele", "Conac", rl.NewColor(150, 160, 200, 255), rl.NewColor(255, 140, 40, 255)},
		}},
}

var realmDecor = map[string]func(rb *RealmBuilder){}

// registerRealmDecor: functia care pune decorul tematic al lumii (vezi RealmBuilder).
func registerRealmDecor(id string, f func(rb *RealmBuilder)) { realmDecor[id] = f }

func realmIndex(id string) int {
	for i, r := range Realms {
		if r.ID == id {
			return i
		}
	}
	return -1
}

// AddCrown: coroana Sefului, pusa de functia personajului cand s.Crown e true
// (y = varful capului).
func AddCrown(mb *MeshBuilder, y float32) {
	mb.Cylinder(v3(0, y-0.03, 0), v3(0, y+0.14, 0), 0.37, 0.37, 12, LegoGold, true)
	for k := 0; k < 8; k++ {
		a := float32(k) * 3.14159 / 4
		x, z := cosf(a)*0.3, sinf(a)*0.3
		mb.Cylinder(v3(x, y+0.1, z), v3(x, y+0.31, z), 0.07, 0.01, 6, LegoGold, true)
	}
}

// numele monstrilor si ale Sefilor (pentru parinte; copilul vede desenul)
func init() {
	names := map[string]struct {
		e [4]string
		b string
	}{
		"mare":      {[4]string{"Crabul-Pirat", "Rechinul în Slip", "Caracatița Marinar", "Piratul Papagal"}, "Căpitanul Barbă-de-Alge"},
		"monstri":   {[4]string{"Fantomița Obraznică", "Dovleacul Săltăreț", "Mumia Zăpăcită", "Garguiul"}, "Vrăjitoarea Ceții"},
		"panze":     {[4]string{"Robo-Gândacul", "Robo-Scorpionul", "Drona-cu-Picioare", "Gorila Mecanică"}, "Profesorul Magnet"},
		"eroi":      {[4]string{"Drona-Robot", "Robo-Sfera", "Omulețul de Metal", "Extraterestrul Verde"}, "Generalul Fulger"},
		"craciun":   {[4]string{"Morocănosul Verde", "Omul de Zăpadă Obraznic", "Pinguinul Poznaș", "Spiridușul-Hoț"}, "Marele Morocănos"},
		"paste":     {[4]string{"Vulpița Poznașă", "Corbul Hoț de Ouă", "Bursucul Morocănos", "Ariciul Țepos"}, "Vulpoiul Șiret"},
		"halloween": {[4]string{"Dovlecelul Săltăreț", "Scheletelul Dansator", "Pisica Vrăjită", "Liliacul Poznaș"}, "Regele Dovleac"},
	}
	for i := range Realms {
		if n, ok := names[Realms[i].ID]; ok {
			Realms[i].Enemies = n.e
			Realms[i].Boss = n.b
		}
	}
}
