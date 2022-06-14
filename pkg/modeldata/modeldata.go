// Package modeldata contains the following model data in gnmi proto struct:
//	openconfig-interfaces 2.0.0,
//	openconfig-if-ip 1.0.0.
package modeldata

import (
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

const (
	// OpenconfigInterfacesModel is the openconfig YANG model for interfaces.
	OpenconfigInterfacesModel = "openconfig-interfaces"
	// OpenconfigSystemModel is the openconfig YANG model for system.
	OpenconfigIfIpModel = "openconfig-if-ip"
)

var (
	// ModelData is a list of supported models.
	ModelData = []*pb.ModelData{{
		Name:         OpenconfigInterfacesModel,
		Organization: "OpenConfig working group",
		Version:      "2.0.0",
	}, {
		Name:         OpenconfigIfIpModel,
		Organization: "OpenConfig working group",
		Version:      "1.0.0",
	}}
)
