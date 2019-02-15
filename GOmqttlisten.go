package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"syscall"
	"time"

	"os"
	"os/signal"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/go-sql-driver/mysql"
)

var f MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func onMessageReceived(client MQTT.Client, message MQTT.Message) {
	//fmt.Printf("Received message on topic: %s\nMessage: %s\n", message.Topic(), message.Payload())
	fmt.Printf("%s\n", message.Topic())
	fmt.Printf("%s\n", message.Payload())
	splitString := strings.Split(message.Topic(), "/") //Split by '/'
	meg := map[string]interface{}{}
	err := json.Unmarshal(message.Payload(), &meg)
	if err != nil {
		panic(err)
	}
	fmt.Println("local: ", meg["local"])
	fmt.Println("blood: ", meg["blood"])
	db, err := sql.Open("mysql", "mqtt:mqtt123@tcp(localhost)/mqtt") //[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	if err != nil {
		fmt.Printf("mqsql Connected felas")
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	} else {
		result, err := db.Exec(
			"INSERT INTO log(local, blood, created_at,  updated_at, topic, a1) VALUES (?,?,?,?,?,?)",
			meg["local"],
			meg["blood"],
			time.Now(),
			nil,
			splitString[3],
			splitString[1],
		)
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Println("Insert complete")
		}

	}
	defer db.Close()
}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	//mqtt.DEBUG = log.New(os.Stdout, "", 0) //顯示debug訊息
	//mqtt.ERROR = log.New(os.Stdout, "", 0)//顯錯誤訊息
	
	hostname, _ := os.Hostname()
	server := flag.String("server", "localhost:1883", "The full url of the MQTT server to connect to ex: tcp://127.0.0.1:1883")
	topic := flag.String("topic", "channle/+/Mqtt/+", "Topic to subscribe to")
	clientid := flag.String("clientid", hostname+strconv.Itoa(time.Now().Second()), "A clientid for the connection")
	connOpts := MQTT.NewClientOptions().AddBroker(*server).SetClientID(*clientid).SetCleanSession(true)

	connOpts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe(*topic, 0, onMessageReceived); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}

	client := MQTT.NewClient(connOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("Mqtt Connected felas")
		panic(token.Error())
	} else {
		fmt.Printf("Connected to %s\n", *server)
	}

	//text := fmt.Sprintf("this is msg!")
	//	if token := c.Publish("go-mqtt/sample", 0, false, text); token.Wait() && token.Error() != nil {
	//	panic(token.Error())
	//}

	//	if token := c.Unsubscribe("go-mqtt/sample"); token.Wait() && token.Error() != nil {
	//		fmt.Println(token.Error())
	//		os.Exit(1)
	//	}
	//	c.Disconnect(250)
	<-c
}
