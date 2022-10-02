package game_map

import (
	"fmt"
	"math"
	"reflect"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/sirupsen/logrus"
	"github.com/steelx/go-rpg-cgm/combat"
	"github.com/steelx/go-rpg-cgm/gui"
	"github.com/steelx/go-rpg-cgm/utilz"
	"github.com/steelx/go-rpg-cgm/world"
)

type CombatChoiceState struct {
	Stack       *gui.StateStack //The internal stack of states from the CombatState object.
	CombatState *CombatState
	World       *combat.WorldExtended
	Actor       *combat.Actor
	Character   *Character
	UpArrow,
	DownArrow,
	Marker *pixel.Sprite
	UpArrowPosition,
	DownArrowPosition,
	MarkerPosition pixel.Vec
	time      float64
	Selection *gui.SelectionMenu
	textbox   *gui.Textbox
	mHide     bool
}

func CombatChoiceStateCreate(combatState *CombatState, owner *combat.Actor) *CombatChoiceState {
	c := &CombatChoiceState{
		CombatState: combatState,
		Stack:       combatState.InternalStack,
		World:       reflect.ValueOf(combatState.GameState.Globals["world"]).Interface().(*combat.WorldExtended),
		Actor:       owner,
		Character:   combatState.ActorCharMap[owner],
		UpArrow:     world.IconsDB.Get(11),
		DownArrow:   world.IconsDB.Get(12),
		Marker:      pixel.NewSprite(gui.ContinueCaretPng, gui.ContinueCaretPng.Bounds()),
	}
	c.MarkerPosition = c.Character.Entity.GetSelectPosition()
	c.CreateActionDialog(owner.Actions)
	return c
}

func (c *CombatChoiceState) Enter() {
	c.CombatState.SelectedActor = c.Actor
}

func (c *CombatChoiceState) Exit() {
	c.CombatState.SelectedActor = nil
}

func (c *CombatChoiceState) Update(dt float64) bool {
	c.textbox.Update(dt)
	c.bounceMarker(dt)
	return true
}

func (c CombatChoiceState) Render(renderer *pixelgl.Window) {
	c.textbox.Render(renderer)

	c.Marker.Draw(renderer, pixel.IM.Moved(c.MarkerPosition))
}

func (c CombatChoiceState) HandleInput(win *pixelgl.Window) {
	c.Selection.HandleInput(win)
}

func (c *CombatChoiceState) OnSelect(index int, str interface{}) {
	actionItem := reflect.ValueOf(str).Interface().(string)
	if actionItem == combat.ActionAttack {
		c.Selection.HideCursor()

		state := CombatTargetStateCreate(c.CombatState, CombatChoiceParams{
			OnSelect: func(targets []*combat.Actor) {
				c.TakeAction(actionItem, targets)
			},
			OnExit: func() {
				c.Selection.ShowCursor()
			},
			SwitchSides:     true,
			DefaultSelector: nil,
			TargetType:      world.CombatTargetTypeONE,
		})
		c.Stack.Push(state)
		return
	}

	if actionItem == combat.ActionFlee {
		c.Stack.Pop() // choice state
		queue := c.CombatState.EventQueue
		event := CEFleeCreate(c.CombatState, c.Actor, CSMoveParams{Dir: 8, Distance: 180, Time: 0.6})
		tp := event.TimePoints(queue)
		queue.Add(event, tp)
		return
	}

	if actionItem == combat.ActionItem {
		c.OnItemAction()
		return
	}

	if actionItem == combat.ActionMagic {
		c.OnMagicAction()
		return
	}

	if actionItem == combat.ActionSpecial && len(c.Actor.Special) > 0 {
		c.OnSpecialAction()
		return
	}
}

