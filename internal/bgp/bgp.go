// Package bgp provides control functions for BGP operations
package bgp

import (
	"context"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	api "github.com/osrg/gobgp/api"
	gobgp "github.com/osrg/gobgp/pkg/server"
	log "github.com/sirupsen/logrus"
	"net"
)

// Daemon wraps a gobgp.BgpServer
type Daemon struct {
	BgpServer *gobgp.BgpServer
}

// Config stores a router's local configuration
type Config struct {
	Asn      uint32
	RouterId string
}

// NewPeer creates a new BGP session
func (d *Daemon) NewPeer(address string, asn uint32) error {
	return d.BgpServer.AddPeer(context.Background(), &api.AddPeerRequest{
		Peer: &api.Peer{
			Conf: &api.PeerConf{
				NeighborAddress: address,
				PeerAs:          asn,
			},
		},
	})
}

// Announce announces a route
func (d Daemon) Announce(prefix net.IPNet, nexthop net.IPAddr) error {
	// Determine prefix address family and set AFI (SAFI is always UNICAST)
	var afi api.Family_Afi
	if prefix.IP.To4() == nil { // If IPv6
		afi = api.Family_AFI_IP6
	} else {
		afi = api.Family_AFI_IP
	}

	family := &api.Family{
		Afi:  afi,
		Safi: api.Family_SAFI_UNICAST,
	}

	// Get prefix netmask
	_, pfxLen := prefix.Mask.Size()

	// BGP route attributes
	// TODO: error handling instead of ignoring
	nlri, _ := ptypes.MarshalAny(&api.IPAddressPrefix{
		Prefix:    prefix.IP.String(),
		PrefixLen: uint32(pfxLen),
	})

	originAttr, _ := ptypes.MarshalAny(&api.OriginAttribute{
		Origin: 0,
	})

	reachabilityAttrs, _ := ptypes.MarshalAny(&api.MpReachNLRIAttribute{
		Family:   family,
		NextHops: []string{nexthop.String()},
		Nlris:    []*any.Any{nlri},
	})

	largeCommunities, _ := ptypes.MarshalAny(&api.LargeCommunitiesAttribute{
		Communities: []*api.LargeCommunity{},
	})

	pathAttrs := []*any.Any{originAttr, reachabilityAttrs, largeCommunities}

	_, err := d.BgpServer.AddPath(context.Background(), &api.AddPathRequest{
		Path: &api.Path{
			Family: family,
			Nlri:   nlri,
			Pattrs: pathAttrs,
		},
	})

	return err
}

// TODO: write Withdraw function

// New constructs a new Daemon
func New(asn uint32, routerId string) *Daemon {
	s := gobgp.NewBgpServer()
	//go s.Serve()

	// global configuration
	if err := s.StartBgp(context.Background(), &api.StartBgpRequest{
		Global: &api.Global{
			As:         asn,
			RouterId:   routerId,
			ListenPort: 179,
		},
	}); err != nil {
		log.Fatal(err)
	}

	// monitor the change of the peer state
	if err := s.MonitorPeer(context.Background(), &api.MonitorPeerRequest{}, func(p *api.Peer) { log.Info(p) }); err != nil {
		// TODO: Send to metrics system
		log.Fatal(err)
	}

	return &Daemon{BgpServer: s}
}
