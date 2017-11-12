package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/howeyc/gopass"
	"github.com/jroimartin/gocui"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	var email string
	var password []byte

	fmt.Println("Gmail CLI")
	fmt.Println("---------")

	for {
		fmt.Print("email: ")
		email, _ = reader.ReadString('\n')
		email = strings.Replace(email, "\n", "", -1)

		fmt.Print("password: ")
		password, _ = gopass.GetPasswdMasked()

		if len(email) > 0 && len(password) > 0 {
			fmt.Printf("email is: %v, password is %v\n", email, string(password))
			break
		}
	}

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(setupLayout(email, string(password)))

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

	fmt.Println("bye!")
}

func setupLayout(email, password string) func(g *gocui.Gui) error {
	return func(g *gocui.Gui) error {
		maxX, maxY := g.Size()
		if v, err := g.SetView("hello", 0, 0, maxX, maxY); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			fmt.Fprintln(v, "Hello world!!!!")
			fmt.Fprintf(v, "Email: %v\n", email)
			fmt.Fprintf(v, "PasswordL %v\n", password)
		}

		return nil
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
