package main

import (
	"net"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	SERVER_PORT = "9000"
	SERVER_TYPE = "tcp"
)

var goColor = tcell.NewHexColor(0x007d9c)
var userNameForm = tview.NewForm()
var app = tview.NewApplication()
var pages = tview.NewPages()
var chatRoom = tview.NewFlex()
var chatFeed = tview.NewBox().SetBorder(true).SetBorderColor(goColor).SetTitle("GOChat")
var messageForm = tview.NewForm()
var ipAddr = ""
var userName = ""
var userMessage = ""

//var conn = net.Conn()

// ----------------------- Networking Functions ----------------------------------------------

func connect() {
	conn, err := net.Dial(SERVER_TYPE, ipAddr+":"+SERVER_PORT)
	if err != nil {
		panic(err) // add failed connection message to page
	}
	conn.Write([]byte("CONNECT|" + userName))
	buff := make([]byte, 1024)
	mLen, err := conn.Read(buff)
	if err != nil {
		panic(err)
	}
	message := string(buff[:mLen])
	mSplit := strings.Split(message, "|")
	switch mSplit[0] {
	case "CONNECTED":
		pages.SwitchToPage("Chat")
		break
	case "REJECTED":
		disconnect()
		//reason := mSplit[2]
		//TODO: add reason to login page
	}

}

func sendMessage() {
	//TODO: this
}

func disconnect() {
	app.Stop()
	// TODO: DC from server and cleanup socket
}

// --------------------------------------- UI Functions ------------------------------------------------------

// Function to populate the form with inputs
func setupUserNameForm() {
	userNameForm.SetButtonsAlign(tview.AlignCenter)
	userNameForm.AddInputField("Username", "", 50, nil, func(enteredUserName string) {
		userName = enteredUserName
	})
	userNameForm.AddInputField("IP", "", 50, nil, func(enteredIp string) {
		ipAddr = enteredIp
	})
	userNameForm.AddButton("Connect", connect)
	userNameForm.AddButton("Exit", func() {
		app.Stop()
	})
	userNameForm.SetBorder(true).SetBorderColor(goColor).SetTitle("GOChat")
}

// Populates the message form
func setupMessageForm() {
	messageForm.SetButtonsAlign(tview.AlignLeft)
	messageForm.AddInputField("", "", 500, func(textToCheck string, lastChar rune) bool {
		return textToCheck != ""
	}, func(message string) {
		userMessage = message
	})
	messageForm.AddButton("Send", sendMessage)
	messageForm.SetButtonsAlign(tview.AlignRight)
	messageForm.AddButton("Exit", disconnect)
}

// Populates the chatroom flexbox
func setupChatRoom() {
	setupMessageForm()
	chatRoom.SetDirection(tview.FlexRow)
	chatRoom.AddItem(chatFeed, 0, 5, false) // This sets up flexbox so that chatFeed is 5 times the size of messageForm
	chatRoom.AddItem(messageForm, 0, 1, false)

}

func main() {
	setupUserNameForm()
	setupChatRoom()
	pages.AddPage("Chat", chatRoom, true, false)
	pages.AddPage("Login", userNameForm, true, true)
	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
