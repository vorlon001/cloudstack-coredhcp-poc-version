package main

import (
	//"bytes"
	"fmt"
	//"os/exec"
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
	Task	*Task
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
	Uuid	string
	ListenName	string
        Listen  *VLAN
	DhcpListenRunner	DhcpListenRunner
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

func main() {

        config, err := Loader("demo3.yaml")
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
