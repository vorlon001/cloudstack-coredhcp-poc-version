// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// This is a generated file, edits should be made in the corresponding source file
// And this file regenerated using `coredhcp-generator --from core-plugins.txt`
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/coredhcp/coredhcp/config"
	"github.com/coredhcp/coredhcp/logger"
	"github.com/coredhcp/coredhcp/server"

	"github.com/coredhcp/coredhcp/plugins"
	pl_autoconfigure "github.com/coredhcp/coredhcp/plugins/autoconfigure"
	pl_dns "github.com/coredhcp/coredhcp/plugins/dns"
	pl_file "github.com/coredhcp/coredhcp/plugins/file"
	pl_ipv6only "github.com/coredhcp/coredhcp/plugins/ipv6only"
	pl_leasetime "github.com/coredhcp/coredhcp/plugins/leasetime"
	pl_mtu "github.com/coredhcp/coredhcp/plugins/mtu"
	pl_nbp "github.com/coredhcp/coredhcp/plugins/nbp"
	pl_netmask "github.com/coredhcp/coredhcp/plugins/netmask"
	pl_prefix "github.com/coredhcp/coredhcp/plugins/prefix"
	pl_range "github.com/coredhcp/coredhcp/plugins/range"
	pl_router "github.com/coredhcp/coredhcp/plugins/router"
	pl_searchdomains "github.com/coredhcp/coredhcp/plugins/searchdomains"
	pl_serverid "github.com/coredhcp/coredhcp/plugins/serverid"
	pl_sleep "github.com/coredhcp/coredhcp/plugins/sleep"
	pl_staticroute "github.com/coredhcp/coredhcp/plugins/staticroute"
        pl_example "github.com/coredhcp/coredhcp/plugins/example"

	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"

        "runtime"
        "time"
        "sync"
        "io/ioutil"
        "github.com/google/uuid"
        "gopkg.in/yaml.v3"
)



func PanicRecover(task *Task) {
        if r := recover(); r != nil {
                fmt.Printf("Internal error: %v", r)
                buf := make([]byte, 1<<16)
                stackSize := runtime.Stack(buf, true)
                fmt.Printf("--------------------------------------------------------------------------------")
                fmt.Printf(fmt.Sprintf("Internal error: %s\n", string(buf[0:stackSize])))
                fmt.Printf("--------------------------------------------------------------------------------")
        }
        task.ErrChannel <- &Error{Listen: task.ListenName, Message: fmt.Sprintf("Error Job PanicRecover. %v", task.Listen), Task: task}
}


type Config struct {
        Config map[string]VLAN
}


type VLAN struct {
        Server4 Server4 `yaml:"server4"`
}

type Server4 struct {
        Listen  []string `yaml:"listen"`
        Plugins []Plugin `json:"plugins"`
}


type Plugin struct {
        LeaseTime     string     `yaml:"lease_time,omitempty"`
        ServerID      string     `yaml:"server_id,omitempty"`
        File          string     `yaml:"file,omitempty"`
        DNS           string     `yaml:"dns,omitempty"`
        MTU           int64      `yaml:"mtu,omitempty"`
        Searchdomains []string    `yaml:"searchdomains,omitempty"`
        Router        string     `yaml:"router,omitempty"`
        Netmask       string     `yaml:"netmask,omitempty"`
        Range         string     `yaml:"range,omitempty"`
        Staticroute   string     `yaml:"staticroute,omitempty"`
}

type Error struct {
        Listen  string
        Message string
        Task    *Task
}

func (e Error) Error() string {
        return e.Message
}


type DhcpListenRunner func (config []byte) error

func dhcpListenRunner(config []byte) error{

        fmt.Printf("RUN dhcpListenRunner %v\n", config)
        time.Sleep(10 * time.Second) // only here so the time import is needed
        fmt.Printf("End dhcpListenRunner %v\n", config)
        return nil

}


type Task struct {
        Uuid    string
        ListenName      string
        Listen  *VLAN
        DhcpListenRunner        DhcpListenRunner
        ErrChannel chan *Error
        Wg *sync.WaitGroup
}



func connect(task *Task) {

        defer PanicRecover(task)
        defer task.Wg.Done()

        fmt.Printf(">>>>>>>>%v\n",task.Listen)
        yamlData, err := yaml.Marshal(task.Listen)

        if err != nil {
                task.ErrChannel <- &Error{Listen: task.ListenName, Message: fmt.Sprintf("Error while Marshaling. %v", task.Listen), Task: task}
        }

        fmt.Println(" --- YAML ---")
        fmt.Println(string(yamlData))

        err = task.DhcpListenRunner(yamlData)
        if err != nil {
                task.ErrChannel <- &Error{Listen: task.ListenName, Message: fmt.Sprintf("Errored on goroutine %v", task.Listen), Task: task}
        }

        fmt.Printf("%s: RUN\n", task.Listen)
        time.Sleep(time.Second * 2)
        fmt.Printf("%s: DONE\n", task.Listen)
}

func listener(task chan *Task) {
    for {
        listen, ok := <-task
        // check channel is closed or not
        if !ok{
            break
        }

        go connect(listen)
        listen.Wg.Add(1)

    }

}


func errorEventListener(errChannel chan *Error, task chan *Task) {
    for {
        event := <-errChannel
        fmt.Printf("%v\n", event)
        go connect(event.Task)
    }

}

func Loader(config string) (*Config, error){
        var blog Config;

        yamlExample, err := ioutil.ReadFile(config)
        if err != nil {
                return nil, err
        }

        fmt.Printf("File contents: %s", yamlExample)

        err = yaml.Unmarshal(yamlExample, &blog)
        if err != nil {
                return nil, err
        }

        return &blog, nil
}


