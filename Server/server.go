package main

/*
* Author: Seth Kroeker
* Course: COSC 4653 Advanced Networks
* Description: Implements the server according to the
* protocol specifications
 */

/*
* Server Specification
* All commands are in ASCII
*
* Server send commands
* CONNECTED|<name>|
* REJECTED|<name>|<reason>|
* PUBLIC|<sending name>|<message>|
* JOINED|<name>|
* LEFT|<name>|
* ERROR|<unknown command>|
* PRIVATE|<sender>|<message>|
* PRIVRR|<recipient>|
* LIST|<count>|<pipe-delimited-list-of-names>|
* TIME|<time string>|
*
* Server recieve commands
* CONNECT|<name requested>|
* SAY|<message>|
* EXIT|
* PRIVATE|<recipient>|<message>|
* LIST|
* TIME|

 */

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	EPOLLET        = 1 << 31
	MaxEpollEvents = 32
)

type Client struct {
	conn net.Conn
	name string
}

var mChan = make(chan struct {
	Client
	string
})
var clientArray []Client

func getClientFromName(clientArray []Client, name string) (Client, int) {
	for i := 0; i < len(clientArray); i++ {
		if clientArray[i].name == name {
			return clientArray[i], i
		}
	}
	return Client{nil, ""}, -1
}

func getNameFromClient(clientArray []Client, c Client) (string, int) {
	for i := 0; i < len(clientArray); i++ {
		if clientArray[i].conn == c.conn {
			return clientArray[i].name, i
		}
	}
	return "", -1
}

func broadcastMessage(clientArray []Client, message string) {
	for i := 0; i < len(clientArray); i++ {
		connRecipient := clientArray[i].conn
		connRecipient.Write([]byte(message))
	}
}
func sendMessage(client Client, meessage string) {
	broadcastMessage([]Client{client}, meessage)
}

func logCommands(message string) {
	file, err := os.OpenFile("serverLog.txt", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		fmt.Println(err)
		return
	}
	time := time.Now().Format("2006-01-02 15:04:05")
	output := "[" + time + "]: " + message
	fmt.Fprintf(file, "%s\n", output)
	fmt.Println(output)

}

func handleConnection() {
	for {
		c := <-mChan
		msg := c.string
		conn := c.conn
		var strSplit = strings.Split(msg, "|")

		//match accoding to command
		switch strings.ToUpper(strSplit[0]) {
		case "CONNECT":
			name := strSplit[1]
			// check if name exists
			clientExist, _ := getClientFromName(clientArray, name)

			if clientExist.name == "" {
				//success
				clientArray = append(clientArray, Client{conn: conn, name: strSplit[1]})
				conn.Write([]byte("CONNECTED|" + strSplit[1] + "|"))
				// sa, err :=
				addr := conn.RemoteAddr().String()
				s := fmt.Sprintf("%v  connected as %v", addr, name)
				logCommands(s)

				//notify users of newly joined user
				message := "JOINED|" + name + "|"
				broadcastMessage(clientArray, message)
			} else {
				//failed
				//name exists
				addr := conn.RemoteAddr().String()
				s := fmt.Sprintf("%v tried connecting as %v.  Request rejected", addr, name)
				logCommands(s)
				conn.Write([]byte("REJECTED|" + strSplit[1] + "|Name is Taken|"))

			}
		case "SAY":
			//broadcast message to all registered users
			name, index := getNameFromClient(clientArray, Client{conn, ""})
			message := "PUBLIC|" + name + "|" + strSplit[1] + "|"
			temp := make([]Client, len(clientArray))
			copy(temp, clientArray)
			test := temp[:index]
			rest := temp[index+1:]
			test = append(test, rest...)
			broadcastMessage(test, message)
			//log
			addr := conn.RemoteAddr().String()
			s := fmt.Sprintf("%v %v said %v. ", addr, name, message)
			logCommands(s)
		case "EXIT":
			name, index := getNameFromClient(clientArray, Client{conn, ""})
			message := "LEFT|" + name + "|"
			//log
			addr := conn.RemoteAddr().String()
			s := fmt.Sprintf("%v [%v] disconnected. ", addr, name)
			logCommands(s)
			//remove from array
			conn.Close()
			(clientArray) = append((clientArray)[:index], (clientArray)[index+1:]...)
			broadcastMessage(clientArray, message)

		case "PRIVATE":
			name := strSplit[1]
			recievedMessage := strSplit[2]
			var message = ""
			sender, senderIndex := getNameFromClient(clientArray, Client{conn, ""})
			recipient, indexOfRecipient := getClientFromName(clientArray, name)
			if indexOfRecipient >= 0 {
				message = "PRIVATE|" + sender + "|" + strSplit[2] + "|"
				broadcastMessage([]Client{(clientArray)[indexOfRecipient]}, message)

				addr := conn.RemoteAddr().String()
				addr_r := recipient.conn.RemoteAddr().String()
				s := fmt.Sprintf("%v messaged %v ", addr, addr_r)
				logCommands(s)

			} else {
				//send error
				message = "PRIVERR|" + name + "|" + recievedMessage + "|"
				broadcastMessage([]Client{(clientArray)[senderIndex]}, message)

				addr := conn.RemoteAddr().String()
				s := fmt.Sprintf("%v attempted to message a nonexistent user", addr)
				logCommands(s)
			}
		case "LIST":
			_, senderIndex := getNameFromClient(clientArray, Client{conn, ""})
			count := len(clientArray)
			message := "LIST|" + strconv.Itoa(count) + "|"
			for i := 0; i < count; i++ {
				message += (clientArray)[i].name + "|"
			}
			broadcastMessage([]Client{(clientArray)[senderIndex]}, message)

		case "TIME":
			_, senderIndex := getNameFromClient(clientArray, Client{conn, ""})
			time := time.Now().Format("2006-01-02 15:04:05")
			message := "TIME|" + time
			broadcastMessage([]Client{(clientArray)[senderIndex]}, message)
		default:
			message := "ERROR|" + strSplit[0] + "|"
			_, senderIndex := getNameFromClient(clientArray, Client{conn, ""})
			broadcastMessage([]Client{(clientArray)[senderIndex]}, message)
			addr := conn.RemoteAddr().String()
			s := fmt.Sprintf("%v attempted to issue invalid command %v", addr, strSplit[0:])
			logCommands(s)

		}
	}

}

func handleConn(c Client) {
	conn := c.conn
	for {
		buff := make([]byte, 1024)
		m, err := conn.Read([]byte(buff))
		if err != nil {
			name, index := getNameFromClient(clientArray, Client{conn, ""})
			if index > -1 {
				addr := conn.RemoteAddr().String()
				s := fmt.Sprintf("%v [%v] disconnected. ", addr, name)
				logCommands(s)
				(clientArray) = append((clientArray)[:index], (clientArray)[index+1:]...)
			}
			return
			// panic(err)
		}
		//pipe the Client and the read message into a channel
		mChan <- struct {
			Client
			string
		}{c, string(buff[:m])}
	}
}

func main() {

	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		// handle error
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		var c Client = Client{conn, ""}
		go handleConn(c)
		go handleConnection()
	}

}
