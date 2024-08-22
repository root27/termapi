package main

import (
	"bytes"
	"fmt"
	"github.com/jroimartin/gocui"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func saveResponse(g *gocui.Gui, v *gocui.View) error {
	responseView, err := g.View("response")
	if err != nil {
		return err
	}

	saveFileView, _ := g.View("saveModal")

	saveFileName := saveFileView.Buffer()

	responseBody := responseView.Buffer()

	os.WriteFile(saveFileName, []byte(responseBody), 0644)

	return nil
}

func nextView(g *gocui.Gui, v *gocui.View) error {
	if v.Name() == "url" {
		_, err := g.SetCurrentView("method")

		return err
	}

	if v.Name() == "method" {
		_, err := g.SetCurrentView("headers")

		return err
	}

	if v.Name() == "headers" {
		_, err := g.SetCurrentView("body")

		return err
	}
	if v.Name() == "body" {
		_, err := g.SetCurrentView("url")
		return err
	}
	return nil
}

var (
	url     string
	headers string
	body    string
	method  string
)

func handleRequest(g *gocui.Gui, v *gocui.View) error {
	urlView, err := g.View("url")
	if err != nil {
		return err
	}
	headersView, err := g.View("headers")
	if err != nil {
		return err
	}
	bodyView, err := g.View("body")
	if err != nil {
		return err
	}

	methodView, err := g.View("method")

	if err != nil {
		return err
	}

	responseView, err := g.View("response")
	if err != nil {
		return err
	}
	url = strings.TrimSpace(urlView.Buffer())
	headers = strings.TrimSpace(headersView.Buffer())
	body = strings.TrimSpace(bodyView.Buffer())
	method = methodView.Buffer()

	client := &http.Client{}

	switch strings.TrimSpace(method) {
	case "POST":

		req, requesterr := http.NewRequest("POST", url, bytes.NewReader([]byte(body)))

		if requesterr != nil {
			fmt.Fprintf(responseView, "Error: %s\n", err)
		}

		response, responseerr := client.Do(req)

		if responseerr != nil {
			fmt.Fprintf(responseView, "Error: %s\n", err)
		}

		resStatus := response.Status

		resbody, bodyerr := io.ReadAll(response.Body)

		if bodyerr != nil {
			fmt.Fprintf(responseView, "Error: %s\n", err)

		}

		fmt.Fprintf(responseView, "Status: %s\n\nResponse Body: %s\n", resStatus, resbody)

	case "GET":

		req, err := http.NewRequest("GET", url, nil)

		if err != nil {
			fmt.Fprintf(responseView, "Error: %s\n", err)
		}

		response, err := client.Do(req)

		if err != nil {
			fmt.Fprintf(responseView, "Error: %s\n", err)
		}

		defer response.Body.Close()
		resStatus := response.Status

		resbody, err := io.ReadAll(response.Body)

		if err != nil {
			fmt.Fprintf(responseView, "Error: %s\n", err)
		}

		fmt.Fprintf(responseView, "Status: %s\n\nResponse Body: %s\n", resStatus, resbody)

	default:
		fmt.Fprintf(responseView, "Method not supported: %s\n", method)
	}

	return nil

}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	g.Highlight = true
	g.SelFgColor = gocui.ColorGreen

	if v, err := g.SetView("url", 0, 0, maxX-150, 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "API Endpoint"
		v.Editable = true

		if _, err := g.SetCurrentView("url"); err != nil {
			return err
		}

	}

	if v, err := g.SetView("method", 0, 3, maxX-150, 5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Method"
		v.Editable = true

	}

	if v, err := g.SetView("headers", 0, 6, maxX-150, 18); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Headers (Key:Value)"
		v.Editable = true
		v.Autoscroll = true
	}

	if v, err := g.SetView("body", 0, 19, maxX-150, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Request Body"
		v.Editable = true
		fmt.Fprintln(v, "{}")
	}

	if v, err := g.SetView("response", maxX-149, 0, maxX-1, 28); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Response"
		v.Wrap = true
	}

	// Help view - always visible at the bottom
	if v, err := g.SetView("help", 0, maxY-4, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Help"
		v.Wrap = true
		v.Frame = true
		v.Autoscroll = true
		fmt.Fprintln(v, "Navigation: Tab: Next View | Send Request: Ctrl+S | Send New Request: Ctrl+R | Quit: Ctrl+C")
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func cleanFields(g *gocui.Gui, v *gocui.View) error {

	for _, view := range g.Views() {

		if view.Name() != "help" {

			view.Clear()
			_ = view.SetCursor(0, 0)
		}

	}

	if _, err := g.SetCurrentView("url"); err != nil {
		return err
	}

	return nil

}

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}

	defer g.Close()

	g.Cursor = true

	g.SetManagerFunc(layout)

	// Keybinding to quit the application
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlS, gocui.ModNone, handleRequest); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlR, gocui.ModNone, cleanFields); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
