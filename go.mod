module github.com/egfanboy/mediapire-manager

go 1.16

require (
	github.com/egfanboy/mediapire-common v0.0.0-20240522004433-cbd1b5041bc7
	github.com/egfanboy/mediapire-media-host v0.0.1-alpha.0.20250209202443-affe1d3552c7
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
