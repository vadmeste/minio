// Copyright (c) 2015-2022 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/minio/madmin-go"
)

func collectLocalMetrics(types madmin.MetricType, hosts map[string]struct{}, disks map[string]struct{}) (m madmin.RealtimeMetrics) {
	if types == madmin.MetricsNone {
		return
	}

	if len(hosts) > 0 {
		if _, ok := hosts[globalMinioAddr]; !ok {
			return
		}
	}

	if types.Contains(madmin.MetricsDisk) && !globalIsGateway {
		m.ByDisk = make(map[string]madmin.DiskMetric)
		aggr := madmin.DiskMetric{
			CollectedAt: time.Now(),
		}
		for name, disk := range collectLocalDisksMetrics(disks) {
			m.ByDisk[name] = disk
			aggr.Merge(&disk)
		}
		m.Aggregated.Disk = &aggr
	}

	if types.Contains(madmin.MetricsScanner) {
		metrics := globalScannerMetrics.report()
		m.Aggregated.Scanner = &metrics
	}
	if types.Contains(madmin.MetricsOS) {
		metrics := globalOSMetrics.report()
		m.Aggregated.OS = &metrics
	}
	// Add types...

	// ByHost is a shallow reference, so careful about sharing.
	m.ByHost = map[string]madmin.Metrics{globalMinioAddr: m.Aggregated}
	m.Hosts = append(m.Hosts, globalMinioAddr)

	return m
}

// Map between disk device id and io stats
type ioStats map[int32]madmin.IOStats

var ioStatsCache timedValue

func getIOStats() (info ioStats, err error) {
	ioStatsCache.Once.Do(func() {
		ioStatsCache.TTL = 10 * time.Second
		ioStatsCache.Update = func() (interface{}, error) {
			proc, err := os.Open("./diskstats")
			if err != nil {
				return nil, err
			}
			defer proc.Close()

			ret := make(ioStats)

			sc := bufio.NewScanner(proc)
			for sc.Scan() {
				line := sc.Text()
				fields := strings.Fields(line)
				if len(fields) < 14 {
					continue
				}
				major, err := strconv.ParseUint((fields[0]), 10, 16)
				if err != nil {
					return ret, err
				}
				minor, err := strconv.ParseUint((fields[1]), 10, 16)
				if err != nil {
					return ret, err
				}

				reads, err := strconv.ParseUint((fields[3]), 10, 64)
				if err != nil {
					return ret, err
				}
				mergedReads, err := strconv.ParseUint((fields[4]), 10, 64)
				if err != nil {
					return ret, err
				}
				rbytes, err := strconv.ParseUint((fields[5]), 10, 64)
				if err != nil {
					return ret, err
				}
				rtime, err := strconv.ParseUint((fields[6]), 10, 64)
				if err != nil {
					return ret, err
				}
				writes, err := strconv.ParseUint((fields[7]), 10, 64)
				if err != nil {
					return ret, err
				}
				mergedWrites, err := strconv.ParseUint((fields[8]), 10, 64)
				if err != nil {
					return ret, err
				}
				wbytes, err := strconv.ParseUint((fields[9]), 10, 64)
				if err != nil {
					return ret, err
				}
				wtime, err := strconv.ParseUint((fields[10]), 10, 64)
				if err != nil {
					return ret, err
				}
				iopsInProgress, err := strconv.ParseUint((fields[11]), 10, 64)
				if err != nil {
					return ret, err
				}
				iotime, err := strconv.ParseUint((fields[12]), 10, 64)
				if err != nil {
					return ret, err
				}
				weightedIO, err := strconv.ParseUint((fields[13]), 10, 64)
				if err != nil {
					return ret, err
				}
				d := madmin.IOStats{
					ReadBlocks:       rbytes,
					WriteBlocks:      wbytes,
					ReadCount:        reads,
					WriteCount:       writes,
					MergedReadCount:  mergedReads,
					MergedWriteCount: mergedWrites,
					ReadTime:         rtime,
					WriteTime:        wtime,
					IopsInProgress:   iopsInProgress,
					IoTime:           iotime,
					WeightedIO:       weightedIO,
				}

				devid := int32(major<<16 + minor)
				ret[devid] = d
			}

			if err := sc.Err(); err != nil {
				return nil, err
			}

			return ret, nil
		}
	})

	v, err := ioStatsCache.Get()
	if v != nil {
		info = v.(ioStats)
	}
	return info, err
}

func collectLocalDisksMetrics(disks map[string]struct{}) map[string]madmin.DiskMetric {
	objLayer := newObjectLayerFn()
	if objLayer == nil {
		return nil
	}

	metrics := make(map[string]madmin.DiskMetric)

	procStats, procErr := getIOStats()
	if procErr != nil {
		fmt.Println(procErr)
	}

	// only need Disks information in server mode.
	storageInfo, errs := objLayer.LocalStorageInfo(GlobalContext)

	for i, disk := range storageInfo.Disks {
		if len(disks) != 0 {
			_, ok := disks[disk.Endpoint]
			if !ok {
				continue
			}
		}

		if errs[i] != nil {
			metrics[disk.Endpoint] = madmin.DiskMetric{NDisks: 1, Offline: 1}
			continue
		}

		var d madmin.DiskMetric
		d.NDisks = 1
		if disk.Healing {
			d.Healing++
		}
		if disk.Metrics != nil {
			d.LifeTimeOps = make(map[string]uint64, len(disk.Metrics.APICalls))
			for k, v := range disk.Metrics.APICalls {
				if v != 0 {
					d.LifeTimeOps[k] = v
				}
			}
			d.LastMinute.Operations = make(map[string]madmin.TimedAction, len(disk.Metrics.APICalls))
			for k, v := range disk.Metrics.LastMinute {
				if v.Count != 0 {
					d.LastMinute.Operations[k] = v
				}
			}
		}

		// get disk
		if procErr == nil {
			d.IOStats = procStats[disk.DevID]
		}

		metrics[disk.Endpoint] = d
	}
	return metrics
}

func collectRemoteMetrics(ctx context.Context, types madmin.MetricType, hosts map[string]struct{}, disks map[string]struct{}) (m madmin.RealtimeMetrics) {
	if !globalIsDistErasure {
		return
	}
	all := globalNotificationSys.GetMetrics(ctx, types, hosts, disks)
	for _, remote := range all {
		m.Merge(&remote)
	}
	return m
}
