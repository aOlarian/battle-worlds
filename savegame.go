package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PlayerSave represents a saved player profile
type PlayerSave struct {
	Name       string  `json:"name"`
	Level      int     `json:"level"`
	Gold       int64   `json:"gold"`
	Silver     int64   `json:"silver"`
	Attack     float64 `json:"attack"`
	Defense    float64 `json:"defense"`
	HP         float64 `json:"hp"`
	MaxHP      float64 `json:"maxHP"`
	Score      int64   `json:"score"`
	Difficulty string  `json:"difficulty"`

	// blocuri progress
	LegoStages   []int `json:"legoStages"`
	LegoUpgrades []int `json:"legoUpgrades"`
	TotalBuilt   int   `json:"totalBuilt"`

	// Upgrades purchased
	BonusDamage float64 `json:"bonusDamage"`
	BonusHP     float64 `json:"bonusHP"`
	BonusShield float64 `json:"bonusShield"`

	// V 4.0
	Achievements   []string `json:"achievements,omitempty"`
	ZonesVisited   []bool   `json:"zonesVisited,omitempty"`
	BridgesCrossed []bool   `json:"bridgesCrossed,omitempty"`

	// V 5.0
	Visited   []bool   `json:"visited,omitempty"`
	Outfits   []string `json:"outfits,omitempty"`
	Outfit    string   `json:"outfit,omitempty"`
	Gun       int      `json:"gun,omitempty"`
	Bricks    []int    `json:"bricks,omitempty"` // lume*5 + index
	BeachPads []bool   `json:"beachPads,omitempty"`
	RodeWheel bool     `json:"rodeWheel,omitempty"`
	RodeCoast bool     `json:"rodeCoast,omitempty"`

	// V 6.0: lumile mari
	Realm    int         `json:"realm,omitempty"`
	Unlocked int         `json:"unlockedRealm,omitempty"`
	Realms   []RealmSave `json:"realms,omitempty"`
}

// RealmSave: progresul dintr-o lume mare.
type RealmSave struct {
	Seen      bool   `json:"seen,omitempty"`
	Level     int    `json:"level"`
	Visited   []bool `json:"visited,omitempty"`
	Bricks    []int  `json:"bricks,omitempty"`
	BeachPads []bool `json:"beachPads,omitempty"`
	RodeWheel bool   `json:"rodeWheel,omitempty"`
	RodeCoast bool   `json:"rodeCoast,omitempty"`
	Stages    []int  `json:"stages,omitempty"`
	Upgrades  []int  `json:"upgrades,omitempty"`
	Stars     int    `json:"stars,omitempty"`
}

// SaveData holds all player saves
type SaveData struct {
	Players   []PlayerSave `json:"players"`
	MouseSens int          `json:"mouseSens,omitempty"` // 1..5 (0 = implicit)
}

func getSavePath() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = "."
	}
	dir := filepath.Join(appData, "BattleWorlds")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "battle_save_v6.json")
}

// V 6.0 are salvarea ei; prima data preia doar numele jucatorilor din V 5 (sau V 4), cu progres nou.
func importV4Names() *SaveData {
	dir := filepath.Dir(getSavePath())
	data, err := os.ReadFile(filepath.Join(dir, "battle_save_v5.json"))
	if err != nil {
		data, err = os.ReadFile(filepath.Join(dir, "battle_save.json"))
	}
	if err != nil {
		return nil
	}
	var old SaveData
	if json.Unmarshal(data, &old) != nil {
		return nil
	}
	out := &SaveData{Players: []PlayerSave{}, MouseSens: old.MouseSens}
	for _, p := range old.Players {
		out.Players = append(out.Players, CreatePlayer(p.Name))
	}
	fmt.Printf("Am preluat %d nume din salvarea V 4 (progres nou)\n", len(out.Players))
	return out
}

func LoadPlayers() *SaveData {
	path := getSavePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if s := importV4Names(); s != nil {
			return s
		}
		fmt.Println("Nu exista salvare, pornim de la zero")
		return &SaveData{Players: []PlayerSave{}}
	}

	var save SaveData
	if err := json.Unmarshal(data, &save); err != nil {
		fmt.Println("Eroare la citirea salvarii:", err)
		return &SaveData{Players: []PlayerSave{}}
	}

	fmt.Printf("Am incarcat %d jucator(i)\n", len(save.Players))
	return &save
}

