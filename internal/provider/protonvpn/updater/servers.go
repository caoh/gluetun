// Package updater contains code to obtain the server information
// for the ProtonVPN provider.
package updater

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/qdm12/gluetun/internal/constants"
	"github.com/qdm12/gluetun/internal/models"
	"github.com/qdm12/gluetun/internal/provider/common"
)

func (u *Updater) FetchServers(ctx context.Context, minServers int) (
	servers []models.Server, err error) {
	data, err := fetchAPI(ctx, u.client)
	if err != nil {
		return nil, err
	}

	countryCodes := constants.CountryCodes()

	var count int
	for _, logicalServer := range data.LogicalServers {
		count += len(logicalServer.Servers)
	}

	if count < minServers {
		return nil, fmt.Errorf("%w: %d and expected at least %d",
			common.ErrNotEnoughServers, count, minServers)
	}

	ipToServer := make(ipToServer, count)
	for _, logicalServer := range data.LogicalServers {
		region := getStringValue(logicalServer.Region)
		city := getStringValue(logicalServer.City)
		// TODO v4 remove `name` field because of
		// https://github.com/qdm12/gluetun/issues/1018#issuecomment-1151750179
		name := logicalServer.Name
		for _, physicalServer := range logicalServer.Servers {
			if physicalServer.Status == 0 { // disabled so skip server
				u.warner.Warn("ignoring server " + physicalServer.Domain + " with status 0")
				continue
			}

			hostname := physicalServer.Domain
			entryIP := physicalServer.EntryIP

			lowerCaseName := strings.ToLower(name)
			var free bool
			if strings.Contains(hostname, "free") || strings.Contains(lowerCaseName, "free") {
				free = true
			}

			// Note: for multi-hop use the server name or hostname
			// instead of the country
			countryCode := logicalServer.ExitCountry
			country, warning := codeToCountry(countryCode, countryCodes)
			if warning != "" {
				u.warner.Warn(warning)
			}

			ipToServer.add(country, region, city, name, hostname, free, entryIP)
		}
	}

	if len(ipToServer) < minServers {
		return nil, fmt.Errorf("%w: %d and expected at least %d",
			common.ErrNotEnoughServers, len(ipToServer), minServers)
	}

	servers = ipToServer.toServersSlice()

	sort.Sort(models.SortableServers(servers))

	return servers, nil
}

func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
