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

const (
	minWidth  = 80
	minHeight = 24
)

var helpText = "Navigation: Tab: Next View | Send Request: Ctrl+S | Send New Request: Ctrl+R | Save Response: Ctrl+E | Quit: Ctrl+C"

var (
	url      string
	headers  string
	body     string
	method   string
	filename string
)

func saveResponse(g *gocui.Gui, v *gocui.View) error {
	responseView, err := g.View("response")
	if err != nil {
		return err
	}

	saveFileView, _ := g.View("saveModal")

	filename = strings.TrimSpace(saveFileView.Buffer())

	responseBody := responseView.Buffer()

	os.WriteFile(filename, []byte(responseBody), 0644)

	_ = g.DeleteView("saveModal")

	_, err = g.SetCurrentView("url")

	if err != nil {

		return err

	}

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

	headersArray := []string{}

	if headers != "" {

		headersArray = strings.Split(headers, "\n")

	}

	//Clear Response View

	responseView.Clear()
	switch strings.TrimSpace(method) {
	case "POST":

		req, requesterr := http.NewRequest("POST", url, bytes.NewReader([]byte(body)))

		if requesterr != nil {
			fmt.Fprintf(responseView, "Error: %s\n", err)

			break
		}

		if body == "" {

			fmt.Fprintf(responseView, "Error: %s\n", "Body is required for this HTTP Method")

			break
		}

		if url == "" {

			fmt.Fprintf(responseView, "Error: %s\n", "Url can not be empty")

			break

		}

		if len(headersArray) > 0 {

			for _, header := range headersArray {

				headerKeyValue := strings.Split(header, ":")
				req.Header.Add(headerKeyValue[0], headerKeyValue[1])
			}
		}

		response, responseerr := client.Do(req)

		if responseerr != nil {
			fmt.Fprintf(responseView, "Error: %s\n", err)
		}

		resStatus := response.Status

		resbody, bodyerr := io.ReadAll(response.Body)

		if bodyerr != nil {
			fmt.Fprintf(responseView, "Error: %s\n", err)

			break
		}

		fmt.Fprintf(responseView, "Status: %s\n\nResponse Body: %s\n", resStatus, resbody)

	case "GET":

		req, err := http.NewRequest("GET", url, nil)

		if err != nil {
			fmt.Fprintf(responseView, "Error: %s\n", err)
		}

		if url == "" {

			fmt.Fprintf(responseView, "Error: %s\n", "Url can not be empty")

			break

		}
		if len(headersArray) > 0 {

			for _, header := range headersArray {

				headerKeyValue := strings.Split(header, ":")
				req.Header.Add(headerKeyValue[0], headerKeyValue[1])
			}
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

	// Check if terminal size is below the minimum required dimensions
	if maxX < minWidth || maxY < minHeight {
		if v, err := g.SetView("error", 0, 0, maxX-1, maxY-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = "Error"
			v.Wrap = true
			v.Frame = true
			v.Autoscroll = true
			v.BgColor = gocui.ColorRed
			v.FgColor = gocui.ColorWhite
			v.Clear()
			fmt.Fprintf(v, "Terminal size is too small! Please resize to at least %dx%d.", minWidth, minHeight)
		}
		return nil // Skip the regular layout if the terminal is too small
	}

	// Remove the error view if terminal is resized to a valid size
	g.DeleteView("error")

	if v, err := g.SetView("url", 0, 0, maxX/2-1, 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "API Endpoint"
		v.Editable = true

		if _, err := g.SetCurrentView("url"); err != nil {
			return err
		}
	}

	if v, err := g.SetView("method", 0, 3, maxX/2-1, 5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Method"
		v.Editable = true
	}

	if v, err := g.SetView("headers", 0, 6, maxX/2-1, maxY/3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Headers (Key:Value)"
		v.Editable = true
		v.Autoscroll = true
	}

	if v, err := g.SetView("body", 0, maxY/3+1, maxX/2-1, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Request Body"
		v.Editable = true
		fmt.Fprintln(v, "{}")
	}

	if v, err := g.SetView("response", maxX/2, 0, maxX-1, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Response"
		v.Wrap = true
	}

	if v, err := g.SetView("help", 0, maxY-4, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Help"
		v.Wrap = true
		v.Frame = true

		fmt.Fprintln(v, helpText)
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

func openModal(g *gocui.Gui, v *gocui.View) error {

	maxX, maxY := g.Size()

	modalWidth := 50
	modalHeight := 2

	startX := (maxX / 2) - (modalWidth / 2)
	startY := (maxY / 2) - (modalHeight / 2)
	endX := startX + modalWidth
	endY := startY + modalHeight

	if v, err := g.SetView("saveModal", startX, startY, endX, endY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Wrap = true
		v.Frame = true
		v.Title = "Save Response (Enter to submit and Esc to quit)"
		v.Editable = true

		if _, err := g.SetCurrentView("saveModal"); err != nil {
			return err
		}
	}

	return nil

}

func closeModal(g *gocui.Gui, v *gocui.View) error {

	g.DeleteView("saveModal")

	_, _ = g.SetCurrentView("url")

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

	g.InputEsc = true
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

	if err := g.SetKeybinding("", gocui.KeyCtrlE, gocui.ModNone, openModal); err != nil {

		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlR, gocui.ModNone, cleanFields); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("saveModal", gocui.KeyEnter, gocui.ModNone, saveResponse); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("saveModal", gocui.KeyEsc, gocui.ModNone, closeModal); err != nil {
		log.Panicln(err)
	}
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
