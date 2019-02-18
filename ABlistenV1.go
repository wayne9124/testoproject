package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/go-sql-driver/mysql"
	"github.com/vmihailenco/msgpack"
)

type Key struct {
	v       string // firmware version
	mid     string // message ID
	time    string // boot time
	ip      string // the IP for gateway
	mac     string // the mac address for gateway
	devices string // an array for BLE advertising datas that gateway collected
}

func onMessageReceived(client MQTT.Client, message MQTT.Message) {
	//fmt.Printf("Received message on topic: %s\nMessage: %s\n", message.Topic(), message.Payload())
	fmt.Println(message.Topic())
	//fmt.Println(message.Payload())
	//splitString := strings.Split(message.Topic(), "/") //Split by '/'
	//var out map[string]interface{}
	var out map[string]interface{}
	err := msgpack.Unmarshal(message.Payload(), &out)
	if err != nil {
		fmt.Printf("mesg felas")
		panic(err)
	}
	//fmt.Println(out)
	fmt.Println("v =", out["v"])
	fmt.Println("mid =", out["mid"])
	fmt.Println("time =", out["time"])
	fmt.Println("ip =", out["ip"])
	fmt.Println("mac =", out["mac"])
	var devices []string
	//strs = fmt.Sprint(out["devices"])
	//fmt.Println(strs[4])
	for k, v := range out {
		switch value := v.(type) {
		case []interface{}:
			fmt.Sprint(k)
			//fmt.Println(k, "is an array:",v)
			for i, u := range value {
				var strs = make([]string, len(value))

				fmt.Sprint(i, "=", u)
				for n := range value {
					//fmt.Println("------------")
					strs[n] = fmt.Sprintf("%X", value[n])

					//fmt.Printf("strs[%d] = %s\n", n, strs[n])
				}
				devices = strs
			}

		}
	}

	for k, v := range devices {
		fmt.Println(k, v[:2], v[2:14], v[14:16], v[16:18], v[18:])

	}

}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	//mqtt.DEBUG = log.New(os.Stdout, "", 0) //顯示debug訊息
	//mqtt.ERROR = log.New(os.Stdout, "", 0)//顯錯誤訊息
	hostname, _ := os.Hostname()
	server := "34.217.228.207:1883"
	topic := "/gw/april_brother/status"
	clientid := hostname + strconv.Itoa(time.Now().Second())
	connOpts := MQTT.NewClientOptions().AddBroker(server).SetClientID(clientid).SetCleanSession(true)

	connOpts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe(topic, 0, onMessageReceived); token.Wait() && token.Error() != nil {

			panic(token.Error())

		}

	}

	client := MQTT.NewClient(connOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("Mqtt Connected felas")
		panic(token.Error())
	} else {
		fmt.Printf("Connected to %s\n", server)
	}

	<-c
}
