package mqtt

import (
	"fmt"
	"os"
	"os/signal"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func Subscribe(broker, topic string, onReceiveFn mqtt.MessageHandler) error {
	opts := mqtt.NewClientOptions().AddBroker(broker).SetClientID("mqtt_subscriber")

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	if token := client.Subscribe(topic, 0, onReceiveFn); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	fmt.Printf("Subscribed to topic: %s\n", topic)

	// Handle interrupt signal to gracefully shut down the subscriber
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Wait for interrupt signal
	<-interrupt

	fmt.Println("Unsubscribing and disconnecting...")
	if token := client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	client.Disconnect(250) // Wait 250 milliseconds for any pending message to be sent

	fmt.Println("Subscriber stopped.")
	return nil
}
