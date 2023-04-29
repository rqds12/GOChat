package main

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
	fmt.Println(clientArray)
	fmt.Println(c.name)
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
		// client := c.Client
		conn := c.conn
		// defer syscall.Close(fd)
		// var buf [1024]byte

		// for {
		// 	nbytes, e := syscall.Read(fd, buf[:])
		// 	//zeroize buf and convert to string
		// 	if nbytes <= 0 {
		// 		break
		// 	}
		// var string = string(buf[:nbytes])
		fmt.Println(msg)
		var strSplit = strings.Split(msg, "|")

		//match accoding to command
		switch strings.ToUpper(strSplit[0]) {
		case "CONNECT":
			name := strSplit[1]
			fmt.Println("Connect")
			// check if name exists
			// fmt.Println(fd)
			// fmt.Println(*clientArray)
			clientExist, _ := getClientFromName(clientArray, name)

			if clientExist.name == "" {
				//success
				clientArray = append(clientArray, Client{conn: conn, name: strSplit[1]})
				conn.Write([]byte("CONNECTED|" + strSplit[1] + "|"))
				// sa, err :=
				addr := conn.RemoteAddr().String()
				s := fmt.Sprintf("%v  connected as %v", addr, name)
				logCommands(s)
				// if err == nil {
				// 	// port := sa.(*syscall.SockaddrInet4).Port
				// 	// addr := sa.(*syscall.SockaddrInet4).Addr

				// }

				//notify users of newly joined user
				message := "JOINED|" + name + "|"
				broadcastMessage(clientArray, message)
			} else {
				//failed
				//name exists
				// sa, err := syscall.Getpeername(fd)

				// if err == nil {
				// 	port := sa.(*syscall.SockaddrInet4).Port
				// 	addr := sa.(*syscall.SockaddrInet4).Addr
				// 	s := fmt.Sprintf("%v.%v.%v.%v:%v tried connecting as %v.  Request rejected", addr[0], addr[1], addr[2], addr[3], port, name)
				// 	logCommands(s)
				// }

				addr := conn.RemoteAddr().String()
				s := fmt.Sprintf("%v tried connecting as %v.  Request rejected", addr, name)
				logCommands(s)
				conn.Write([]byte("REJECTED|" + strSplit[1] + "|name in use|"))
				// syscall.Write(fd, []byte("REJECTED|"+strSplit[1]+"name in use|"))
				conn.Close()
			}
		case "SAY":
			fmt.Println("Say")
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
			// sa, err := syscall.Getpeername(fd)

			// if err == nil {
			// 	port := sa.(*syscall.SockaddrInet4).Port
			// 	addr := sa.(*syscall.SockaddrInet4).Addr
			// 	s := fmt.Sprintf("%v.%v.%v.%v:%v [%v] said %v. ", addr[0], addr[1], addr[2], addr[3], port, name, message)
			// 	logCommands(s)
			// }

			addr := conn.RemoteAddr().String()
			s := fmt.Sprintf("%v %v said %v. ", addr, name, message)
			logCommands(s)
		case "EXIT":
			fmt.Println("Exit")
			name, index := getNameFromClient(clientArray, Client{conn, ""})
			message := "LEFT|" + name + "|"
			//log
			addr := conn.RemoteAddr().String()
			s := fmt.Sprintf("%v [%v] disconnected. ", addr, name)
			logCommands(s)
			// sa, err := syscall.Getpeername(fd)

			// if err == nil {
			// 	port := sa.(*syscall.SockaddrInet4).Port
			// 	addr := sa.(*syscall.SockaddrInet4).Addr
			// 	s := fmt.Sprintf("%v.%v.%v.%v:%v [%v] disconnected. ", addr[0], addr[1], addr[2], addr[3], port, name)
			// 	logCommands(s)
			// }

			//remove from array
			// syscall.Close(fd)
			conn.Close()
			fmt.Println(index)
			(clientArray) = append((clientArray)[:index], (clientArray)[index+1:]...)
			broadcastMessage(clientArray, message)

		case "PRIVATE":
			fmt.Println("Private")
			name := strSplit[1]
			recievedMessage := strSplit[2]
			var message = ""
			sender, senderIndex := getNameFromClient(clientArray, Client{conn, ""})
			recipient, indexOfRecipient := getClientFromName(clientArray, name)
			if indexOfRecipient >= 0 {
				message = "PRIVATE|" + sender + "|" + strSplit[2] + "|"
				broadcastMessage([]Client{(clientArray)[indexOfRecipient]}, message)
				// sa_r, err1 := syscall.Getpeername(recipient)
				// port_r := sa_r.(*syscall.SockaddrInet4).Port
				// addr_r := sa_r.(*syscall.SockaddrInet4).Addr
				// sa, err := syscall.Getpeername(fd)

				// if err == nil && err1 == nil {
				// 	port := sa.(*syscall.SockaddrInet4).Port
				// 	addr := sa.(*syscall.SockaddrInet4).Addr
				// 	s := fmt.Sprintf("%v.%v.%v.%v:%v messaged %v.%v.%v.%v:%v", addr[0], addr[1], addr[2], addr[3], port, addr_r[0], addr_r[1], addr_r[2], addr_r[3], port_r)
				// 	logCommands(s)
				// }
				addr := conn.RemoteAddr().String()
				addr_r := recipient.conn.RemoteAddr().String()
				s := fmt.Sprintf("%v messaged %v ", addr, addr_r[0], addr_r[1], addr_r[2], addr_r[3])
				logCommands(s)

			} else {
				//send error
				message = "PRIVERR|" + name + "|" + recievedMessage + "|"
				broadcastMessage([]Client{(clientArray)[senderIndex]}, message)
				// sa, err := syscall.Getpeername(fd)

				// if err == nil {
				// 	port := sa.(*syscall.SockaddrInet4).Port
				// 	addr := sa.(*syscall.SockaddrInet4).Addr
				// 	s := fmt.Sprintf("%v.%v.%v.%v:%v attempted to message a nonexistent user", addr[0], addr[1], addr[2], addr[3], port)
				// 	logCommands(s)
				// }
				addr := conn.RemoteAddr().String()
				s := fmt.Sprintf("%v attempted to message a nonexistent user", addr)
				logCommands(s)
			}
		case "LIST":
			fmt.Println("List")
			_, senderIndex := getNameFromClient(clientArray, Client{conn, ""})
			count := len(clientArray)
			message := "LIST|" + strconv.Itoa(count) + "|"
			for i := 0; i < count; i++ {
				message += (clientArray)[i].name + "|"
			}
			broadcastMessage([]Client{(clientArray)[senderIndex]}, message)

		case "TIME":
			fmt.Println("Time")
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
			// sa, err := syscall.Getpeername(fd)
			// if err == nil {
			// 	port := sa.(*syscall.SockaddrInet4).Port
			// 	addr := sa.(*syscall.SockaddrInet4).Addr
			// 	s := fmt.Sprintf("%v.%v.%v.%v:%v attempted to issue invalid command %v", addr[0], addr[1], addr[2], addr[3], port, strSplit[0:])
			// 	logCommands(s)
			// }

		}
	}

	// if e != nil {
	// 	break
	// }
	// }
}

