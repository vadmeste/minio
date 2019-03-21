package cmd

import (
	"context"

	"github.com/minio/minio/cmd/logger"
	"github.com/minio/minio/pkg/madmin"
)

func init() {
	initHealBackground()
}

type healTask struct {
	path       string
	opts       madmin.HealOpts
	responseCh chan healResult
}

type healResult struct {
	result madmin.HealResultItem
	err    error
}

type healBackground struct {
	ctx   context.Context
	tasks chan healTask
}

func (h *healBackground) queueHealTask(task healTask) {
	h.tasks <- task
}

func (h *healBackground) run() {
	ctx := context.Background()
	for task := range h.tasks {
		var res madmin.HealResultItem
		var err error
		bucket, object := urlPath2BucketObjectName(task.path)
		switch {
		case bucket == "" && object == "":
			res, err = bgHealDiskFormat(ctx, task.opts)
		case bucket != "" && object == "":
			res, err = bgHealBucket(ctx, bucket, task.opts)
		case bucket != "" && object != "":
			res, err = bgHealObject(ctx, bucket, object, task.opts)
		}
		task.responseCh <- healResult{result: res, err: err}
	}
}

var globalHealBackground *healBackground

func initHealBackground() {
	healBg := &healBackground{
		tasks: make(chan healTask),
	}
	go healBg.run()

	globalHealBackground = healBg
}

// bgHealDiskFormat - heals format.json, return value indicates if a
// failure error occurred.
func bgHealDiskFormat(ctx context.Context, opts madmin.HealOpts) (madmin.HealResultItem, error) {
	// Get current object layer instance.
	objectAPI := newObjectLayerFn()
	if objectAPI == nil {
		return madmin.HealResultItem{}, errServerNotInitialized
	}

	res, err := objectAPI.HealFormat(ctx, opts.DryRun)

	// return any error, ignore error returned when disks have
	// already healed.
	if err != nil && err != errNoHealRequired {
		return madmin.HealResultItem{}, err
	}

	// Healing succeeded notify the peers to reload format and re-initialize disks.
	// We will not notify peers only if healing succeeded.
	if err == nil {
		for _, nerr := range globalNotificationSys.ReloadFormat(opts.DryRun) {
			if nerr.Err != nil {
				logger.GetReqInfo(ctx).SetTags("peerAddress", nerr.Host.String())
				logger.LogIf(ctx, nerr.Err)
			}
		}
	}

	return res, nil
}

// bghealBucket - traverses and heals given bucket
func bgHealBucket(ctx context.Context, bucket string, opts madmin.HealOpts) (madmin.HealResultItem, error) {
	// Get current object layer instance.
	objectAPI := newObjectLayerFn()
	if objectAPI == nil {
		return madmin.HealResultItem{}, errServerNotInitialized
	}

	return objectAPI.HealBucket(ctx, bucket, opts.DryRun, opts.Remove)
}

// bgHealObject - heal the given object and record result
func bgHealObject(ctx context.Context, bucket, object string, opts madmin.HealOpts) (madmin.HealResultItem, error) {
	// Get current object layer instance.
	objectAPI := newObjectLayerFn()
	if objectAPI == nil {
		return madmin.HealResultItem{}, errServerNotInitialized
	}
	return objectAPI.HealObject(ctx, bucket, object, opts.DryRun, opts.Remove, opts.ScanMode)
}
