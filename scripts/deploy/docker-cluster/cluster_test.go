// Copyright 2024 Amiasys Corporation and/or its affiliates. All rights reserved.

package cluster_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

const n = 100

type ASNC struct {
	Log         Log         `yaml:"log"`
	Mongodb     DB          `yaml:"mongodb"`
	Influx      DB          `yaml:"influxdbv1"`
	Iam         Iam         `yaml:"iam"`
	Grpc        GRPC        `yaml:"grpc"`
	Network     Network     `yaml:"network"`
	ServiceNode ServiceNode `yaml:"servicenode"`
	Service     Service     `yaml:"service"`
}

type Log struct {
	Demo   bool      `yaml:"demo"`
	Path   string    `yaml:"path"`
	Prefix string    `yaml:"prefix"`
	ALog   LogConfig `yaml:"api_log"`
	RLog   LogConfig `yaml:"runtime_log"`
	ELog   LogConfig `yaml:"entity_log"`
	PLog   LogConfig `yaml:"perf_log"`
}

type LogConfig struct {
	FileName string `yaml:"filename"`
	Level    string `yaml:"level"`
}

type DB struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Database string `yaml:"database_name"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Iam struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Network struct {
	Id       string `yaml:"id"`
	TopoFile string `yaml:"topo_file"`
}

type GRPC struct {
	Port uint64 `yaml:"port"`
}

type ServiceNode struct {
	KeepAlive int `yaml:"keepalive"`
}

type Service struct {
	Dir       string               `yaml:"dir"`
	Supported map[string]MyService `yaml:"supported"`
}

type MyService struct {
	Min string `yaml:"min"`
	Max string `yaml:"max"`
}

type MyNetwork struct {
	NetworkID   string     `json:"network_id"`
	NetworkName string     `json:"network_name"`
	Topology    []Topology `json:"topology"`
}

type Topology struct {
	NodeName       string   `json:"node_name"`
	NodeType       string   `json:"nodeType"`
	Location       Location `json:"location"`
	Label          string   `json:"label"`
	ExternalLinked []string `json:"external_linked"`
	SubNodes       []Node   `json:"sub_nodes"`
}

type Location struct {
	Coordinates Coordinate `json:"coordinates"`
	Address     string     `json:"address"`
}

type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Node struct {
	NodeName       string   `json:"node_name"`
	NodeType       string   `json:"nodeType"`
	Location       Location `json:"location"`
	Label          string   `json:"label"`
	ExternalLinked []string `json:"external_linked"`
	InternalLinked []string `json:"internal_linked"`
}

type asncDocker struct {
	Services map[string]DockerService `yaml:"services"`
	Volumes  map[string]Volume        `yaml:"volumes,omitempty"`
}

type DockerService struct {
	ContainerName string            `yaml:"container_name,omitempty"`
	Image         string            `yaml:"image,omitempty"`
	Restart       string            `yaml:"restart,omitempty"`
	Ulimits       map[string]int    `yaml:"ulimits,omitempty"`
	Environment   map[string]string `yaml:"environment,omitempty"`
	Ports         []string          `yaml:"ports,omitempty"`
	Volumes       []string          `yaml:"volumes,omitempty"`
	Command       string            `yaml:"command,omitempty"`
	DependsOn     []string          `yaml:"depends_on,omitempty"`
}

type Volume struct {
	Driver string `yaml:"driver,omitempty"`
}

type ASNSN struct {
	Log        Log               `yaml:"log"`
	General    General           `yaml:"general"`
	Controller Controller        `yaml:"controller"`
	Tsdb       TSDB              `yaml:"tsdb"`
	Service    SNService         `yaml:"service"`
	NetIf      map[string]string `yaml:"netif"`
}

type General struct {
	Mode            string `yaml:"mode"`
	ID              string `yaml:"id"`
	NetworkPath     string `yaml:"network_path"`
	NodeName        string `yaml:"node_name"`
	Type            string `yaml:"type"`
	NetworkCapacity int    `yaml:"network_capacity"`
	CliPort         int    `yaml:"cli_port"`
}

type Controller struct {
	IP            string `yaml:"ip"`
	Port          int    `yaml:"port"`
	RetryInterval int    `yaml:"retry_interval"`
}

type TSDB struct {
	Type     string `yaml:"type"`
	Name     string `yaml:"name"`
	IP       string `yaml:"ip"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type SNService struct {
	Dir           string `yaml:"dir"`
	ConfigTimeout int    `yaml:"config_timeout"`
}

