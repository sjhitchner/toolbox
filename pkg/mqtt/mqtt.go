package mqtt

import (
	"fmt"

	"github.com/sjhitchner/toolbox/pkg/metrics"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTConfig struct {
	Broker, Topic, ClientID string
}

// Creates New MQTT Client
// Remember to add the defer below to make sure things are gracefully shutdown
// defer client.Disconnect(250)
func New(broker, clientID string) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions().AddBroker(broker).SetClientID(clientID)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return client, nil
}

type Unmarshalable interface {
	UnmarshalMQTT([]byte) error
}

type Marshalable interface {
	MarshalMQTT() ([]byte, error)
}

func Subscribe[T Unmarshalable](done <-chan struct{}, client mqtt.Client, topic string) (<-chan T, <-chan error) {
	out := make(chan T)
	errCh := make(chan error)

	go func() {
		defer close(out)
		defer close(errCh)

		token := client.Subscribe(topic, 0, func(client mqtt.Client, raw mqtt.Message) {
			defer metrics.CounterAt("mqtt.receive", 1, "topic", topic).Emit()
			errCount := metrics.CounterAt("mqtt.receive_error", 1, "topic", topic)
			defer errCount.Emit()

			// TODO Metrics
			var msg T
			msg = *new(T)
			if err := msg.UnmarshalMQTT(raw.Payload()); err != nil {
				errCount.Incr()
				errCh <- err
				return
			}

			out <- msg
		})
		if token.Wait() && token.Error() != nil {
			errCh <- token.Error()
			return
		}

		fmt.Printf("Subscribed to topic: %s\n", topic)
		<-done

		fmt.Println("Unsubscribing and disconnecting...")
		if token := client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
			errCh <- token.Error()
			return
		}
	}()

	return out, errCh
}

func Publish[T Marshalable](done <-chan struct{}, client mqtt.Client, topic string, inCh <-chan T) <-chan error {
	errCh := make(chan error)

	go func() {
		defer close(errCh)

		for {
			select {
			case msg := <-inCh:
				if err := publish(client, topic, msg); err != nil {
					errCh <- err
					continue
				}

			case <-done:
				return
			}
		}
	}()
	return errCh
}

func publish[T Marshalable](client mqtt.Client, topic string, msg T) error {
	defer metrics.CounterAt("mqtt.send", 1, "topic", topic).Emit()
	errCount := metrics.CounterAt("mqtt.send_error", 1, "topic", topic)
	defer errCount.Emit()

	var qos byte
	var retain bool

	payload, err := msg.MarshalMQTT()
	if err != nil {
		errCount.Incr()
		return err
	}

	token := client.Publish(topic, qos, retain, payload)

	if token.Wait() && token.Error() != nil {
		errCount.Incr()
		return fmt.Errorf("Failed to publish to topic %s: %v\n", topic, token.Error())
	}

	return nil
}

/*
func main() {
	broker := "tcp://broker.emqx.io:1883" // Replace with your broker
	topic := "test/topic"

	// MQTT Client options
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID("golang-client")
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalln("Error connecting to MQTT broker:", token.Error())
	}
	defer client.Disconnect(250)

	// Channel for messages
	msgChannel := make(chan string)

	// Subscribe using the channel
	go subscribeWithChannel(client, topic, msgChannel)

	// Publish some test messages
	go func() {
		for i := 0; i < 5; i++ {
			publish(client, topic, fmt.Sprintf("Message %d", i))
			time.Sleep(1 * time.Second)
		}
	}()

	// Read messages from the channel
	for i := 0; i < 5; i++ {
		msg := <-msgChannel
		fmt.Println("Received:", msg)
	}
}



func SubscribeFn(client mqtt.Client, topic string, onReceiveFn mqtt.MessageHandler) error {

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

*/
