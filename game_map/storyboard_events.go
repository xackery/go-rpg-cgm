package game_map

import (
	"image/color"
	"reflect"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/sirupsen/logrus"
	"github.com/steelx/go-rpg-cgm/gui"
	"github.com/steelx/go-rpg-cgm/sound"
	"github.com/steelx/go-rpg-cgm/state_machine"
)

var queue sound.Queue
var queueBG sound.Queue
var sr beep.SampleRate

func init() {
	sr = beep.SampleRate(44100)
	err := speaker.Init(sr, sr.N(time.Second/10))
	logFatalErr(err)
	speaker.Play(&queue)
	speaker.Play(&queueBG)
}

func Wait(seconds float64) *WaitEvent {
	return WaitEventCreate(seconds)
}

// BlackScreen - end to call KillState("blackscreen") once done
func BlackScreen(id string) func(storyboard *Storyboard) *WaitEvent {
	return func(storyboard *Storyboard) *WaitEvent {
		screen := ScreenStateCreate(storyboard.Stack, color.RGBA{R: 255, G: 255, B: 255, A: 1})
		storyboard.PushState(id, screen)
		return WaitEventCreate(0)
	}
}

func TitleCaptionScreen(id string, txt string, duration float64) func(storyboard *Storyboard) *TweenEvent {
	return func(storyboard *Storyboard) *TweenEvent {
		captions := gui.CaptionScreenCreate(txt, pixel.V(0, 100), 3)
		storyboard.PushState(id, &captions)

		return TweenEventCreate(
			1, 0, duration,
			&captions,
			func(e *TweenEvent) {
				captions.Update(e.Tween.Value())
			},
		)
	}
}

func SubTitleCaptionScreen(id string, txt string, duration float64) func(storyboard *Storyboard) *TweenEvent {

	return func(storyboard *Storyboard) *TweenEvent {
		captions := gui.CaptionScreenCreate(txt, pixel.V(0, 50), 1)
		storyboard.PushState(id, &captions)

		return TweenEventCreate(
			1, 0, duration,
			&captions,
			func(e *TweenEvent) {
				captions.Update(e.Tween.Value())
			},
		)
	}
}

func Scene(mapName string, hideHero bool, win *pixelgl.Window) func(storyboard *Storyboard) *NonBlockEvent {

	return func(storyboard *Storyboard) *NonBlockEvent {
		if win == nil {
			win = storyboard.Stack.Win
		}
		mapInfo := MapsDB[mapName](storyboard.Stack)
		exploreState := ExploreStateCreate(storyboard.Stack, mapInfo, win)
		if hideHero {
			exploreState.HideHero()
		}

		storyboard.PushState(mapName, &exploreState)

		return NonBlockEventCreate(0)
	}
}

// player_house, def = "sleeper", x = 14, y = 19
func RunActionAddNPC(mapName, entityDef string, x, y, seconds float64) func(storyboard *Storyboard) *WaitEvent {
	return func(storyboard *Storyboard) *WaitEvent {
		exploreState := getExploreState(storyboard, mapName)
		char := Characters[entityDef](exploreState.Map)
		exploreState.SetFollowCam(true, char)

		runFunc := AddNPC(exploreState.Map, x, y)
		runFunc(char)
		return WaitEventCreate(seconds)
	}
}

func getExploreState(storyboard *Storyboard, stateId string) *ExploreState {
	exploreStateI := storyboard.States[stateId]
	exploreStateV := reflect.ValueOf(exploreStateI)
	exploreState := exploreStateV.Interface().(*ExploreState)
	return exploreState
}

func KillState(id string) func(storyboard *Storyboard) *WaitEvent {
	return func(storyboard *Storyboard) *WaitEvent {
		storyboard.RemoveState(id)
		return WaitEventCreate(0)
	}
}

func MoveNPC(npcId, mapName string, path []string) func(storyboard *Storyboard) *BlockUntilEvent {

	return func(storyboard *Storyboard) *BlockUntilEvent {
		exploreState := getExploreState(storyboard, mapName)
		npc := exploreState.Map.NPCbyId[npcId]
		npc.FollowPath(path)

		return BlockUntilEventCreate(func() bool {
			return npc.PathIndex >= len(path)
		})
	}
}

func Say(mapName, npcId, textMessage string, time float64) func(storyboard *Storyboard) *TimedTextboxEvent {
	return func(storyboard *Storyboard) *TimedTextboxEvent {
		exploreState := getExploreState(storyboard, mapName)
		npc := exploreState.Map.NPCbyId[npcId]
		tileX, tileY := npc.GetFacedTileCoords()
		posX, posY := exploreState.Map.GetTileIndex(tileX, tileY)
		tBox := storyboard.InternalStack.PushFitted(posX, posY+32, textMessage)
		return TimedTextboxEventCreate(tBox, time)
	}
}

