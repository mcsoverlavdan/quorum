package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"time"
)

//the addr variable has the address to the broker in this case localhost port 8000
var addr = flag.String("addr", "localhost:8000", "http service address")

func main(){
	//the url uses the addr variable we created in line 22 and point to the path /broker

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/broker"}
	log.Printf("connecting to %s", u.String())

	//the dial function connects the publisher to the broker and if returned error dial err is printed

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	//c holds the *websocket.Conn websocket connection variable with which we can read incoming data

	_, message, err := c.ReadMessage()
	if err != nil {
		log.Println("read:", err)
		return
	}
	fmt.Println(message)
	type Message struct {
		Action  string  `json:"action"`
		RequestType   string `json:"request"`
		Message string `json:"message"`
		RequestId string `json:"requestid"`
		DataVersion int `json:"dataversion"`
		TableName string `json:"tablename"`
		TableValue string `json:"tablevalue"`
	}


	msg:="{\"action\":\""+"publish"+"\",\"request\":\""+"write"+"\",\"message\":\"70\",\"tablename\":\""+"completionstable"+"\",\"tablevalue\":\""+"aravind"+"\"}"


	fmt.Println(msg)
	if err = c.WriteMessage(1, []byte(msg)); err != nil {
		fmt.Println(err)
	}

	time.Sleep(5000*time.Millisecond)



	msg="{\"action\":\""+"publish"+"\",\"request\":\""+"read"+"\",\"message\":\"70\",\"tablename\":\""+"completionstable"+"\",\"tablevalue\":\""+"aravind"+"\"}"


	fmt.Println(msg)
	if err = c.WriteMessage(1, []byte(msg)); err != nil {
		fmt.Println(err)
	}

	_, message, err = c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		fmt.Println(string(message))

	time.Sleep(5000*time.Millisecond)
	time.Sleep(5000*time.Millisecond)

	msg="{\"action\":\""+"publish"+"\",\"request\":\""+"write"+"\",\"message\":\"89\",\"tablename\":\""+"completionstable"+"\",\"tablevalue\":\""+"krishnan"+"\"}"


	fmt.Println(msg)
	if err = c.WriteMessage(1, []byte(msg)); err != nil {
		fmt.Println(err)
	}

	time.Sleep(5000*time.Millisecond)



	msg="{\"action\":\""+"publish"+"\",\"request\":\""+"read"+"\",\"message\":\"89\",\"tablename\":\""+"completionstable"+"\",\"tablevalue\":\""+"krishnan"+"\"}"


	fmt.Println(msg)
	if err = c.WriteMessage(1, []byte(msg)); err != nil {
		fmt.Println(err)
	}

	for {

		_, newmessage, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		fmt.Println(string(newmessage))
	}





}
