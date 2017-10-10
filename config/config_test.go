package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	path := "testdata/good.yaml"
	cfg, err := LoadFile(path)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "TestGraphName", cfg.GraphName)

	// expectedCluster := []*Cluster{
	// 	&Cluster{
	// 		Name: "test-cluster-1",
	// 		ServiceConnections: []*Connection{
	// 			&Connection{
	// 				Name:          "http",
	// 				Query:         "prom_http_query",
	// 				PrometheusURL: "http://prometheus1:9090",
	// 				Source: NodeMapping{
	// 					Replacement: "INTERNET",
	// 				},
	// 				Target: NodeMapping{
	// 					Label: "service",
	// 				},
	// 			},
	// 		},
	// 	},
	// }
	fmt.Println(cfg.Clusters[0].ServiceConnections[0].Target)
}
