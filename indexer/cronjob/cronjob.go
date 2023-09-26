package cronjob

import (
	"flare-indexer/database"
	"flare-indexer/indexer/config"
	"flare-indexer/logger"
	"flare-indexer/utils"
	"flare-indexer/utils/staking"
	"time"
)

type Cronjob interface {
	Name() string
	Enabled() bool
	Timeout() time.Duration
	Call() error
	OnStart() error
}

func RunCronjob(c Cronjob) {
	if !c.Enabled() {
		logger.Debug("%s cronjob disabled", c.Name())
		return
	}

	err := c.OnStart()
	if err != nil {
		logger.Error("%s cronjob on start error %v", c.Name(), err)
		return
	}

	logger.Debug("starting %s cronjob", c.Name())

	ticker := time.NewTicker(c.Timeout())
	for range ticker.C {
		err := c.Call()
		if err != nil {
			logger.Error("%s cronjob error %s", c.Name(), err.Error())
		}
	}
}

const (
	defaultEpochBatchSize int64 = 100
)

type epochCronjob struct {
	enabled   bool
	timeout   time.Duration // call cronjob every "timeout"
	epochs    staking.EpochInfo
	delay     time.Duration // voting delay
	batchSize int64
}

type epochRange struct {
	start int64
	end   int64
}

func newEpochCronjob(cronjobCfg *config.CronjobConfig, epochs staking.EpochInfo) epochCronjob {
	return epochCronjob{
		enabled:   cronjobCfg.Enabled,
		timeout:   cronjobCfg.Timeout,
		epochs:    epochs,
		batchSize: cronjobCfg.BatchSize,
		delay:     cronjobCfg.Delay,
	}
}

func (c *epochCronjob) Enabled() bool {
	return c.enabled
}

func (c *epochCronjob) Timeout() time.Duration {
	return c.timeout
}

// Get processing range (closed interval)
func (c *epochCronjob) getEpochRange(start int64, now time.Time) *epochRange {
	return c.getTrimmedEpochRange(start, c.epochs.GetEpochIndex(now)-1)
}

// Get trimmed processing range (closed interval)
func (c *epochCronjob) getTrimmedEpochRange(start, end int64) *epochRange {
	start = utils.Max(start, c.epochs.First)
	batchSize := c.batchSize
	if batchSize == 0 {
		batchSize = defaultEpochBatchSize
	} else if batchSize < 0 {
		batchSize = end - start + 1
	}
	if end >= start+batchSize {
		end = batchSize + start - 1
	}
	return &epochRange{start, end}
}

func (c *epochCronjob) indexerBehind(idxState *database.State, epoch int64) bool {
	epochEnd := c.epochs.GetEndTime(epoch)
	return epochEnd.After(idxState.Updated.Add(-c.delay)) || idxState.NextDBIndex <= idxState.LastChainIndex
}
