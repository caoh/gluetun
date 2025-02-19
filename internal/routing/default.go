package routing

import (
	"errors"
	"fmt"
	"net"

	"github.com/qdm12/gluetun/internal/netlink"
)

var (
	ErrRouteDefaultNotFound = errors.New("default route not found")
)

type DefaultRoute struct {
	NetInterface string
	Gateway      net.IP
	AssignedIP   net.IP
	Family       int
}

func (d DefaultRoute) String() string {
	return fmt.Sprintf("interface %s, gateway %s and assigned IP %s",
		d.NetInterface, d.Gateway, d.AssignedIP)
}

func (r *Routing) DefaultRoutes() (defaultRoutes []DefaultRoute, err error) {
	routes, err := r.netLinker.RouteList(nil, netlink.FAMILY_ALL)
	if err != nil {
		return nil, fmt.Errorf("cannot list routes: %w", err)
	}

	for _, route := range routes {
		if route.Dst == nil {
			defaultRoute := DefaultRoute{
				Gateway: route.Gw,
				Family:  route.Family,
			}
			linkIndex := route.LinkIndex
			link, err := r.netLinker.LinkByIndex(linkIndex)
			if err != nil {
				return nil, fmt.Errorf("cannot obtain link by index: for default route at index %d: %w", linkIndex, err)
			}
			attributes := link.Attrs()
			defaultRoute.NetInterface = attributes.Name

			defaultRoute.AssignedIP, err = r.assignedIP(defaultRoute.NetInterface)
			if err != nil {
				return nil, fmt.Errorf("cannot get assigned IP of %s: %w", defaultRoute.NetInterface, err)
			}

			r.logger.Info("default route found: " + defaultRoute.String())
			defaultRoutes = append(defaultRoutes, defaultRoute)
		}
	}

	if len(defaultRoutes) == 0 {
		return nil, fmt.Errorf("%w: in %d route(s)", ErrRouteDefaultNotFound, len(routes))
	}

	return defaultRoutes, nil
}

func DefaultRoutesInterfaces(defaultRoutes []DefaultRoute) (interfaces []string) {
	interfaces = make([]string, len(defaultRoutes))
	for i := range defaultRoutes {
		interfaces[i] = defaultRoutes[i].NetInterface
	}
	return interfaces
}
