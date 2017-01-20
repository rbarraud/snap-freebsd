/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package control

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/vrischmann/jsonutil"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
)

// default configuration values
const (
	defaultListenAddr        string        = "127.0.0.1"
	defaultListenPort        int           = 8082
	defaultMaxRunningPlugins int           = 3
	defaultPluginLoadTimeout int           = 3
	defaultPluginTrust       int           = 1
	defaultAutoDiscoverPath  string        = ""
	defaultKeyringPaths      string        = ""
	defaultCacheExpiration   time.Duration = 500 * time.Millisecond
)

type pluginConfig struct {
	All         *cdata.ConfigDataNode `json:"all"`
	Collector   *pluginTypeConfigItem `json:"collector"`
	Publisher   *pluginTypeConfigItem `json:"publisher"`
	Processor   *pluginTypeConfigItem `json:"processor"`
	pluginCache map[string]*cdata.ConfigDataNode
}

type pluginTypeConfigItem struct {
	Plugins map[string]*pluginConfigItem
	All     *cdata.ConfigDataNode `json:"all"`
}

type pluginConfigItem struct {
	*cdata.ConfigDataNode
	Versions map[int]*cdata.ConfigDataNode `json:"versions"`
}

// holds the configuration passed in through the SNAP config file
//   Note: if this struct is modified, then the switch statement in the
//         UnmarshalJSON method in this same file needs to be modified to
//         match the field mapping that is defined here
type Config struct {
	MaxRunningPlugins int               `json:"max_running_plugins"yaml:"max_running_plugins"`
	PluginLoadTimeout int               `json:"plugin_load_timeout"yaml:"plugin_load_timeout"`
	PluginTrust       int               `json:"plugin_trust_level"yaml:"plugin_trust_level"`
	AutoDiscoverPath  string            `json:"auto_discover_path"yaml:"auto_discover_path"`
	KeyringPaths      string            `json:"keyring_paths"yaml:"keyring_paths"`
	CacheExpiration   jsonutil.Duration `json:"cache_expiration"yaml:"cache_expiration"`
	Plugins           *pluginConfig     `json:"plugins"yaml:"plugins"`
	ListenAddr        string            `json:"listen_addr,omitempty"yaml:"listen_addr"`
	ListenPort        int               `json:"listen_port,omitempty"yaml:"listen_port"`
}

const (
	CONFIG_CONSTRAINTS = `
			"control" : {
				"type": ["object", "null"],
				"properties": {
					"plugin_trust_level": {
						"type": "integer",
						"minimum": 0,
						"maximum": 2
					},
					"auto_discover_path": {
						"type": "string"
					},
					"cache_expiration": {
						"type": "string"
					},
					"max_running_plugins": {
						"type": "integer",
						"minimum": 1
					},
					"plugin_load_timeout": {
						"type": "integer",
						"minimum": 3,
						"maximum": 60
					},
					"keyring_paths" : {
						"type": "string"
					},
					"plugins": {
						"type": ["object", "null"],
						"properties" : {},
						"additionalProperties": true
					},
					"listen_addr": {
						"type": "string"
					},
					"listen_port": {
						"type": "integer"
					}
				},
				"additionalProperties": false
			}
	`
)

// get the default snapd configuration
func GetDefaultConfig() *Config {
	return &Config{
		ListenAddr:        defaultListenAddr,
		ListenPort:        defaultListenPort,
		MaxRunningPlugins: defaultMaxRunningPlugins,
		PluginLoadTimeout: defaultPluginLoadTimeout,
		PluginTrust:       defaultPluginTrust,
		AutoDiscoverPath:  defaultAutoDiscoverPath,
		KeyringPaths:      defaultKeyringPaths,
		CacheExpiration:   jsonutil.Duration{defaultCacheExpiration},
		Plugins:           newPluginConfig(),
	}
}