func SavePlayers(save *SaveData) error {
	path := getSavePath()
	data, err := json.MarshalIndent(save, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling save data: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("nu pot scrie salvarea: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("nu pot inlocui salvarea: %w", err)
	}

	return nil
}

func CreatePlayer(name string) PlayerSave {
	return PlayerSave{
		Name:         name,
		Level:        1,
		Gold:         1000,
		Silver:       500,
		Attack:       45,
		Defense:      30,
		HP:           250,
		MaxHP:        250,
		Score:        0,
		Difficulty:   "medium",
		LegoStages:   make([]int, 13),
		LegoUpgrades: make([]int, 13),
		TotalBuilt:   0,
	}
}

func FindPlayer(save *SaveData, name string) (int, *PlayerSave) {
	for i := range save.Players {
		if save.Players[i].Name == name {
			return i, &save.Players[i]
		}
	}
	return -1, nil
}

func UpdatePlayerData(p *PlayerSave, g *Game) {
	p.Level = g.Level
	p.Gold = g.Player.Gold
	p.Silver = g.Player.Silver
	p.Attack = g.Player.Attack
	p.Defense = g.Player.Defense
	p.HP = g.Player.HP
	p.MaxHP = g.Player.MaxHP
	p.Score = g.Score
	p.Difficulty = g.Difficulty
	p.BonusDamage = g.BonusDamage
	p.BonusHP = g.BonusHP
	p.BonusShield = g.BonusShield
	if g.City != nil {
		n := len(g.City.LegoSites)
		p.LegoStages = make([]int, n)
		p.LegoUpgrades = make([]int, n)
		for i := range g.City.LegoSites {
			p.LegoStages[i] = g.City.LegoSites[i].CurrentStage
			p.LegoUpgrades[i] = g.City.LegoSites[i].UpgradeLevel
		}
	}
	p.TotalBuilt = g.TotalBuilt
	p.Achievements = p.Achievements[:0]
	for _, a := range g.Achieve.List {
		if a.Unlocked {
			p.Achievements = append(p.Achievements, a.ID)
		}
	}
	p.ZonesVisited = append([]bool{}, g.ZonesVisited[:]...)
	p.BridgesCrossed = append([]bool{}, g.BridgesCrossed[:]...)
	saveProg(p, &g.Prog)
	saveRealms(p, g)
}

func LoadPlayerIntoGame(p *PlayerSave, g *Game) {
	g.backToIslands()
	g.Level = p.Level
	if g.Level < 1 {
		g.Level = 1
	}
	g.Player.Name = p.Name
	g.Player.Gold = p.Gold
	g.Player.Silver = p.Silver
	g.Player.Attack = p.Attack
	g.Player.Defense = p.Defense
	g.Player.MaxHP = p.MaxHP
	g.Player.HP = p.MaxHP
	g.Score = p.Score
	g.Difficulty = p.Difficulty
	if g.Difficulty == "" {
		g.Difficulty = "medium"
	}
	for i, d := range difficultyOptions {
		if d.Key == g.Difficulty {
			g.DifficultyIdx = i
		}
	}
	g.BonusDamage = p.BonusDamage
	g.BonusHP = p.BonusHP
	g.BonusShield = p.BonusShield
	if g.Player.MaxHP <= 0 || g.Player.Attack <= 0 {
		g.applyDifficulty()
	}
	if g.City != nil {
		for i := range g.City.LegoSites {
			s := &g.City.LegoSites[i]
			s.CurrentStage = 0
			s.UpgradeLevel = 1
			if i < len(p.LegoStages) {
				s.CurrentStage = p.LegoStages[i]
				if s.CurrentStage > s.TotalStages {
					s.CurrentStage = s.TotalStages
				}
			}
			if i < len(p.LegoUpgrades) && p.LegoUpgrades[i] > 1 {
				s.UpgradeLevel = p.LegoUpgrades[i]
				if s.UpgradeLevel > 5 {
					s.UpgradeLevel = 5
				}
			}
			s.TargetScale = float32(s.UpgradeLevel)
			s.CurrentScale = s.TargetScale
			s.HammerActive = false
		}
	}
	g.TotalBuilt = p.TotalBuilt
	g.Achieve = InitAchievements()
	for _, id := range p.Achievements {
		g.Achieve.MarkUnlocked(id)
	}
	g.ZonesVisited = [5]bool{}
	for i := 0; i < len(p.ZonesVisited) && i < 5; i++ {
		g.ZonesVisited[i] = p.ZonesVisited[i]
	}
	g.BridgesCrossed = [4]bool{}
	for i := 0; i < len(p.BridgesCrossed) && i < 4; i++ {
		g.BridgesCrossed[i] = p.BridgesCrossed[i]
	}
	loadProg(p, g)
	loadRealms(p, g)
	g.loadRealmState()
	if p.Realm > 0 && p.Realm < len(Realms) {
		g.switchRealm(p.Realm)
	}
}

func saveProg(p *PlayerSave, pr *ProgState) {
	p.Visited = append([]bool{}, pr.Visited[:]...)
	p.Outfits = p.Outfits[:0]
	for i, own := range pr.OwnOutfit {
		if own {
			p.Outfits = append(p.Outfits, Outfits[i].ID)
		}
	}
	p.Outfit = Outfits[clampInt(pr.Outfit, 0, len(Outfits)-1)].ID
	p.Gun = pr.Gun
	p.Bricks = p.Bricks[:0]
	for wi := range pr.Bricks {
		for i, b := range pr.Bricks[wi] {
			if b {
				p.Bricks = append(p.Bricks, wi*BricksPerWorld+i)
			}
		}
	}
	p.BeachPads = append([]bool{}, pr.BeachPads[:]...)
	p.RodeWheel = pr.RodeWheel
	p.RodeCoast = pr.RodeCoast
}

func loadProg(p *PlayerSave, g *Game) {
	g.Prog = newProgState()
	g.Prog.GPS.World = -1
	pr := &g.Prog
	for i := 0; i < len(p.Visited) && i < NumWorlds; i++ {
		pr.Visited[i] = p.Visited[i]
	}
	for _, id := range p.Outfits {
		for i, o := range Outfits {
			if o.ID == id {
				pr.OwnOutfit[i] = true
			}
		}
	}
	for i, o := range Outfits {
		if o.ID == p.Outfit && pr.OwnOutfit[i] {
			pr.Outfit = i
		}
	}
	for _, b := range p.Bricks {
		if wi, i := b/BricksPerWorld, b%BricksPerWorld; wi >= 0 && wi < NumWorlds {
			pr.Bricks[wi][i] = true
		}
	}
	for i := 0; i < len(p.BeachPads) && i < 2; i++ {
		pr.BeachPads[i] = p.BeachPads[i]
	}
	pr.RodeWheel = p.RodeWheel
	pr.RodeCoast = p.RodeCoast
	g.syncFog()
	pr.LastStars = g.totalStars()
	pr.Gun = 0
	if g.gunUnlocked(p.Gun) {
		pr.Gun = p.Gun
	}
	g.Player.Pose.Gun = pr.Gun
	g.Player.Meshes = GetCharMeshes(Outfits[pr.Outfit].Style)
}

// V 6.0: lumile mari in salvare
func saveRealms(p *PlayerSave, g *Game) {
	g.saveRealmState()
	p.Realm = curRealm
	p.Unlocked = g.UnlockedRealm
	p.Realms = p.Realms[:0]
	for _, rs := range g.RealmStates {
		s := RealmSave{Seen: rs.Seen, Level: rs.Level, BeachPads: append([]bool{}, rs.BeachPads[:]...), RodeWheel: rs.RodeWheel, RodeCoast: rs.RodeCoast,
			Stages: append([]int{}, rs.Stages...), Upgrades: append([]int{}, rs.Upgrades...), Stars: rs.Stars}
		s.Visited = append([]bool{}, rs.Visited[:]...)
		for wi := range rs.Bricks {
			for i, b := range rs.Bricks[wi] {
				if b {
					s.Bricks = append(s.Bricks, wi*BricksPerWorld+i)
				}
			}
		}
		p.Realms = append(p.Realms, s)
	}
}

func loadRealms(p *PlayerSave, g *Game) {
	g.RealmStates = nil
	g.ensureRealmStates()
	for ri, s := range p.Realms {
		if ri >= len(g.RealmStates) {
			break
		}
		rs := &g.RealmStates[ri]
		rs.Seen, rs.Level, rs.RodeWheel, rs.RodeCoast, rs.Stars = s.Seen, maxInt(s.Level, 1), s.RodeWheel, s.RodeCoast, s.Stars
		for i := 0; i < len(s.Visited) && i < MaxZones; i++ {
			rs.Visited[i] = s.Visited[i]
		}
		for i := 0; i < len(s.BeachPads) && i < 2; i++ {
			rs.BeachPads[i] = s.BeachPads[i]
		}
		for _, b := range s.Bricks {
			if wi, i := b/BricksPerWorld, b%BricksPerWorld; wi >= 0 && wi < MaxZones {
				rs.Bricks[wi][i] = true
			}
		}
		rs.Stages = append([]int{}, s.Stages...)
		rs.Upgrades = append([]int{}, s.Upgrades...)
	}
	if len(p.Realms) == 0 {
		// salvare fara lumi: progresul vechi e al lumii 0
		rs := &g.RealmStates[0]
		rs.Level = maxInt(p.Level, 1)
		rs.Visited = g.Prog.Visited
		rs.Bricks = g.Prog.Bricks
		rs.Stages = append([]int{}, p.LegoStages...)
		rs.Upgrades = append([]int{}, p.LegoUpgrades...)
	}
	g.UnlockedRealm = p.Unlocked
}
