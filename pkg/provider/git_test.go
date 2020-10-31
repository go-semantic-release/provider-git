package provider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRepository(t *testing.T) {
	require := require.New(t)
	repo := &Repository{}
	err := repo.Init(map[string]string{})
	require.EqualError(err, "repository does not exist")
}
