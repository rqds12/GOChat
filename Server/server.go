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
	"strings"
	"syscall"
)

const (
	EPOLLET        = 1 << 31
	MaxEpollEvents = 32
)

type Client struct {
	fd   int
	name string
}

func getFdFromName(clientArray *[]Client, name string) int {
	for i := 0; i < len(*clientArray); i++ {
		if (*clientArray)[i].name == name {
			return (*clientArray)[i].fd
		}
	}
	return -1
}

func getNameFromFd(clientArray *[]Client, fd int) string {
	for i := 0; i < len(*clientArray); i++ {
		if (*clientArray)[i].fd == fd {
			return (*clientArray)[i].name
		}
	}
	return ""
}

func handleConnection(fd int, clientArray *[]Client) {
	defer syscall.Close(fd)
	var buf [1024]byte
	for {
		nbytes, e := syscall.Read(fd, buf[:])
		//zeroize buf and convert to string
		var string = string(buf[:nbytes])
		var strSplit = strings.Split(string, "|")

		//match accoding to command
		switch strSplit[0] {
		case "CONNECT":
			fmt.Println("Connect")
			// check if name exists
			var name = getNameFromFd(clientArray, fd)
			if name == "" {
				//success
				*clientArray = append(*clientArray, Client{fd: fd, name: strSplit[1]})
				syscall.Write(fd, []byte("CONNECTED|"+strSplit[1]))
			} else {
				//failed
				//name exists
				syscall.Write(fd, []byte("REJECTED|"+strSplit[1]+"name in use"))
			}
		case "SAY":
			fmt.Println("Say")

		case "EXIT":
			fmt.Println("Exit")
		case "PRIVATE":
			fmt.Println("Private")
		case "LIST":
			fmt.Println("List")
		case "TIME":
			fmt.Println("Time")

		}

		if nbytes > 0 {
			fmt.Printf(">>> %s", buf)
			syscall.Write(fd, buf[:nbytes])
			fmt.Printf("<<< %s", buf)
		}
		if e != nil {
			break
		}
	}
}

func main() {
	var event syscall.EpollEvent
	var events [MaxEpollEvents]syscall.EpollEvent
	var clientArray []Client

	fd, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer syscall.Close(fd)

	if err = syscall.SetNonblock(fd, true); err != nil {
		fmt.Println("setnonblock1: ", err)
		os.Exit(1)
	}

	addr := syscall.SockaddrInet4{Port: 2000}
	copy(addr.Addr[:], net.ParseIP("0.0.0.0").To4())

	syscall.Bind(fd, &addr)
	syscall.Listen(fd, 10)

	epfd, e := syscall.EpollCreate1(0)
	if e != nil {
		fmt.Println("epoll_create1: ", e)
		os.Exit(1)
	}
	defer syscall.Close(epfd)

	event.Events = syscall.EPOLLIN
	event.Fd = int32(fd)
	if e = syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, fd, &event); e != nil {
		fmt.Println("epoll_ctl: ", e)
		os.Exit(1)
	}

	for {
		//populates events and returns the number of events
		nevents, e := syscall.EpollWait(epfd, events[:], -1)
		if e != nil {
			fmt.Println("epoll_wait: ", e)
			break
		}

		for ev := 0; ev < nevents; ev++ {
			if int(events[ev].Fd) == fd {
				connFd, _, err := syscall.Accept(fd)
				if err != nil {
					fmt.Println("accept: ", err)
					continue
				}
				syscall.SetNonblock(fd, true)
				event.Events = syscall.EPOLLIN | EPOLLET
				event.Fd = int32(connFd)
				if err := syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, connFd, &event); err != nil {
					fmt.Print("epoll_ctl: ", connFd, err)
					os.Exit(1)
				}
			} else {
				go handleConnection(int(events[ev].Fd), &clientArray)
			}
		}

	}
}
