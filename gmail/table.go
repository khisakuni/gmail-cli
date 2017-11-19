package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"log"
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

var subjectsList []string = make([]string, 0)
var ids []string = make([]string, 0)

func main() {

	// GETTING SUBJECTS
	list, err := getMessageIds()
	ids = list
	if err != nil {
		log.Fatalf("noooo %v", err)
	}

	/*
		ch := make(chan string)
		getSubjects(ids, ch)
		for subject := range ch {
			subjectsList = append(subjectsList, subject)
			if len(ids) == len(subjectsList) {
				close(ch)
			}
		}
	*/

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

	if v, err := g.SetView("table", 0, 0, maxX, maxY); err != nil {

		/*
			for _, e := range items {
				time.Sleep(1000 * time.Millisecond)
				fmt.Printf("SUBJECT>>> %v\n", e)
			}
		*/

		if len(subjectsList) < len(ids) {
			ch := make(chan string)
			getSubjects(ids, ch)
			for subject := range ch {
				fmt.Printf("SUBJECT>>> %v\n", subject)
				subjectsList = append(subjectsList, subject)
				// fmt.Fprintf(v, "%v\n", subject)

				if len(ids) == len(subjectsList) {
					close(ch)
				}
			}
		}

		/*
			for i, e := range subjectsList {
				fmt.Fprintf(v, "%v.) %v\n", i, e)
			}
		*/

		v.Frame = false

		if err := v.SetOrigin(0, 0); err != nil {
			return err
		}
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
