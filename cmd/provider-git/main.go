package main

import (
	gitProvider "github.com/go-semantic-release/provider-git/pkg/provider"
	"github.com/go-semantic-release/semantic-release/v2/pkg/plugin"
	"github.com/go-semantic-release/semantic-release/v2/pkg/provider"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		Provider: func() provider.Provider {
			return &gitProvider.Repository{}
		},
	})
}
