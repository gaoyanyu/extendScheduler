package disk

import (
	"context"
	"extendScheduler/pkg/promethus"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	fruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

const Name = "DiskIO"

var _ = framework.ScorePlugin(&DiskIO{})

type DiskIO struct {
	prometheus *promethus.PromDiskIOHandle
	// FrameworkHandle 提供插件可以使用的数据和一些工具。
	// 它在插件初始化时传递给 plugin 工厂类。
	// plugin 必须存储和使用这个handle来调用framework函数。
	handle framework.FrameworkHandle
}

// FitArgs holds the args that are used to configure the plugin.
type DiskIOArgs struct {
	IP         string `json:"ip"`
	TimeRange  int    `json:"timeRange"`
}

// New initializes a new plugin and returns it.
func New(plArgs runtime.Object, h framework.FrameworkHandle) (framework.Plugin, error) {
	args := &DiskIOArgs{}
	if err := fruntime.DecodeInto(plArgs, args); err != nil {
		return nil, err
	}

	klog.Infof("[DiskIO Plugin] args received. TimeRange: %d, Address: %s", args.TimeRange, args.IP)

	return &DiskIO{
		handle:     h,
		prometheus: promethus.NewDiskIOProme(args.IP, time.Duration(args.TimeRange)),
	}, nil
}

// Name returns name of the plugin. It is used in logs, etc.
func (d *DiskIO) Name() string {
	return Name
}

// 如果返回framework.ScoreExtensions 就需要实现framework.ScoreExtensions
func (d *DiskIO) ScoreExtensions() framework.ScoreExtensions {
	return d
}

// NormalizeScore与ScoreExtensions是固定格式
func (d *DiskIO) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, scores framework.NodeScoreList) *framework.Status {
	var higherScore int64
	for _, node := range scores {
		if higherScore < node.Score {
			higherScore = node.Score
		}
	}
	// 计算公式为，满分 - (当前DiskIO / 最高DiskIO * 100)
	// 公式的计算结果为，DiskIO越大的机器，分数越低
	for i, node := range scores {
		scores[i].Score = framework.MaxNodeScore - (node.Score * 100 / higherScore)
	}

	klog.Infof("[DiskIO Plugin] Nodes final score: %v", scores)
	return nil
}

func (d *DiskIO) Score(ctx context.Context, state *framework.CycleState, p *corev1.Pod, nodeName string) (int64, *framework.Status) {
	nodeDiskIO, err := d.prometheus.GetGauge(nodeName)
	if err != nil {
		//ignore it when getting node disk io err from prometheus, pod scheduling can't be blocked here.
		klog.Errorf("[DiskIO] score for pod %s in namespace %s, err: %s", p.Name, p.Namespace, err)
		return framework.MinNodeScore + 1, framework.NewStatus(framework.Success, fmt.Sprintf("get node disk io err: %s, but ignore", err))
	}
	diskIO := int64(nodeDiskIO.Value)
	klog.Infof("[DiskIO Plugin] node '%s' diskIO: %d", nodeName, diskIO)
	return diskIO, nil
}
