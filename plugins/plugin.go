// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package plugins

import (
	"errors"
	"fmt"
	"github.com/coredhcp/coredhcp/config"
	"github.com/coredhcp/coredhcp/handler"
	"github.com/coredhcp/coredhcp/logger"
)

var log = logger.GetLogger("plugins")

// Plugin represents a plugin object.
// Setup6 and Setup4 are the setup functions for DHCPv6 and DHCPv4 handlers
// respectively. Both setup functions can be nil.
type Plugin struct {
	Name   string
	Setup6 SetupFunc6
	Setup4 SetupFunc4
}

// RegisteredPlugins maps a plugin name to a Plugin instance.
//var RegisteredPlugins = make(map[string]*Plugin)

// SetupFunc6 defines a plugin setup function for DHCPv6
type SetupFunc6 func(Listiner string, args ...string) (handler.Handler6, error)

// SetupFunc4 defines a plugin setup function for DHCPv6
type SetupFunc4 func(Listiner string, args ...string) (handler.Handler4, error)

// RegisterPlugin registers a plugin.
func RegisterPlugin(Listiner string, RegisteredPlugins map[string]*Plugin, plugin *Plugin) error {
	if plugin == nil {
		return errors.New("cannot register nil plugin")
	}
	log.Printf("Listiner: %s, Registering plugin '%s'", Listiner, plugin.Name)
	if _, ok := RegisteredPlugins[plugin.Name]; ok {
		// TODO this highlights that asking the plugins to register themselves
		// is not the right approach. Need to register them in the main program.
		log.Panicf("Listiner: %s, Plugin '%s' is already registered", Listiner, plugin.Name)
	}
	RegisteredPlugins[plugin.Name] = plugin
	return nil
}

// LoadPlugins reads a Config object and loads the plugins as specified in the
// `plugins` section, in order. For a plugin to be available, it must have been
// previously registered with plugins.RegisterPlugin. This is normally done at
// plugin import time.
// This function returns the list of loaded v6 plugins, the list of loaded v4
// plugins, and an error if any.
func LoadPlugins(Listiner string, RegisteredPlugins map[string]*Plugin, conf *config.Config) ([]handler.Handler4, []handler.Handler6, error) {
	log.Print("Loading plugins...")
	handlers4 := make([]handler.Handler4, 0)
	handlers6 := make([]handler.Handler6, 0)

	if conf.Server6 == nil && conf.Server4 == nil {
		return nil, nil, errors.New("no configuration found for either DHCPv6 or DHCPv4")
	}

	// now load the plugins. We need to call its setup function with
	// the arguments extracted above. The setup function is mapped in
	// plugins.RegisteredPlugins .

	// Load DHCPv6 plugins.
	if conf.Server6 != nil {
		for _, pluginConf := range conf.Server6.Plugins {
			if plugin, ok := RegisteredPlugins[pluginConf.Name]; ok {
				log.Printf("Listiner: %s, DHCPv6: loading plugin `%s`", Listiner, pluginConf.Name)
				if plugin.Setup6 == nil {
					log.Warningf("Listiner: %s, DHCPv6: plugin `%s` has no setup function for DHCPv6", Listiner, pluginConf.Name)
					continue
				}
				h6, err := plugin.Setup6(Listiner,pluginConf.Args...)
				if err != nil {
					return nil, nil, err
				} else if h6 == nil {
					return nil, nil, config.ConfigErrorFromString("Listiner: %s, no DHCPv6 handler for plugin %s", Listiner, pluginConf.Name)
				}
				handlers6 = append(handlers6, h6)
			} else {
				return nil, nil, config.ConfigErrorFromString("Listiner: %s, DHCPv6: unknown plugin `%s`", Listiner, pluginConf.Name)
			}
		}
	}
	// Load DHCPv4 plugins. Yes, duplicated code, there's not really much that
	// can be deduplicated here.
	log.Info(fmt.Sprintf("Listiner: %s, EEEEEEE>%v\n", Listiner, conf.Server4))
	if conf.Server4 != nil {
		for _, pluginConf := range conf.Server4.Plugins {
			if plugin, ok := RegisteredPlugins[pluginConf.Name]; ok {
				log.Printf("Listiner: %s, DHCPv4: loading plugin `%s`", Listiner, pluginConf.Name)
				if plugin.Setup4 == nil {
					log.Warningf("Listiner: %s, DHCPv4: plugin `%s` has no setup function for DHCPv4", Listiner, pluginConf.Name)
					continue
				}
                                log.Info(fmt.Sprintf("Listiner: %s, DHCPv4, run Setup4: %v, %v.", Listiner, pluginConf.Name, pluginConf))
				h4, err := plugin.Setup4(Listiner, pluginConf.Args...)
				if err != nil {
					return nil, nil, err
				} else if h4 == nil {
					return nil, nil, config.ConfigErrorFromString("Listiner: %s, no DHCPv4 handler for plugin %s", Listiner, pluginConf.Name)
				}
				handlers4 = append(handlers4, h4)
			} else {
				return nil, nil, config.ConfigErrorFromString("Listiner: %s, DHCPv4: unknown plugin `%s`", Listiner, pluginConf.Name)
			}
		}
	}

	return handlers4, handlers6, nil
}
