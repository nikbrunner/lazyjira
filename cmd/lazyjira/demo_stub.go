//go:build !demo

package main

import (
	"errors"

	"github.com/nikbrunner/lazyjira/v2/pkg/config"
	"github.com/nikbrunner/lazyjira/v2/pkg/jira"
	"github.com/nikbrunner/lazyjira/v2/pkg/tui"
)

func startDemo(_ *config.Config) (jira.ClientInterface, tui.AuthMethod, func(), error) {
	return nil, "", nil, errors.New("demo mode not available (rebuild with: go build -tags demo)")
}
