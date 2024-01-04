package main

import (
	"crypto/tls"
	//"flag"
	"fmt"
	"log"
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)


func ConnectMqtt(mqttaddress string, mqttuser string, mqttpwd string) MQTT.Client {
	//MQTT.DEBUG = log.New(os.Stdout, "", 0)
	//MQTT.ERROR = log.New(os.Stdout, "", 0)
	//hostname, _ := os.Hostname()
	//topic := flag.String("topic", "alarm/", "Topic to publish the messages on")
	//qos := flag.Int("qos", 0, "The QoS to send the messages at")
	//retained := flag.Bool("retained", false, "Are the messages sent with the retained flag")
	//clientid := flag.String("clientid", "utcar", "A clientid for the connection")
	//flag.Parse()
	
	connOpts := MQTT.NewClientOptions().AddBroker(mqttaddress).SetClientID("utcar").SetCleanSession(true)
	connOpts.SetUsername(mqttuser)
	connOpts.SetPassword(mqttpwd)
	
	tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
	connOpts.SetTLSConfig(tlsConfig)

	client := MQTT.NewClient(connOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
		return nil
	}
	log.Printf("Connected to %s\n", mqttaddress)
	return client
}

func PublishMqtt(client MQTT.Client, sia SIA) error {
	var body string
	switch sia.command {
	//information pour les zones
	case "UA":
		body = "ON"
		itemUrl := strings.Join([]string{"alarm/zone_", sia.zone, "/state"}, "")
		client.Publish(itemUrl, 0, false, body)
		log.Printf("Publish %s to %s", body, itemUrl)
	case "UR":
		body = "OFF"
		itemUrl := strings.Join([]string{"alarm/zone_", sia.zone, "/state"}, "")
		client.Publish(itemUrl, 0, false, body)
		log.Printf("Publish %s to %s", body, itemUrl)
	case "BA":
		body = "ALARME_ACTIVE"
		itemUrl := strings.Join([]string{"alarm/zone_", sia.zone, "/state"}, "")
		client.Publish(itemUrl, 0, false, body)
		log.Printf("Publish %s to %s", body, itemUrl)
	case "BR":
		body = "ALARME_RESTAURE"
		itemUrl := strings.Join([]string{"alarm/zone_", sia.zone, "/state"}, "")
		client.Publish(itemUrl, 0, false, body)
		log.Printf("Publish %s to %s", body, itemUrl)
	case "BB":
		body = "EXCLUE"
		itemUrl := strings.Join([]string{"alarm/zone_", sia.zone, "/state"}, "")
		client.Publish(itemUrl, 0, false, body)
		log.Printf("Publish %s to %s", body, itemUrl)
	case "BU":
	    	body = "INCLUE"
		itemUrl := strings.Join([]string{"alarm/zone_", sia.zone, "/state"}, "")
		client.Publish(itemUrl, 0, false, body)
		log.Printf("Publish %s to %s", body, itemUrl)
	
	//information pour les groupes	
	case "CL":
		body = "ARMEE"
		itemUrl := strings.Join([]string{"alarm/groupe_", sia.account, "/state"}, "")
		client.Publish(itemUrl, 0, false, body)
		log.Printf("Publish %s to %s", body, itemUrl)
	case "OP":
		body = "DESARME"
		itemUrl := strings.Join([]string{"alarm/groupe_", sia.account, "/state"}, "")
       		client.Publish(itemUrl, 0, false, body)
        	log.Printf("Publish %s to %s", body, itemUrl)
   	case "OR":
		body = "DESARME"
		itemUrl := strings.Join([]string{"alarm/groupe_", sia.account, "/state"}, "")
        	client.Publish(itemUrl, 0, false, body)
        	log.Printf("Publish %s to %s", body, itemUrl)
	case "CG":
		body = "PARTIEL"
		itemUrl := strings.Join([]string{"alarm/groupe_", sia.account, "/state"}, "")
        	client.Publish(itemUrl, 0, false, body)
        	log.Printf("Publish %s to %s", body, itemUrl)
	
	//information pour les CS
	case "RP":
		body = "TEST_AUTO"
		itemUrl := strings.Join([]string{"alarm/CS", sia.account, "/state"}, "")
        	client.Publish(itemUrl, 0, false, body)
        	log.Printf("Publish %s to %s", body, itemUrl)
	case "NR":
		body = "CS_OK"
		itemUrl := strings.Join([]string{"alarm/CS", sia.account, "/state"}, "")
        	client.Publish(itemUrl, 0, false, body)
        	log.Printf("Publish %s to %s", body, itemUrl)
    	case "NC":
		body = "CS_KO"
		itemUrl := strings.Join([]string{"alarm/CS", sia.account, "/state"}, "")
        	client.Publish(itemUrl, 0, false, body)
	        log.Printf("Publish %s to %s", body, itemUrl)
		
	default:
	    itemUrl := strings.Join([]string{"alarm/erreur" , "/state"}, "")
	    client.Publish(itemUrl, 0, false, sia.command)
	    log.Printf("Publish %s to %s", sia.command, itemUrl)
		return fmt.Errorf("Unsupported SIA command for pusher (%s)\n", sia.command)
	}
	return nil
}
