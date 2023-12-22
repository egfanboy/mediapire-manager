module github.com/egfanboy/mediapire-manager

go 1.16

require (
	github.com/egfanboy/mediapire-common v0.0.0-20231219000342-fbb6228cf11c
	github.com/egfanboy/mediapire-media-host v0.0.1-alpha.0.20231222181916-4af0e32008a0
	github.com/google/uuid v1.4.0
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/consul/api v1.15.3
	github.com/rabbitmq/amqp091-go v1.8.1
	github.com/rs/zerolog v1.27.0
	go.mongodb.org/mongo-driver v1.13.0
	gopkg.in/yaml.v3 v3.0.1

)

// uncomment for local development
// replace github.com/egfanboy/mediapire-common => ../mediapire-common
// replace github.com/egfanboy/mediapire-media-host => ../mediapire-media-host
