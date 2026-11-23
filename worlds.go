package main

import rl "github.com/gen2brain/raylib-go/raylib"

// ====================================================================
// V 5.0: LUMILE (insulele tematice) si ordinea in care se deschid.
// Nivelul 1 = doar piata din mijloc; fiecare nivel trecut deschide
// urmatoarea insula (norii se risipesc, apare podul / tunelul spre ea).
// Indicii de sector sunt cei din Sectors (world.go).
// ====================================================================

type WorldDef struct {
	Sector int      // indexul din Sectors
	Name   string   // pentru parinte (copilul vede pictograma si culorile)
	Short  string   // nume scurt pentru harta
	Main   rl.Color // culoarea principala a lumii (harta, banner, cabina)
	Accent rl.Color
	Outfit int // costumul primit la prima intrare (index in Outfits)
}

// Ordinea de deschidere: WorldOrder[nivel-1].
var Worlds = []WorldDef{
	{Sector: 4, Name: "Portul Soarelui", Short: "Port", Main: rl.NewColor(240, 160, 40, 255), Accent: rl.NewColor(40, 120, 200, 255), Outfit: -1},
	{Sector: 1, Name: "Golful Piraților", Short: "Pirati", Main: rl.NewColor(150, 90, 50, 255), Accent: rl.NewColor(230, 40, 40, 255), Outfit: 1},
	{Sector: 0, Name: "Jungla Papagalilor", Short: "Jungla", Main: rl.NewColor(60, 170, 70, 255), Accent: rl.NewColor(255, 120, 40, 255), Outfit: -1},
	{Sector: 3, Name: "Insula Farului", Short: "Far", Main: rl.NewColor(230, 60, 60, 255), Accent: rl.NewColor(250, 250, 250, 255), Outfit: -1},
	{Sector: 2, Name: "Recifele de Corali", Short: "Corali", Main: rl.NewColor(255, 120, 150, 255), Accent: rl.NewColor(60, 200, 220, 255), Outfit: -1},
	{Sector: 5, Name: "Plaja Soarelui", Short: "Plaja", Main: rl.NewColor(40, 190, 210, 255), Accent: rl.NewColor(255, 150, 60, 255), Outfit: -1},
	{Sector: 6, Name: "Parcul Acvatic", Short: "Aqua", Main: rl.NewColor(60, 140, 240, 255), Accent: rl.NewColor(255, 220, 50, 255), Outfit: -1},
}

const MaxZones = 7 // cele mai multe insule-niveluri intr-o lume mare

// NumWorlds: cate insule-niveluri are lumea mare curenta (7 la mare, 5 in rest).
var NumWorlds = 7

// WorldOfSector: indexul lumii (0..NumWorlds-1) pentru un sector, sau -1.
func WorldOfSector(si int) int {
	for i, w := range Worlds {
		if w.Sector == si {
			return i
		}
	}
	return -1
}

// numele zonelor (banner, harta) devin numele lumilor
func init() {
	for _, w := range Worlds {
		Sectors[w.Sector].Name = w.Name
	}
}
