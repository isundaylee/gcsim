package lisa

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

const conductiveTag = "lisa-conductive-stacks"

func (c *char) ChargeAttack(p map[string]int) action.ActionInfo {
	f, a := c.ActionFrames(action.ActionCharge, p)

	//TODO: assumes this applies every time per
	//[7:53 PM] Hold ₼KLEE like others hold GME: CHarge is pyro every charge
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Charge Attack",
		AttackTag:  combat.AttackTagExtra,
		ICDTag:     combat.ICDTagNone,
		ICDGroup:   combat.ICDGroupDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       charge[c.TalentLvlAttack()],
	}

	done := false
	cb := func(a combat.AttackCB) {
		if done {
			return
		}
		count := a.Target.GetTag(conductiveTag)
		if count < 3 {
			a.Target.SetTag(conductiveTag, count+1)
		}
		done = true
	}

	c.Core.Combat.QueueAttack(ai, combat.NewDefCircHit(0.1, false, combat.TargettableEnemy), 0, f, cb)

	return f, a
}

var skillHitmarks = []int{22, 117}

//p = 0 for no hold, p = 1 for hold
func (c *char) Skill(p map[string]int) action.ActionInfo {
	hold := p["hold"]
	if hold == 1 {
		return c.skillHold(p)
	}
	return c.skillPress(p)
}

//TODO: how long do stacks last?
func (c *char) skillPress(p map[string]int) (int, int) {
	f, a := c.ActionFrames(action.ActionSkill, p)
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Violet Arc",
		AttackTag:  combat.AttackTagElementalArt,
		ICDTag:     combat.ICDTagLisaElectro,
		ICDGroup:   combat.ICDGroupDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       skillPress[c.TalentLvlSkill()],
	}

	done := false
	cb := func(a combat.AttackCB) {
		if done {
			return
		}
		count := a.Target.GetTag(conductiveTag)
		if count < 3 {
			a.Target.SetTag(conductiveTag, count+1)
		}
		done = true
	}

	c.Core.Combat.QueueAttack(ai, combat.NewDefSingleTarget(1, combat.TargettableEnemy), 0, skillHitmarks[0], cb)

	if c.Core.Rand.Float64() < 0.5 {
		c.QueueParticle("Lisa", 1, attributes.Electro, f+100)
	}

	c.SetCDWithDelay(action.ActionSkill, 60, 17)
	return f, a
}

func (c *char) skillHold(p map[string]int) (int, int) {
	f, a := c.ActionFrames(action.ActionSkill, p)
	//no multiplier as that's target dependent
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Violet Arc (Hold)",
		AttackTag:  combat.AttackTagElementalArt,
		ICDTag:     combat.ICDTagNone,
		ICDGroup:   combat.ICDGroupDefault,
		Element:    attributes.Electro,
		Durability: 50,
	}

	//c2 add defense? no interruptions either way
	if c.Base.Cons >= 2 {
		//increase def for the duration of this abil in however many frames
		val := make([]float64, attributes.EndStatType)
		val[attributes.DEFP] = 0.25
		c.AddStatMod("lisa-c2",

			c.Core.F+126, attributes.NoStat, func() ([]float64, bool) { return val, true })

	}

	count := 0
	var c1cb func(a combat.AttackCB)
	if c.Base.Cons > 0 {
		c1cb = func(a combat.AttackCB) {
			if count == 5 {
				return
			}
			count++
			c.AddEnergy("lisa-c1", 2)
		}
	}

	//[8:31 PM] ArchedNosi | Lisa Unleashed: yeah 4-5 50/50 with Hold
	//[9:13 PM] ArchedNosi | Lisa Unleashed: @gimmeabreak actually wait, xd i noticed i misread my sheet, Lisa Hold E always gens 5 orbs
	c.Core.Combat.QueueAttack(ai, combat.NewDefCircHit(3, false, combat.TargettableEnemy), 0, skillHitmarks[1], c1cb)

	// count := 4
	// if c.Core.Rand.Float64() < 0.5 {
	// 	count = 5
	// }
	c.QueueParticle("Lisa", 5, attributes.Electro, f+100)

	// c.CD[def.SkillCD] = c.Core.F + 960 //16seconds, starts after 114 frames
	c.SetCDWithDelay(action.ActionSkill, 960, 114)
	return f, a
}

func (c *char) Burst(p map[string]int) action.ActionInfo {

	f, a := c.ActionFrames(action.ActionBurst, p)

	//first zap has no icd
	targ := c.Core.RandomTargetIndex(combat.TargettableEnemy)
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Lightning Rose (Initial)",
		AttackTag:  combat.AttackTagElementalBurst,
		ICDTag:     combat.ICDTagNone,
		ICDGroup:   combat.ICDGroupDefault,
		Element:    attributes.Electro,
		Durability: 0,
		Mult:       0.1,
	}
	c.Core.Combat.QueueAttack(ai, combat.NewDefSingleTarget(targ, combat.TargettableEnemy), f, f, a4cb)

	//duration is 15 seconds, tick every .5 sec
	//30 zaps once every 30 frame, starting at 119

	ai = combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Lightning Rose (Tick)",
		AttackTag:  combat.AttackTagElementalBurst,
		ICDTag:     combat.ICDTagElementalBurst,
		ICDGroup:   combat.ICDGroupDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()],
	}

	for i := 119; i <= 119+900; i += 30 { //first tick at 119

		var cb core.AttackCBFunc
		if c.Base.Cons >= 4 {
			//random 1 to 3 jumps
			count := c.Rand.Intn(3) + 1
			cb = func(a combat.AttackCB) {
				if count == 0 {
					return
				}
				//generate additional attack, random target
				//if we get -1 for a target then that just means there's no target
				//to jump to so that's fine; chain will terminate
				count++
				//grab a list of enemies by range; we assume it'll just hit the closest?

			}
		}
		c.Core.Combat.QueueAttack(ai, combat.NewDefSingleTarget(c.Core.RandomEnemyTarget(), combat.TargettableEnemy), f-1, i, cb, a4cb)
	}

	//add a status for this just in case someone cares
	c.AddTask(func() {
		c.Core.Status.AddStatus("lisaburst", 119+900)
	}, "lisa burst status", f)

	//on lisa c4
	//[8:11 PM] gimmeabreak: er, what does lisa c4 do?
	//[8:11 PM] ArchedNosi | Lisa Unleashed: allows each pulse of the ult to be 2-4 arcs
	//[8:11 PM] ArchedNosi | Lisa Unleashed: if theres enemies given
	//[8:11 PM] gimmeabreak: oh so it jumps 2 to 4 times?
	//[8:11 PM] gimmeabreak: i guess single target it does nothing then?
	//[8:12 PM] ArchedNosi | Lisa Unleashed: yeah single does nothing

	//burst cd starts 53 frames after executed
	//energy usually consumed after 63 frames
	c.ConsumeEnergy(63)
	// c.CD[def.BurstCD] = c.Core.F + 1200
	c.SetCDWithDelay(action.ActionBurst, 1200, 53)
	return f, a
}

func a4cb(a combat.AttackCB) {
	a.Target.AddDefMod("lisa-a4", -0.15, 600)
}
