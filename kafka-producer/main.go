package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
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

func getNewKafkaWriter(x509Source *workloadapi.X509Source, kafkaURL string, topic string, spiffeID string) *kafka.Writer {

	serverID := spiffeid.RequireFromString(spiffeID)

	config := tlsconfig.MTLSClientConfig(x509Source, x509Source, tlsconfig.AuthorizeID(serverID))

	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
		TLS:       config,
	}

	log.Printf("Configuring Broker : %v and Topic : %v", kafkaURL, topic)

	return kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{kafkaURL},
		Topic:    topic,
		Balancer: &kafka.Hash{},
		Dialer:   dialer,
	})
}

func handleSendMessage(kafkaWriter *kafka.Writer) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Printf("Error reading request body: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("Received message : ", fmt.Sprintf("%s", body))

		msg := kafka.Message{
			Key:   []byte(uuid.New().String()),
			Value: body,
		}

		err = kafkaWriter.WriteMessages(req.Context(), msg)
		if err != nil {
			log.Printf("Error writing to topic: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("Successfully written message to Kafka topic!!")
		w.WriteHeader(http.StatusOK)
	})
}

func getRequiredEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("%v env variable not provided", key)
	}
	log.Printf("Received %v : %v", key, value)

	return value
}

func main() {
	kafkaURL := getRequiredEnv("KAFKA_BOOTSTRAP_SERVER")
	topic := getRequiredEnv("KAFKA_TOPIC")
	spiffeID := getRequiredEnv("SPIFFE_ID")

	x509Source := getNewX509Source()
	defer x509Source.Close()

	kafkaWriter := getNewKafkaWriter(x509Source, kafkaURL, topic, spiffeID)
	defer kafkaWriter.Close()

	log.Println("Starting producer...")

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/sendMessage", handleSendMessage(kafkaWriter))

	// Run the web server.
	log.Fatal(http.ListenAndServe(":8080", httpMux))
}
