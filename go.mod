module github.com/egfanboy/mediapire-manager

go 1.16

require (
	github.com/egfanboy/mediapire-common v0.0.0-20220828193527-bd4d006a2cb6
	github.com/egfanboy/mediapire-media-host v0.0.0-20220828193959-b4703c0b0de3
	github.com/go-redis/redis/v9 v9.0.0-beta.2
	github.com/gorilla/mux v1.8.0
	github.com/rs/zerolog v1.27.0

)

// uncomment for local development
// replace github.com/egfanboy/mediapire-common => ../mediapire-common
// replace github.com/egfanboy/mediapire-media-host => ../mediapire-media-host
