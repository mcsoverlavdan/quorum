package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)


//A server application calls the Upgrader.Upgrade method from an HTTP request handler to get a *Conn:

var Controller = flag.String("Controller", "localhost:8000", "http service address")


var Leader string

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//the client id which is auto generated id using autoid function and the connection variable stores the connection value
type Client struct {
	Id         string
	Connection *websocket.Conn
}
//the main Leader client value

//autoid generates a unique id to differentiate
func autoId() string {

	return uuid.Must(uuid.NewV4()).String()
}
type CompletionsTable struct{
	UserNameValue []UserName

}
type UserName struct{
	UniqueUserName string
	Completions string
	Status string
	Version int
}
//----------------------for passing data to channels creating struct (start).....
type QueueData struct{
	clinetdetails Client
	payloadmessage []byte
}
type DataPasser struct {
	queue chan QueueData
}

//this Function is used for websocket connections
func (p *DataPasser) ConnectNodes(w http.ResponseWriter, r *http.Request) {
	//Upgrader check orgin is returned true so that any websocket connections arent refsed
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true

	}

	fmt.Println("starting connec.......")
	//The Conn type represents a WebSocket connection.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	//the client details are added to the client Val with struct Client
	ClientVal := Client{
		Id:         autoId(),
		Connection: conn,
	}


	//reading incoming data from the client in and infinte loop until the connection from the client is closed
	for {
		_, payload, err := conn.ReadMessage()
		fmt.Println("payload msg",payload)
		//the data where payload is the messafe and client val is the client details are passed to the channel queue
		p.queue<-QueueData{ClientVal,payload}


		if err != nil {
			log.Println(err)
			return
		}
	}


}
//......................... struct  to unmarshal JSON tupe----------------
type Message struct {
	Request  string    `json:"request"`
	RequestId string   `json:"requestid"`
	MessageValue string `json:"message"`
	Table  string        `json:"tablename"`
	TableValue string   `json:"tablevalue"`
}


//the log method is run as a go routine and check for data in queue channel and does the publish subscribe methods accordingly
//checking the message sent by the client
func (p *DataPasser) log() {
	//for range in the channel queue
	for item := range p.queue {
		//m of tpe Message used for unmarshalling the JSON
		m := Message{}
		//the message is unmarshaled to m
		err := json.Unmarshal(item.payloadmessage, &m)
		if err != nil {
			fmt.Println("This is not from mesage message payload",err,"payload : ",item.payloadmessage)
		}
		fmt.Println("Client Id", item.clinetdetails.Id)
		fmt.Println("Messsage :",m.Request,m.RequestId,m.MessageValue)

		if m.Request=="read"{
			log.Println("inside read")

			jsonFile, err := os.Open("database4.json")
			byteValue, _ := ioutil.ReadAll(jsonFile)
			comStructData := CompletionsTable{}
			err = json.Unmarshal(byteValue, &comStructData)
			if err != nil {
				fmt.Println("This is not correctbn message payload",err)
			}
			for _,data:= range comStructData.UserNameValue{
				log.Println("inside read data",data)
				log.Println("inside read data",m.MessageValue)
				log.Println("inside read data",data.UniqueUserName)
				if data.UniqueUserName==m.TableValue{
					log.Println("inside if data",data)

					//the url uses the controller addr variable

					u := url.URL{Scheme: "ws", Host: *Controller, Path: "/broker"}
					log.Printf("connecting to %s", u.String())

					//the dial function connects the publisher to the broker and if returned error dial err is printed

					c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
					if err != nil {
						log.Fatal("dial:", err)
					}


					msg:="{\"action\":\""+"readquorum"+"\",\"message\":\""+data.Completions+"\",\"requestid\":\""+m.RequestId+"\",\"dataversion\":\""+strconv.Itoa(data.Version)+"\"}"

					fmt.Println(msg)
					if err = c.WriteMessage(1, []byte(msg)); err != nil {
						fmt.Println(err)
					}

					break

				}
			}










		}else if m.Request=="write"{
			fmt.Println("inside,.....write..........")
			jsonFile, err := os.Open("database4.json")
			byteValue, _ := ioutil.ReadAll(jsonFile)
			comStructData := CompletionsTable{}
			err = json.Unmarshal(byteValue, &comStructData)

			var completedInt int
			var messageToSent string

			if err != nil {
				fmt.Println("This is not correct message payload",err)
			}
			if m.TableValue=="krishnan"{
				messageToSent = "fail"
			}else {
				for index,data:= range comStructData.UserNameValue{
					if data.UniqueUserName==m.TableValue{
						comStructData.UserNameValue[index].Completions=m.MessageValue
						comStructData.UserNameValue[index].Version=comStructData.UserNameValue[index].Version+1
						completedInt,err=strconv.Atoi(m.MessageValue)
						if completedInt>70{
							comStructData.UserNameValue[index].Status="COMPLIANT"
						}
						break
					}
				}


				file, _ := json.MarshalIndent(comStructData, "", " ")

				err = ioutil.WriteFile("database4.json", file, 0644)
				messageToSent = "ack"
				if err != nil {
					log.Fatal("dial:", err)
					messageToSent= "fail"

				}


			}













			//the url uses the controller addr variable

			u := url.URL{Scheme: "ws", Host: *Controller, Path: "/broker"}
			log.Printf("connecting to %s", u.String())

			//the dial function connects the publisher to the broker and if returned error dial err is printed

			c2, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				log.Fatal("dial:", err)
			}


			msg:="{\"action\":\""+"writequorum"+"\",\"message\":\""+messageToSent+"\",\"requestid\":\""+m.RequestId+"\"}"

			fmt.Println(msg)
			if err = c2.WriteMessage(1, []byte(msg)); err != nil {
				fmt.Println(err)
			}

			log.Println("writeeeeeeeeeeee...............")









		}else if m.Request=="kill"{
			os.Exit(0)


		}

		}


	}






func main(){

	//the url uses the controller addr variable

	u := url.URL{Scheme: "ws", Host: *Controller, Path: "/broker"}
	log.Printf("connecting to %s", u.String())

	//the dial function connects the publisher to the broker and if returned error dial err is printed

	c3, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	msg:="{\"action\":\"subscribe\",\"message\":\""+strconv.Itoa(8004)+"\"}"
	fmt.Println(msg)
	if err = c3.WriteMessage(1, []byte(msg)); err != nil {
		fmt.Println(err)
	}

	//the channel queue is used to pass data to the log funtion when new messages are recieved from the client
	queue:= make(chan QueueData)

	passer := &DataPasser{queue: queue}
	//log function sorts the request to publish and subscribe
	go passer.log()

	http.HandleFunc("/ws",passer.ConnectNodes)
	//addr=rand.Intn(10000 - 8000) + 8000
	err=http.ListenAndServe(":8004",nil)

	if err !=nil{
		fmt.Println("error in listen and serve",err)
	}
}
