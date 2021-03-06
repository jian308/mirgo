package main

import (
	"fmt"
	"time"

	"github.com/yenkeia/mirgo/common"
	"github.com/yenkeia/mirgo/proto/server"
)

type Monster struct {
	MapObject
	Image       common.Monster
	AI          int
	Effect      int
	Poison      common.PoisonType
	Light       uint8
	Target      IMapObject
	Level       uint16
	HP          uint32
	MaxHP       uint32
	MinAC       uint16
	MaxAC       uint16
	MinMAC      uint16
	MaxMAC      uint16
	MinDC       uint16
	MaxDC       uint16
	MinMC       uint16
	MaxMC       uint16
	MinSC       uint16
	MaxSC       uint16
	Accuracy    uint8
	Agility     uint8
	MoveSpeed   uint16
	AttackSpeed int32
	ArmourRate  float32
	DamageRate  float32
	Master      *Player
	SearchTime  *time.Time // 怪物下一次搜索目标的时间
}

func (m *Monster) String() string {
	return fmt.Sprintf("Monster: %s, (%v), ID: %d, ptr: %p\n", m.Name, m.CurrentLocation, m.ID, m)
}

func NewMonster(mp *Map, p common.Point, mi *common.MonsterInfo) (m *Monster) {
	m = new(Monster)
	m.ID = mp.Env.NewObjectID()
	m.Map = mp
	m.Name = mi.Name
	m.NameColor = common.Color{R: 255, G: 255, B: 255}
	m.Image = common.Monster(mi.Image)
	m.AI = mi.AI
	m.Effect = mi.Effect
	m.Light = uint8(mi.Light)
	m.Target = nil
	m.Poison = common.PoisonTypeNone
	m.CurrentLocation = p
	m.CurrentDirection = common.MirDirection(G_Rand.RandInt(0, 7))
	m.Level = uint16(mi.Level)
	m.HP = uint32(mi.HP)
	m.MaxHP = uint32(mi.HP)
	m.MinAC = uint16(mi.MinAC)
	m.MaxAC = uint16(mi.MaxAC)
	m.MinMAC = uint16(mi.MinMAC)
	m.MaxMAC = uint16(mi.MaxMAC)
	m.MinDC = uint16(mi.MinDC)
	m.MaxDC = uint16(mi.MaxDC)
	m.MinMC = uint16(mi.MinMC)
	m.MaxMC = uint16(mi.MaxMC)
	m.MinSC = uint16(mi.MinSC)
	m.MaxSC = uint16(mi.MaxSC)
	m.Accuracy = uint8(mi.Accuracy)
	m.Agility = uint8(mi.Agility)
	m.MoveSpeed = uint16(mi.MoveSpeed)
	m.AttackSpeed = int32(mi.AttackSpeed)
	m.ArmourRate = 1.0
	m.DamageRate = 1.0
	t := time.Now()
	m.SearchTime = &t
	return m
}

func (m *Monster) GetID() uint32 {
	return m.ID
}

func (m *Monster) GetRace() common.ObjectType {
	return common.ObjectTypeMonster
}

func (m *Monster) GetPoint() common.Point {
	return m.CurrentLocation
}

func (m *Monster) GetCell() *Cell {
	return m.Map.GetCell(m.CurrentLocation)
}

func (m *Monster) GetDirection() common.MirDirection {
	return m.CurrentDirection
}

func (m *Monster) GetInfo() interface{} {
	res := &server.ObjectMonster{
		ObjectID:          m.ID,
		Name:              m.Name,
		NameColor:         m.NameColor.ToInt32(),
		Location:          m.GetPoint(),
		Image:             m.Image,
		Direction:         m.GetDirection(),
		Effect:            uint8(m.Effect),
		AI:                uint8(m.AI),
		Light:             m.Light,
		Dead:              m.IsDead(),
		Skeleton:          m.IsSkeleton(),
		Poison:            m.Poison,
		Hidden:            m.IsHidden(),
		ShockTime:         0,     // TODO
		BindingShotCenter: false, // TODO
		Extra:             false, // TODO
		ExtraByte:         0,     // TODO
	}
	return res
}

