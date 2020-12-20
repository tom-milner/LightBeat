// Package iot is for controlling the IOT devices.
package iot

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var client mqtt.Client

// MQTTConnInfo holds all the info needed to connect to an MQTT broker
type MQTTConnInfo struct {
	Username string
	Password string
	ClientID string
	Broker   MQTTBroker
}

// MQTTBroker holds all the information about a single broker
type MQTTBroker struct {
	Address string
	Port    string
}

// Handle any default messages.
func messagePubHandler(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())

}

// Function called when connected
func onConnectHandler(client mqtt.Client) {
	log.Println("Connected as ", client.OptionsReader())
}

func onConnectionLostHandler(client mqtt.Client, err error) {
	log.Println("Connection lost.", err)
}

// ConnectToMQTTBroker connects to the specified MQTT broker.
func ConnectToMQTTBroker(info MQTTConnInfo) (mqtt.Client, error) {

	// Add connection settings
	opts := mqtt.NewClientOptions()
	opts.SetClientID(info.ClientID)
	opts.AddBroker("tcp://" + info.Broker.Address + ":" + info.Broker.Port)
	opts.SetClientID(info.ClientID)
	opts.SetUsername(info.Username)
	opts.SetPassword(info.Password)

	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = onConnectHandler
	opts.OnConnectionLost = onConnectionLostHandler

	client = mqtt.NewClient(opts)

	// Connect!
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return client, token.Error()
	}
	return client, nil
}

func SendMessage(topic string, payload interface{}) {
	token := client.Publish(topic, 0, false, payload)
	token.Wait()
}
