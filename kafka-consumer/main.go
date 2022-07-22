package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

const agentSocketPath = "unix:///run/spire/sockets/agent.sock"

func getNewX509Source() *workloadapi.X509Source {
	x509Source, err := workloadapi.NewX509Source(context.Background(), workloadapi.WithClientOptions(workloadapi.WithAddr(agentSocketPath)))
	if err != nil {
		log.Fatal("Unable to create X509Source: ", err)
	}

	return x509Source
}

func getNewKafkaReader(x509Source *workloadapi.X509Source, kafkaURL string, topic string, spiffeID string) *kafka.Reader {

	serverID := spiffeid.RequireFromString(spiffeID)

	config := tlsconfig.MTLSClientConfig(x509Source, x509Source, tlsconfig.AuthorizeID(serverID))

	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
		TLS:       config,
	}

	log.Printf("Configuring Broker : %v and Topic : %v", kafkaURL, topic)

	brokers := strings.Split(kafkaURL, ",")
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Dialer:   dialer,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
}

func getRequiredEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("%v env variable not provided", key)
	}
	log.Printf("Kafka bootstrap Server : %v", value)

	return value
}

func main() {
	kafkaURL := getRequiredEnv("KAFKA_BOOTSTRAP_SERVER")
	topic := getRequiredEnv("KAFKA_TOPIC")
	spiffeID := getRequiredEnv("SPIFFE_ID")

	x509Source := getNewX509Source()
	defer x509Source.Close()

	kafkaReader := getNewKafkaReader(x509Source, kafkaURL, topic, spiffeID)
	defer kafkaReader.Close()

	log.Println("Starting consuming...")

	for {
		m, err := kafkaReader.ReadMessage(context.Background())
		if err != nil {
			log.Print("Error reading message : ", err)
			return
		}

		log.Printf("Message at topic:%v partition:%v offset:%v	%s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
	}
}