// ReplaceScene will remove mapName and add newMapName with a Hero at given Tile X, Y
func ReplaceScene(mapName string, newMapName string, tileX, tileY float64, hideHero bool, win *pixelgl.Window) func(storyboard *Storyboard) *NonBlockEvent {
	return func(storyboard *Storyboard) *NonBlockEvent {
		if win == nil {
			win = storyboard.Stack.Win
		}
		storyboard.RemoveState(mapName) //remove previous map (exploreState)

		mapInfo := MapsDB[newMapName](storyboard.Stack)
		newExploreState := ExploreStateCreate(storyboard.Stack, mapInfo, win)

		if hideHero {
			newExploreState.HideHero()
		} else {
			newExploreState.ShowHero(tileX, tileY)
		}

		storyboard.PushState(newMapName, &newExploreState) //ADD new map (exploreState)

		return NonBlockEventCreate(0)
	}
}

// HandOffToMainStack will remove the exploreState from Storyboard and push it to main stack
func HandOffToMainStack(mapName string) func(storyboard *Storyboard) *WaitEvent {
	return func(storyboard *Storyboard) *WaitEvent {
		exploreState := getExploreState(storyboard, mapName)
		storyboard.Stack.Pop()
		exploreState.Stack = storyboard.Stack

		storyboard.Stack.Push(exploreState)

		return WaitEventCreate(1)
	}
}

var fragmentShader = `
#version 330 core

in vec2  vTexCoords;

out vec4 fragColor;

uniform vec4 uTexBounds;
uniform sampler2D uTexture;

void main() {
	// Get our current screen coordinate
	vec2 t = (vTexCoords - uTexBounds.xy) / uTexBounds.zw;

	// Sum our 3 color channels
	float sum  = texture(uTexture, t).r;
	      sum += texture(uTexture, t).g;
	      sum += texture(uTexture, t).b;

	// Divide by 3, and set the output to the result
	vec4 color = vec4( sum/3, sum/3, sum/3, 1.0);
	fragColor = color;
}
`

func FadeOutMap(mapName string, duration float64) func(storyboard *Storyboard) *TweenEvent {

	return func(storyboard *Storyboard) *TweenEvent {
		exploreState := getExploreState(storyboard, mapName)

		return TweenEventCreate(
			1, 0, duration,
			exploreState,
			func(e *TweenEvent) {
				exploreState.Map.Canvas.SetFragmentShader(fragmentShader)
			},
		)
	}
}

func FadeOutCharacter(mapName, npcId string, duration float64) func(storyboard *Storyboard) *TweenEvent {
	//pic, _ := utilz.LoadPicture("../resources/universal-lpc-sprite_male_01_walk-3frame.png")
	//frames := utilz.LoadAsFrames(pic, 32, 32)
	return func(storyboard *Storyboard) *TweenEvent {
		exploreState := getExploreState(storyboard, mapName)
		var npc *Character
		if npcId == "hero" {
			exploreState.SetFollowCam(false, exploreState.Hero)
			exploreState.SetManualCam(20, 20)
			exploreState.HideHero()
			npc = exploreState.Hero
		} else {
			npc = exploreState.Map.NPCbyId[npcId]
		}

		return TweenEventCreate(
			1, 0, duration,
			exploreState,
			func(e *TweenEvent) {
				//npc.Entity.Sprite.Set(pic, frames[1])
				npc.Entity.SetTilePos(0, 0)
				npc.Entity.TeleportAndDraw(exploreState.Map, exploreState.Map.Canvas)
			},
		)
	}
}

func WriteTile(mapName string, tileX, tileY float64, collision bool) func(storyboard *Storyboard) *WaitEvent {
	return func(storyboard *Storyboard) *WaitEvent {
		exploreState := getExploreState(storyboard, mapName)
		exploreState.Map.WriteTile(tileX, tileY, collision)

		return WaitEventCreate(0)
	}
}
func SetHiddenTileVisible(mapName string, tileX, tileY float64) func(storyboard *Storyboard) *WaitEvent {
	return func(storyboard *Storyboard) *WaitEvent {
		exploreState := getExploreState(storyboard, mapName)
		exploreState.Map.SetHiddenTileVisible(int(tileX), int(tileY))

		return WaitEventCreate(0)
	}
}