// UnmarshalJSON unmarshals valid json into a Config.  An example Config can be found
// at github.com/intelsdi-x/snap/blob/master/examples/configs/snap-config-sample.json
func (c *Config) UnmarshalJSON(data []byte) error {
	// construct a map of strings to json.RawMessages (to defer the parsing of individual
	// fields from the unmarshalled interface until later) and unmarshal the input
	// byte array into that map
	t := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	// loop through the individual map elements, parse each in turn, and set
	// the appropriate field in this configuration
	for k, v := range t {
		switch k {
		case "max_running_plugins":
			if err := json.Unmarshal(v, &(c.MaxRunningPlugins)); err != nil {
				return fmt.Errorf("%v (while parsing 'control::max_running_plugins')", err)
			}
		case "plugin_load_timeout":
			if err := json.Unmarshal(v, &(c.PluginLoadTimeout)); err != nil {
				return fmt.Errorf("%v (while parsing 'control::plugin_load_timeout')", err)
			}
		case "plugin_trust_level":
			if err := json.Unmarshal(v, &(c.PluginTrust)); err != nil {
				return fmt.Errorf("%v (while parsing 'control::plugin_trust_level')", err)
			}
		case "auto_discover_path":
			if err := json.Unmarshal(v, &(c.AutoDiscoverPath)); err != nil {
				return fmt.Errorf("%v (while parsing 'control::auto_discover_path')", err)
			}
		case "keyring_paths":
			if err := json.Unmarshal(v, &(c.KeyringPaths)); err != nil {
				return fmt.Errorf("%v (while parsing 'control::keyring_paths')", err)
			}
		case "cache_expiration":
			if err := json.Unmarshal(v, &(c.CacheExpiration)); err != nil {
				return fmt.Errorf("%v (while parsing 'control::cache_expiration')", err)
			}
		case "plugins":
			if err := json.Unmarshal(v, c.Plugins); err != nil {
				return err
			}
		case "listen_addr":
			if err := json.Unmarshal(v, &(c.ListenAddr)); err != nil {
				return err
			}
		case "listen_port":
			if err := json.Unmarshal(v, &(c.ListenPort)); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Unrecognized key '%v' in global config file while parsing 'control'", k)
		}
	}
	return nil
}

// NewPluginsConfig returns a map of *pluginConfigItems where the key is the plugin name.
func NewPluginsConfig() map[string]*pluginConfigItem {
	return map[string]*pluginConfigItem{}
}

// NewPluginConfigItem returns a *pluginConfigItem.
func NewPluginConfigItem() *pluginConfigItem {
	return &pluginConfigItem{
		cdata.NewNode(),
		map[int]*cdata.ConfigDataNode{},
	}
}

func newPluginTypeConfigItem() *pluginTypeConfigItem {
	return &pluginTypeConfigItem{
		make(map[string]*pluginConfigItem),
		cdata.NewNode(),
	}
}

func newPluginConfig() *pluginConfig {
	return &pluginConfig{
		All:         cdata.NewNode(),
		Collector:   newPluginTypeConfigItem(),
		Processor:   newPluginTypeConfigItem(),
		Publisher:   newPluginTypeConfigItem(),
		pluginCache: make(map[string]*cdata.ConfigDataNode),
	}
}

func (p *Config) GetPluginConfigDataNode(pluginType core.PluginType, name string, ver int) cdata.ConfigDataNode {
	return *p.Plugins.getPluginConfigDataNode(pluginType, name, ver)
}

func (p *Config) MergePluginConfigDataNode(pluginType core.PluginType, name string, ver int, cdn *cdata.ConfigDataNode) cdata.ConfigDataNode {
	p.Plugins.mergePluginConfigDataNode(pluginType, name, ver, cdn)
	return *p.Plugins.getPluginConfigDataNode(pluginType, name, ver)
}

