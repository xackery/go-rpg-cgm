package game_map

import (
	"reflect"

	"github.com/faiface/pixel/pixelgl"
	"github.com/steelx/go-rpg-cgm/animation"
	"github.com/steelx/go-rpg-cgm/state_machine"
)

type CSStandBy struct {
	Name        string
	Character   *Character
	CombatState *CombatState
	Entity      *Entity
	Anim        animation.Animation
	AnimId      string
}

//char *Character, cs *CombatState
func CSStandByCreate(args ...interface{}) state_machine.State {
	charV := reflect.ValueOf(args[0])
	char := charV.Interface().(*Character)
	csV := reflect.ValueOf(args[1])
	cs := csV.Interface().(*CombatState)

	return &CSStandBy{
		Name:        csStandby,
		Character:   char,
		CombatState: cs,
		Entity:      char.Entity,
		Anim:        animation.Create([]int{char.Entity.StartFrame}, true, 0.16),
	}
}

func (s CSStandBy) IsFinished() bool {
	return true
}

func (s *CSStandBy) Enter(data ...interface{}) {
	s.AnimId = reflect.ValueOf(data[0]).Interface().(string)
	frames := s.Character.GetCombatAnim(s.AnimId)
	s.Anim.SetFrames(frames)
}

func (s *CSStandBy) Render(win *pixelgl.Window) {
	//The *CombatState will do the render for us
}

func (s *CSStandBy) Exit() {
}

func (s *CSStandBy) Update(dt float64) {
	s.Anim.Update(dt)
	s.Entity.SetFrame(s.Anim.Frame())
}
