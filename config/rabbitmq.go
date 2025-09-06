package config

import (
	"github.com/Ermi9s/Anubis/model"
	"github.com/Ermi9s/Anubis/repository"
	"encoding/json"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"log"
	amqp "github.com/streadway/amqp"
)



func StartRabbitMQConsumer(configuration *model.Configuration, repository *repository.Repository) {
	
	conn, err := amqp.Dial(configuration.RabbitMQ.Address)
	if err != nil {
		log.Fatalf("[Anubis Error] failed to connect to RabbitMQ: %s", err)
		defer conn.Close()
	}


	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("[Anubis Error] failed to open channel: %s", err)
		defer ch.Close()
	}

	
	argsTable := amqp.Table{}
	for k, v := range configuration.RabbitMQ.Args {
		argsTable[k] = v
	}


	queue, err := ch.QueueDeclare(
		configuration.RabbitMQ.QueueName,
		configuration.RabbitMQ.Durable,
		configuration.RabbitMQ.AutoDelete,
		configuration.RabbitMQ.Exclusive,
		configuration.RabbitMQ.NoWait,
		argsTable,
	)

	if err != nil {
		log.Fatalf("[Anubis Error] failed to declare queue: %s", err)
	}
	
	workers := configuration.Concurency
	if workers <= 0 {
		workers = 5
	} 


	err = ch.Qos(workers, 0, false)
	if err != nil {
		log.Fatalf("[Anubis Error] failed to set QoS: %s", err)
	}


	messages, err := ch.Consume(
		queue.Name,
		"",    
		false, 
		false, 
		false, 
		false, 
		nil,
	)

	if err != nil {
		log.Fatalf("[Anubis Error] failed to register consume: %s", err)
	}

	log.Printf("[Anubis] Waiting for messages from queue %s, number of workers %d", configuration.RabbitMQ.QueueName, workers)


	var wg sync.WaitGroup


	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			log.Printf("[Anubis] Worker %d deployed", workerID)
			for msg := range messages {
				var event model.AuditEvent
				if err := json.Unmarshal(msg.Body, &event); err != nil {
					log.Printf("[Anubis Error] Worker %d: Error decoding message: %v", workerID, err)
					msg.Nack(false, false) 
					continue
				}


				if err := repository.CreateAudit(event); err != nil {
					log.Printf("[Anubis Error] Worker %d: Error Saving message: %v to database", workerID, err)
					msg.Nack(false, true) 
					continue
				}

				msg.Ack(false) 
			}
		}(i)
	}

	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan // waits for a program break
	log.Println("[Anubis] Shutdown signal received, waiting for workers to finish...")
	ch.Close()
	conn.Close()


	wg.Wait() //might have stuck worker problems fix later
	log.Println("[Anubis] Consumer stopped gracefully")
}
