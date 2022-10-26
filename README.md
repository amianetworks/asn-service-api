# ASN Service API

## Description
ASN(AI-Driving Secure Networking) is a distributed framework of secure network functions.
This API package is shared by all ANS severices built as plugins. To build a ASN Distributed Service, use the template package to start.


The latest stable version is 1.8.0.

## API Layout
    .
    ├── controller    // Local API for Service Controller, which manage distributed services.
    ├── servicenode   // Local API for Service Node, which connects to the controller or run in standalone mode.
    └── logger        // Formatted logger

## Development Preparations
1. Check the versions of ASN Controller and Service Nodes, and check out the corresponding version of API.
2. It's strongly suggested to use the latest stable version of ASN framework.
3. A ASN service will be built as a plugins. (Please search and learn how Golang plugin works)

## NEED TO KNOW
This is a sub-project and has API only.
