# ASN Service API

## Description
The repository is a shared package ASN Services built as plugins.
The latest release is 1.0.0.

## API Layout
    .
    ├── controller    // API for Service Controller
    ├── logger        // ASN formatted logger
    └── servicenode   // API for Service Node

## Preparations
1. Make sure you're using the lastest stable versions of ASN Controller and Service Node, which should have been built with the latest tagged API.
2. Build an example project, e.g. firewall, and try to load it as a plugin module.
3. Check with ASN Controller/Service Node developer about the API version in use.

## Which version to use
1. It's recommended starting a new project based on a template project, so that you don't need to specify the API version.
2. The latest tagged release is always recommended, unless you know which specific version, tagged, you need
