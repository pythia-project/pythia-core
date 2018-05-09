package frontend

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

//Conf struct for the conf.json file
type Conf struct {
	IP []string
}

//GetConf to retreive the data from the conf.json file
func GetConf() Conf {
	var conf Conf
	c, err := ioutil.ReadFile("conf.json")
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
	conf := Conf{[]string{"127.0.0.1"}}
	//append(Conf.IP, "127.0.0.1")
	return conf
}
