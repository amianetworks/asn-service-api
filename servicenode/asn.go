// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package snapi

import (
	commonapi "asn.amiasys.com/asn-service-api/v26/common"
	"asn.amiasys.com/asn-service-api/v26/log"
)

// ASNServiceNode contains the APIs provided by ASN Service Node.
//
// 1. Initialization
// 2. Node Info
// 3. SendMessageToController
type ASNServiceNode interface {
	/*
		Initialization and Resource Allocation
	*/

	// InitLogger returns a logger dedicated to the service.
	//
	// ASN Framework manages logging for all services, and the default log files are <servicename>-*.log.
	// SHOULD ONLY call once. Further calls will get an error.
	InitLogger() (*log.Logger, error)

	// InitDocDB returns a doc DB handle.
	//
	// The DB is connected and ready for use through the DocDBHandler upon return.
	// SHOULD ONLY call once for each name. Further calls will get an error.
	InitDocDB(name string) (commonapi.DocDBHandler, error)

	// InitTSDB returns a doc DB handle.
	//
	// The DB is connected and ready for use through the TSDBHandler upon return.
	// SHOULD ONLY call once for each name. Further calls will get an error.
	InitTSDB(name string) (commonapi.TSDBHandler, error)

	// Placeholder for Locker, in case it's necessary.
	// Placeholder for IAM, in case it's necessary.

	/*
		Node Info
	*/

	// GetNodeType returns the service node's type.
	GetNodeType() commonapi.NodeType

	// GetNodeInfo returns the service node's info.
	GetNodeInfo() *NodeInfo

	/*
		SendMessageToController
	*/

	// SendMessageToController
	//
	// Service Node may send a formated message to its controller, which may handle the message by
	// implementing HandleMessageFromNode(). NO DIRECT RESPONSE to the message should be expected.
	SendMessageToController(messageType, payload string) error

	// GetSharedData returns the shared data's keys that a service provides.
	//
	// `aggregated` are the data's keys that can be queried.
	// `subscribable` are the data's keys that can be subscribed to.
	//
	// If the given serviceName does not exist, ErrServiceNotFound will be returned.
	GetSharedData(serviceName string) (aggregated, subscribable []string, err error)

	// QueryServiceSharedData asks a service for data's values based on the given keys.
	//
	// This function works in pairs with `ASNService.OnQuerySharedData`.
	// It is used for fetching data from another service.
	// If you wish to subscribe to TS-like data, use `ASNServiceNode.SubscribeServiceSharedData` instead.
	//
	// In some cases, services need to query another service for data.
	// If the service you are implementing is a data CONSUMER, i.e., you will ask another service for data,
	// this is the function you should call when you need to get the values.
	// Otherwise, you can ignore this function.
	//
	// The keys should be negotiated between services beforehand so that other services know what you are providing.
	//
	// If the given serviceName does not exist, ErrServiceNotFound will be returned.
	// If one of the given keys does not exist, ErrKeyNotFound will be returned.
	QueryServiceSharedData(serviceName string, keys []string) (values map[string]any, err error)

	// SubscribeServiceSharedData subscribes to data's values from a service based on the given keys.
	//
	// This function works in pairs with `ASNService.OnSubscribeSharedData`.
	// It is used for subscribing to TS-like data.
	// If you wish to fetch data, use `ASNServiceNode.QueryServiceSharedData` instead.
	//
	// In some cases, services need to subscribe to data from another service.
	// If the service you are implementing is a data CONSUMER, i.e., you will ask another service for data,
	// this is the function you should call when you need to subscribe to the values.
	// Otherwise, you can ignore this function.
	//
	// The keys should be negotiated between services beforehand so that other services know what you are providing.
	//
	// If the given serviceName does not exist, ErrServiceNotFound will be returned.
	// If one of the given keys does not exist, ErrKeyNotFound will be returned.
	// For each key, after all values are dumped, the returned channel will be closed.
	//
	// CAUTION: for each key, only one subscription is allowed.
	// Multiple subscriptions before the previous one closes will lead to unexpected results.
	SubscribeServiceSharedData(serviceName string, keys []string) (values map[string]<-chan any, err error)
}
