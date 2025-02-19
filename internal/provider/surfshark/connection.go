package surfshark

import (
	"github.com/qdm12/gluetun/internal/configuration/settings"
	"github.com/qdm12/gluetun/internal/models"
	"github.com/qdm12/gluetun/internal/provider/utils"
)

func (p *Provider) GetConnection(selection settings.ServerSelection) (
	connection models.Connection, err error) {
	defaults := utils.NewConnectionDefaults(1443, 1194, 0) //nolint:gomnd
	return utils.GetConnection(p.Name(),
		p.storage, selection, defaults, p.randSource)
}
