package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"log"
	"net/http"
	"net/url"
	"strconv"
)


var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//the clients connected will be appended to the clients slice and the subscribers to the subscription slice
type PubSub struct {
	Clients       []Client
	Subscriptions []Subscription
	Publisher []PublisherClinet
}
type Client struct {
	Id         string
	Connection *websocket.Conn
}

type Subscription struct {
	Client Client
	portNumber string
}
type PublisherClinet struct{
	Client Client
	requestId string
	Data []DataVersion
}

type DataVersion struct{
	DataFrom Subscription
	DataReceived string
	Version string

}
//autoid generates a unique id to differentiate clients
func autoId() string {

	return uuid.Must(uuid.NewV4()).String()
}
//addclient method is used to append new c;ient details to the client slice
func (ps *PubSub) AddClient(client Client) *PubSub {

	ps.Clients = append(ps.Clients, client)

	fmt.Println("adding new client to the list", client.Id, len(ps.Clients))

	payload := []byte("Hello Client ID:" +
		client.Id)
	//sent the client id to the connection
	err := client.Connection.WriteMessage(1, payload)
	if err != nil {
		log.Println(err)
	}
	return ps

}
//when connection is cloed remove client from the subscriptions and the client list
func (ps *PubSub) RemoveClient(client Client) (*PubSub) {

	// first remove all subscriptions by this client

	for index, sub := range ps.Subscriptions {

		if client.Id == sub.Client.Id {
			//slice data is deleted using append data till that value and after that value
			ps.Subscriptions = append(ps.Subscriptions[:index], ps.Subscriptions[index+1:]...)
		}
	}

	// remove client from the list

	for index, c := range ps.Clients {

		if c.Id == client.Id {
			ps.Clients = append(ps.Clients[:index], ps.Clients[index+1:]...)
		}

	}

	return ps
}

//a variable ps of tpe pub sub is created
var ps PubSub
//----------------------for passing data to channels creating struct (start).....
type QueueData struct{
	clinetdetails Client
	payloadmessage []byte
}
type DataPasser struct {
	queue chan QueueData
}

//this funtion is used for websocket connections
func (p *DataPasser) ConnectNodes(w http.ResponseWriter, r *http.Request) {
	//upgrader checkorgin is returned true so that any websocket connections arent refsed
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
	fmt.Println("thissssssss issssssss theeee conn valueeee.....\n",*conn)
	//client is then added to the client list using Addclinet function
	ps.AddClient(ClientVal)

	//reading incoming data from the client in and infinte loop until the connection from the client is closed
	for {
		_, payload, err := conn.ReadMessage()
		fmt.Println("payload msg",payload)
		//the data where payload is the messafe and client val is the client details are passed to the channel queue
		p.queue<-QueueData{ClientVal,payload}


		if err != nil {
			log.Println(err)
			ps.RemoveClient(ClientVal)
			fmt.Println("after removing client list", len(ps.Clients))
			return
		}
	}


}