//TakeAction function pops the CombatTargetState and CombatChoiceState off the
//stack. This leaves the CombatState internal stack empty and causes the EventQueue
//to start updating again.
func (c *CombatChoiceState) TakeAction(id string, targets []*combat.Actor) {
	c.Stack.Pop() //select state
	c.Stack.Pop() //action state

	if id == combat.ActionAttack {
		logrus.Info("Entered TakeAction 'attack'")
		attack := CEAttackCreate(c.CombatState, c.Actor, targets, AttackOptions{})
		tp := attack.TimePoints(c.CombatState.EventQueue)
		c.CombatState.EventQueue.Add(attack, tp)
		return
	}
}

func (c *CombatChoiceState) SetArrowPosition() {
	x, y := c.textbox.Position.X, c.textbox.Position.Y
	width, height := c.textbox.Width, c.textbox.Height

	arrowPad := 9.0
	arrowX := x + width - arrowPad
	c.UpArrowPosition = pixel.V(arrowX, y-arrowPad)
	c.DownArrowPosition = pixel.V(arrowX, y-height+arrowPad)
}
func (c *CombatChoiceState) CreateActionDialog(choices interface{}) {
	selectionMenu := gui.SelectionMenuCreate(20, 0, 0,
		choices,
		false,
		pixel.ZV,
		c.OnSelect,
		nil,
	)
	c.Selection = &selectionMenu

	x := c.Stack.Win.Bounds().W() / 2
	y := c.Stack.Win.Bounds().H() / 2

	height := c.Selection.GetHeight() + 18
	//width := c.Selection.GetWidth() + 16

	y = y - height
	x = x - 90

	c.textbox = gui.TextboxFITPassedMenuCreate(
		c.Stack,
		x, y, "",
		c.Selection,
	)
	c.textbox.Panel.BGColor = utilz.HexToColor("#3c2f2f")
}

func (c *CombatChoiceState) bounceMarker(dt float64) {
	c.time = c.time + dt
	bounce := pixel.V(c.MarkerPosition.X, c.MarkerPosition.Y+math.Sin(c.time*5))
	c.MarkerPosition = bounce
}

func (c *CombatChoiceState) OnItemAction() {
	// 1. Get the filtered item list
	filter := world.Usable
	filteredItems := c.World.FilterItems(filter)
	if len(filteredItems) == 0 {
		return
	}

	// 2. Create the selection box
	itemsSelectionWidth := 120.0
	x := c.Selection.X - (itemsSelectionWidth / 2)
	y := c.Selection.Y - (itemsSelectionWidth / 2)
	c.Selection.HideCursor()

	OnFocus := func(itemI interface{}) {
		item := reflect.ValueOf(itemI).Interface().(world.ItemIndex)
		def := world.ItemsDB[item.Id]
		c.CombatState.ShowTip(def.Description)
	}
	OnExit := func() {
		c.CombatState.HideTip()
		c.Selection.ShowCursor()
	}

	OnRenderItem := func(a ...interface{}) {
		//renderer pixel.Target, x, y float64, item world.ItemIndex
		rendererV := reflect.ValueOf(a[0])
		renderer := rendererV.Interface().(pixel.Target)
		xV := reflect.ValueOf(a[1])
		x := xV.Interface().(float64)
		yV := reflect.ValueOf(a[2])
		y := yV.Interface().(float64)
		itemIdxV := reflect.ValueOf(a[3])
		itemIdx := itemIdxV.Interface().(world.ItemIndex)

		def := world.ItemsDB[itemIdx.Id]
		txt := def.Name
		if itemIdx.Count > 1 {
			txt = fmt.Sprintf("%s x%00d", def.Name, itemIdx.Count)
		}

		pos := pixel.V(x, y)
		textBase := text.New(pos, gui.BasicAtlasAscii)
		fmt.Fprintln(textBase, txt)
		textBase.Draw(renderer, pixel.IM)
	}

	//selection *BrowseListState, index int, itemIdx interface{}
	OnSelection := func(selection *BrowseListState, index int, itemIdxI interface{}) {
		itemIdx := reflect.ValueOf(itemIdxI).Interface().(world.ItemIndex)
		def := world.ItemsDB[itemIdx.Id]
		targeter := c.CreateItemTargeter(def, selection)
		c.Stack.Push(targeter)
	}

	state := BrowseListStateCreate(
		c.Stack, x+24, y+24, itemsSelectionWidth+10, 100, "ITEMS",
		OnFocus,
		OnExit,
		filteredItems,
		OnSelection,
		OnRenderItem,
	)
	c.Stack.Push(state)
}

