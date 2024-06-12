// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package dns

import (
	"errors"
	"net"
	"fmt"
	"github.com/coredhcp/coredhcp/handler"
	"github.com/coredhcp/coredhcp/logger"
	"github.com/coredhcp/coredhcp/plugins"
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv6"
)

var log = logger.GetLogger("plugins/dns")

// Plugin wraps the DNS plugin information.
var Plugin = plugins.Plugin{
	Name:   "dns",
	Setup6: setup6,
	Setup4: setup4,
}

var (
	dnsServers6 map[string][]net.IP
	dnsServers4 map[string][]net.IP
)


func init() {
        dnsServers6 = make(map[string][]net.IP)
        dnsServers4 = make(map[string][]net.IP)
}



func setup6(Listiner string, args ...string) (handler.Handler6, error) {
	if len(args) < 1 {
		return nil, errors.New("need at least one DNS server")
	}
	dnsServers6[Listiner] = make([]net.IP, 0)
	for _, arg := range args {
		server := net.ParseIP(arg)
		if server.To16() == nil {
			return Handler6, errors.New("expected an DNS server address, got: " + arg)
		}
		dnsServers6[Listiner] = append(dnsServers6[Listiner], server)
	}
	log.Infof("loaded %d DNS servers.", len(dnsServers6[Listiner]))
	return Handler6, nil
}

func setup4(Listiner string, args ...string) (handler.Handler4, error) {
	log.Printf("loaded plugin for DHCPv4.")
        dnsServers4[Listiner] = make([]net.IP, 0)
	if len(args) < 1 {
		return nil, errors.New("need at least one DNS server")
	}
	for _, arg := range args {
		DNSServer := net.ParseIP(arg)
		if DNSServer.To4() == nil {
			return Handler4, errors.New("expected an DNS server address, got: " + arg)
		}
		dnsServers4[Listiner] = append(dnsServers4[Listiner], DNSServer)
	}
	log.Infof("loaded %d DNS servers.", len(dnsServers4[Listiner]))
	return Handler4, nil
}

// Handler6 handles DHCPv6 packets for the dns plugin
func Handler6(Listiner string, req, resp dhcpv6.DHCPv6) (dhcpv6.DHCPv6, bool) {
        log.Infof(fmt.Sprintf("DNS Server: Handler4: %v,%v\n\t%v\n", Listiner, dnsServers6[Listiner],req))
	decap, err := req.GetInnerMessage()
	if err != nil {
		log.Errorf("Could not decapsulate relayed message, aborting: %v", err)
		return nil, true
	}

	if decap.IsOptionRequested(dhcpv6.OptionDNSRecursiveNameServer) {
		resp.UpdateOption(dhcpv6.OptDNS(dnsServers6[Listiner]...))
	}
	return resp, false
}

//Handler4 handles DHCPv4 packets for the dns plugin
func Handler4(Listiner string, req, resp *dhcpv4.DHCPv4) (*dhcpv4.DHCPv4, bool) {
        log.Infof(fmt.Sprintf("DNS Server: Handler4: %v,%v\n\t%v\n", Listiner, dnsServers6[Listiner],req))
	if req.IsOptionRequested(dhcpv4.OptionDomainNameServer) {
		resp.Options.Update(dhcpv4.OptDNS(dnsServers4[Listiner]...))
	}
	return resp, false
}
