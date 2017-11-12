package main

import (
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
)

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("table", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("table", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		log.Panicln(err)
	}

	// err = g.SetKeybinding("table", gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
	// 	fmt.Fprintf(v, "up!")
	// 	return nil
	// })
	// if err != nil {
	// 	log.Panicln(err)
	// }
	// err = g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
	// 	dx := 0
	// 	dy := 1
	// 	x0, y0, x1, y1, err := g.ViewPosition("table")
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if _, err := g.SetView("table", x0+dx, y0+dy, x1+dx, y1+dy); err != nil {
	// 		return err
	// 	}
	// 	return nil
	// })
	g.Cursor = true

	if err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

	fmt.Println("bye!")
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// rows := []string{"one", "two", "three"}

	if v, err := g.SetView("table", 0, 0, maxX, maxY); err != nil {
		subjects, _ := getSubjects()
		for i, e := range subjects {
			fmt.Fprintf(v, "%v.) %v\n", i, e)
		}
		v.Frame = false

		v.SetOrigin(0, 0)
		v.Wrap = true
		// v.Editable = true
		// if editable is false, need to implement own cursor up and down methods
	}

	// delta := 2

	// v, _ := g.SetView(fmt.Sprintf("row-%v", i), 1, delta*(i+1), maxX, (i+1)*delta+delta)
	// v.Autoscroll = true

	g.SetCurrentView("table")

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
