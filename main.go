package main

import (
	"bufio"
	"log"
	"net"
	"os"
)

const (
	port     string = "5555"
	address  string = ""
	protocol string = "tcp"
	filename string = "openWrtLog.txt"
)

var (
	home     string
	listener net.Listener
	file     *os.File
)

func init() {
	var err error

	listener, err = net.Listen(protocol, address+":"+port)
	if err != nil {
		log.Fatalf("Couldn't get TCP listener -> %v\n", err)
	}

	home, err = os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting user's home dir -> %v\n", err)
	}

	err = os.Chdir(home)
	if err != nil {
		log.Fatalf("Error cding to user's home dir -> %v\n", err)
	}

	file, err = os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("Error opening|creating file -> %v\n", err)
	}
}

func main() {
	defer listener.Close()
	defer file.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error accepting connection -> %v\n", err)
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		appendToFile(file, msg)
	}
}

func appendToFile(file *os.File, data string) {
	_, err := file.WriteString(data)
	if err != nil {
		log.Fatalf("Error appending to file -> %v\n", err)
	}
}
