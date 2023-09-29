package client

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/http/httptrace"
	"os"
	"strings"
	"time"
)

var prefix string = "!Client:"
var ip string
var traceCtx context.Context
var connected bool = true

// commands
var pinginterval []time.Time = []time.Time{time.Now(), time.Now()}

func Connectto(serverip string) {
	ip = serverip
	fmt.Println(prefix, "Trying to connect to server "+serverip+"...")
	clientTrace := &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) {},
	}
	traceCtx = httptrace.WithClientTrace(context.Background(), clientTrace)
	CheckConnection()
}

func CheckConnection() {
	req, err := http.NewRequestWithContext(traceCtx, http.MethodGet, "http://"+ip+"/checkconnection", nil)
	if err != nil {
		fmt.Print(prefix, " Error - Error formulating request\n\n")
		fmt.Println(prefix, "Error info:", err.Error())
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Print(prefix, " Error - couldn't connect to server (")
		if strings.Contains(err.Error(), "no such host") {
			fmt.Print("server not found)\n\n")
		} else {
			fmt.Print("unknown error)\n\n")
		}
		fmt.Println(prefix, "Error info:", err.Error())
		return
	}
	bs := make([]byte, 128)
	resp.Body.Read(bs)
	fmt.Println(prefix, string(bs))
	go MaintainConnection()
	go ReceiveData(true, []string{})
	SendCommands()
}

func SendCommands() {
	for {
		var command []string
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		command = strings.Split(text, " ")
		// remove everything after "!!"
		for i, v := range command {
			if strings.Contains(v, "!!") {
				command = command[:i]
				break
			}
		}
		// if last char of last element is "\n"
		if command[len(command)-1][len(command[len(command)-1])-1] == '\n' {
			fmt.Println(prefix, "Error - command did not end properly")
			continue
		} else {
			if command[0] == "ping" {
				// start timer and wait for response
				pinginterval[0] = time.Now()
			}
			if command[0] == "leave" {
				connected = false
			}
			ReceiveData(false, command)
		}
		if !connected {
			return
		}
	}
}

func ReceiveData(permanent bool, command []string) {
	var query string
	if permanent {
		query = "receive/regular"
	} else {
		query = "receive/custom/" + strings.Join(command, ";;")
	}
	for {
		// all connection errors are handled in maintainconnection
		req, err := http.NewRequestWithContext(traceCtx, http.MethodGet, "http://"+ip+"/"+query, nil)
		resp, err := http.DefaultClient.Do(req)
		_ = err
		// 5mb limit
		bs := make([]byte, 5242880)
		resp.Body.Read(bs)
		// response to string slice separated by ";;"
		var datareceived []string = strings.Split(string(bs), ";;")
		if datareceived[0] == "regular" {
			// add here what to do with regular data
		}
		if datareceived[0] == "ping" {
			pinginterval[1] = time.Now()
			fmt.Println(prefix, "pong!", (pinginterval[1].Sub(pinginterval[0]).String()))
		}
		if datareceived[0] == "id" {
			fmt.Println(prefix, "ID:", datareceived[1])
		}
		if datareceived[0] == "where" {
			fmt.Println(prefix, "Current server is:", datareceived[1])
		}
		if datareceived[0] == "clients" {
			if datareceived[len(datareceived)-2] == "(ever)"{
				fmt.Println(prefix, datareceived[1], "clients ever connected:", datareceived[2])
			}
			if datareceived[len(datareceived)-2] == "(connected)"{
				fmt.Println(prefix, datareceived[1], "clients connected:", datareceived[2])
			}
		}
		if datareceived[0] == "null" {
			fmt.Println(prefix, "Empty response (probably unknown command or bad syntax)")
		}
		if !permanent || !connected {
			return
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func MaintainConnection() {
	for {
		req, err := http.NewRequestWithContext(traceCtx, http.MethodGet, "http://"+ip+"/renew", nil)
		if err != nil {
			fmt.Print(prefix, " Error - connection lost\n\n")
			fmt.Println(prefix, "Error info:", err.Error())
			connected = false
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Print(prefix, " Error - connection lost\n\n")
			fmt.Println(prefix, "Error info:", err.Error())
			connected = false
		}
		if resp.StatusCode != 200 {
			fmt.Println(prefix, "Error - couldn't renew connection, connection not accepted or doesn't exist")
			connected = false
		}
		if resp.ContentLength == 0 {
			fmt.Println(prefix, "Error - couldn't renew connection, server didn't respond")
			connected = false
		}
		if !connected {
			fmt.Println(prefix, "Disconnected from server")
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}
