package world

import (
	"fmt"
	"math"
	"reflect"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/text"
	log "github.com/sirupsen/logrus"
	"golang.org/x/image/font/basicfont"
)

type World struct {
	Time, Gold      float64
	Items, KeyItems []ItemIndex
	//Party check world_extended.go
	Icons Icons
}

type ItemIndex struct {
	Id, Count int
}

func Create() *World {
	w := &World{
		Time:     0,
		Gold:     0,
		Items:    make([]ItemIndex, 0),
		KeyItems: make([]ItemIndex, 0),
		Icons:    IconsDB,
	}

	//temp user items in inventory
	//w.Items = append(w.Items, ItemIndex{Id: 1, Count: 2})
	//w.Items = append(w.Items, ItemIndex{Id: 2, Count: 1})
	//w.Items = append(w.Items, ItemIndex{Id: 3, Count: 1})
	//w.KeyItems = append(w.KeyItems, ItemIndex{Id: 4, Count: 1})

	return w
}

func (w *World) AddItem(itemId, count int) {
	if _, ok := ItemsDB[itemId]; !ok {
		log.Fatal(fmt.Sprintf("Item ID {%v} does not exists in DB", itemId))
	}

	for i := range w.Items {
		//Does it already exist in World
		if w.Items[i].Id == itemId {
			w.Items[i].Count += count
			return
		}
	}

	//Add new
	w.Items = append(w.Items, ItemIndex{
		Id:    itemId,
		Count: count,
	})
}

func (w *World) RemoveItem(itemId, count int) {
	if _, ok := ItemsDB[itemId]; !ok {
		log.Fatal(fmt.Sprintf("Item ID {%v} does not exists in DB", itemId))
	}

	for i := len(w.Items) - 1; i >= 0; i-- {
		//Does it already exist in World
		if w.Items[i].Id == itemId {
			w.Items[i].Count -= count
		}

		if w.Items[i].Count <= 0 {
			w.removeItemFromArray(i)
			return
		}
	}
}

func (w *World) removeItemFromArray(index int) {
	if len(w.Items) == 1 {
		w.Items = make([]ItemIndex, 0)
		return
	}
	w.Items[index], w.Items[0] = w.Items[0], w.Items[index]
	w.Items = w.Items[1:]
}

func (w World) hasKeyItem(itemId int) bool {
	for _, v := range w.KeyItems {
		if v.Id == itemId {
			return true
		}
	}
	return false
}

func (w *World) AddKeyItem(itemId int) {
	if w.hasKeyItem(itemId) {
		//if already exists we dont add again
		return
	}

	w.KeyItems = append(w.KeyItems, ItemIndex{Id: itemId, Count: 1})
}
func (w *World) RemoveKeyItem(itemId int) {
	if !w.hasKeyItem(itemId) {
		return
	}

	w.removeKeyItemFromArray(itemId)
}
func (w *World) removeKeyItemFromArray(index int) {
	if len(w.KeyItems) == 1 {
		w.Items = make([]ItemIndex, 0)
		return
	}
	w.KeyItems[index], w.KeyItems[0] = w.KeyItems[0], w.KeyItems[index]
	w.KeyItems = w.KeyItems[1 : len(w.KeyItems)-1]
}

func (w *World) Update(dt float64) {
	w.Time = w.Time + dt
}

func (w World) TimeAsString() string {
	time := w.Time
	hours := math.Floor(time / 3600)
	minutes := math.Ceil(math.Mod(time, 3600)/60) - 1
	seconds := int(time) % 60
	return fmt.Sprintf("%v:%v:%v", hours, minutes, seconds)
}

func (w World) GoldAsString() string {
	return fmt.Sprintf("%v", w.Gold)
}

func (w World) GetItemsAsStrings() []string {
	var items []string
	for _, item := range w.Items {
		items = append(items, fmt.Sprintf("%s, (%v)", ItemsDB[item.Id].Name, item.Count))
	}
	return items
}

func (w World) GetKeyItemsAsStrings() []string {
	var items []string
	for _, item := range w.KeyItems {
		items = append(items, fmt.Sprintf("%s, (%v)", ItemsDB[item.Id].Name, item.Count))
	}
	return items
}

func (w World) DrawItem(a ...interface{}) {
	//renderer pixel.Target, x, y float64, itemIdx ItemIndex
	rendererV := reflect.ValueOf(a[0])
	renderer := rendererV.Interface().(pixel.Target)
	xV := reflect.ValueOf(a[1])
	x := xV.Interface().(float64)
	yV := reflect.ValueOf(a[2])
	y := yV.Interface().(float64)
	itemIdxV := reflect.ValueOf(a[3])
	itemIdx := itemIdxV.Interface().(ItemIndex)

	itemDef := ItemsDB[itemIdx.Id]
	iconsSize := 16.0

	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	textPos := pixel.V(x+iconsSize, y)
	textBase := text.New(textPos, basicAtlas)
	fmt.Fprintln(textBase, fmt.Sprintf("%-6s (%v)", itemDef.Name, itemIdx.Count))
	textBase.Draw(renderer, pixel.IM)

	iconPos := pixel.V(x+5, y+(iconsSize/2))
	iconSprite := w.Icons.Get(itemDef.Icon)
	iconSprite.Draw(renderer, pixel.IM.Moved(iconPos))
}

func (w *World) HasKey(id int) bool {
	for _, v := range w.KeyItems {
		if v.Id == id {
			return true
		}
	}
	return false
}
func (w *World) Get(idx ItemIndex) Item {
	return ItemsDB[idx.Id]
}

func (w *World) FilterItems(predicate ItemType) []ItemIndex {
	list := make([]ItemIndex, 0)
	for _, v := range w.Items {
		item := ItemsDB[v.Id]
		if item.ItemType == predicate {
			list = append(list, v)
		}
	}

	return list
}
