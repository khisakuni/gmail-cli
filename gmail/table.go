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

func onClick(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		table, _ := g.SetCurrentView("table")
		table.Clear()
		ids, _ := getNext()
		subjectsList = make([]string, 0)
		fmt.Fprintf(table, "next page token: %v\n", nextPageToken)
		if len(subjectsList) < len(ids) {
			ch := make(chan string)
			getSubjects(ids, ch)
			for subject := range ch {
				fmt.Fprintf(table, "> next: %v\n", subject)
				subjectsList = append(subjectsList, subject)

				if len(ids) == len(subjectsList) {
					close(ch)
				}
			}
		}
	}
	return nil
}

func onClickPrev(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		table, _ := g.SetCurrentView("table")
		table.Clear()
		tokens := &prevPageTokens
		fmt.Fprintf(table, "prev page token: %v\n", *tokens)
		ids, _ := getPrev()
		subjectsList = make([]string, 0)
		if len(subjectsList) < len(ids) {
			ch := make(chan string)
			getSubjects(ids, ch)
			for subject := range ch {
				fmt.Fprintf(table, "> next: %v\n", subject)
				subjectsList = append(subjectsList, subject)

				if len(ids) == len(subjectsList) {
					close(ch)
				}
			}
		}
	}
	return nil
}

var subjectsList []string = make([]string, 0)
var ids []string = make([]string, 0)

func main() {

	// GETTING SUBJECTS
	// query := []googleapi.CallOption{option{key: "maxResults", value: "20"}}
	list, err := getNext()

	ids = list

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
		if len(subjectsList) < len(ids) {
			ch := make(chan string)
			getSubjects(ids, ch)
			for subject := range ch {
				fmt.Fprintf(v, "> next: %v\n", subject)
				subjectsList = append(subjectsList, subject)

				if len(ids) == len(subjectsList) {
					close(ch)
				}
			}
		}

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
