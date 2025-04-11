package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var token string

func main() {
	for {
		fmt.Println("\n1. Register\n2. Login\n3. Check-in\n4. View Check-ins\n5. Exit")
		fmt.Print("Choose: ")
		var choice int
		fmt.Scan(&choice)

		switch choice {
		case 1:
			register()
		case 2:
			login()
		case 3:
			checkin()
		case 4:
			viewCheckins()
		case 5:
			os.Exit(0)
		}
	}
}

func register() {
	user := inputUser()
	jsonData, _ := json.Marshal(user)
	resp, _ := http.Post("http://localhost:8080/register", "application/json", bytes.NewBuffer(jsonData))
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func login() {
	user := inputUser()
	jsonData, _ := json.Marshal(user)
	resp, _ := http.Post("http://localhost:8080/login", "application/json", bytes.NewBuffer(jsonData))
	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)
	if t, ok := result["token"]; ok {
		token = t
		fmt.Println("Login successful.")
	} else {
		fmt.Println(string(body))
	}
}

func checkin() {
	req, _ := http.NewRequest("POST", "http://localhost:8080/checkin", nil)
	req.Header.Set("Authorization", token)
	client := &http.Client{}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func viewCheckins() {
	req, _ := http.NewRequest("GET", "http://localhost:8080/checkins", nil)
	req.Header.Set("Authorization", token)
	client := &http.Client{}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func inputUser() User {
	var u User
	fmt.Print("Username: ")
	fmt.Scan(&u.Username)
	fmt.Print("Password: ")
	fmt.Scan(&u.Password)
	return u
}
