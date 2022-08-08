package mediahost

import (
	"fmt"
	"net/http"
)

type mediaHostIntegration interface {
	VerifyConnectivity(scheme string, host string, port int) error
}

type integration struct {
}

func (i *integration) VerifyConnectivity(scheme string, host string, port int) error {
	hostUri := fmt.Sprintf("%s:%v", host, port)

	resp, err := http.Get(fmt.Sprintf("%s://%s/api/v1/health", scheme, hostUri))

	if err != nil {

		return err
	}

	if resp != nil && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("media-host %s returned status code %q instead of 200", hostUri, resp.StatusCode)
	}

	return nil
}

func NewMediaHostIntegration() mediaHostIntegration {
	return &integration{}
}
