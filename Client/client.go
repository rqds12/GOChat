package main

import (
	"fmt"
	"net"
)

func main() {
	//connect
	conn, err := net.Dial("tcp", "127.0.0.1:2000")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	go print_Recieve(conn)
	for {
		command := getUserInput()
		conn.Write([]byte(command))

	}

}

func print_Recieve(conn net.Conn) {
	for {
		reply := readString(conn)
		fmt.Println(reply)
	}
}

func getUserInput() string {
	var name string
	fmt.Scan(&name)
	return name
}

func readString(conn net.Conn) string {
	reply := make([]byte, 1024)
	n, _ := conn.Read(reply)
	reply_string := string(reply[:n])
	return reply_string
}
