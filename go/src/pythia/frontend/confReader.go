package frontend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

//Conf struct for the config.json file
type Conf struct {
	IP []string
}

var myConf Conf

//GetConf that was or will be loaded from the config.json file
func GetConf() Conf {
	//check if not initialized
	if cap(myConf.IP) == 0 {
		fmt.Println("ip = 0")
		myConf = LoadConf()
	}
	fmt.Println("sending conf")
	return myConf
}

//LoadConf to retreive the data from the conf.json file
func LoadConf() Conf {
	var conf Conf
	//Problem with the congif.json file: it must be in the go/bin/ directory to be loaded
	c, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Println("Could not open conf file, default configuration loaded")
		return LoadDefaultConf()
	}

	if err := json.Unmarshal(c, &conf); err != nil {
		log.Println("Could not retreive data from conf file default configuration loaded")
		return LoadDefaultConf()
	}

	return conf
}

func LoadDefaultConf() Conf {
	conf := Conf{[]string{"127.0.0.1", "::1"}}
	return conf
}
