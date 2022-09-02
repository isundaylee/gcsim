package tighnari

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

var skillFrames []int

const (
	skillHitmark           = 20
	vijnanasuffusionStatus = "vijnanasuffusion"
	wreatharrows           = "wreatharrows"
)

func init() {
	skillFrames = frames.InitAbilSlice(30)
	skillFrames[action.ActionAttack] = 20
	skillFrames[action.ActionAim] = 20
	skillFrames[action.ActionBurst] = 22
	skillFrames[action.ActionDash] = 23
	skillFrames[action.ActionJump] = 23
	skillFrames[action.ActionSwap] = 21
}

func (c *char) Skill(p map[string]int) action.ActionInfo {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Vijnana-Phala Mine",
		AttackTag:  combat.AttackTagElementalArt,
		ICDTag:     combat.ICDTagNone,
		ICDGroup:   combat.ICDGroupDefault,
		StrikeType: combat.StrikeTypeDefault,
		Element:    attributes.Dendro,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(c.Core.Combat.Player(), 2, false, combat.TargettableEnemy),
		0,
		skillHitmark,
	)

	var count float64 = 3
	if c.Core.Rand.Float64() < 0.5 {
		count++
	}
	c.Core.QueueParticle("tighnari", count, attributes.Dendro, skillHitmark+c.ParticleDelay)
	c.SetCDWithDelay(action.ActionSkill, 12*60, 13)

	c.AddStatus(vijnanasuffusionStatus, 12*60, false)
	c.SetTag(wreatharrows, 3)

	if c.Base.Cons >= 2 {
		c.QueueCharTask(c.c2, skillHitmark+1)
	}

	return action.ActionInfo{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionSwap], // earliest cancel
		State:           action.SkillState,
	}
}
