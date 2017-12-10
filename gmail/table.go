package main

import (
	"fmt"
	"log"
	"sort"

	"github.com/jroimartin/gocui"
)

var m *messages

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

func onClick(g *gocui.Gui, v *gocui.View) error {
	// if loading {
	// 	return nil
	// }
	if v != nil {
		loading = true
		table, _ := g.SetCurrentView("table")
		ids, _ := m.getNext()
		populateTable(table, getMessages(ids))
	}
	return nil
}

func onClickPrev(g *gocui.Gui, v *gocui.View) error {
	// if loading {
	// 	return nil
	// }
	if v != nil {
		loading = true
		table, _ := g.SetCurrentView("table")
		ids, _ := m.getPrev()
		populateTable(table, getMessages(ids))
		subjectsList = make([]message, 0)
	}
	return nil
}

func populateTable(table *gocui.View, messagesList []message) {
	table.Clear()
	sort.Sort(byDate(messagesList))
	for _, message := range messagesList {
		fmt.Fprintf(table, "> %v\n", message.subject)
	}
}

var subjectsList []message
var ids []string
var loading bool

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("nextBtn", gocui.MouseLeft, gocui.ModNone, onClick); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("prevBtn", gocui.MouseLeft, gocui.ModNone, onClickPrev); err != nil {
		log.Panicln(err)
	}

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
	g.Mouse = true

	if err != nil {
		log.Panicln(err)
	}

	m, _ = newMessages()

	// Initial messages
	ids, _ := m.getNext()
	subjectsList = make([]message, 0)
	if len(subjectsList) < len(ids) {
		ch := make(chan message)
		getSubjects(ids, ch)
		for subject := range ch {
			subjectsList = append(subjectsList, subject)
			// populateTable(table, subjectsList)

			if len(ids) == len(subjectsList) {
				close(ch)
				loading = false
				// fmt.Fprintf(table, "LOADING OFF %v\n", loading)
			}
		}
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

	fmt.Println("bye!")
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if v, err := g.SetView("nextBtn", 0, 0, 10, 4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		fmt.Fprintln(v, "NEXT")
	}

	if v, err := g.SetView("prevBtn", 12, 0, maxX, 4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		fmt.Fprintln(v, "PREV")
	}

	if v, err := g.SetView("table", 0, 5, maxX, maxY); err != nil {
		populateTable(v, subjectsList)
		fmt.Fprintf(v, "LOADED")

		v.Frame = true

		if err := v.SetOrigin(0, 0); err != nil {
			return err
		}
		v.Wrap = true
	}

	g.SetCurrentView("table")

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