func TestGenerateCluster(t *testing.T) {
	// generate controller file
	err := os.MkdirAll("controller", 0755)
	err = os.MkdirAll("controller/config", 0755)
	err = os.MkdirAll("controller/log", 0755)
	if err != nil {
		panic(err)
	}

	asnConf := ASNC{
		Log: Log{
			Demo:   true,
			Path:   "./log",
			Prefix: "asn",
			ALog: LogConfig{
				FileName: "api.log",
				Level:    "info",
			},
			RLog: LogConfig{
				FileName: "runtime.log",
				Level:    "info",
			},
			ELog: LogConfig{
				FileName: "entity.log",
				Level:    "info",
			},
			PLog: LogConfig{
				FileName: "perf.log",
				Level:    "info",
			},
		},
		Mongodb: DB{
			Host:     "172.17.0.1",
			Port:     "27017",
			Database: "asn",
			Username: "amia",
			Password: "2022",
		},
		Influx: DB{
			Host:     "172.17.0.1",
			Port:     "8086",
			Database: "asn",
			Username: "amia",
			Password: "2022",
		},
		Iam: Iam{
			Host: "localhost",
			Port: "17930",
		},
		Grpc: GRPC{50051},
		Network: Network{
			Id:       "network1",
			TopoFile: "./config/100nodes-topology.json",
		},
		ServiceNode: ServiceNode{3},
		Service: Service{
			Dir: "./plugins/",
			Supported: map[string]MyService{
				"myservice": {Min: "v2.2.0", Max: "v2.2.0"},
			},
		},
	}

	asnYaml, err := yaml.Marshal(asnConf)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("controller/config/asn.conf", asnYaml, 0644)
	if err != nil {
		panic(err)
	}

	cliConf := map[string]string{
		"server": "localhost",
		"port":   "50051",
		"token":  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NDcwMTk0NDUsInVzZXJuYW1lIjoiYXNuLXN1cGVydmlzb3IifQ.8UlBi9qlL3NxXYllKp3NN2WUBwSs4Q1sqKvfMk3MRwI",
	}
	cliYaml, err := yaml.Marshal(cliConf)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("controller/config/cli.conf", cliYaml, 0644)
	if err != nil {
		panic(err)
	}

	// You can decide the topology network
	network := MyNetwork{
		NetworkID:   "network1",
		NetworkName: "Network with 100 nodes",
		Topology:    []Topology{},
	}

	for i := 1; i <= n; i++ {
		location := Location{
			Coordinates: Coordinate{
				Latitude:  -90.0 + rand.Float64()*180,
				Longitude: -180.0 + rand.Float64()*360,
			},
			Address: fmt.Sprintf("%d street", i),
		}
		network.Topology = append(network.Topology, Topology{
			NodeName:       fmt.Sprintf("node%d", i),
			NodeType:       "networkNode",
			Location:       location,
			Label:          "CORE",
			ExternalLinked: []string{},
			SubNodes: []Node{{
				NodeName:       fmt.Sprintf("switch%d", i),
				NodeType:       "switch",
				Location:       location,
				Label:          "CORE",
				ExternalLinked: []string{},
				InternalLinked: []string{},
			}},
		})
	}

	bytes, err := json.MarshalIndent(network, "", "  ")
	if err != nil {
		panic(err)
	}

	fileName := "controller/config/100nodes-topology.json"
	err = os.WriteFile(fileName, bytes, 0644)
	if err != nil {
		panic(err)
	}

	asncD := asncDocker{
		Services: map[string]DockerService{},
		Volumes: map[string]Volume{
			"influxdb_data": {Driver: "local"},
		},
	}
	asncD.Services["asn-mdb"] = DockerService{
		ContainerName: "asn-mdb",
		Image:         "mongo:7.0",
		Restart:       "always",
		Ulimits: map[string]int{
			"nofile": 100000,
		},
		Environment: map[string]string{
			"MONGO_INITDB_ROOT_USERNAME": "amia",
			"MONGO_INITDB_ROOT_PASSWORD": "2022",
		},
		Ports:   []string{"27017:27017"},
		Volumes: []string{"./data/:/data/db"},
		Command: "--bind_ip_all --auth",
	}
	asncD.Services["asn-idb"] = DockerService{
		ContainerName: "asn-idb",
		Image:         "influxdb:1.11.8",
		Ports:         []string{"8086:8086"},
		Environment: map[string]string{
			"INFLUXDB_DB":             "asn",
			"INFLUXDB_ADMIN_USER":     "amia",
			"INFLUXDB_ADMIN_PASSWORD": "2022",
			"INFLUXDB_USER":           "amia",
			"INFLUXDB_USER_PASSWORD":  "2022",
		},
	}
	asncD.Services["asnc"] = DockerService{
		Image:     "registry.amiasys.com/asnc:v25.0.0",
		Restart:   "always",
		DependsOn: []string{"asn-mdb", "asn-idb"},
		Ports:     []string{"50051:50051"},
		Volumes:   []string{"./cert/:/asn/cert/", "./config/:/asn/config/", "./log/:/asn/log/"},
	}

	asncYaml, err := yaml.Marshal(asncD)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("controller/asnc.yml", asncYaml, 0644)
	if err != nil {
		panic(err)
	}

	// make service node file
	err = os.MkdirAll("servicenode", 0755)
	if err != nil {
		panic(err)
	}

	for i := 1; i <= n; i++ {
		fileName = fmt.Sprintf("sn%d", i)
		err = os.MkdirAll("servicenode/"+fileName, 0755)
		err = os.MkdirAll(fmt.Sprintf("servicenode/%s/config", fileName), 0755)
		err = os.MkdirAll(fmt.Sprintf("servicenode/%s/log", fileName), 0755)
		if err != nil {
			panic(err)
		}

		asnC := ASNSN{
			Log: Log{
				Path:   "./log",
				Prefix: "asn",
				ALog: LogConfig{
					FileName: "api.log",
					Level:    "info",
				},
				RLog: LogConfig{
					FileName: "runtime.log",
					Level:    "info",
				},
				ELog: LogConfig{
					FileName: "entity.log",
					Level:    "info",
				},
				PLog: LogConfig{
					FileName: "perf.log",
					Level:    "info",
				},
			},
			General: General{
				Mode:            "cluster",
				ID:              "",
				NetworkPath:     fmt.Sprintf("network1.node%d.switch%d", i, i),
				NodeName:        fmt.Sprintf("switch%d", i),
				Type:            "server",
				NetworkCapacity: 1024,
				CliPort:         50052,
			},
			Controller: Controller{
				IP:            "172.17.0.1",
				Port:          50051,
				RetryInterval: 5,
			},
			Tsdb: TSDB{
				Type:     "influxdbv1",
				Name:     "asn-dev",
				IP:       "172.17.0.1",
				Port:     8086,
				Username: "amia",
				Password: "2022",
			},
			Service: SNService{
				Dir:           "./plugins/",
				ConfigTimeout: 20,
			},
			NetIf: map[string]string{
				"data":       "eth0",
				"control":    "eth0",
				"management": "eth0",
			},
		}
		asnCFYaml, err := yaml.Marshal(asnC)
		if err != nil {
			panic(err)
		}
		err = os.WriteFile(fmt.Sprintf("servicenode/%s/config/asn.conf", fileName), asnCFYaml, 0644)
		if err != nil {
			panic(err)
		}

		asnD := asncDocker{
			Services: map[string]DockerService{
				"asnsn": {
					Image:         "registry.amiasys.com/asnsn:v25.0.0",
					ContainerName: fmt.Sprintf("network-node%d-switch%d", i, i),
					Restart:       "always",
					Volumes:       []string{"./config/:/asn/config/", "./log/:/asn/log/"},
				}},
		}
		asnDY, err := yaml.Marshal(asnD)
		if err != nil {
			panic(err)
		}
		err = os.WriteFile(fmt.Sprintf("servicenode/%s/asnsn.yml", fileName), asnDY, 0644)
		if err != nil {
			panic(err)
		}
	}

	shellUp := `#!/bin/bash
# 进入 controller 文件夹并启动 Docker Compose
cd controller || { echo "Failed to enter controller folder"; exit 1; }
echo "Starting Docker Compose in controller folder..."
docker compose -f asnc.yml up -d || { echo "Failed to execute docker compose in controller folder"; exit 1; }
echo "Docker Compose started in controller folder."
cd - || { echo "Failed to return to the previous directory"; exit 1; }
# 进入 servicenode 文件夹并逐个启动 sn1 到 sn100
cd servicenode || { echo "Failed to enter servicenode folder"; exit 1; }
for i in $(seq 1 100); do
folder="sn$i"
if [ -d "$folder" ]; then
cd "$folder" || { echo "Failed to enter $folder folder"; exit 1; }
echo "Starting Docker Compose in $folder..."
docker compose -f asnsn.yml up -d || { echo "Failed to execute docker compose in $folder"; exit 1; }
echo "Docker Compose started in $folder."
cd - >/dev/null || { echo "Failed to return to servicenode folder"; exit 1; }
else
echo "Folder $folder does not exist, skipping."
fi
done
echo "All tasks completed."`

	if err := os.WriteFile("up.sh", []byte(shellUp), 0755); err != nil {
		panic(err)
	}

	shellDown := `#!/bin/bash
cd controller || { echo "Failed to enter controller folder"; exit 1; }
echo "Starting Docker Compose in controller folder..."
docker compose -f asnc.yml down || { echo "Failed to execute docker compose in controller folder"; exit 1; }
echo "Docker Compose started in controller folder."
cd - || { echo "Failed to return to the previous directory"; exit 1; }
cd servicenode || { echo "Failed to enter servicenode folder"; exit 1; }
for i in $(seq 1 100); do
folder="sn$i"
if [ -d "$folder" ]; then
cd "$folder" || { echo "Failed to enter $folder folder"; exit 1; }
echo "Starting Docker Compose in $folder..."
docker compose -f asnsn.yml down || { echo "Failed to execute docker compose in $folder"; exit 1; }
echo "Docker Compose started in $folder."
cd - >/dev/null || { echo "Failed to return to servicenode folder"; exit 1; }
else
echo "Folder $folder does not exist, skipping."
fi
done
echo "All tasks completed."`
	if err := os.WriteFile("down.sh", []byte(shellDown), 0755); err != nil {
		panic(err)
	}
}