func (p *Config) MergePluginConfigDataNodeAll(cdn *cdata.ConfigDataNode) cdata.ConfigDataNode {
	p.Plugins.mergePluginConfigDataNodeAll(cdn)
	return *p.Plugins.All
}

func (p *Config) DeletePluginConfigDataNodeField(pluginType core.PluginType, name string, ver int, fields ...string) cdata.ConfigDataNode {
	for _, field := range fields {
		p.Plugins.deletePluginConfigDataNodeField(pluginType, name, ver, field)
	}
	return *p.Plugins.getPluginConfigDataNode(pluginType, name, ver)
}

func (p *Config) DeletePluginConfigDataNodeFieldAll(fields ...string) cdata.ConfigDataNode {
	for _, field := range fields {
		p.Plugins.deletePluginConfigDataNodeFieldAll(field)
	}
	return *p.Plugins.All
}

func (p *Config) GetPluginConfigDataNodeAll() cdata.ConfigDataNode {
	return *p.Plugins.All
}

// UnmarshalJSON unmarshals valid json into pluginConfig.  An example Config
// github.com/intelsdi-x/snap/blob/master/examples/configs/snap-config-sample.json
func (p *pluginConfig) UnmarshalJSON(data []byte) error {
	t := map[string]interface{}{}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&t); err != nil {
		return err
	}

	//process the key value pairs for ALL plugins
	if v, ok := t["all"]; ok {
		jv, err := json.Marshal(v)
		if err != nil {
			return err
		}
		cdn := &cdata.ConfigDataNode{}
		dec = json.NewDecoder(bytes.NewReader(jv))
		dec.UseNumber()
		if err := dec.Decode(&cdn); err != nil {
			return err
		}
		p.All = cdn
	}

	//process the hierarchy of plugins
	for _, typ := range []string{"collector", "processor", "publisher"} {
		if err := unmarshalPluginConfig(typ, p, t); err != nil {
			return err
		}
	}

	return nil
}

