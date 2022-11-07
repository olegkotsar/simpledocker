package simpledocker

import (
	"context"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSimpleDocker(t *testing.T) {
	name := uuid.New().String()
	cc, err := CreateContainer(
		name,
		"mysql:latest",
		[]string{"3307:3306"},
		map[string]string{"MYSQL_ROOT_PASSWORD": "test_pass"},
	)
	require.NoError(t, err)
	defer cc.Remove()

	// Check.
	c, err := client.NewEnvClient()
	require.NoError(t, err)
	ctx := context.Background()
	list, err := c.ContainerList(ctx, types.ContainerListOptions{})
	require.NoError(t, err)
	exists := false
	for _, v := range list {
		for _, n := range v.Names {
			if strings.Contains(n, name) {
				exists = true
				break
			}
		}
	}
	require.True(t, exists)
}
