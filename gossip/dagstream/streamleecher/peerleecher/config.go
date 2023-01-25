package peerleecher

import (
	"time"

	"github.com/Ncog-Earth-Chain/forest-base/inter/dag"
)

type EpochDownloaderConfig struct {
	RecheckInterval        time.Duration
	DefaultChunkSize       dag.Metric
	ParallelChunksDownload int
}