//......................... struct  to unmarshal JSON tupe----------------
type Message struct {
	Action  string  `json:"action"`
	RequestType   string `json:"request"`
	Message string `json:"message"`
	RequestId string `json:"requestid"`
	DataVersion string `json:"dataversion"`
	TableName string `json:"tablename"`
	TableValue string `json:"tablevalue"`
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
			fmt.Println("This is not correct message payload",err,"payload : ",item.payloadmessage)
		}
		fmt.Println("Client Id", item.clinetdetails.Id)
		fmt.Println("Messsage : ",m.Message,m.Action)

		//for action mentioned in the action parameter publish and subscribe methods are done
		if m.Action =="publish"{
			PublisherValue := PublisherClinet{
				Client: item.clinetdetails,
				requestId:         autoId(),
			}
			ps.Publisher=append(ps.Publisher,PublisherValue )

			//sending reuest data to subscribers..................
			msg:="{\"request\":\""+m.RequestType+"\",\"requestid\":\""+PublisherValue.requestId+"\",\"message\":\""+m.Message+"\",\"tablename\":\""+m.TableName+"\",\"tablevalue\":\""+m.TableValue+"\"}"
			print(msg)
			for _, subscription := range ps.Subscriptions {
				fmt.Printf("Sending to client id %s message is %s \n", subscription.Client.Id, m.Message)
				//sub.Client.Connection.WriteMessage(1, message)

				//------------------------------------------

				//the url uses the controller addr variable

				//the dial function connects the publisher to the broker and if returned error dial err is printed
				tempaddress:= "localhost:"+subscription.portNumber
				u := url.URL{
					Scheme: "ws",
					Host:   tempaddress,
					Path:   "/ws",
				}



				c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
				if err != nil {
					log.Fatal("dial:", err)
				}
				fmt.Println(msg)
				if err = c.WriteMessage(1, []byte(msg)); err != nil {
					fmt.Println(err)
				}
				c.Close()






				//-----------------------------------------------------?
				//
				//subscription.Client.Connection.WriteMessage(1, []byte(msg))
				//
				//if err !=nil{
				//	fmt.Println("not able to send ",err)
				//}
				//
				//
				}





			//subscribstions----------------------------------------------------

		}else if m.Action =="subscribe"{
			fmt.Println("inside subscribe")
			var subscriptions  []Subscription
			//check if subscription already exixsts for the same client
			for _, subscription := range ps.Subscriptions {
				if subscription.Client.Id == item.clinetdetails.Id  {
					subscriptions = append(subscriptions, subscription)

				}
			}
			if len(subscriptions) > 0 {

				// client is subscribed

				continue
			}

			newSubscription := Subscription{
				Client: item.clinetdetails,
				portNumber:m.Message,
			}
			//append the client to the subscritption list
			ps.Subscriptions = append(ps.Subscriptions, newSubscription)
		}else if m.Action =="readquorum"{
			log.Println("inside write quorum................")
			var subscriptionValue Subscription
			subscriptionValue.Client=item.clinetdetails

			dataValue := DataVersion{
				DataFrom:     subscriptionValue,
				DataReceived: m.Message,
				Version:      m.DataVersion,
			}
			for index, data := range ps.Publisher {
				if data.requestId== m.RequestId {
					ps.Publisher[index].Data=append(ps.Publisher[index].Data, dataValue)
					if len(ps.Publisher[index].Data)>=(int(len(ps.Subscriptions))/2+1){
						maximumVersion:=0
						version:=0
						for _,dataversion :=range ps.Publisher[index].Data{
							version, _ =strconv.Atoi(dataversion.Version)
							if version>maximumVersion{
								maximumVersion=version
									}
								}
						for _,dataversion :=range ps.Publisher[index].Data{
							version, _ =strconv.Atoi(dataversion.Version)

							if version==maximumVersion{

								msg:=dataversion.DataReceived
								fmt.Println("sending messgae back to server.....",msg)
								ps.Publisher[index].Client.Connection.WriteMessage(1, []byte(msg))

								if err !=nil{
									fmt.Println("not able to send ",err)
									}
								break


							}
						}

						break
					}
					break
				}

			}



//testing one


///write ack sould be received
		}else if m.Action=="writequorum"{
			log.Println("inside write quorum................")
			log.Println(ps.Publisher)
			var subscriptionValue Subscription
			subscriptionValue.Client=item.clinetdetails

			dataValue := DataVersion{
				DataFrom:     subscriptionValue,
				DataReceived: m.Message,
				Version:      m.DataVersion,
			}
			for index, data := range ps.Publisher {
				fmt.Println("inside write quorum first for................",data)
				if data.requestId== m.RequestId {
					ps.Publisher[index].Data=append(ps.Publisher[index].Data, dataValue)
					fmt.Println("inside write calculation................",int(len(ps.Subscriptions)/2)+1)
					fmt.Println("put write condition................",len(ps.Publisher[index].Data))
					if len(ps.Publisher[index].Data)>=(int(len(ps.Subscriptions)/2)+1){

						fmt.Println("inside write condition................",len(ps.Publisher[index].Data))
						maximumack:=0
						Failed:=0
						var FaliedSubscribers []Subscription
						for _,dataversion :=range ps.Publisher[index].Data{
							if dataversion.DataReceived=="ack"{
								maximumack+=1
							}
							if dataversion.DataReceived=="fail"{
								Failed+=1
								FaliedSubscribers=append(FaliedSubscribers, dataversion.DataFrom)
							}
						}

						fmt.Println("the write was sucessfull")
						fmt.Println("failed ",Failed,"list ",FaliedSubscribers)

					}


					}
					break
				}


			}

			
		} 


	}






func main(){
	//the channel queue is used to pass data to the log funtion when new messages are recieved from the client
	queue:= make(chan QueueData)
	passer := &DataPasser{queue: queue}
	//log funtion sorts the request to publish and subscribe
	go passer.log()




	//connect nodes established the websocket connection between clients
	http.HandleFunc("/broker",passer.ConnectNodes)
	err:=http.ListenAndServe(":8000",nil)

	if err !=nil{
		fmt.Println("error in listen and serve",err)
	}
}