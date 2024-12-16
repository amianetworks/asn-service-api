# ASN Service API

## Description
ASN(AI-Driving Secure Networking) is a distributed framework of secure network functions.
This API package is shared by all ANS services built as plugins. To build an ASN Distributed Service, use the template package to start.


The latest stable version is 1.8.0. 2.0.0 is also launched, but with a slightly different API design.

## API Layout
    .
    ├── controller    // Local API for Service Controller, which manage distributed services.
    ├── servicenode   // Local API for Service Node, which connects to the controller or run in standalone mode.
    └── logger        // Formatted logger

## Development Preparations
1. Check the versions of ASN Controller and Service Nodes, and check out the corresponding version of API.
2. It's strongly suggested to use the latest stable version of ASN framework.
3. An ASN service will be built as a plugins. (Please search and learn how Golang plugin works)

## NEED TO KNOW
This is a subproject and has API only.
