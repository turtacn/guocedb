package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// StorageEngine interface for collecting storage metrics
type StorageEngine interface {
	Size() (lsm int64, vlog int64)
	KeyCount() int64
	TableCount() map[int]int
}

// StorageCollector collects storage-related metrics
type StorageCollector struct {
	storage StorageEngine

	lsmSizeDesc   *prometheus.Desc
	vlogSizeDesc  *prometheus.Desc
	numKeysDesc   *prometheus.Desc
	numTablesDesc *prometheus.Desc
}

// NewStorageCollector creates a new StorageCollector
func NewStorageCollector(storage StorageEngine) *StorageCollector {
	return &StorageCollector{
		storage: storage,
		lsmSizeDesc: prometheus.NewDesc(
			"guocedb_storage_lsm_bytes",
			"LSM tree size in bytes",
			nil, nil,
		),
		vlogSizeDesc: prometheus.NewDesc(
			"guocedb_storage_vlog_bytes",
			"Value log size in bytes",
			nil, nil,
		),
		numKeysDesc: prometheus.NewDesc(
			"guocedb_storage_keys_total",
			"Total number of keys",
			nil, nil,
		),
		numTablesDesc: prometheus.NewDesc(
			"guocedb_storage_tables_total",
			"Total number of SST tables",
			[]string{"level"}, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors
func (c *StorageCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.lsmSizeDesc
	ch <- c.vlogSizeDesc
	ch <- c.numKeysDesc
	ch <- c.numTablesDesc
}

// Collect is called by the Prometheus registry when collecting metrics
func (c *StorageCollector) Collect(ch chan<- prometheus.Metric) {
	if c.storage == nil {
		return
	}

	// Get storage sizes
	lsm, vlog := c.storage.Size()
	ch <- prometheus.MustNewConstMetric(c.lsmSizeDesc, prometheus.GaugeValue, float64(lsm))
	ch <- prometheus.MustNewConstMetric(c.vlogSizeDesc, prometheus.GaugeValue, float64(vlog))

	// Get key count
	numKeys := c.storage.KeyCount()
	ch <- prometheus.MustNewConstMetric(c.numKeysDesc, prometheus.GaugeValue, float64(numKeys))

	// Get table count per level
	tables := c.storage.TableCount()
	for level, count := range tables {
		ch <- prometheus.MustNewConstMetric(
			c.numTablesDesc,
			prometheus.GaugeValue,
			float64(count),
			fmt.Sprintf("%d", level),
		)
	}
}

// RegisterStorageCollector registers the storage collector with Prometheus
func RegisterStorageCollector(storage StorageEngine) {
	prometheus.MustRegister(NewStorageCollector(storage))
}
