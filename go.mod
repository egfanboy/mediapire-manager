module github.com/egfanboy/mediapire-manager

go 1.16

require (
	github.com/egfanboy/mediapire-common v0.0.0-20220905143518-e3d4d7ef0bac
	github.com/egfanboy/mediapire-media-host v0.0.1-alpha.0.20230502001517-88234b1fc6ce
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/consul/api v1.15.3
	github.com/rs/zerolog v1.27.0
	gopkg.in/yaml.v3 v3.0.1

)

// uncomment for local development
// replace github.com/egfanboy/mediapire-common => ../mediapire-common
// replace github.com/egfanboy/mediapire-media-host => ../mediapire-media-host
