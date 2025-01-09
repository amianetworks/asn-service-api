// Copyright 2024 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

const (
	DBTypeDoc = "docdb"
	DBTypeTS  = "tsdb"
)

const (
	DBProviderInfluxDBV1 string = "influxdbv1"
	DBProviderInfluxDBV2 string = "influxdbv2"
	DBProviderMongoDB    string = "mongodb"
	DBProviderFileDB     string = "filedb"
	DBProviderPrometheus string = "prometheus"
	DBProviderTSFilDB    string = "tsfiledb"
)

type DBConf struct {
	DBType       string // database type: influxdbv1, influxdbv2, mongodb, filedb, prometheus
	DBName       string // database name
	Url          string // database url address
	Username     string // username for login
	Password     string // password for login
	Organization string // only needed by influxdbv2
	Token        string // only needed by influxdbv2
}
