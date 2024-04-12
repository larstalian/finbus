package mqtt

import (
	"finbus/internal/models"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"strings"
)

type BusDataSubscriber interface {
	SubscribeToTopic(topic string) error
	mqttMessageHandler(client mqtt.Client, msg mqtt.Message)
	ListenToAllTopics()
}

// busDataSubscriber is an MQTT client that subscribes to a specific topic and sends the data to a channel
type busDataSubscriber struct {
	client      mqtt.Client
	dataChannel chan models.BusData
}

// NewBusDataSubscriber creates a new busDataSubscriber and connects to the MQTT broker
func NewBusDataSubscriber(mqttBroker string, dataChannel chan models.BusData) (BusDataSubscriber, error) {
	opts := mqtt.NewClientOptions().AddBroker(mqttBroker).SetClientID("go_mqtt_client").SetAutoReconnect(true)

	opts.OnConnect = func(c mqtt.Client) {
		fmt.Println("Connected to MQTT broker")
	}

	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		fmt.Printf("Connection lost: %v. Reconnecting...\n", err)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("error connecting to MQTT broker: %v", token.Error())
	}

	return &busDataSubscriber{client: client, dataChannel: dataChannel}, nil
}

// mqttMessageHandler handles incoming MQTT messages and sends the data to the data channel
func (m *busDataSubscriber) mqttMessageHandler(_ mqtt.Client, msg mqtt.Message) {
	busData := parseTopic(msg.Topic())
	m.dataChannel <- busData
}

// parseTopic parses the MQTT topic and returns a BusData struct
func parseTopic(topic string) models.BusData {
	parts := strings.Split(topic, "/")
	return models.BusData{
		FeedFormat:       parts[1],
		Type:             parts[2],
		FeedID:           parts[3],
		AgencyID:         parts[4],
		AgencyName:       parts[5],
		Mode:             parts[6],
		RouteID:          parts[7],
		DirectionID:      parts[8],
		TripHeadsign:     parts[9],
		TripID:           parts[10],
		NextStop:         parts[11],
		StartTime:        parts[12],
		VehicleID:        parts[13],
		GeohashHead:      parts[14],
		GeohashFirstDeg:  parts[15],
		GeohashSecondDeg: parts[16],
		GeohashThirdDeg:  parts[17],
		ShortName:        parts[18],
		Color:            parts[19],
	}
}

// SubscribeToTopic subscribes to a specific MQTT topic
func (m *busDataSubscriber) SubscribeToTopic(topic string) error {
	if token := m.client.Subscribe(topic, 0, m.mqttMessageHandler); token.Wait() && token.Error() != nil {
		return fmt.Errorf("error subscribing to topic %s: %v", topic, token.Error())
	}
	fmt.Printf("Subscribed to topic: %s\n", topic)
	return nil
}

// ListenToAllTopics subscribes to all topics
func (m *busDataSubscriber) ListenToAllTopics() {
	if token := m.client.Subscribe("#", 0, m.mqttMessageHandler); token.Wait() && token.Error() != nil {
		fmt.Printf("Error subscribing to all topics: %v\n", token.Error())
	}
}

var _ BusDataSubscriber = (*busDataSubscriber)(nil)
