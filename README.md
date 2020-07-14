# Estes
RabbitMQ wrapper for Golang. we using github.com/streadway/amqp
# How to Use
### Publish Event/Message
first create file `publisher.go` 
```go
package pubsub

import (
	"encoding/json"
	"fmt"
	"os"
	"github.com/hualoqueros/estes"
)

func Publish(messageEvent MessageEvent) (*estes.RMQ, error) {
        rmq, err := estes.InitRMQ(os.Getenv("RABBIT_MQ_URL"), os.Getenv("RABBIT_MQ_EXCHANGE"))
    	if err != nil {
    		fmt.Printf("ERROR RMQ : +%v", err)
    	}
    
    	msgByte, _ := json.Marshal(messageEvent)
    	err = rmq.PublishEvent(msgByte)
    
    	if err != nil {
    		fmt.Printf("ERROR RMQ : +%v", err)
    	}
}
```
let assume we have file `user_controller.go` and we want to push event `user_created`, so you can call the `publisher.go` on your logic script  like this:
```go
  msg := pubsub.MessageEvent{
	Event: "user_created",
	Data: map[string]interface{}{
	        "username": "jhonDoe",
	 		"email":    "jhon.doe@email.com",
	    },
	}

	rmqPublisher, err := pubsub.Publish(msg)
	if err != nil {
	 	fmt.Printf("ERROR PUBLISH EVENT : %+v\n", err)
    }
	defer rmqPublisher.Channel.Close()
	defer rmqPublisher.Connection.Close()
```

### Consume Event/Message
create file `consumer.go`
```go
package pubsub

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/hualoqueros/estes"
	"github.com/streadway/amqp"
)

type MessageEvent struct {
	Event string                 `json:"event"`
	Data  map[string]interface{} `json:"data"`
}

func StartConsumer(queueName string) (*estes.RMQ, error) {
	fmt.Println("Start the consumer...")

	rmq, err := estes.InitRMQ(os.Getenv("RABBIT_MQ_URL"), os.Getenv("RABBIT_MQ_EXCHANGE"))
	if err != nil {
		fmt.Printf("ERROR RMQ : +%v", err)
	}

	ds, err := rmq.ConsumeEvent(queueName)
	if err != nil {
		fmt.Printf("ERROR RMQ : +%v", err)
	}
	go consume(ds)
	return rmq, err
}

func consume(ds <-chan amqp.Delivery) {
	for d := range ds {
		log.Printf(" [x] %s", d.Body)
		var msg MessageEvent
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			log.Printf(err.Error())
		}
		log.Printf(" [x] Messages  : %s", msg)
		switch msg.Event {
		case "user_created":
			log.Println("Process User Created Event")
			// do logic here
			break
		default:
			log.Println("Default here")
		}

	}
}
```

now we can run the consumer :
```go
	rmqConsumer, err := pubsub.StartConsumer(os.Getenv("QUEUE_NAME"))
	if err != nil {
		fmt.Printf("ERROR PUBLISH EVENT : %+v\n", err)
	}
	defer rmqConsumer.Channel.Close()
	defer rmqConsumer.Connection.Close()
```

