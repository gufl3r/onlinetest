package main

import (
	"server"
	"client"
	"fmt"
	"time"
)

func Manageas_server() {
	var port string
	fmt.Print("Enter port: ")
	fmt.Scanln(&port)
	server.Createserver(port)
}

func Manageas_client() {
	var ip string
	fmt.Print("Enter server ip: ")
	fmt.Scanln(&ip)
	client.Connectto(ip)
}

func main() {
	for {
		time.Sleep(500*time.Millisecond)
		var input string
		fmt.Print("Enter 'server' or 'client': ")
		fmt.Scanln(&input)
		if input == "server" {
			Manageas_server()
		} else if input == "client" {
			Manageas_client()
		} else {
			fmt.Println("Invalid input")
		}
	}
}