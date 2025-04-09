package config

import (
	"context"
)

type Reconciler struct {
	sourcePath string
	outputPath string
}

func NewReconciller(src, out string) *Reconciler {
	return &Reconciler{
		sourcePath: src,
		outputPath: out,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, leader bool) error {
	cfg, err := loadConfiguration(r.sourcePath)
	if err != nil {
		return err
	}

	targetCfg := cfg

	if !leader {
		delete(targetCfg, "scrape_configs")
	}

	return writeConfiguration(r.outputPath, targetCfg)
}