var (
	flagLogFile     = flag.StringP("logfile", "l", "", "Name of the log file to append to. Default: stdout/stderr only")
	flagLogNoStdout = flag.BoolP("nostdout", "N", false, "Disable logging to stdout/stderr")
	flagLogLevel    = flag.StringP("loglevel", "L", "info", fmt.Sprintf("Log level. One of %v", getLogLevels()))
	flagConfig      = flag.StringP("conf", "c", "", "Use this configuration file instead of the default location")
	flagPlugins     = flag.BoolP("plugins", "P", false, "list plugins")
)

var logLevels = map[string]func(*logrus.Logger){
	"none":    func(l *logrus.Logger) { l.SetOutput(io.Discard) },
	"debug":   func(l *logrus.Logger) { l.SetLevel(logrus.DebugLevel) },
	"info":    func(l *logrus.Logger) { l.SetLevel(logrus.InfoLevel) },
	"warning": func(l *logrus.Logger) { l.SetLevel(logrus.WarnLevel) },
	"error":   func(l *logrus.Logger) { l.SetLevel(logrus.ErrorLevel) },
	"fatal":   func(l *logrus.Logger) { l.SetLevel(logrus.FatalLevel) },
}

func getLogLevels() []string {
	var levels []string
	for k := range logLevels {
		levels = append(levels, k)
	}
	return levels
}

var desiredPlugins = []*plugins.Plugin{
        &pl_autoconfigure.Plugin,
        &pl_dns.Plugin,
        &pl_file.Plugin,
        &pl_ipv6only.Plugin,
        &pl_leasetime.Plugin,
        &pl_mtu.Plugin,
        &pl_nbp.Plugin,
        &pl_netmask.Plugin,
        &pl_prefix.Plugin,
        &pl_range.Plugin,
        &pl_router.Plugin,
        &pl_searchdomains.Plugin,
        &pl_serverid.Plugin,
        &pl_sleep.Plugin,
        &pl_staticroute.Plugin,
        &pl_example.Plugin,
}


func RunJob(configByte []byte, logs *logrus.Entry) {

        config, err := config.Load("yml", configByte, *flagConfig)
        if err != nil {
                logs.Fatalf("Failed to load configuration: %v", err)
        }
        // register plugins

	var RegisteredPlugins = make(map[string]*plugins.Plugin)

        for _, plugin := range desiredPlugins {
                if err := plugins.RegisterPlugin(RegisteredPlugins, plugin); err != nil {
                        logs.Fatalf("Failed to register plugin '%s': %v", plugin.Name, err)
                }
        }

        // start server
        srv, err := server.Start("vlan200", config, RegisteredPlugins)
        if err != nil {
                logs.Fatal(err)
        }
        if err := srv.Wait(); err != nil {
                logs.Error(err)
        }
}

/*
func main() {
	flag.Parse()

	if *flagPlugins {
		for _, p := range desiredPlugins {
			fmt.Println(p.Name)
		}
		os.Exit(0)
	}

	log := logger.GetLogger("main")
	fn, ok := logLevels[*flagLogLevel]
	if !ok {
		log.Fatalf("Invalid log level '%s'. Valid log levels are %v", *flagLogLevel, getLogLevels())
	}
	fn(log.Logger)
	log.Infof("Setting log level to '%s'", *flagLogLevel)
	if *flagLogFile != "" {
		log.Infof("Logging to file %s", *flagLogFile)
		logger.WithFile(log, *flagLogFile)
	}
	if *flagLogNoStdout {
		log.Infof("Disabling logging to stdout/stderr")
		logger.WithNoStdOutErr(log)
	}


        var configByte = []byte(`
server4:
    listen:
        - '%vlan200'
    plugins:
        - lease_time: 3600s
        - server_id: 192.168.1.10
        - file: file_leases.txt
        - mtu: 1500
        - dns: 192.168.1.10
        - router: 192.168.200.1
        - netmask: 255.255.255.255
        - range: leases.txt 192.168.200.10 192.168.200.50 60s
        - staticroute: 10.20.20.0/24,192.168.200.1 0.0.0.0/0,192.168.200.1
`)


	fmt.Printf("%T\n",log)
	RunJob(configByte, log)

}
*/


func main() {
        flag.Parse()

        if *flagPlugins {
                for _, p := range desiredPlugins {
                        fmt.Println(p.Name)
                }
                os.Exit(0)
        }

        log := logger.GetLogger("main")
        fn, ok := logLevels[*flagLogLevel]
        if !ok {
                log.Fatalf("Invalid log level '%s'. Valid log levels are %v", *flagLogLevel, getLogLevels())
        }
        fn(log.Logger)
        log.Infof("Setting log level to '%s'", *flagLogLevel)
        if *flagLogFile != "" {
                log.Infof("Logging to file %s", *flagLogFile)
                logger.WithFile(log, *flagLogFile)
        }
        if *flagLogNoStdout {
                log.Infof("Disabling logging to stdout/stderr")
                logger.WithNoStdOutErr(log)
        }

        config, err := Loader(*flagConfig)
        if err != nil {
                panic(err)
        }

        for i, v := range config.Config {
                fmt.Printf("%v = %v\n", i, v)
        }


        var wg sync.WaitGroup

        errChannel := make(chan *Error, 1)

        var task = make(chan *Task)

        go listener(task)
        go errorEventListener(errChannel, task)

        for i, v := range config.Config {
                taskListen := Task{ListenName: i, Listen: &v, Uuid: uuid.NewString(), ErrChannel: errChannel, Wg: &wg, DhcpListenRunner: dhcpListenRunner}
                task <- &taskListen
        }

        wg.Wait()
        close(task)

}