func (m *Monster) GetBaseStats() BaseStats {
	return BaseStats{
		MinAC:    m.MinAC,
		MaxAC:    m.MaxAC,
		MinMAC:   m.MinMAC,
		MaxMAC:   m.MaxMAC,
		MinDC:    m.MinDC,
		MaxDC:    m.MaxDC,
		MinMC:    m.MinMC,
		MaxMC:    m.MaxMC,
		MinSC:    m.MinSC,
		MaxSC:    m.MaxSC,
		Accuracy: m.Accuracy,
		Agility:  m.Agility,
	}
}

func (m *Monster) Broadcast(msg interface{}) {
	m.Map.Submit(NewTask(func(args ...interface{}) {
		grids := m.Map.AOI.GetSurroundGrids(m.CurrentLocation)
		for i := range grids {
			areaPlayers := grids[i].GetAllPlayer()
			for i := range areaPlayers {
				areaPlayers[i].Enqueue(msg)
			}
		}
	}))
}

func (m *Monster) BroadcastDamageIndicator(typ common.DamageType, dmg int) {
	m.Broadcast(ServerMessage{}.DamageIndicator(int32(dmg), typ, m.GetID()))
}

func (m *Monster) IsDead() bool {
	return m.HP <= 0
}

func (m *Monster) IsSkeleton() bool {
	return false
}

func (m *Monster) IsHidden() bool {
	return false
}

func (m *Monster) IsAttackTarget(attacker IMapObject) bool {
	return true
}

func (m *Monster) IsFriendlyTarget(attacker IMapObject) bool {
	return false
}

func (m *Monster) CanMove() bool {
	return true
}

func (m *Monster) CanAttack() bool {
	return true
}

func (m *Monster) InAttackRange() bool {
	return true
}

func (m *Monster) Process() {
	m.ProcessAI()
	m.ProcessBuffs()
	m.ProcessRegan()
	m.ProcessPoison()
}

func (m *Monster) ProcessAI() {
	if m.IsDead() {
		return
	}
	if m.Master != nil {

	}
	m.ProcessSearch()
	m.ProcessRoam()
	m.ProcessTarget()
}

// ProcessSearch 寻找目标
func (m *Monster) ProcessSearch() {
	now := time.Now()
	if m.SearchTime.After(now) {
		return
	}
	*m.SearchTime = now.Add(1 * time.Second)
	if m.CanMove() {
		// walk randomly
		// ok := m.Walk()
	}
	if m.Target == nil {
		m.FindTarget()
	}
}

func (m *Monster) ProcessRoam() {

}

func (m *Monster) ProcessTarget() {
	if m.Target == nil || !m.CanAttack() {
		return
	}
	if m.InAttackRange() {
		m.Attack()
		if m.Target.IsDead() {
			m.FindTarget()
		}
		return
	}
	m.MoveTo(m.Target.GetPoint())
}

func (m *Monster) ProcessBuffs() {

}

// processRegan 怪物自身回血
func (m *Monster) ProcessRegan() {

}

func (m *Monster) ProcessPoison() {

}

func (m *Monster) GetDefencePower(min, max int) int {
	if min < 0 {
		min = 0
	}
	if min > max {
		max = min
	}
	return G_Rand.RandInt(min, max+1)
}

func (m *Monster) Die() {
	if m.IsDead() {
		return
	}
	m.HP = 0
	m.Broadcast(ServerMessage{}.ObjectDied(m.GetID(), m.GetDirection(), m.GetPoint()))
	// EXPOwner.WinExp(Experience, Level);
	m.Drop()
}

// ChangeHP 怪物改变血量 amount 可以是负数(扣血)
func (m *Monster) ChangeHP(amount int) {
	if m.IsDead() {
		return
	}
	value := int(m.HP) + amount
	if value == int(m.HP) {
		return
	}
	if value <= 0 {
		m.Die()
		m.HP = 0
	} else {
		m.HP = uint32(value)
	}
	percent := uint8(float32(m.HP) / float32(m.MaxHP) * 100)
	log.Debugf("monster HP: %d, MaxHP: %d, percent: %d\n", m.HP, m.MaxHP, percent)
	m.Broadcast(ServerMessage{}.ObjectHealth(m.GetID(), percent, 5))
}

