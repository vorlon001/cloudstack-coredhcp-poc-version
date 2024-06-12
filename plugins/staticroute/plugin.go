// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package staticroute

import (
	"errors"
	"net"
	"strings"
	"fmt"
	"github.com/coredhcp/coredhcp/handler"
	"github.com/coredhcp/coredhcp/logger"
	"github.com/coredhcp/coredhcp/plugins"
	"github.com/insomniacslk/dhcp/dhcpv4"
)

var log = logger.GetLogger("plugins/staticroute")

// Plugin wraps the information necessary to register a plugin.
var Plugin = plugins.Plugin{
	Name:   "staticroute",
	Setup4: setup4,
}

var routes map[string]dhcpv4.Routes

func init() {
	routes = make(map[string]dhcpv4.Routes)
}


func setup4(Listiner string, args ...string) (handler.Handler4, error) {
	log.Printf("loaded plugin for DHCPv4.")

	if len(args) < 1 {
		return nil, errors.New("need at least one static route")
	}

	var err error
	routes[Listiner] = make(dhcpv4.Routes, 0)
	for _, arg := range args {
		fields := strings.Split(arg, ",")
		if len(fields) != 2 {
			return Handler4, errors.New("expected a destination/gateway pair, got: " + arg)
		}

		route := &dhcpv4.Route{}
		_, route.Dest, err = net.ParseCIDR(fields[0])
		if err != nil {
			return Handler4, errors.New("expected a destination subnet, got: " + fields[0])
		}

		route.Router = net.ParseIP(fields[1])
		if route.Router == nil {
			return Handler4, errors.New("expected a gateway address, got: " + fields[1])
		}

		routes[Listiner] = append(routes[Listiner], route)
		log.Debugf("adding static route %s", route)
	}

	log.Printf("loaded %d static routes.", len(routes))

	return Handler4, nil
}

// Handler4 handles DHCPv4 packets for the static routes plugin
func Handler4(Listiner string, req, resp *dhcpv4.DHCPv4) (*dhcpv4.DHCPv4, bool) {
	log.Infof(fmt.Sprintf("ROUTER: Handler4: %v,%v\n\t%v\n", Listiner, routes[Listiner],req))
	if len(routes) > 0 {
		resp.Options.Update(dhcpv4.Option{
			Code:  dhcpv4.OptionCode(dhcpv4.OptionClasslessStaticRoute),
			Value: routes[Listiner],
		})
	}

	return resp, false
}