func handleConn(c Client) {
	conn := c.conn
	for {
		// msg := ""
		buff := make([]byte, 1024)
		m, err := conn.Read([]byte(buff))
		// fmt.Println(buff[:m])
		if err != nil {
			return
			// panic(err)
		}

		mChan <- struct {
			Client
			string
		}{c, string(buff[:m])}
	}

}

// // Process to wait for messages to be entered, then sends them to the server
// func messageSender(conn net.Conn, m *sync.Mutex) {
// 	for {
// 		message := <-mChan
// 		if len(message) == 0 {
// 			continue
// 		}
// 		// Lock for writing to channel
// 		m.Lock()
// 		parseMessage(conn, message)
// 		// conn.Write([]byte("SAY|" + message + "|"))
// 		m.Unlock()
// 	}
// }

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

	// var event syscall.EpollEvent
	// var events [MaxEpollEvents]syscall.EpollEvent
	// var clientArray []Client

	// fd, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	// defer syscall.Close(fd)

	// if err = syscall.SetNonblock(fd, true); err != nil {
	// 	fmt.Println("setnonblock1: ", err)
	// 	os.Exit(1)
	// }

	// addr := syscall.SockaddrInet4{Port: 9000}
	// copy(addr.Addr[:], net.ParseIP("127.0.0.1").To4())

	// syscall.Bind(fd, &addr)
	// syscall.Listen(fd, 10)

	// epfd, e := syscall.EpollCreate1(0)
	// if e != nil {
	// 	fmt.Println("epoll_create1: ", e)
	// 	os.Exit(1)
	// }
	// defer syscall.Close(epfd)

	// event.Events = syscall.EPOLLIN
	// event.Fd = int32(fd)
	// if e = syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, fd, &event); e != nil {
	// 	fmt.Println("epoll_ctl: ", e)
	// 	os.Exit(1)
	// }

	// for {
	// 	//populates events and returns the number of events
	// 	nevents, e := syscall.EpollWait(epfd, events[:], -1)
	// 	if e != nil {
	// 		fmt.Println("epoll_wait: ", e)
	// 		//try again
	// 		nevents, e = syscall.EpollWait(epfd, events[:], -1)
	// 		if e != nil {
	// 			fmt.Println("epoll_wait final: ", e)
	// 			break
	// 		}
	// 	}

	// 	for ev := 0; ev < nevents; ev++ {
	// 		if int(events[ev].Fd) == fd {
	// 			connFd, _, err := syscall.Accept(fd)

	// 			if err != nil {
	// 				fmt.Println("accept: ", err)
	// 				continue
	// 			}

	// 			syscall.SetNonblock(fd, true)
	// 			event.Events = syscall.EPOLLIN | EPOLLET
	// 			event.Fd = int32(connFd)
	// 			if err := syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, connFd, &event); err != nil {
	// 				fmt.Print("epoll_ctl: ", connFd, err)
	// 				os.Exit(1)
	// 			}
	// 			sa, _ := syscall.Getpeername(connFd)
	// 			port := sa.(*syscall.SockaddrInet4).Port
	// 			addr := sa.(*syscall.SockaddrInet4).Addr
	// 			s := fmt.Sprintf("Client connected: %v.%v.%v.%v:%v", addr[0], addr[1], addr[2], addr[3], port)
	// 			logCommands(s)
	// 		} else {
	// 			go handleConnection(int(events[ev].Fd), &clientArray)
	// 		}
	// 	}
	//
	// }
}
