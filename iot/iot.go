// Package iot is for controlling the IOT devices.
package iot

import (
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/tom-milner/LightBeatGateway/iot/topics"
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

type IOTMessage mqtt.Message
type MessageHandler func(IOTMessage)

var messageHandlers map[topics.TopicName]MessageHandler

// Handle any default messages.
func mqttMessageHandler(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())

	// Call the correct function for the topic.
	messageHandlers[topics.TopicName(msg.Topic())](IOTMessage(msg))
}

func OnReceive(topic topics.TopicName, handler MessageHandler) {
	client.Subscribe(string(topic), 0, mqttMessageHandler)
	messageHandlers[topic] = handler
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

	opts.SetDefaultPublishHandler(mqttMessageHandler)
	opts.OnConnect = onConnectHandler
	opts.OnConnectionLost = onConnectionLostHandler

	client = mqtt.NewClient(opts)

	// Connect!
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return client, token.Error()
	}

	// Every time we connect to a new MQTT broker, we'll need to respecify the topics to subscribe to.
	messageHandlers = map[topics.TopicName]MessageHandler{}

	return client, nil
}

func SendMessage(topic topics.TopicName, payload interface{}) {
	token := client.Publish(string(topic), 0, false, payload)
	token.Wait()
}