func (c *CombatChoiceState) OnSpecialAction() {
	actor := c.Actor

	// Create the selection box
	itemsSelectionWidth := 150.0
	x := c.Selection.X - (itemsSelectionWidth / 2)
	y := c.Selection.Y - (itemsSelectionWidth / 2)
	c.Selection.HideCursor()

	OnExit := func() {
		c.CombatState.HideTip()
		c.Selection.ShowCursor()
	}

	OnRenderItem := func(a ...interface{}) {
		//renderer pixel.Target, x, y float64, item string
		renderer := reflect.ValueOf(a[0]).Interface().(pixel.Target)
		x := reflect.ValueOf(a[1]).Interface().(float64)
		y := reflect.ValueOf(a[2]).Interface().(float64)
		elementStr := reflect.ValueOf(a[3]).Interface().(string)

		mpNow := actor.Stats.Get("MpNow")
		color_ := utilz.HexToColor("#bbbbbb")
		def, ok := world.SpecialsDB[elementStr]
		if !ok {
			panic(fmt.Sprintf("Key '%s' not found in SpecialsDB", elementStr))
		}

		txtWithCost := fmt.Sprintf("%s (%v)", def.Name, def.MpCost)

		if mpNow >= def.MpCost {
			color_ = utilz.HexToColor("#ffffff")
		}

		pos := pixel.V(x, y)
		textBase := text.New(pos, gui.BasicAtlasAscii)
		textBase.Color = color_
		fmt.Fprintln(textBase, txtWithCost)
		textBase.Draw(renderer, pixel.IM)
	}

	OnSelection := func(selection *BrowseListState, index int, spellStringI interface{}) {
		spellString := reflect.ValueOf(spellStringI).Interface().(string)
		def, ok := world.SpecialsDB[spellString]
		if !ok {
			panic(fmt.Sprintf("Key '%s' not found in SpecialsDB", spellString))
		}

		mpNow := actor.Stats.Get("MpNow")
		if mpNow < def.MpCost {
			return //not enough mp
		}

		var combatEventFunc func(scene *CombatState, owner *combat.Actor, targets []*combat.Actor, spellI interface{}) CombatEvent
		if def.Action == world.ElementSlash {
			combatEventFunc = CESlashCreate
		} else if def.Action == world.ElementSteal {
			combatEventFunc = CEStealCreate
		}

		targeter := c.CreateActionTargeter(def, selection, combatEventFunc)
		c.Stack.Push(targeter)
	}

	specialItemsState := BrowseListStateCreate(
		c.Stack, x+24, y+24, itemsSelectionWidth, 100, "SPECIAL",
		func(item interface{}) {
			//onFocus do nothing
		},
		OnExit,
		actor.Special,
		OnSelection,
		OnRenderItem,
	)
	c.Stack.Push(specialItemsState)
}

