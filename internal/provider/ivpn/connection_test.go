package ivpn

import (
	"errors"
	"math/rand"
	"net"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/qdm12/gluetun/internal/configuration/settings"
	"github.com/qdm12/gluetun/internal/constants"
	"github.com/qdm12/gluetun/internal/constants/providers"
	"github.com/qdm12/gluetun/internal/constants/vpn"
	"github.com/qdm12/gluetun/internal/models"
	"github.com/qdm12/gluetun/internal/provider/common"
	"github.com/stretchr/testify/assert"
)

func Test_Provider_GetConnection(t *testing.T) {
	t.Parallel()

	const provider = providers.Ivpn

	errTest := errors.New("test error")
	boolPtr := func(b bool) *bool { return &b }

	testCases := map[string]struct {
		filteredServers []models.Server
		storageErr      error
		selection       settings.ServerSelection
		connection      models.Connection
		errWrapped      error
		errMessage      string
	}{
		"error": {
			storageErr: errTest,
			errWrapped: errTest,
			errMessage: "cannot filter servers: test error",
		},
		"default OpenVPN TCP port": {
			filteredServers: []models.Server{
				{IPs: []net.IP{net.IPv4(1, 1, 1, 1)}},
			},
			selection: settings.ServerSelection{
				OpenVPN: settings.OpenVPNSelection{
					TCP: boolPtr(true),
				},
			}.WithDefaults(provider),
			connection: models.Connection{
				Type:     vpn.OpenVPN,
				IP:       net.IPv4(1, 1, 1, 1),
				Port:     443,
				Protocol: constants.TCP,
			},
		},
		"default OpenVPN UDP port": {
			filteredServers: []models.Server{
				{IPs: []net.IP{net.IPv4(1, 1, 1, 1)}},
			},
			selection: settings.ServerSelection{
				OpenVPN: settings.OpenVPNSelection{
					TCP: boolPtr(false),
				},
			}.WithDefaults(provider),
			connection: models.Connection{
				Type:     vpn.OpenVPN,
				IP:       net.IPv4(1, 1, 1, 1),
				Port:     1194,
				Protocol: constants.UDP,
			},
		},
		"default Wireguard port": {
			filteredServers: []models.Server{
				{IPs: []net.IP{net.IPv4(1, 1, 1, 1)}, WgPubKey: "x"},
			},
			selection: settings.ServerSelection{
				VPN: vpn.Wireguard,
			}.WithDefaults(provider),
			connection: models.Connection{
				Type:     vpn.Wireguard,
				IP:       net.IPv4(1, 1, 1, 1),
				Port:     58237,
				Protocol: constants.UDP,
				PubKey:   "x",
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			storage := common.NewMockStorage(ctrl)
			storage.EXPECT().FilterServers(provider, testCase.selection).
				Return(testCase.filteredServers, testCase.storageErr)
			randSource := rand.NewSource(0)

			client := (*http.Client)(nil)
			warner := (common.Warner)(nil)
			parallelResolver := (common.ParallelResolver)(nil)
			provider := New(storage, randSource, client, warner, parallelResolver)

			connection, err := provider.GetConnection(testCase.selection)

			assert.ErrorIs(t, err, testCase.errWrapped)
			if testCase.errWrapped != nil {
				assert.EqualError(t, err, testCase.errMessage)
			}

			assert.Equal(t, testCase.connection, connection)
		})
	}
}
