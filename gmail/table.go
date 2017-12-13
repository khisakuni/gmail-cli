package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"

	"golang.org/x/net/html"

	"github.com/jroimartin/gocui"
)

var m *messages

func parseHTML(reader io.Reader) string {
	tokenizer := html.NewTokenizer(reader)
	var buffer bytes.Buffer
	depth := 0
	for {
		token := tokenizer.Next()
		switch token {
		case html.ErrorToken:
			return buffer.String()
		case html.TextToken:
			if depth > 0 {
				text := strings.TrimSpace(string(tokenizer.Text()))
				if len(text) > 0 {
					buffer.WriteString(text + "\n")
				}
			}
		case html.StartTagToken, html.EndTagToken:
			t := tokenizer.Token().Data
			if t == "p" || t == "td" {
				if token == html.StartTagToken {
					depth++
				} else {
					depth--
				}
			}
		}
	}
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		if messageIndex < len(subjectsList)-1 {
			messageIndex++
		}

		messagePane, _ := g.View("message")
		messagePane.Clear()
		body := subjectsList[messageIndex].body
		b := parseHTML(bytes.NewReader(body))
		if len(strings.TrimSpace(b)) > 0 {
			fmt.Fprintln(messagePane, b)
		} else {
			fmt.Fprintln(messagePane, "Content Unavailable")
		}

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
		if messageIndex > 0 {
			messageIndex--
		}
		messagePane, _ := g.View("message")
		messagePane.Clear()
		body := subjectsList[messageIndex].body
		b := parseHTML(bytes.NewReader(body))
		if len(strings.TrimSpace(b)) > 0 {
			fmt.Fprintln(messagePane, b)
		} else {
			fmt.Fprintln(messagePane, "Content Unavailable")
		}

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
	if loading {
		return nil
	}
	if v != nil {
		loading = true
		table, _ := g.SetCurrentView("table")
		ids, _ := m.getNext()
		ms := getMessages(ids)
		populateTable(table, ms)
		loading = false
	}
	return nil
}

func onClickPrev(g *gocui.Gui, v *gocui.View) error {
	if loading {
		return nil
	}
	if v != nil {
		loading = true
		table, _ := g.SetCurrentView("table")
		ids, _ := m.getPrev()
		populateTable(table, getMessages(ids))
		loading = false
	}
	return nil
}

func onClickMessagePane(g *gocui.Gui, v *gocui.View) error {
	v.Clear()
	fmt.Fprintln(v, "HI THERE")
	// g.SetCurrentView("message")

	return nil
}

func populateTable(table *gocui.View, messagesList []message) {
	table.Clear()
	table.SetCursor(0, 0)
	messageIndex = 0
	subjectsList = messagesList
	sort.Sort(byDate(messagesList))
	for _, message := range messagesList {
		fmt.Fprintf(table, "> %v: %v\n", message.sender, message.subject)
	}
}

var subjectsList []message
var ids []string
var loading bool
var messageIndex int

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

	if err := g.SetKeybinding("message", gocui.MouseLeft, gocui.ModNone, onClickMessagePane); err != nil {
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
	subjectsList = getMessages(ids)

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
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		fmt.Fprintln(v, "NEXT")
	}

	if v, err := g.SetView("prevBtn", 12, 0, maxX, 4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		fmt.Fprintln(v, "PREV")
	}

	if v, err := g.SetView("table", 0, 5, maxX-1, maxY/2); err != nil {
		populateTable(v, subjectsList)

		v.Frame = true

		if err := v.SetOrigin(0, 0); err != nil {
			return err
		}
		v.Wrap = true
	}

	if v, err := g.SetView("message", 0, maxY/2+1, maxX-1, maxY-1); err != nil {
		v.Frame = true
		v.Wrap = true
		v.Editable = true
		v.Autoscroll = true
	}

	g.SetCurrentView("table")

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