func (c *CombatChoiceState) OnMagicAction() {
	actor := c.Actor

	// Create the selection box
	itemsSelectionWidth := 150.0
	x := c.Selection.X - (itemsSelectionWidth / 2)
	y := c.Selection.Y - (itemsSelectionWidth / 2)
	c.Selection.HideCursor()

	OnExit := func() {
		c.CombatState.HideTip()
		c.Selection.ShowCursor()
	}

	OnRenderItem := func(a ...interface{}) {
		//renderer pixel.Target, x, y float64, item string
		renderer := reflect.ValueOf(a[0]).Interface().(pixel.Target)
		x := reflect.ValueOf(a[1]).Interface().(float64)
		y := reflect.ValueOf(a[2]).Interface().(float64)
		elementStr := reflect.ValueOf(a[3]).Interface().(string)

		mpNow := actor.Stats.Get("MpNow")
		color_ := utilz.HexToColor("#bbbbbb")
		def, ok := world.SpellsDB[elementStr]
		if !ok {
			panic(fmt.Sprintf("Key '%s' not found in SpellsDB", elementStr))
		}

		txtWithCost := fmt.Sprintf("%s (%v)", def.Name, def.MpCost)

		if mpNow >= def.MpCost {
			color_ = utilz.HexToColor("#ffffff")
		}

		pos := pixel.V(x, y)
		textBase := text.New(pos, gui.BasicAtlasAscii)
		textBase.Color = color_
		fmt.Fprintln(textBase, txtWithCost)
		textBase.Draw(renderer, pixel.IM)
	}

	OnSelection := func(selection *BrowseListState, index int, spellStringI interface{}) {
		spellString := reflect.ValueOf(spellStringI).Interface().(string)
		def, ok := world.SpellsDB[spellString]
		if !ok {
			panic(fmt.Sprintf("Key '%s' not found in SpellsDB", spellString))
		}

		mpNow := actor.Stats.Get("MpNow")
		if mpNow < def.MpCost {
			return //not enough mp
		}

		targeter := c.CreateActionTargeter(def, selection, CECastSpellCreate)
		c.Stack.Push(targeter)
	}

	magicItemsState := BrowseListStateCreate(
		c.Stack, x+24, y+24, itemsSelectionWidth, 100, "MAGIC",
		func(item interface{}) {
			//onFocus do nothing
		},
		OnExit,
		actor.Magic,
		OnSelection,
		OnRenderItem,
	)
	c.Stack.Push(magicItemsState)
}

func (c *CombatChoiceState) CreateItemTargeter(def world.Item, browseState *BrowseListState) *CombatTargetState {
	targetDef := def.Use.Target
	c.CombatState.ShowTip(def.Use.Hint)
	browseState.Hide()
	c.Hide()

	OnSelect := func(targets []*combat.Actor) {
		c.Stack.Pop() // target state
		c.Stack.Pop() // item box state
		c.Stack.Pop() // action state

		queue := c.CombatState.EventQueue
		event := CEUseItemCreate(c.CombatState, c.Actor, def, targets)
		tp := event.TimePoints(queue)
		queue.Add(event, tp)
	}

	OnExit := func() {
		browseState.Show()
		c.Show()
	}

	combatFunc, ok := CombatSelectorMap[targetDef.Selector]
	if !ok {
		panic(fmt.Sprintln("Please declare CombatSelectorFunc", targetDef.Selector))
	}

	return CombatTargetStateCreate(c.CombatState, CombatChoiceParams{
		OnSelect:        OnSelect,
		OnExit:          OnExit,
		SwitchSides:     targetDef.SwitchSides,
		DefaultSelector: combatFunc,
		TargetType:      targetDef.Type,
	})
}

func (c *CombatChoiceState) Hide() {
	c.mHide = true
}
func (c *CombatChoiceState) Show() {
	c.mHide = false
}

func (c *CombatChoiceState) CreateActionTargeter(def world.SpecialItem, browseState *BrowseListState, combatEventF func(scene *CombatState, owner *combat.Actor, targets []*combat.Actor, spellI interface{}) CombatEvent) *CombatTargetState {
	targetDef := def.Target
	browseState.Hide()
	c.Hide()

	OnSelect := func(targets []*combat.Actor) {
		c.Stack.Pop() // target state
		c.Stack.Pop() // spell browse state
		c.Stack.Pop() // action state

		queue := c.CombatState.EventQueue
		event := combatEventF(c.CombatState, c.Actor, targets, def)
		tp := event.TimePoints(queue)
		queue.Add(event, tp)
	}

	OnExit := func() {
		browseState.Show()
		c.Show()
	}

	return CombatTargetStateCreate(c.CombatState, CombatChoiceParams{
		OnSelect:        OnSelect,
		OnExit:          OnExit,
		SwitchSides:     targetDef.SwitchSides,
		DefaultSelector: CombatSelectorMap[targetDef.Selector],
		TargetType:      targetDef.Type,
	})
}
