package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var goColor = tcell.NewHexColor(0x007d9c)
var userNameForm = tview.NewForm()
var app = tview.NewApplication()
var pages = tview.NewPages()

func connect() {

}

// Function to populate the form with inputs
func setupUserNameForm() {
	userName := ""
	userNameForm.SetButtonsAlign(tview.AlignCenter)
	userNameForm.AddInputField("Username", "", 50, nil, func(enteredUserName string) {
		userName = enteredUserName
	})
	_ = userName
	userNameForm.AddButton("Connect", connect)
	userNameForm.AddButton("Exit", func() {
		app.Stop()
	})
	userNameForm.SetBorder(true).SetBorderColor(goColor).SetTitle("GOChat")
}

func main() {
	setupUserNameForm()
	pages.AddPage("Login", userNameForm, true, true)
	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
