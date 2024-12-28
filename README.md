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
1，Implement all functions in controller.ASNService in [YOUR SERVICE] controller module. This is API provided by [YOUR SERVICE] and called by ASN controller.\
2，Provide the init controller function
```
    func NewASNService(asnController ASNController) (controller.ASNService, error)
```
in [YOUR SERVICE] controller module. The function name, parameter and return type should be EXACTLY THE SAME with the function above. This is used for ASN controller to recognize [YOUR SERVICE] controller.
ASNController and ASNService are both the interface defined in `./controller/controller.go`\
3, Implement all functions in servicenode.ASNService in [YOUR SERVICE] servicenode module. This is API provided by [YOUR SERVICE] and called by ASN service node.
4, Provide the init service node function
```
    func NewASNService(asnServiceNode ASNServiceNode) (servicenode.ASNService, error)
```
in [YOUR SERVICE]'s service node module. The function name, parameter and return type should be EXACTLY THE SAME with the function above. This is used for ASN service node to recognize [YOUR SERVICE] node.
ASNServiceNode and ASNService are both the interface defined in `./servicenode/servicenode.go`\
5, Finish [YOUR SERVICE] code. Use the function defined in `ASNController` and `ASNServiceNode` instead to manage the relation between [YOUR SERVICE] controller and servicenode.

## How to compile YOUR SERVICE
1, Check the version of ASN API is corresponding with the version in ASN and [YOUR SERVICE].\
2, Check the dependency in `./go.mod.asn`. If you use the dependency list there, make sure the version is EXACTLY THE SAME one. If you are building multiple plugins in one system,
make sure the dependency they use have no different version.\
3, You can choose to build as plugin or build with ASN project.

To build as plugin.
1. git submodule `asn-service-api` in [YOUR SERVICE] and export SERVICE_API_PATH = `YOUR SUBMODULE PATH`
2. include `./scripts/builder/Makefile` in [YOUR SERVICE] Makefile
3. provide "make build" command in [YOUR SERVICE] Makefile
4. run `make compile-plugin` and can get the build directory like
    ```
        build
        ├── controller    
            ├── YOUR_SERVICE_NAME.so 
            └── conf
        └── servicenode
            ├── YOUR_SERVICE_NAME.so 
            └── conf
    ```
5. move `controller/YOUR_SERVICE_NAME.so` file to your ASN controller plugins directory and `servicenode/YOUR_SERVICE_NAME.so` file to your ASN servicenode plugins directory

To build with ASN project.
1. load [YOUR SERVICE] to `services` directory to the ASN project
2. provide "make build" command in [YOUR SERVICE] Makefile.
3. run `make build-dev` or `make build-pro` and can get the same build directory with the env you choose

## Deploy
1, To deploy [YOUR SERVICE] as one controller and one service node structure, just refer to `./scripts/docker`. Make your project as the structure below
```
    project name
    ├── controller    
        ├── config
        ├── plugins
            └── YOUR_SERVICE_NAME.so 
        └── asnc.yml
    └── servicenode
        ├── config
        ├── plugins
            └── YOUR_SERVICE_NAME.so 
        └── asnsn.yml
```
Then, use docker compose to start both controller and service node server.\
2, To deploy [YOUR SERVICE] as one controller and multiple service nodes structure. Can run the scripts in the `./scripts/docker-cluster`. This will generate the file similar to the structure as `./scripts/docker`.
Move the `.so` files under plugins directory and then use docker compose to start the service.\
3, To deploy [YOUR SERVICE] in your own topology. Please carefully read the Introduction in `ASN.25`. Move your `*-topology.json` file under `project name/controller/config`.
Move the `.so` files under plugins directory. Then can use docker compose to start the service.


## NEED TO KNOW
This is a subproject and has API only.
