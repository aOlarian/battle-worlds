package main

import (
	"fmt"
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// V 5.0: MONSTRII PE LUMI
//   - nivelul N (N <= 7): monstrii stau in lumea N (cea deschisa acum);
//   - dupa ce ii invingi pe toti, apare SEFUL (mare, cu coroana);
//   - Seful invins = nivel trecut -> se deschide lumea urmatoare;
//   - dupa nivelul 7: monstri pe toate insulele, Seful in piata.
// ====================================================================

func (g *Game) levelWorld(level int) int {
	if level <= NumWorlds {
		return level - 1
	}
	return -1
}

func (g *Game) spawnEnemies(level int) {
	hpM, atkM, _ := g.enemyMultipliers()
	g.Enemies = g.Enemies[:0]
	g.AimEnemy = -1 // tinta veche ar arata spre un monstru care nu mai exista
	g.Attacker = -1
	g.Prog.BossOut = false
	rng := rand.New(rand.NewSource(int64(level*977 + 13)))
	wl := g.levelWorld(level)
	count := 5 + level
	if wl < 0 {
		count = 12 + level
	}
	if count > 22 {
		count = 22
	}
	w := g.City.W
	placeIn := func(si int) (rl.Vector3, rl.Vector2, float32) {
		home := Sectors[si].Center
		hr := float32(44)
		if curRealm > 0 {
			hr = realmZoneR(si) - 16
		} else {
			switch si {
			case 4:
				home, hr = rl.NewVector2(0, 6), 34
			case SecWest:
				hr = 40
			case SecPark:
				hr = 46
			}
		}
		for tries := 0; tries < 200; tries++ {
			a := rng.Float32() * 2 * math.Pi
			d := 10 + rng.Float32()*(hr-10)
			x := home.X + cosf(a)*d
			z := home.Y + sinf(a)*d
			if !w.CanStand(x, z, 0.3, 0.8) || w.LandSD(x, z) > -8 || SectorAt(x, z) != si {
				continue
			}
			gy, k := w.GroundAt(x, z, 0.5)
			if k == SurfWater || absf(gy) > 0.6 {
				continue
			}
			return v3(x, gy, z), home, hr
		}
		return v3(home.X, 0.3, home.Y+12), home, hr
	}
	mk := func(i int, si int, boss bool) Enemy {
		wi := WorldOfSector(si)
		if wi < 0 {
			wi = 0
		}
		pos, home, hr := placeIn(si)
		style := ThemedEnemyStyle(wi, i)
		if boss && realmChars[RealmCharCode(curRealm, RoleBoss)] != nil {
			style = RealmCharStyle(curRealm, RoleBoss)
		}
		k := (wi*3 + i) % len(enemyNames)
		lv := level + curRealm*4
		baseHp := float64(260+lv*45) * hpM
		scale := 1.15 + float32(i%3)*0.12
		e := Enemy{
			Name:    fmt.Sprintf("%s (Niv. %d)", realmEnemyName(wi, i, k), level),
			Kind:    k,
			Pos:     pos,
			Yaw:     rng.Float32() * 6.28,
			HP:      baseHp,
			MaxHP:   baseHp,
			Attack:  (22 + float64(lv*4)) * atkM,
			Defense: 16 + float64(level*3),
			Gold:    int64(300 + level*120),
			Silver:  int64(150 + level*60),
			Alive:   true,
			Home:    home,
			HomeR:   hr,
			Ranged:  enemyRanged[k],
			AtkCD:   1 + rng.Float32()*2,
			Pose:    Pose{Walk: rng.Float32() * 6},
		}
		if boss {
			style.Crown = true
			e.Boss = true
			e.Alive = false
			e.HP = baseHp * 5
			e.MaxHP = e.HP
			e.Attack *= 1.4
			e.Gold *= 4
			scale = 2.4
			e.Name = "Șeful " + Worlds[wi].Name
			if b := Realms[curRealm].Boss; b != "" {
				e.Name = b
			}
			e.Ranged = true
		}
		e.Meshes = GetCharMeshes(style)
		e.Scale = scale
		return e
	}
	for i := 0; i < count; i++ {
		si := 4
		if wl >= 0 {
			si = Worlds[wl].Sector
		} else {
			si = Worlds[i%NumWorlds].Sector
		}
		g.Enemies = append(g.Enemies, mk(i, si, false))
	}
	bs := 4
	if wl >= 0 {
		bs = Worlds[wl].Sector
	}
	g.Enemies = append(g.Enemies, mk(count, bs, true))
}

// regularLeft: cati monstri obisnuiti mai sunt (fara Sef).
func (g *Game) regularLeft() int {
	n := 0
	for i := range g.Enemies {
		if g.Enemies[i].Alive && !g.Enemies[i].Boss {
			n++
		}
	}
	return n
}

func (g *Game) bossIndex() int {
	for i := range g.Enemies {
		if g.Enemies[i].Boss {
			return i
		}
	}
	return -1
}

// afterDefeat: dupa fiecare monstru invins (inlocuieste "aliveEnemies()==0").
func (g *Game) afterDefeat() {
	if g.aliveEnemies() > 0 {
		return
	}
	bi := g.bossIndex()
	if bi >= 0 && !g.Prog.BossOut {
		g.releaseBoss(bi)
		return
	}
	g.levelComplete()
}

// releaseBoss: Seful apare (la ~20 m de Aventurierul, in lumea nivelului).
func (g *Game) releaseBoss(bi int) {
	e := &g.Enemies[bi]
	g.Prog.BossOut = true
	w := g.City.W
	p := g.Player.Pos
	si := SectorAt(e.Pos.X, e.Pos.Z)
	best := e.Pos
	for k := 0; k < 24; k++ {
		a := float32(k) / 24 * 2 * math.Pi
		x, z := p.X+cosf(a)*20, p.Z+sinf(a)*20
		if SectorAt(x, z) != si || !w.CanStand(x, z, 0.3, 1.6) {
			continue
		}
		gy, kk := w.GroundAt(x, z, p.Y+2)
		if kk == SurfWater || absf(gy-p.Y) > 3 {
			continue
		}
		best = v3(x, gy, z)
		break
	}
	e.Pos = best
	e.Home = rl.NewVector2(best.X, best.Z)
	e.Alive = true
	e.Alert = true
	e.Stun = 1.2
	g.sayIcon(IconCrown, "A venit Șeful! Învinge-l și se deschide o insulă nouă!")
	g.Play(g.Audio.Engage)
	g.CamShake = 0.5
	for n := 0; n < 70; n++ {
		c := []rl.Color{LegoGoldBright, LegoPurple, LegoWhite}[n%3]
		g.spawnParticle(v3(best.X, best.Y+2, best.Z), v3((rand.Float32()-0.5)*10, rand.Float32()*10, (rand.Float32()-0.5)*10), c, 0.3, 1.5)
	}
}

// realmEnemyName: numele monstrului dupa lumea mare (acelasi rol ca in ThemedEnemyStyle).
func realmEnemyName(wi, i, k int) string {
	R := &Realms[curRealm]
	if R.Enemies[0] != "" {
		return R.Enemies[((wi+i)%4+4)%4]
	}
	return enemyNames[k]
}
