package consul

import (
	"github.com/hashicorp/consul/api"
	"github.com/rs/zerolog/log"
)

func WatchService(serviceName string, serviceRemoved chan<- struct{}) {
	client, err := GetClient()
	if err != nil {
		return
	}

	var lastIndex uint64
	for {
		// do not only return healthy services, we do not want to treat a disconnected service as it being removed
		services, meta, err := client.Health().Service(serviceName, "", false, &api.QueryOptions{
			WaitIndex: lastIndex, // Long polling
		})
		if err != nil {
			log.Err(err).Msgf("Failed to query health for service %s from consul.", serviceName)
			continue
		}

		lastIndex = meta.LastIndex

		if len(services) == 0 {
			serviceRemoved <- struct{}{}
			return
		}
	}
}
