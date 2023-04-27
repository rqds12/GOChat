package main

import (
	"fmt"
	"net"
)

// func connect(ip string, port string) {
// 	//connect
// 	conn, err := net.Dial("tcp", ip+":"+port)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer conn.Close()
// 	go printRecieveStdio(conn)
// 	for {
// 		command := getUserInput()
// 		conn.Write([]byte(command))

// 	}

// }

// printRecieve
// blocking function to perform an action on recieved data
// particularly for recieving and printing data
func printRecieve(conn net.Conn, print func(string)) {
	for {
		reply := readString(conn)
		print(reply)
	}
}

func printRecieveStdio(conn net.Conn) {
	printRecieve(conn, func(s string) { fmt.Println(s) })
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
