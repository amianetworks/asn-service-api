// Copyright 2024 Amiasys Corporation and/or its affiliates. All rights reserved.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

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
	NetworkMode   string            `yaml:"network_mode,omitempty"`
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

func main() {
	var n int
	var err error
	if len(os.Args) != 2 {
		log.Println("args error, using default value: 100")
		n = 1
	} else {
		n, err = strconv.Atoi(os.Args[1])
		if err != nil {
			log.Println("args error, using default value: 100")
			n = 1
		}
	}

	// generate controller file
	err = os.MkdirAll("controller", 0755)
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
			Host:     "localhost",
			Port:     "27017",
			Database: "asn",
			Username: "amia",
			Password: "2022",
		},
		Influx: DB{
			Host:     "localhost",
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
			"ldap_slap":     {Driver: "local"},
			"ldap_data":     {Driver: "local"},
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
	asncD.Services["sapphire-ldap"] = DockerService{
		ContainerName: "sapphire-ldap",
		Image:         "registry.amiasys.com/iam:v1.1.0",
		Restart:       "always",
		Ports:         []string{"18389:389/udp", "18389:389/tcp"},
		Volumes:       []string{"ldap_slap:/etc/ldap/slapd.d/", "ldap_data:/var/lib/ldap/"},
		Command:       `sh -c "/etc/init.d/slapd start && tail -f /dev/null"`,
	}
	asncD.Services["sapphire-iam"] = DockerService{
		ContainerName: "sapphire-iam",
		Image:         "registry.amiasys.com/sapphire.iam:v25.0.2",
		Restart:       "always",
		Ports:         []string{"17930:17930", "17931:17931"},
		Volumes:       []string{"./config/:/usr/local/sapphire/"},
		DependsOn:     []string{"sapphire-ldap"},
	}
	asncD.Services["asnc"] = DockerService{
		Image:       "registry.amiasys.com/asnc:v25.0.13",
		Restart:     "always",
		DependsOn:   []string{"asn-mdb", "asn-idb", "sapphire-ldap", "sapphire-iam"},
		NetworkMode: "host",
		Volumes:     []string{"./cert/:/asn/cert/", "./config/:/asn/config/", "./log/:/asn/log/", "./plugins/:/asn/plugins/"},
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

	err = os.MkdirAll("servicenode/plugins", 0755)
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
					Image:         "registry.amiasys.com/asnsn:v25.0.8",
					ContainerName: fmt.Sprintf("network-node%d-switch%d", i, i),
					Restart:       "always",
					Volumes:       []string{"./config/:/asn/config/", "./log/:/asn/log/", "../plugins/:/asn/plugins/"},
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

	ymlIam := `###################### IAM Configurations #####################
## This file contains the following parts：
## Part1: General Configurations
##   This part contains basic configuration information: log level and log file
## Part2: API Configurations
##   This part mainly configures the ldap service provided to the outside world,
##   including configurations such as port, encryption, and synchronization.
## Part3: Account Configurations
##   This part mainly configures account information.
## Part4: Authentication Configurations
##   This part is mainly used for account authentication and other information, such as limiting login frequency, device binding, and mfa, etc.
## Part5: Authorization Configurations
##   This part of the configuration mainly manages user authentication related configurations, such as token, role, etc.
#
## General Configurations
## Supported log level: panic | fatal | error | warning | info | debug | trace. Default: info
## If the log level is not configured or is configured to an undefined value, the default info level is used.
## The default log file is "/var/log/iam/iam.log". If not configured, the default path will be used.
#general:
#  loglevel: "info"
#  logfile: "/var/log/iam/iam.log"
#
## API Configurations
api:
  #  grpc:
  #    port: 17930 # gRPC API port. Default:17930
  #    tls: true
  ldap: # LDAP Service
    #    # The API Service enable to provide LDAP Server as an exposed service. Default: false
    #    # If set to false, the LDAP service will not be exposed externally.
    #    # If set to true, the LDAP service will be exposed externally by port 18389.
    #    api_service: false
    #    tls: true
    #    # The bind dn is used for binding (signing on) to the LDAP server.
    #    bind_dn: "cn=admin,dc=asn,dc=vpn,dc=com"
    #    credentials: "@ASN2021!" # Password of the bind dn.
    host: 172.17.0.1
    port: 18389
#
#    # LDAP Server to sync up with.
#    external_ldap:
#      # LDAP URL is a string that can be used to encapsulate the address and port of a directory serve.
#      # Example: ldap://example.ldap.com. If url is not configured, ldap synchronization will not be performed.
#      url: ""
#      bind_dn: "cn=example,dc=com"
#      credentials: "password"
#      # Supported sync mode: inbound|outbound|bidirectional. Default: inbound
#      # If sync mode is not configured, the default inbound mode will be used.
#      # In the inbound synchronization mode, data is synchronized into the IAM LDAP server from the external LDAP server.
#      # In the outbound synchronization mode, data is synchronized from the IAM LDAP server to the external LDAP server.
#      # In the bidirectional synchronization mode, a two-way data synchronization will be performed the IAM LDAP server and the external LDAP server.
#      # In the update synchronization mode, data will be updated synchronously into the IAM LDAP server from the external LDAP server.
#      sync_mode: "inbound"
#      # Synchronization will be performed according to the mapping relationship specified by the file.
#      # Example: /user/local/ldap_mapping_template.yml. If format file is not configured or not available, ldap synchronization will not be performed.
#      format_file: "/user/local/ldap_mapping_template.yml"
#      # Sync interval in minutes，if you want to sync only once, set it less than or equal to 0. Default: 15 minutes
#      sync_interval: 15
#
#
#
## Account Configurations
#account:
#  # Regular expressions are used here to specify the format of username, password, and user group.
#  # ^, $: start-of-line and end-of-line respectively.
#  # [...]: Accept ANY ONE of the character within the square bracket, e.g., [aeiou] matches "a", "e", "i", "o" or "u".
#  # [.-.] (Range Expression): Accept ANY ONE of the character in the range, e.g., [0-9] matches any digit; [A-Za-z] matches any uppercase or lowercase letters.
#  # {m,n}: m to n (both inclusive)
#  # If you need to customize the format, please refer to the regular expression syntax.
#  format:
#    name: "^[0-9a-zA-Z\u4e00-\u9fa5!@$._-]{2,36}$"
#    password: "^[0-9a-zA-Z!@$._-]{6,128}$"
#
#  smtp: # SMTP email server config
#    host: "smtp.aliyun.com"
#    username: ""
#    password: ""
#    send_with_tls: false
#
#  # You can recover the account through email or SMS(TBD).
#  # If none of the above methods are available, the admin can reset the account through the cli command line.
#  # TBD: recover account by SMS
#  recover:
#    email: # Sending a code to your recovery email # Code format:  "^[0-9a-zA-Z-:.]{6,128}$"
#      expire: 5 # Expiration time in minutes. Default 5 minutes.
#      resend_interval: 1 # Resending interval in minutes. Default 1 minute
#
#
## Authentication Configurations
#authentication:
#  # Attempt frequency can limit the frequency of user attempts to log in.
#  attempt_frequency:
#    wait_min: 1 # Minimum wait time after an attempt in second. Default: 1 Seconds
#    wait_max: 43200 # Maximum waiting time after an attempt in second. Default: 43200 Seconds
#    amp_factor: 2 # The waiting time after each failure will be extended according to the amplification factor. Default: 3
#
#  # Config the concurrent authentication to limit the number of concurrent authentication entity allowed per user.
#  # You can configure the maximum number of entities allowed to log in, and if exceeded, it will be handled according to the auto replacement policy.
#  # Supported auto replacement policies: disable | random | oldest | latest. Default: disable
#  concurrent_authentication:
#    entity_allowed: 3 # maximum number of entities allowed, 0 means disable;  less than 0 means login is prohibited. Default: 3
#    auto_replacement: "disable"
#
#  # Configure MFA(Multi-Factor Authentication) information
#  # By using the TOTP(Time-Based One Time Password) method, one time password is created on the user side through a smartphone application.
#  # Applications that are known to work with TOTP： Microsoft Authenticator、Google Authenticator）
#  # TBD: SMS
#  mfa:
#    totp:
#      # The issuer indicates the provider or service this account is associated with, URL-encoded according to RFC 3986.
#      issuer: "Amianetworks"
#
#
## Authorization Configurations
#authorization:
#  # JWT（JSON Web Token）is an open source standard (RFC 7519) that defines the format for how communicating parties can exchange information securely.
#  # TBD: JWT provides different token strategies for different entity types
#  jwt:
#    access_token:
#      expire: 60   # Access token expiration time in minutes. Default: 60 minutes.
#    refresh_token:
#      enable: true # Support automatically obtaining access token through refresh token. Default: true
#      expire: 600  # Refresh token expiration time in minutes. Default: 600 minutes.
#
#  oauth2:
#    enable: false
#    port: 17931
#    jwt_signed_key: "g6ckR89RRIolp0i"
#    session:
#      id: "iam_session_id"
#      secret: "j6oftR8TTYolp0i"
#      expire: 1200
#    client:
#      - id: "SWAN"
#        secret: "f@4vfgR8RRIolp0i"
#        name: "SWAN"   # display name of target client
#        domain: "localhost:17926"
#
#  access_control:
#    format:
#      role:
#        name: "^[0-9a-zA-Z!@$._-]{6,128}$"
#      policy:
#        name: "^[0-9a-zA-Z\u4e00-\u9fa5!@$._-]{2,36}$"
#        scope: "^[0-9a-zA-Z!@$._-/*?%&]{6,128}$"
#        operation: "^[0-9a-zA-Z!@$._-]{6,128}$"
#        time: "^[0-9a-zA-Z-:.]{6,128}$"
#
#
#
#
#
#
#`
	if err := os.WriteFile("controller/config/iam.yml", []byte(ymlIam), 0644); err != nil {
		panic(err)
	}

	ymlLdap := `#This is an example file of the LDAP synchronization format.
#You can refer to this file to fill in the format mapping of the external LDAP.
ldap_mapping:
  base_dc: "dc=example,dc=com" # Base search node of LDAP
  entry:
  - ou: "ou=user" # OU information of target LDAP user
    filter: "(title=example)" # Filter of external LDAP. Example: (&(objectClass=organizationalPerson)(title=example)). Default: ""
    external_rdn: "cn"
    external_obj:
      - "inetOrgPerson"
    attributes:
      - external: "cn"  # Attribute name of external LDAP Server
        local: "cn" # Attribute name of local LDAP Server
      - external: "mobile"  
        local: "mobile"
      - external: "email" 
        local: "email" 
    fixed: #Fields that need to be written by default
      - attribute: "title"
        value: "example"
`
	if err := os.WriteFile("controller/config/ldap_mapping_template.yml", []byte(ymlLdap), 0644); err != nil {
		panic(err)
	}
}
