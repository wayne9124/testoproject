package main

import (
	"database/sql"
	"encoding/json"
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

type DeviceInfo struct {
	DevMac  string
	Devtype string
}

func onMessageReceived(client MQTT.Client, message MQTT.Message) {
	//fmt.Printf("Received message on topic: %s\nMessage: %s\n", message.Topic(), message.Payload())
	fmt.Println(message.Topic())
	//fmt.Println(message.Payload())
	//	splitString := strings.Split(message.Topic(), "/") //Split by '/'
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

	////////////////////////////////////////
	var devices []string
	//var dev [][]string
	for _, v := range out {
		switch value := v.(type) {
		case []interface{}: //interface slice
			for i, u := range value { //對格式為interface slice的資料做 rang處理
				fmt.Sprint(u)
				devices = append(devices, fmt.Sprintf("%X", value[i])) //將切片value[i]轉換成HAX字串後加入devices切片中
			}
		}
	}
	////////////////////////////////與資料庫對街和藍牙資料長度檢測
	for _, List := range getDeviceList() {
		for _, Data := range devices {
			if List.DevMac == Data[2:14] {
				AdvDataDecodToJson(List.Devtype, Data[16:], List.DevMac)

			}
		}
	}
}

func AdvDataDecodToJson(devType string, data string, DevMac string) { ///針對單獨adv封包分析的工具
	switch devType {
	case "beacon":
		if l := len(data); l == 60 {
			mapD := map[string]string{
				"Major":     data[50:54],
				"Minor":     data[54:58],
				"TX-Power ": data[58:60],
			}
			mapB, _ := json.Marshal(mapD)
			fmt.Println(string(mapB))
			ToDB(string(mapB), DevMac)
		} else {
			fmt.Println("data err")
		}
	case "bluetooth":
		if l := len(data); l == 74 {
			mapD := map[string]int64{
				"Heartrate": toInt(data[26:28]),
				"Footstep":  toInt(data[28:32]),
				"Power":     toInt(data[32:34]),
			}
			mapB, _ := json.Marshal(mapD)
			fmt.Println(string(mapB))
			ToDB(string(mapB), DevMac)
		} else {
			fmt.Println("data err")
		}

	case "sensor":
		if l := len(data); l == 52 {
			mapD := map[string]int64{
				"BatteryLevel": toInt(data[38:40]),
				"Temperature":  toInt(linkit(data[42:44], data[40:42])) / 8,
				"Humidity ":    toInt(linkit(data[46:48], data[44:46])) / 4,
				"Light":        toInt(linkit(data[50:52], data[48:50])),
			}
			mapB, _ := json.Marshal(mapD)
			fmt.Println(string(mapB))
			ToDB(string(mapB), DevMac)
		} else {
			fmt.Println("data err")
		}

	default:

	}

}

func getDeviceList() []DeviceInfo {
	var devList []DeviceInfo
	db, err := sql.Open("mysql", "iots:iots#TECH@tcp(34.217.228.207)/RichSystem_production") //[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	if err != nil {
		fmt.Printf("mqsql Connected felas")
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	} else {
		rows, err := db.Query("SELECT did,device_type FROM devices")
		if err != nil {
			log.Fatal(err)
		}

		//var Mac string
		defer rows.Close()
		var dataSlice DeviceInfo
		for rows.Next() {
			if err := rows.Scan(&dataSlice.DevMac, &dataSlice.Devtype); err != nil {
				log.Fatal(err)
			}
			devList = append(devList, dataSlice)
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}
		for _, u := range devList {
			fmt.Println(u.DevMac, u.Devtype)
		}
	}
	defer db.Close()
	return devList
}

func ToDB(data string, TargetMac string) {
	db, err := sql.Open("mysql", "iots:iots#TECH@tcp(34.217.228.207)/RichSystem_production") //[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	if err != nil {
		fmt.Printf("mqsql Connected felas")
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	} else {

		row := db.QueryRow("SELECT did FROM devices WHERE did = ?", TargetMac)
		var devicetype string
		if err := row.Scan(&devicetype); err != nil {
			fmt.Printf("can't find")
			log.Fatal(err)
		} else {
			result, err := db.Exec(
				"UPDATE devices SET device_data = ?,updated_at = ? WHERE did = ?",
				data,
				time.Now(),
				TargetMac,
			)
			result2, err := db.Exec(
				"INSERT INTO device_logs (device_id, user_id, protocol, protocol_description, device_data, result, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?)",
				1,
				1,
				"mqtt",
				"Script subscribe Mqtt topic",
				data,
				1,
				time.Now(),
				time.Now(),
			)
			if err != nil {
				log.Fatal(err)
			} else {

				fmt.Printf("%s Update success\n", TargetMac)
				fmt.Println("-------------------------")
				fmt.Sprint(result)
				fmt.Sprint(result2)

			}

		}
	}
	defer db.Close()
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

