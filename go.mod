module github.com/egfanboy/mediapire-manager

go 1.16

require (
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/egfanboy/mediapire-common v0.0.0-20220905143518-e3d4d7ef0bac
	github.com/egfanboy/mediapire-media-host v0.0.0-20220905144209-41b677b2a6a6
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/consul/api v1.15.3
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.3.1 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/rs/zerolog v1.27.0
	golang.org/x/sys v0.1.0 // indirect
	gopkg.in/yaml.v3 v3.0.1

)

// uncomment for local development
// replace github.com/egfanboy/mediapire-common => ../mediapire-common
// replace github.com/egfanboy/mediapire-media-host => ../mediapire-media-host
