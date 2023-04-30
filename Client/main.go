package main

/*
* Author: Alex Wildman
* Course: COSC 4653 Advanced Networks
* Description: Implements a client that is compatable with the
* protocol specifications
 */
import (
	"io"
	"net"
	"strconv"
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
var messageForm = tview.NewForm()
var userMessage = ""
var ipAddr = ""
var userName = ""
var mChan = make(chan string)
var disconChan = make(chan int)

// ----------------------- Networking Functions ----------------------------------------------

func sendMessage() {
	if len(userMessage) > 0 {
		mChan <- userMessage
		chatFeed.Write([]byte("[" + userName + "]: " + userMessage + "\n"))
		userMessage = ""
		setupMessageForm()
	}
}

func disconnect() {
	disconChan <- 0
}

func disconnectAndClose() {
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
		//deal with unknown command()
		_, err := chatFeed.Write([]byte("[SERVER]: UNKNOWN COMMAND " + mSplit[1] + "\n"))
		if err != nil {
			panic(err)
		}
	case "PRIVATE":
		_, err := chatFeed.Write([]byte("[" + mSplit[1] + "]:" + mSplit[2] + "\n"))
		if err != nil {
			panic(err)
		}
	case "PRIVERR":
		//user doesn't exist
		_, err := chatFeed.Write([]byte("[SERVER]: USER " + mSplit[1] + " is not in the server.\n"))
		if err != nil {
			panic(err)
		}
	case "TIME":
		_, err := chatFeed.Write([]byte("Server time is: " + mSplit[1] + "\n"))
		if err != nil {
			panic(err)
		}
	case "LIST":
		s := ""
		count, err := strconv.Atoi(mSplit[1])
		if err != nil {
			panic(err)
		}
		for i := 0; i < count; i++ {
			s += mSplit[i+2] + "|"
		}
		_, err = chatFeed.Write([]byte("[SERVER]: List of users: " + s + "\n"))
		if err != nil {
			panic(err)
		}

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
// If it receives a 1 on the disconnect signal, also stops app
func disconnector(conn net.Conn, m *sync.Mutex) {
	code := <-disconChan
	if code == 1 {
		app.Stop()
	}
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
	if userName == "" {
		writeError("Name must not be empty.")
		return
	}
	// Check if name contains whitespace
	if strings.Contains(userName, " ") {
		writeError("Name must not contain spaces.")
		return
	}
	conn, err := net.Dial(SERVER_TYPE, ipAddr+":"+SERVER_PORT)
	if err != nil {
		writeError("Failed to connect to that IP.")
		return
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
		reason := mSplit[2]
		writeError("Connection refused by server. Reason: " + reason)
		return
	}

	// Spins up the server and message handling threads if connection successful
	go readServer(conn)
	go messageSender(conn, &m)

}

// --------------------------------------- UI Functions ------------------------------------------------------

// Function to populate the form with inputs
func setupUserNameForm(err string) {
	userNameForm.Clear(true)
	userNameForm.SetButtonsAlign(tview.AlignLeft)
	userName = ""
	ipAddr = ""
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
	userNameForm.AddTextView("", err, 0, 0, false, false)
	userNameForm.SetBorder(true).SetBorderColor(goColor).SetTitle("GOChat")
}

// Populates the message form
func setupMessageForm() {
	messageForm.Clear(true)
	messageForm.SetButtonsAlign(tview.AlignRight)
	messageForm.AddInputField("", "", 200, nil, func(text string) {
		userMessage = text
	})
	messageForm.AddButton("Send", sendMessage)
	messageForm.AddButton("Exit", disconnectAndClose)
}

// Function to write connection errors to the connection form
func writeError(err string) {
	setupUserNameForm("ERROR: " + err)
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
	setupUserNameForm("")
	setupChatRoom()
	pages.AddPage("Chat", chatRoom, true, false)
	pages.AddPage("Login", userNameForm, true, true)
	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