func newPluginConfigItem(opts ...pluginConfigOpt) *pluginConfigItem {
	p := &pluginConfigItem{
		ConfigDataNode: cdata.NewNode(),
		Versions:       make(map[int]*cdata.ConfigDataNode),
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

type pluginConfigOpt func(*pluginConfigItem)

func optAddPluginConfigItem(key string, value ctypes.ConfigValue) pluginConfigOpt {
	return func(p *pluginConfigItem) {
		p.AddItem(key, value)
	}
}

func (p *pluginConfig) mergePluginConfigDataNodeAll(cdn *cdata.ConfigDataNode) {
	// clear cache
	p.pluginCache = make(map[string]*cdata.ConfigDataNode)

	p.All.Merge(cdn)
	return
}

func (p *pluginConfig) deletePluginConfigDataNodeFieldAll(key string) {
	// clear cache
	p.pluginCache = make(map[string]*cdata.ConfigDataNode)

	p.All.DeleteItem(key)
	return
}

func (p *pluginConfig) mergePluginConfigDataNode(pluginType core.PluginType, name string, ver int, cdn *cdata.ConfigDataNode) {
	// clear cache
	p.pluginCache = make(map[string]*cdata.ConfigDataNode)

	// merge new config into existing
	switch pluginType {
	case core.CollectorPluginType:
		if res, ok := p.Collector.Plugins[name]; ok {
			if res2, ok2 := res.Versions[ver]; ok2 {
				res2.Merge(cdn)
				return
			}
			res.Merge(cdn)
			return
		}
		if name != "" {
			cn := cdata.NewNode()
			cn.Merge(cdn)
			p.Collector.Plugins[name] = newPluginConfigItem()
			if ver > 0 {
				p.Collector.Plugins[name].Versions = map[int]*cdata.ConfigDataNode{ver: cn}
				return
			}
			p.Collector.Plugins[name].ConfigDataNode = cn
			return
		}
		p.Collector.All.Merge(cdn)
	case core.ProcessorPluginType:
		if res, ok := p.Processor.Plugins[name]; ok {
			if res2, ok2 := res.Versions[ver]; ok2 {
				res2.Merge(cdn)
				return
			}
			res.Merge(cdn)
			return
		}
		if name != "" {
			cn := cdata.NewNode()
			cn.Merge(cdn)
			p.Processor.Plugins[name] = newPluginConfigItem()
			if ver > 0 {
				p.Processor.Plugins[name].Versions = map[int]*cdata.ConfigDataNode{ver: cn}
				return
			}
			p.Processor.Plugins[name].ConfigDataNode = cn
			return
		}
		p.Processor.All.Merge(cdn)
	case core.PublisherPluginType:
		if res, ok := p.Publisher.Plugins[name]; ok {
			if res2, ok2 := res.Versions[ver]; ok2 {
				res2.Merge(cdn)
				return
			}
			res.Merge(cdn)
			return
		}
		if name != "" {
			cn := cdata.NewNode()
			cn.Merge(cdn)
			p.Publisher.Plugins[name] = newPluginConfigItem()
			if ver > 0 {
				p.Publisher.Plugins[name].Versions = map[int]*cdata.ConfigDataNode{ver: cn}
				return
			}
			p.Publisher.Plugins[name].ConfigDataNode = cn
			return
		}
		p.Publisher.All.Merge(cdn)
	}
}

func (p *pluginConfig) deletePluginConfigDataNodeField(pluginType core.PluginType, name string, ver int, key string) {
	// clear cache
	p.pluginCache = make(map[string]*cdata.ConfigDataNode)

	switch pluginType {
	case core.CollectorPluginType:
		if res, ok := p.Collector.Plugins[name]; ok {
			if res2, ok2 := res.Versions[ver]; ok2 {
				res2.DeleteItem(key)
				return
			}
			res.DeleteItem(key)
			return
		}
		p.Collector.All.DeleteItem(key)
	case core.ProcessorPluginType:
		if res, ok := p.Processor.Plugins[name]; ok {
			if res2, ok2 := res.Versions[ver]; ok2 {
				res2.DeleteItem(key)
				return
			}
			res.DeleteItem(key)
			return
		}
		p.Processor.All.DeleteItem(key)
	case core.PublisherPluginType:
		if res, ok := p.Publisher.Plugins[name]; ok {
			if res2, ok2 := res.Versions[ver]; ok2 {
				res2.DeleteItem(key)
				return
			}
			res.DeleteItem(key)
			return
		}
		p.Publisher.All.DeleteItem(key)
	}
}

func (p *pluginConfig) getPluginConfigDataNode(pluginType core.PluginType, name string, ver int) *cdata.ConfigDataNode {
	// check cache
	key := fmt.Sprintf("%d"+core.Separator+"%s"+core.Separator+"%d", pluginType, name, ver)
	if res, ok := p.pluginCache[key]; ok {
		return res
	}

	//todo process/interpolate values

	p.pluginCache[key] = cdata.NewNode()
	p.pluginCache[key].Merge(p.All)

	// check for plugin config
	switch pluginType {
	case core.CollectorPluginType:
		p.pluginCache[key].Merge(p.Collector.All)
		if res, ok := p.Collector.Plugins[name]; ok {
			p.pluginCache[key].Merge(res.ConfigDataNode)
			if res2, ok2 := res.Versions[ver]; ok2 {
				p.pluginCache[key].Merge(res2)
			}
		}
	case core.ProcessorPluginType:
		p.pluginCache[key].Merge(p.Processor.All)
		if res, ok := p.Processor.Plugins[name]; ok {
			p.pluginCache[key].Merge(res.ConfigDataNode)
			if res2, ok2 := res.Versions[ver]; ok2 {
				p.pluginCache[key].Merge(res2)
			}
		}
	case core.PublisherPluginType:
		p.pluginCache[key].Merge(p.Publisher.All)
		if res, ok := p.Publisher.Plugins[name]; ok {
			p.pluginCache[key].Merge(res.ConfigDataNode)
			if res2, ok2 := res.Versions[ver]; ok2 {
				p.pluginCache[key].Merge(res2)
			}
		}
	}

	log.WithFields(log.Fields{
		"_block_":            "getPluginConfigDataNode",
		"_module":            "config",
		"config-cache-key":   key,
		"config-cache-value": p.pluginCache[key],
	}).Debug("Getting plugin config")

	return p.pluginCache[key]
}

func unmarshalPluginConfig(typ string, p *pluginConfig, t map[string]interface{}) error {
	if v, ok := t[typ]; ok {
		switch plugins := v.(type) {
		case map[string]interface{}:
			for name, c := range plugins {
				if name == "all" {
					jv, err := json.Marshal(c)
					if err != nil {
						return err
					}
					cdn := cdata.NewNode()
					dec := json.NewDecoder(bytes.NewReader(jv))
					dec.UseNumber()
					if err := dec.Decode(&cdn); err != nil {
						return err
					}
					switch typ {
					case "collector":
						p.Collector.All = cdn
					case "processor":
						p.Processor.All = cdn
					case "publisher":
						p.Publisher.All = cdn
					}
					continue
				}
				switch typ {
				case "collector":
					p.Collector.Plugins[name] = newPluginConfigItem()
				case "processor":
					p.Processor.Plugins[name] = newPluginConfigItem()
				case "publisher":
					p.Publisher.Plugins[name] = newPluginConfigItem()
				}
				switch col := c.(type) {
				case map[string]interface{}:
					if v, ok := col["all"]; ok {
						jv, err := json.Marshal(v)
						if err != nil {
							return err
						}
						cdn := cdata.NewNode()
						dec := json.NewDecoder(bytes.NewReader(jv))
						dec.UseNumber()
						if err := dec.Decode(&cdn); err != nil {
							return err
						}
						switch typ {
						case "collector":
							p.Collector.Plugins[name].ConfigDataNode = cdn
						case "processor":
							p.Processor.Plugins[name].ConfigDataNode = cdn
						case "publisher":
							p.Publisher.Plugins[name].ConfigDataNode = cdn
						}
					}
					if vs, ok := col["versions"]; ok {
						switch versions := vs.(type) {
						case map[string]interface{}:
							for ver, version := range versions {
								switch v := version.(type) {
								case map[string]interface{}:
									jv, err := json.Marshal(v)
									if err != nil {
										return err
									}
									cdn := cdata.NewNode()
									dec := json.NewDecoder(bytes.NewReader(jv))
									dec.UseNumber()
									if err := dec.Decode(&cdn); err != nil {
										return err
									}
									ver, err := strconv.Atoi(ver)
									if err != nil {
										return err
									}
									switch typ {
									case "collector":
										p.Collector.Plugins[name].Versions[ver] = cdn
									case "processor":
										p.Processor.Plugins[name].Versions[ver] = cdn
									case "publisher":
										p.Publisher.Plugins[name].Versions[ver] = cdn
									}
								default:
									return fmt.Errorf("Error unmarshalling %v'%v' expected '%v' got '%v'", typ, name, map[string]interface{}{}, reflect.TypeOf(v))
								}
							}

						default:
							return fmt.Errorf("Error unmarshalling %v '%v' expected '%v' got '%v'", typ, name, map[string]interface{}{}, reflect.TypeOf(versions))
						}
					}
				default:
					return fmt.Errorf("Error unmarshalling %v '%v' expected '%v' got '%v'", typ, name, map[string]interface{}{}, reflect.TypeOf(col))
				}
			}
		default:
			return fmt.Errorf("Error unmarshalling %v expected '%v' got '%v'", typ, map[string]interface{}{}, reflect.TypeOf(plugins))
		}
	}
	return nil
}
