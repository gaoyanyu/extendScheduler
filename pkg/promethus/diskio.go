package promethus

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"k8s.io/klog"
	"time"
)

const (
	// nodeDiskIOQueryTemplate is the template string to get the query for the node used bandwidth
	nodeDiskIOQueryTemplate = "max(rate(node_disk_written_bytes_total{nodename=\"%s\"}[%ds])) + max(rate(node_disk_read_bytes_total{nodename=\"%s\"}[%ds]))"
)

type PromDiskIOHandle struct {
	timeRange  time.Duration
	ip         string
	client     v1.API
}

func NewDiskIOProme(ip string, timeRange time.Duration) *PromDiskIOHandle {
	client, err := api.NewClient(api.Config{Address: ip})
	if err != nil {
		klog.Fatalf("[DiskIO] FatalError creating prometheus client: %s", err.Error())
	}
	return &PromDiskIOHandle{
		ip:         ip,
		timeRange:  timeRange,
		client:     v1.NewAPI(client),
	}
}

func (p *PromDiskIOHandle) GetGauge(node string) (*model.Sample, error) {
	value, err := p.query(fmt.Sprintf(nodeDiskIOQueryTemplate, node, p.timeRange, node, p.timeRange))
	if err != nil {
		return nil, fmt.Errorf("[DiskIO] GetGauge err: %w", err)
	}

	nodeMeasure := value.(model.Vector)
	if len(nodeMeasure) != 1 {
		return nil, fmt.Errorf("[DiskIO] Invalid response, expected 1 value, got %d", len(nodeMeasure))
	}
	return nodeMeasure[0], nil
}

func (p *PromDiskIOHandle) query(promQL string) (model.Value, error) {
	results, warnings, err := p.client.Query(context.Background(), promQL, time.Now())
	if err != nil {
		klog.Errorf("[DiskIO] Error querying: %v\n", err)
	}
	if len(warnings) > 0 {
		klog.Warningf("[DiskIO] Warnings querying: %v\n", warnings)
	}

	return results, err
}