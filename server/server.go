package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"strconv"
)

var prefix string = "!Server:"
var clients []string
var clientsnew []string
var clientsid [][]string
var idtracker int = 0
var idtoset string

func CheckConnection(w http.ResponseWriter, r *http.Request) {
	if strings.Split(r.Header.Get("Accept"),",")[0] == "text/html" {
		fmt.Println(prefix, "Client tried to connect with a browser",r.RemoteAddr)
		r.Close = true
	} else {
		var clientaddresssplit []string = strings.Split(r.RemoteAddr, ":")
		var clientip string = strings.Join(clientaddresssplit[:len(clientaddresssplit)-1], ":")
		// check if client is in list of clientsid (means that they were connected before)
		var found bool = false
		for _, client := range clientsid {
			if client[0] == clientip {
				found = true
				break
			}
		}
		if found {
			idtoset = clientsid[0][1]
			fmt.Println(prefix, "Client connected:", r.RemoteAddr, "id:", idtoset)
		} else {
			idtoset = fmt.Sprint(idtracker)
			fmt.Println(prefix, "New client connected:", r.RemoteAddr, "id:", idtoset)
		}
		clients = append(clients, clientip)
		clientsid = append(clientsid, []string{clientip, idtoset})
		fmt.Println(prefix, "(Async) Clients amount changed:", len(clients)-1, "->", len(clients))
		idtracker++
		fmt.Fprint(w, "Connection OK, id:", idtoset)
	}
}

func RenewConnection(w http.ResponseWriter, r *http.Request) {
	var clientaddresssplit []string = strings.Split(r.RemoteAddr, ":")
	var clientip string = strings.Join(clientaddresssplit[:len(clientaddresssplit)-1], ":")
	for _, client := range clients {
		if client == clientip {
			clientsnew = append(clientsnew, clientip)
			break
		}
	}
	var found bool = false
	for _, client := range clients {
		if client == clientip {
			found = true
			break
		} else {
			fmt.Println(client, "-", clientip)
		}
	}
	if found {
		fmt.Fprint(w, "OK")
	} else {
		fmt.Println(prefix, clientip+"'s connection couldn't be checked")
	}
}

func CheckConnected() {
	for {
		for i := 0; i < len(clientsnew); i++ {
			for j := i + 1; j < len(clientsnew); j++ {
				if clientsnew[i] == clientsnew[j] {
					clientsnew = append(clientsnew[:j], clientsnew[j+1:]...)
					j--
				}
			}
		}
		if len(clientsnew)-len(clients) != 0 {
			fmt.Println(prefix, "Clients amount changed:", len(clients), "->", len(clientsnew))
		}
		clients = clientsnew
		clientsnew = []string{}
		time.Sleep(3 * time.Second)
	}
}

func SendRegularData (w http.ResponseWriter, r *http.Request) {
	var clientaddresssplit []string = strings.Split(r.RemoteAddr, ":")
	var clientip string = strings.Join(clientaddresssplit[:len(clientaddresssplit)-1], ":")
	datatosend := "regular;;"
	
	var found bool = false
	for _, client := range clients {
		if client == clientip {
			found = true
			break
		}
	}
	if found {
		fmt.Fprint(w, datatosend)
	}
}

func SendCustomData (w http.ResponseWriter, r *http.Request) {
	var clientaddresssplit []string = strings.Split(r.RemoteAddr, ":")
	var clientip string = strings.Join(clientaddresssplit[:len(clientaddresssplit)-1], ":")
	var command []string = strings.Split(r.URL.Path[len("/receive/custom/"):], ";;")
	var datatosend string
	var verified bool = false
	fmt.Println(prefix, "Command received from", clientip+":", r.URL.Path[len("/receive/custom/"):])
	if command[0] == "ping" {
		// send pong
		datatosend = "ping;;"
		verified = true
	}
	if command[0] == "id" {
		// send client id
		for _, client := range clientsid {
			if client[0] == clientip {
				datatosend = "id;;"+client[1]
				break
			}
		}
		verified = true
	}
	if command[0] == "where" {
		// send server address
		datatosend = "where;;"+r.Host
		verified = true
	}
	if command[0] == "clients" {
		// send clients id separated by ";;"
		datatosend = "clients;;"+strconv.Itoa(len(clients))+";;"
		if len(command) < 2 {
			command = append(command, "")
		}
		if command[1] == "ever" {
			// send all clients id separated by ";;"
			for _, client := range clientsid {
				datatosend = datatosend+client[1]+",;;"
			}
			verified = true
			datatosend = datatosend+"(ever);;"
		}
		if command[1] == "connected" {
		// send connected clients id separated by ";;"
		// loop through clients and then through clientsid to find its id, if it is then add it to datatosend
			for _, client := range clients {
				for _, clientid := range clientsid {
					if client == clientid[0] {
						datatosend = datatosend+clientid[1]+",;;"
						break
					}
				}
			}
			verified = true
			datatosend = datatosend+"(connected);;"
		}
	}
	if !verified {
		datatosend = "null;;"
	}
	var found bool = false
	for _, client := range clients {
		if client == clientip {
			found = true
			break
		}
	}
	if found {
		fmt.Fprint(w, datatosend)
	} else {
		fmt.Println(prefix, clientip+"'s command request rejected")
	}
}

func Createserver(port string) {
	fmt.Println(prefix, "Server starting at port", port)
	go CheckConnected()
	http.HandleFunc("/checkconnection", CheckConnection)
	http.HandleFunc("/renew", RenewConnection)
	http.HandleFunc("/receive/regular", SendRegularData)
	http.HandleFunc("/receive/custom/", SendCustomData)
	http.ListenAndServe(":"+port, nil)
}