// Attacked 被攻击
func (m *Monster) Attacked(attacker IMapObject, damage int, defenceType common.DefenceType, damageWeapon bool) {
	if m.Target == nil && attacker.IsAttackTarget(m) {
		m.Target = attacker
	}
	armor := 0
	switch defenceType {
	case common.DefenceTypeACAgility:
		if G_Rand.RandInt(0, int(m.Agility)+1) > int(attacker.GetBaseStats().Accuracy) {
			m.BroadcastDamageIndicator(common.DamageTypeMiss, 0)
			return
		}
		armor = m.GetDefencePower(int(m.MinAC), int(m.MaxAC))
	case common.DefenceTypeAC:
		armor = m.GetDefencePower(int(m.MinAC), int(m.MaxAC))
	case common.DefenceTypeMACAgility:
		if G_Rand.RandInt(0, int(m.Agility)+1) > int(attacker.GetBaseStats().Accuracy) {
			m.BroadcastDamageIndicator(common.DamageTypeMiss, 0)
			return
		}
		armor = m.GetDefencePower(int(m.MinMAC), int(m.MaxMAC))
	case common.DefenceTypeMAC:
		armor = m.GetDefencePower(int(m.MinMAC), int(m.MaxMAC))
	case common.DefenceTypeAgility:
		if G_Rand.RandInt(0, int(m.Agility)+1) > int(attacker.GetBaseStats().Accuracy) {
			m.BroadcastDamageIndicator(common.DamageTypeMiss, 0)
			return
		}
	}
	armor = int(float32(armor) * m.ArmourRate)
	damage = int(float32(damage) * m.DamageRate)
	value := damage - armor
	log.Debugf("attacker damage: %d, monster armor: %d\n", damage, armor)
	if value <= 0 {
		m.BroadcastDamageIndicator(common.DamageTypeMiss, 0)
		return
	}
	// TODO 还有很多没做
	m.Broadcast(ServerMessage{}.ObjectStruck(m, attacker.GetID()))
	m.BroadcastDamageIndicator(common.DamageTypeHit, value)
	m.ChangeHP(-value)
}

// Drop 怪物掉落物品
func (m *Monster) Drop() {
	value, ok := m.Map.Env.GameDB.DropInfoMap.Load(m.Name)
	if !ok {
		return
	}
	dropInfos := value.([]common.DropInfo)
	mapItems := make([]Item, 0)
	for i := range dropInfos {
		drop := dropInfos[i]
		if G_Rand.RandInt(1, drop.Chance+1) != 1 {
			continue
		}
		if drop.Gold > 0 {
			mapItems = append(mapItems, Item{
				MapObject: MapObject{
					ID:  m.Map.Env.NewObjectID(),
					Map: m.Map,
				},
				Gold:     uint64(drop.Gold),
				UserItem: nil,
			})
			continue
		}
		info := m.Map.Env.GameDB.GetItemInfoByName(drop.ItemName)
		if info == nil {
			continue
		}
		mapItems = append(mapItems, Item{
			MapObject: MapObject{
				ID:  m.Map.Env.NewObjectID(),
				Map: m.Map,
			},
			Gold:     0,
			UserItem: m.Map.Env.NewUserItem(info),
		})
	}
	for i := range mapItems {
		if msg, ok := mapItems[i].Drop(m.GetPoint(), 3); !ok {
			log.Warnln(msg)
		}
	}
}

// FindTarget 怪物寻找攻击目标
func (m *Monster) FindTarget() {

}

// Walk 移动，成功返回 true
func (m *Monster) Walk(dir common.MirDirection) bool {
	return true
}

func (m *Monster) Attack() {

}

func (m *Monster) MoveTo(p common.Point) {

}