// MoveCamToTile not working as intended. pending
func MoveCamToTile(stateId string, fromTileX, fromTileY, tileX, tileY, duration float64) func(storyboard *Storyboard) *TweenEvent {

	return func(storyboard *Storyboard) *TweenEvent {
		exploreState := getExploreState(storyboard, stateId)
		exploreState.SetFollowCam(false, exploreState.Hero)

		exploreState.ManualCamX = fromTileX
		exploreState.ManualCamY = fromTileY
		startX := exploreState.ManualCamX
		startY := exploreState.ManualCamY
		endX, endY := tileX, tileY
		xDistance := endX - startX
		yDistance := endY - startY

		return TweenEventCreate(
			0, 1, duration,
			exploreState,
			func(e *TweenEvent) {
				dX := startX + (xDistance * e.Tween.Value())
				dY := startY + (yDistance * e.Tween.Value())
				exploreState.ManualCamX = dX
				exploreState.ManualCamY = dY
			},
		)
	}
}

// PlaySound will stop after the given duration
func PlaySound(path string, duration float64) func(storyboard *Storyboard) *NonBlockingTimer {
	path = strings.TrimPrefix(path, "../sound/")
	f, err := sound.FS.Open(path)
	logFatalErr(err)

	// Decode it.
	streamer, format, err := mp3.Decode(f)
	logFatalErr(err)

	return func(storyboard *Storyboard) *NonBlockingTimer {
		logrus.Info("Playing sound: ", path)

		// The speaker's sample rate is fixed at 44100. Therefore, we need to
		// resample the file in case it's in a different sample rate.
		resampled := beep.Resample(3, format.SampleRate, sr, streamer)

		// And finally, we add the song to the queue.
		speaker.Lock()
		queue.Add(resampled)
		speaker.Unlock()

		return NonBlockingTimerCreate(
			duration,
			func(e *NonBlockingTimer) {
				if e.TimeUp() {
					queue.Pop()
					logrus.Info("Removing sound: ", path)
				}
			},
		)
	}
}

// PlayBGSound will stop after track has finished
func PlayBGSound(path string) func() {
	path = strings.TrimPrefix(path, "../sound/")
	f, err := sound.FS.Open(path)
	logFatalErr(err)

	// Decode it.
	streamer, format, err := mp3.Decode(f)
	logFatalErr(err)

	//load audio into memory
	//buffer := beep.NewBuffer(format)
	//buffer.Append(streamer)
	//streamer.Close()
	//f.Close()

	return func() {
		logrus.Info("Playing BG sound: ", path)
		//bufferedSound := buffer.Streamer(0, buffer.Len())

		// The speaker's sample rate is fixed at 44100. Therefore, we need to
		// resample the file in case it's in a different sample rate.
		resampled := beep.Resample(3, format.SampleRate, sr, streamer)

		// And finally, we add the song to the queue.
		speaker.Lock()
		queueBG.Add(resampled)
		speaker.Unlock()

	}
}

// StopBGSound will pop out last queueBG item
func StopBGSound() func() {
	return func() {
		queueBG.Pop()
	}
}

func RunState(stateMachine *state_machine.StateMachine, stateId string, params ...interface{}) func(storyboard *Storyboard) *BlockUntilEvent {

	return func(storyboard *Storyboard) *BlockUntilEvent {
		stateMachine.Change(stateId, params...)
		return BlockUntilEventCreate(func() bool {
			return stateMachine.Current.IsFinished()
		})
	}
}

func RunFunction(fn func()) func(storyboard *Storyboard) *WaitEvent {

	return func(storyboard *Storyboard) *WaitEvent {
		fn()
		return Wait(0)
	}
}

func UpdateState(state gui.StackInterface, time float64) func(storyboard *Storyboard) *TweenEvent {

	return func(storyboard *Storyboard) *TweenEvent {

		return TweenEventCreate(
			0, 1, time,
			state,
			func(e *TweenEvent) {
				state.Update(storyboard.Stack.DeltaTime)
			},
		)
	}
}

func ReplaceState(current, new gui.StackInterface) func(storyboard *Storyboard) {

	return func(storyboard *Storyboard) {
		logrus.Info("Being asked to replace a state")
		stack := storyboard.Stack

		for k, v := range stack.States {
			if v == current {
				stack.States[k].Exit()
				stack.States[k] = new
				stack.States[k].Enter()
				return
			}
		}

		panic("You should have found a Current state.")
	}
}

func RemoveState(state gui.StackInterface) func(storyboard *Storyboard) {
	return func(storyboard *Storyboard) {
		logrus.Info("Being asked to Remove a state")
		stack := storyboard.Stack

		for i := len(stack.States) - 1; i >= 0; i-- {
			v := stack.States[i]
			if v == state {
				v.Exit()
				stack.RemoveStateAtIndex(i)
				return
			}
		}

	}
}
