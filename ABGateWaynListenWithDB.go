package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/go-sql-driver/mysql"
	"github.com/vmihailenco/msgpack"
)

type Dev struct {
	Temperature  int
	Humidity     string
	Light        string
	 
}

func onMessageReceived(client MQTT.Client, message MQTT.Message) {
	//fmt.Printf("Received message on topic: %s\nMessage: %s\n", message.Topic(), message.Payload())
	fmt.Println(message.Topic())
	//fmt.Println(message.Payload())
	splitString := strings.Split(message.Topic(), "/") //Split by '/'
	//var out map[string]interface{}
	var out map[string]interface{}
	err := msgpack.Unmarshal(message.Payload(), &out)
	if err != nil {
		fmt.Printf("mesg felas")
		panic(err)
	}
	fmt.Println("------------Gate way inf --------------")
	fmt.Println("v =", out["v"])
	fmt.Println("mid =", out["mid"])
	fmt.Println("time =", out["time"])
	fmt.Println("ip =", out["ip"])
	fmt.Println("mac =", out["mac"])
	//////////////////////////////////////
	db, err := sql.Open("mysql", "mqtt:mqtt123@tcp(localhost)/mqtt") //[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	if err != nil {
		fmt.Printf("mqsql Connected felas")
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	} else {
		result, err := db.Exec(
			"INSERT INTO GateWayMesginf(GateV, GateMID,  GateTime, GateIP, GateMac,Incoming_at) VALUES (?,?,?,?,?,?)",
			out["v"],
			out["mid"],
			out["time"],
			out["ip"],
			out["mac"],
			time.Now(),
		)
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Println(result)
		}

	}
	defer db.Close()
	////////////////////////////////////////
	var devices []string
	//var dev [][]string
	for k, v := range out {
		switch value := v.(type) {
		case []interface{}: //interface slice
			fmt.Sprint(k)
			for i, u := range value { //對格式為interface slice的資料做 rang處理
				fmt.Sprint(u)
				devices = append(devices, fmt.Sprintf("%X", value[i])) //將切片value[i]轉換成HAX字串後加入devices切片中
			}
		}
	}
	for k, v := range devices {

		if d := v[2:14]; d == "F4E9E6551BBC" || d == "DFBABEDA2FAF" {
			if l := len(v); l == 68 { //鎖定的裝置MAC
				fmt.Sprint(k)
				fmt.Printf("--------------Device[%s]--------------\n", v[2:14])
				fmt.Printf("UUID = %s\n", v[26:30])
				fmt.Printf("MAC = %s\n", v[2:14])
				fmt.Printf("BatteryLevel = %d\n", toInt(v[54:56]))
				fmt.Printf("Temperature = %d\n", toInt(linkit(v[58:60], v[56:58]))/8)
				fmt.Printf("Humidity = %d\n", toInt(linkit(v[62:64], v[60:62]))/4)
				fmt.Printf("Light = %d\n", toInt(linkit(v[66:68], v[64:66])))
				db, err := sql.Open("mysql", "mqtt:mqtt123@tcp(localhost)/mqtt") //[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
				if err != nil {
					fmt.Printf("mqsql Connected felas")
					panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
				} else {
					result, err := db.Exec(
						"REPLACE INTO AdvData(UUID,Mac, BatteryLv,  Temperature, Humidity, Light,Created_MID,GateWay) VALUES (?,?,?,?,?,?,?,?)",
						v[26:30],
						v[2:14],
						toInt(v[54:56]),
						toInt(linkit(v[58:60], v[56:58]))/8,
						toInt(linkit(v[62:64], v[60:62]))/4,
						toInt(linkit(v[66:68], v[64:66])),
						out["mid"],
						splitString[2],
					)
					if err != nil {
						log.Fatal(err)
					} else {
						fmt.Println(result)
					}
				}
				defer db.Close()
				///////////////////////////////////////////////////
			} else {
				fmt.Println("data err")
			}
		}
	}

}
func toInt(Hax string) int64 {
	intdata, err := strconv.ParseInt(Hax, 16, 64)
	if err != nil {
		fmt.Printf("flase")
	}

	return intdata
}
func linkit(params ...interface{}) string {
	var paramSlice []string
	for _, param := range params {
		paramSlice = append(paramSlice, param.(string))
	}
	aa := strings.Join(paramSlice, "") // Join 方法第2个参数是 string 而不是 rune
	return aa
}

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	//mqtt.DEBUG = log.New(os.Stdout, "", 0) //顯示debug訊息
	//mqtt.ERROR = log.New(os.Stdout, "", 0)//顯錯誤訊息
	hostname, _ := os.Hostname()
	server := "34.217.228.207:1883"
	topic := "/gw/+/status"
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
