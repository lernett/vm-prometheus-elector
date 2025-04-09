package watcher

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"k8s.io/klog/v2"
	"path/filepath"

	"github.com/jlevesy/prometheus-elector/config"
	"github.com/jlevesy/prometheus-elector/election"
	"github.com/jlevesy/prometheus-elector/notifier"
)

type FileWatcher struct {
	fsWatcher     *fsnotify.Watcher
	reconciler    *config.Reconciler
	leaderChecker election.LeaderChecker
	notifier      notifier.Notifier
	configPath    string
}

func New(path string, reconciler *config.Reconciler, notifier notifier.Notifier, leaderChecker election.LeaderChecker) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	configDir := filepath.Dir(path)
	if err != nil {
		return nil, fmt.Errorf("unable to create fsnotify watcher: %w", err)
	}

	if err := watcher.Add(configDir); err != nil {
		return nil, fmt.Errorf("unable to create watch config directory: %w", err)
	}

	klog.InfoS("Watching config directory", "path", configDir)

	return &FileWatcher{
		fsWatcher:     watcher,
		leaderChecker: leaderChecker,
		reconciler:    reconciler,
		notifier:      notifier,
		configPath:    filepath.Base(path),
	}, nil
}

func (f *FileWatcher) Watch(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.Canceled {
				return nil
			}
			return ctx.Err()
		case evt, ok := <-f.fsWatcher.Events:
			if !ok {
				return nil
			}

			if !evt.Has(fsnotify.Create) || !(filepath.Base(evt.Name) == f.configPath) {
				continue
			}

			klog.Info("Configuration changed, reconciling...")

			if err := f.reconciler.Reconcile(ctx, f.leaderChecker.IsLeader()); err != nil {
				klog.ErrorS(err, "Reconciler reported an error")
				continue
			}

			if err := f.notifier.Notify(ctx); err != nil {
				klog.ErrorS(err, "Unable to notify prometheus")
				continue
			}
		case err, ok := <-f.fsWatcher.Errors:
			if !ok {
				return nil
			}

			klog.ErrorS(err, "Watcher reported an error")
		}
	}
}

func (f *FileWatcher) Close() error {
	return f.fsWatcher.Close()
}
