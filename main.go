package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
)

type InfluxDB_alert_msg struct {

	// Layout of an InfluxDB v2.0 alert notification

	Check_id                   string  `json:"_check_id"`
	Check_name                 string  `json:"_check_name"`
	Level                      string  `json:"_level"`
	Measurement                string  `json:"_measurement"`
	Message                    string  `json:"_message"`
	Notification_endpoint_id   string  `json:"_notification_endpoint_id"`
	Notification_endpoint_name string  `json:"_notification_endpoint_name"`
	Notification_rule_id       string  `json:"_notification_rule_id"`
	Notification_rule_name     string  `json:"_notification_rule_name"`
	Source_measurement         string  `json:"_source_measurement"`
	Source_timestamp           float64 `json:"_source_timestamp"`
	Start                      string  `json:"_start"`
	Status_timestamp           float64 `json:"_status_timestamp"`
	Stop                       string  `json:"_stop"`
	Time                       string  `json:"_time"`
	Type                       string  `json:"_type"`
	Version                    float64 `json:"_version"`
	Alert                      bool    `json:"alert"`
}

func handleWebhook(w http.ResponseWriter, r *http.Request, mycmd *string, debug_flg *bool) {

	if mycmd == nil {
		http.Error(w, "invalid alert script name", http.StatusBadRequest)
		return
	}

	if *debug_flg == true {
		fmt.Printf("\n")
		fmt.Printf("headers: %v\n\n", r.Header)
	}

	var influx_alert InfluxDB_alert_msg

	err := json.NewDecoder(r.Body).Decode(&influx_alert)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if *debug_flg == true {
		fmt.Printf("Influx buffer: %+v\n", influx_alert)
		fmt.Printf("Received alert from InfluxDB, message is: %s.\n", influx_alert.Message)
		fmt.Printf("\n")
	}

	if *mycmd != "#NOTSET" {

		if *debug_flg == true {
			fmt.Printf("Sending to Zabbix \n")
		}

		arg0 := influx_alert.Message
		cmd := exec.Command(*mycmd, arg0)

		var outm bytes.Buffer
		var errm bytes.Buffer
		cmd.Stdout = &outm
		cmd.Stderr = &errm

		err1 := cmd.Run()

		if err1 != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if *debug_flg == true {
			fmt.Printf("Alert send stdout: %q\n", outm.String())
			fmt.Printf("Alert send stderr: %q\n", errm.String())
		}
	}

}

func main() {

	// parse cmd parameters
	myPortPtr := flag.String("port", ":55556", "webhook listening port, default :55556")
	alertSendCmd := flag.String("alert_sender", "#NOTSET", "Alert sender script")
	debugFlag := flag.Bool("debug", false, "Display debug messages")

	flag.Parse()

	if *debugFlag == true {
		fmt.Println("my port:", *myPortPtr)
		fmt.Println("Alert sender script:", *alertSendCmd)
		fmt.Println("tail:", flag.Args())
	}

	log.Println("Zabbix webhook server started !")

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		handleWebhook(w, r, alertSendCmd, debugFlag)
	})

	err := http.ListenAndServe(*myPortPtr, nil)
	if err != nil {
		log.Fatal("HTTP ListenAndServe: ", err)
	}

}
