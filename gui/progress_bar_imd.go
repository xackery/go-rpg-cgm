package gui

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/steelx/go-rpg-cgm/utilz"
)

type ProgressBarIMD struct {
	x, y       float64
	Background color.RGBA
	Foreground color.RGBA
	foregroundPosition,
	backgroundPosition pixel.Vec
	Value, Maximum, HalfWidth,
	Height, Width float64
	parentImd *imdraw.IMDraw
}

func ProgressBarIMDCreate(x, y float64, value, max float64, background, foreground string, height, width float64, parentImd *imdraw.IMDraw) ProgressBarIMD {

	pb := ProgressBarIMD{
		x:          x,
		y:          y,
		Background: utilz.HexToColor(background),
		Foreground: utilz.HexToColor(foreground),
		Value:      value,
		Maximum:    max,
		Width:      width,
		Height:     height,
		parentImd:  parentImd,
	}

	pb.HalfWidth = pb.Width / 2
	pb.SetValue(value)

	return pb
}

func (pb *ProgressBarIMD) SetMax(maxHealth float64) {
	pb.Maximum = maxHealth
}
func (pb *ProgressBarIMD) SetValue(health float64) {
	pb.Value = health
}

func (pb *ProgressBarIMD) SetPosition(x, y float64) {
	pb.x = x
	pb.y = y
}
func (pb ProgressBarIMD) GetPosition() (x, y float64) {
	return pb.x, pb.y
}

func (pb ProgressBarIMD) GetPercentWidth() float64 {
	percent := (pb.Value / pb.Maximum) * 100
	return (pb.Width * percent) / 100
}

func (pb ProgressBarIMD) Render(renderer pixel.Target) {
	imd_ := imd
	if pb.parentImd != nil {
		imd_ = pb.parentImd
	}

	imd_.Clear()

	leftPos := pixel.V(pb.x, pb.y)
	imd_.Color = pb.Background
	imd_.Push(leftPos, leftPos.Add(pixel.V(pb.Width, pb.Height)))
	imd_.Rectangle(0)

	imd_.Color = pb.Foreground
	imd_.Push(leftPos, leftPos.Add(pixel.V(pb.GetPercentWidth(), pb.Height)))
	imd_.EndShape = imdraw.RoundEndShape
	imd_.Rectangle(0)
	imd_.Draw(renderer)
}

/*
TO MATCH StackInterface below
*/
func (pb ProgressBarIMD) HandleInput(win *pixelgl.Window) {
}
func (pb ProgressBarIMD) Enter() {}
func (pb ProgressBarIMD) Exit()  {}
func (pb ProgressBarIMD) Update(dt float64) bool {
	return true
}
