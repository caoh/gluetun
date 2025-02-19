package env

import (
	"fmt"
	"time"

	"github.com/qdm12/gluetun/internal/configuration/settings"
)

func (r *Reader) readPublicIP() (publicIP settings.PublicIP, err error) {
	publicIP.Period, err = readPublicIPPeriod()
	if err != nil {
		return publicIP, err
	}

	publicIP.IPFilepath = r.readPublicIPFilepath()

	return publicIP, nil
}

func readPublicIPPeriod() (period *time.Duration, err error) {
	s := getCleanedEnv("PUBLICIP_PERIOD")
	if s == "" {
		return nil, nil //nolint:nilnil
	}

	period = new(time.Duration)
	*period, err = time.ParseDuration(s)
	if err != nil {
		return nil, fmt.Errorf("environment variable PUBLICIP_PERIOD: %w", err)
	}

	return period, nil
}

func (r *Reader) readPublicIPFilepath() (filepath *string) {
	_, s := r.getEnvWithRetro("PUBLICIP_FILE", "IP_STATUS_FILE")
	if s != "" {
		return &s
	}
	return nil
}
