// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package router

import (
	"errors"
	"net"
	"fmt"
	"github.com/coredhcp/coredhcp/handler"
	"github.com/coredhcp/coredhcp/logger"
	"github.com/coredhcp/coredhcp/plugins"
	"github.com/insomniacslk/dhcp/dhcpv4"
)

var log = logger.GetLogger("plugins/router")

// Plugin wraps plugin registration information
var Plugin = plugins.Plugin{
	Name:   "router",
	Setup4: setup4,
}

var (
	routers map[string][]net.IP
)

func init() {
	routers = make(map[string][]net.IP)
}

func setup4(Listiner string, args ...string) (handler.Handler4, error) {
	log.Printf("Loaded plugin for DHCPv4.")
	if len(args) < 1 {
		return nil, errors.New("need at least one router IP address")
	}
	routers[Listiner] = []net.IP{}
	for _, arg := range args {
		router := net.ParseIP(arg)
		if router.To4() == nil {
			return Handler4, errors.New("expected an router IP address, got: " + arg)
		}
		routers[Listiner] = append(routers[Listiner], router)
	}
	log.Infof("loaded %d router IP addresses.", len(routers))
	return Handler4, nil
}

//Handler4 handles DHCPv4 packets for the router plugin
func Handler4(Listiner string, req, resp *dhcpv4.DHCPv4) (*dhcpv4.DHCPv4, bool) {
	log.Infof(fmt.Sprintf("ROUTER: Handler4: %v,%v\n\t%v\n", Listiner, routers[Listiner],req))
	resp.Options.Update(dhcpv4.OptRouter(routers[Listiner]...))
	return resp, false
}
