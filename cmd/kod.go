package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gdamore/tcell"
)

var (
	screen tcell.Screen
	buf    *Buffer

	defaultStyle tcell.Style

	events chan tcell.Event
)

func HandleEvent(ev tcell.Event) {
	switch ev.(type) {
	case *tcell.EventKey:
		// TODO: Check if normal mode, if so check for
		// "global" keybindings which aren't bound to the buffer
		// and pass on buffer-specific keybindings
		buf.HandleEvent(ev)
	}
}

func InitScreen() {
	var err error

	screen, err = tcell.NewScreen()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err = screen.Init(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	screen.SetStyle(defaultStyle)

	screen.Clear()
}

func main() {
	f, _ := os.OpenFile("out.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	log.SetOutput(f)

	defer f.Close()
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	InitScreen()
	defer screen.Fini()

	// screen event channel
	events = make(chan tcell.Event, 100)
	quit := make(chan bool, 1)

	go func() {
		for {
			if screen != nil {
				// feed events into channel
				events <- screen.PollEvent()
			}
		}
	}()

	buf = NewBuffer()

	// main loop
	for {
		screen.Clear()
		for y, line := range buf.lines {
			for x, char := range line.data {
				// TODO: Highlight line and use line.Style as style
				// TODO: UTF-8, for now only single byte chars(ASCII)
				screen.SetCell(x, y, defaultStyle, rune(char))
			}
		}
		screen.ShowCursor(buf.Cursor.x, buf.Cursor.y)
		screen.Show()
		var event tcell.Event

		select {
		case event = <-events:
		case <-quit:
			screen.Fini()
			log.Println("bye")
			os.Exit(0)
		}

		for event != nil {
			switch e := event.(type) {
			case *tcell.EventKey:
				switch e.Key() {
				case tcell.KeyEscape:
					close(quit)
				}
			case *tcell.EventResize:
				screen.Sync()
			}

			HandleEvent(event)

			// continue handling events
			select {
			case event = <-events:
			default:
				event = nil
			}
		}
	}
}