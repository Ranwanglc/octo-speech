package asrlog

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Cleaner struct {
	baseDir       string
	retentionDays int
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	logger        *zap.Logger
}

func NewCleaner(baseDir string, retentionDays int, logger *zap.Logger) *Cleaner {
	ctx, cancel := context.WithCancel(context.Background())
	return &Cleaner{
		baseDir:       baseDir,
		retentionDays: retentionDays,
		ctx:           ctx,
		cancel:        cancel,
		logger:        logger,
	}
}

func (c *Cleaner) Start() {
	c.wg.Add(1)
	go c.run()
}

func (c *Cleaner) run() {
	defer c.wg.Done()

	c.clean()
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.clean()
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *Cleaner) Close() {
	c.cancel()
	c.wg.Wait()
}

func (c *Cleaner) clean() {
	entries, err := os.ReadDir(c.baseDir)
	if err != nil {
		if !os.IsNotExist(err) && c.logger != nil {
			c.logger.Error("read asr log dir failed", zap.Error(err))
		}
		return
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -c.retentionDays)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		t, err := time.Parse("2006-01-02", entry.Name())
		if err != nil {
			continue
		}
		if t.Before(cutoff) {
			dir := filepath.Join(c.baseDir, entry.Name())
			if err := os.RemoveAll(dir); err != nil {
				if c.logger != nil {
					c.logger.Error("remove old asr log dir failed",
						zap.Error(err), zap.String("dir", dir))
				}
			} else if c.logger != nil {
				c.logger.Info("removed old asr log dir", zap.String("dir", dir))
			}
		}
	}
}
