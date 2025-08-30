package apiconfig

import (
	"context"
	"fmt"
	"github.com/casbin/casbin/v2"
	"hideout/config"
)

func CreateCasBinEnforcer(ctx context.Context, config config.CasBinConfig) (*casbin.Enforcer, error) {
	switch config.AdapterType {
	case CasBinAdapterType_File:
		{
			if Settings.CasBin.ModelPath != "" && Settings.CasBin.PolicyPath != "" {
				return casbin.NewEnforcer(Settings.CasBin.ModelPath, Settings.CasBin.PolicyPath)
			}
			return nil, fmt.Errorf("Missing model or policy files for CasBin with file adapter")
		}
	}

	return nil, fmt.Errorf("Unknown adapter type: %v", config.AdapterType)
}
