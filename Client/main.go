package main

import (
	"io"
	"net"
	"strings"
	"sync"

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
var chatFeed = tview.NewTextView().SetWrap(true).SetWordWrap(true)
var messageForm = tview.NewFlex()
var userMessage = ""
var messageField = tview.NewInputField().SetFieldWidth(500).SetChangedFunc(func(enteredUserName string) {
	userMessage = enteredUserName
})
var ipAddr = ""
var userName = ""
var mChan = make(chan string)
var disconChan = make(chan int)

// ----------------------- Networking Functions ----------------------------------------------

func sendMessage() {
	if len(userMessage) > 0 {
		mChan <- userMessage
		chatFeed.Write([]byte("[" + userName + "]: " + userMessage + "\n"))
	}
}

func disconnect() {
	disconChan <- 1
}

func handleServerMessage(message string) {
	mSplit := strings.Split(message, "|")
	switch mSplit[0] {
	case "PUBLIC":
		_, err := chatFeed.Write([]byte("[" + mSplit[1] + "]: " + mSplit[2] + "\n"))
		if err != nil {
			panic(err)
		}
	case "JOINED":
		_, err := chatFeed.Write([]byte(mSplit[1] + " has joined the chat.\n"))
		if err != nil {
			panic(err)
		}
	case "LEFT":
		_, err := chatFeed.Write([]byte(mSplit[1] + " has left the chat.\n"))
		if err != nil {
			panic(err)
		}
	case "ERROR":
		//deal with unknown command
	case "PRIVATE":
		_, err := chatFeed.Write([]byte("[" + mSplit[1] + "]:" + mSplit[2] + "\n"))
		if err != nil {
			panic(err)
		}
	case "PRIVRR":
		//user doesn't exist

	case "TIME":
		_, err := chatFeed.Write([]byte("Server time is: " + mSplit[1] + "\n"))
		if err != nil {
			panic(err)
		}
	case "LIST":

	}
	// Redraws the screen (Writing does not do this automatically for some reason :(
	app.Draw()
}

// ---------------------------------------- Threading Stuff -------------------------------------------------

// Process to check for new data from the socket and pipe it into the channel
func readServer(conn net.Conn) {
	for {
		buff := make([]byte, 1024)
		mLen, err := conn.Read(buff)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if mLen > 0 {
			handleServerMessage(string(buff[:mLen]))
		}
	}
}

// Process to wait for messages to be entered, then sends them to the server
func messageSender(conn net.Conn, m *sync.Mutex) {
	for {
		message := <-mChan
		if len(message) == 0 {
			continue
		}
		// Lock for writing to channel
		m.Lock()
		parseMessage(conn, message)
		// conn.Write([]byte("SAY|" + message + "|"))
		m.Unlock()
	}
}

func parseMessage(conn net.Conn, message string) {
	send := ""

	if message[0:1] == "/" {
		//special command
		//pipe delimited command
		cmd := strings.Split(message[1:], "|")
		switch cmd[0] {
		case "private":
			if len(cmd) != 3 {
				//invalid command
				//TODO
				break
			}
			send = "PRIVATE|" + cmd[1] + "|" + cmd[2] + "|"
		case "list":
			send = "LIST|"
		case "time":
			send = "TIME|"
		}
	} else {
		send = "SAY|" + message + "|"
	}
	conn.Write([]byte(send))

}

// Waits for any input on the disconnect channel then disconnects
// This is probably a weird way to do this
func disconnector(conn net.Conn, m *sync.Mutex) {
	<-disconChan
	app.Stop()
	m.Lock()
	_, err := conn.Write([]byte("EXIT|"))
	if err != nil {
		panic(err)
	}
	conn.Close()
	m.Unlock()
}

// Function to connect to the server, then pass off socket reads and socket writes to the appropriate threads
func handleConn() {
	conn, err := net.Dial(SERVER_TYPE, ipAddr+":"+SERVER_PORT)
	if err != nil {
		panic(err) // add failed connection message to page
	}
	conn.Write([]byte("CONNECT|" + userName + "|"))
	buff := make([]byte, 1024)
	mLen, err := conn.Read(buff)
	if err != nil {
		panic(err)
	}
	message := string(buff[:mLen])
	mSplit := strings.Split(message, "|")
	var m sync.Mutex
	// Spins up the disconnector
	go disconnector(conn, &m)

	// Parses the message
	switch mSplit[0] {
	case "CONNECTED":
		pages.SwitchToPage("Chat")
		break
	case "REJECTED":
		disconnect()
		return
		//reason := mSplit[2]
		//TODO: add reason to login page
	}

	// Spins up the server and message handling threads if connection successful
	go readServer(conn)
	go messageSender(conn, &m)

}

// --------------------------------------- UI Functions ------------------------------------------------------

// Function to populate the form with inputs
func setupUserNameForm() {
	userNameForm.SetButtonsAlign(tview.AlignCenter)
	userNameForm.AddInputField("Username", "", 50, func(textToCheck string, lastChar rune) bool {
		return textToCheck != ""
	}, func(enteredUserName string) {
		userName = enteredUserName

	})
	userNameForm.AddInputField("IP", "", 50, nil, func(enteredIp string) {
		ipAddr = enteredIp
	})
	userNameForm.AddButton("Connect", handleConn)
	userNameForm.AddButton("Exit", func() {
		app.Stop()
	})
	userNameForm.SetBorder(true).SetBorderColor(goColor).SetTitle("GOChat")
}

// Populates the message form
func setupMessageForm() {
	sendButton := tview.NewButton("Send").SetSelectedFunc(sendMessage)
	exitButton := tview.NewButton("Exit").SetSelectedFunc(disconnect)
	messageForm.SetDirection(tview.FlexRow)
	messageForm.AddItem(messageField, 1, 0, true)
	messageForm.AddItem(sendButton, 1, 0, true)
	messageForm.AddItem(exitButton, 1, 0, true)
}

// Populates the chatroom flexbox
func setupChatRoom() {
	setupMessageForm()
	chatRoom.SetDirection(tview.FlexRow)
	chatFeed.SetBorder(true).SetBorderColor(goColor).SetTitle("GOChat")
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
