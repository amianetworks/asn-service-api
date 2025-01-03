# ASN Service API

## Description
ASN(AI-Driving Secure Networking) is a distributed framework of secure network functions.\
This API package is shared by all ASN services built as plugins. To build an ASN Distributed Service, you can refer asn-service-template to get start.\
The latest version is v25.1.5

## API Layout
    ├── common        // common structs and functions for both controller and service node use. 
    ├── controller    // API defined between ASN contoller and [YOUR SERVICE] controller.
    ├── servicenode   // API defined between ASN service node and [YOUR SERVICE] service node.
    └── log           // Formatted logger provided for [YOUR SERVICE] to use

## How to use the API
1. Implement all functions in controller.ASNService in [YOUR SERVICE] controller module. This is API provided by [YOUR SERVICE] and called by ASN controller.
2. Provide the init controller function
   ```
       func NewASNService(asnController ASNController) (controller.ASNService, error)
   ```
   in [YOUR SERVICE] controller module. The function name, parameter and return type should be FULLY IDENTICAL with the above function. This is used for ASN controller to recognize [YOUR SERVICE] controller.
   ASNController and ASNService are both the interface defined in `./controller/controller.go`
3. Implement all functions in servicenode.ASNService in [YOUR SERVICE] servicenode module. This is API provided by [YOUR SERVICE] and called by ASN service node.
4. Provide the init service node function
   ```
       func NewASNService(asnServiceNode ASNServiceNode) (servicenode.ASNService, error)
   ```
   in [YOUR SERVICE]'s service node module. The function name, parameter and return type should be FULLY IDENTICAL with the function above. This is used for ASN service node to recognize [YOUR SERVICE] node.
   ASNServiceNode and ASNService are both the interface defined in `./servicenode/servicenode.go`
5, Finish [YOUR SERVICE] code. Use the function defined in `ASNController` and `ASNServiceNode` instead to manage the relation between [YOUR SERVICE] controller and servicenode.

## How to compile YOUR SERVICE
1. Check the version of ASN API is corresponding with the version in ASN and [YOUR SERVICE].
2. Check the dependency in `./go.mod.asn`. If you use the dependency list there, make sure the version is FULLY IDENTICAL with dependency listed. If you are building multiple plugins in one system,
make sure the dependencies shared by them are fully matched.
3. You can then choose to build as plugin or build with ASN project.

To build as plugin [RECOMMENDED]
1. introduce `asn-service-api` as the submodule of [YOUR SERVICE]. run the following command
   ```
   git submodule add git@github.com:amianetworks/asn-service-api.git service-api
   git submodule update --init
   ```
2. add `include service-api/scripts/builder/Makefile` in [YOUR SERVICE] Makefile
3. provide "build local" command in [YOUR SERVICE] Makefile. This command is used to build [YOUR SERVICE] in local environment
4. run `make build` and you can get the build directory like
    ```
   build
   ├── controller    
       ├── YOUR SERVICE.so 
       ├── *.yml
       └── *.conf
   └── servicenode
       └── YOUR SERVICE.so
   ```

To build with ASN project.
1. load [YOUR SERVICE] to the `services` directory to the ASN project
2. provide "make build" command in both [YOUR SERVICE] controller and node Makefile. This command is used to build [YOUR SERVICE] controller or node in local environment
3. run `make build-dev` or `make build-pro` according to the env you want to use. You will also get the build directory like above.

## Simple Deploy
For initial deployment, you can just use deploy the basic network structure with one controller and one service node. To achieve that
1. In the deployed environment. Place the directory structure like 
   ```
       you project name
       ├── controller    
           ├── config
           ├── plugins
           └── asnc.yml
       └── servicenode
           ├── config
           ├── plugins
           └── asnsn.yml
   ```
   `asnc.yml` is in `./scripts/deploy/docker/controller` and `asnsn.yml` is in `./scripts/deploy/docker/servicenode`
2. in your build file,  move `controller/swan.so` to `<you project name>/controller/plugins`, move `controller/*.conf` and `controller/*.yml` to `<you project name>/controller/config`,
   move `servicenode/swan.so` to `<you project name>/servicenode/plugins`, move `servicenode/*.conf`to `<you project name>/servicenode/config`,
3. `cd controller` and run `docker compose -f asnc.yml up -d` to start SWAN controller
4. `cd servicenode` and run `docker compose -f asnsn.yml up -d` to start SWAN service node

### For complex Network
To deploy SWAN with a complex networks, like multiple controllers which multiple service nodes. You need to
1. Prepare the `*-topology.json` file. This is json that defines how the network is composed. The json structure is like
   ```
   {
      "network_id": "",                      // network ID, MUST BE UNIQUE 
      "network_name": "",                    // network display name
      "topology": [                          // all the service nodes in the network
         {
            "node_name": "node1",            // node display name, MUST BE UNIQUE
            "nodeType": "",                  // node type 
            "location": {},                  // node pyhscial location, include coordinates and address                
            "group": "",                     // group belongs to the node
            "external_linked": [],           // all the external linked node ids
            "internal_linked": [],           // all the internal inked node ids
            "sub_nodes": [],                 // all child nodes
         }
      ]
   }
   ```
   one controller corresponds to one network, all service nodes it connects are all defined in `topology` field. The number of service nodes equals to the length of topology list.
   For detail information about topology networks and nodes defined, please refer to README.md in ASN.25.
2. move the prepared `*-topology.json` file to `<you project name>/controller/config`.
   update `<you project name>/controller/config/asn.conf` and set `network.topo_file: ./config/*-topology.json`
3. For each controller, set directory structure as
   ```
   ├── controller    
       ├── config
           ├── *.conf
           ├── *.yml
           └── *-topology.json
       ├── plugins
           └── swan.so
       └── asnc.yml
   ```
   For each controller, run `docker compose -f asnc.yml up -d` to start SWAN controller
4. For each servicenode, set directory structure as
   ```
   ├── servicenode    
       ├── config
           ├── *.conf
           └── *.yml
       ├── plugins
           └── swan.so
       └── asnsn.yml
   ```
   For each servicenode, go to the root directory and run `docker compose -f asnsn.yml up -d` to start SWAN service nodes separately


## NEED TO KNOW
This is a subproject and has API only.
