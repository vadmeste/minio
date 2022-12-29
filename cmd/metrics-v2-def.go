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

// MetricNamespace is top level grouping of metrics to create the metric name.
type MetricNamespace string

// MetricSubsystem is the sub grouping for metrics within a namespace.
type MetricSubsystem string

const (
	bucketMetricNamespace    MetricNamespace = "minio_bucket"
	clusterMetricNamespace   MetricNamespace = "minio_cluster"
	healMetricNamespace      MetricNamespace = "minio_heal"
	interNodeMetricNamespace MetricNamespace = "minio_inter_node"
	nodeMetricNamespace      MetricNamespace = "minio_node"
	minioMetricNamespace     MetricNamespace = "minio"
	s3MetricNamespace        MetricNamespace = "minio_s3"
)

const (
	cacheSubsystem            MetricSubsystem = "cache"
	capacityRawSubsystem      MetricSubsystem = "capacity_raw"
	capacityUsableSubsystem   MetricSubsystem = "capacity_usable"
	diskSubsystem             MetricSubsystem = "disk"
	fileDescriptorSubsystem   MetricSubsystem = "file_descriptor"
	goRoutines                MetricSubsystem = "go_routine"
	ioSubsystem               MetricSubsystem = "io"
	nodesSubsystem            MetricSubsystem = "nodes"
	objectsSubsystem          MetricSubsystem = "objects"
	processSubsystem          MetricSubsystem = "process"
	replicationSubsystem      MetricSubsystem = "replication"
	requestsSubsystem         MetricSubsystem = "requests"
	requestsRejectedSubsystem MetricSubsystem = "requests_rejected"
	timeSubsystem             MetricSubsystem = "time"
	trafficSubsystem          MetricSubsystem = "traffic"
	softwareSubsystem         MetricSubsystem = "software"
	sysCallSubsystem          MetricSubsystem = "syscall"
	usageSubsystem            MetricSubsystem = "usage"
	quotaSubsystem            MetricSubsystem = "quota"
	ilmSubsystem              MetricSubsystem = "ilm"
	scannerSubsystem          MetricSubsystem = "scanner"
	iamSubsystem              MetricSubsystem = "iam"
	kmsSubsystem              MetricSubsystem = "kms"
	notifySubsystem           MetricSubsystem = "notify"
	auditSubsystem            MetricSubsystem = "audit"
)

// MetricName are the individual names for the metric.
type MetricName string

const (
	authTotal      MetricName = "auth_total"
	canceledTotal  MetricName = "canceled_total"
	errorsTotal    MetricName = "errors_total"
	headerTotal    MetricName = "header_total"
	healTotal      MetricName = "heal_total"
	hitsTotal      MetricName = "hits_total"
	inflightTotal  MetricName = "inflight_total"
	invalidTotal   MetricName = "invalid_total"
	limitTotal     MetricName = "limit_total"
	missedTotal    MetricName = "missed_total"
	waitingTotal   MetricName = "waiting_total"
	incomingTotal  MetricName = "incoming_total"
	objectTotal    MetricName = "object_total"
	offlineTotal   MetricName = "offline_total"
	onlineTotal    MetricName = "online_total"
	openTotal      MetricName = "open_total"
	readTotal      MetricName = "read_total"
	timestampTotal MetricName = "timestamp_total"
	writeTotal     MetricName = "write_total"
	total          MetricName = "total"
	freeInodes     MetricName = "free_inodes"

	failedCount     MetricName = "failed_count"
	failedBytes     MetricName = "failed_bytes"
	freeBytes       MetricName = "free_bytes"
	readBytes       MetricName = "read_bytes"
	rcharBytes      MetricName = "rchar_bytes"
	receivedBytes   MetricName = "received_bytes"
	latencyMilliSec MetricName = "latency_ms"
	sentBytes       MetricName = "sent_bytes"
	totalBytes      MetricName = "total_bytes"
	usedBytes       MetricName = "used_bytes"
	writeBytes      MetricName = "write_bytes"
	wcharBytes      MetricName = "wchar_bytes"

	latencyMicroSec MetricName = "latency_us"
	latencyNanoSec  MetricName = "latency_ns"

	usagePercent MetricName = "update_percent"

	commitInfo  MetricName = "commit_info"
	usageInfo   MetricName = "usage_info"
	versionInfo MetricName = "version_info"

	sizeDistribution = "size_distribution"
	ttfbDistribution = "ttfb_seconds_distribution"

	lastActivityTime = "last_activity_nano_seconds"
	startTime        = "starttime_seconds"
	upTime           = "uptime_seconds"
	memory           = "resident_memory_bytes"
	cpu              = "cpu_total_seconds"

	expiryPendingTasks     MetricName = "expiry_pending_tasks"
	transitionPendingTasks MetricName = "transition_pending_tasks"
	transitionActiveTasks  MetricName = "transition_active_tasks"

	transitionedBytes    MetricName = "transitioned_bytes"
	transitionedObjects  MetricName = "transitioned_objects"
	transitionedVersions MetricName = "transitioned_versions"

	kmsOnline          = "online"
	kmsRequestsSuccess = "request_success"
	kmsRequestsError   = "request_error"
	kmsRequestsFail    = "request_failure"
	kmsUptime          = "uptime"
)

// MetricType for the types of metrics supported
type MetricType string

const (
	gaugeMetric     = "gaugeMetric"
	counterMetric   = "counterMetric"
	histogramMetric = "histogramMetric"
)

// MetricDescription describes the metric
type MetricDescription struct {
	Namespace MetricNamespace `json:"MetricNamespace"`
	Subsystem MetricSubsystem `json:"Subsystem"`
	Name      MetricName      `json:"MetricName"`
	Help      string          `json:"Help"`
	Type      MetricType      `json:"Type"`
}

func getClusterCapacityTotalBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: capacityRawSubsystem,
		Name:      totalBytes,
		Help:      "Total capacity online in the cluster",
		Type:      gaugeMetric,
	}
}

func getClusterCapacityFreeBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: capacityRawSubsystem,
		Name:      freeBytes,
		Help:      "Total free capacity online in the cluster",
		Type:      gaugeMetric,
	}
}

func getClusterCapacityUsageBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: capacityUsableSubsystem,
		Name:      totalBytes,
		Help:      "Total usable capacity online in the cluster",
		Type:      gaugeMetric,
	}
}

func getClusterCapacityUsageFreeBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: capacityUsableSubsystem,
		Name:      freeBytes,
		Help:      "Total free usable capacity online in the cluster",
		Type:      gaugeMetric,
	}
}

func getNodeDriveAPILatencyMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      latencyMicroSec,
		Help:      "Average last minute latency in µs for drive API storage operations",
		Type:      gaugeMetric,
	}
}

func getNodeDriveUsedBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      usedBytes,
		Help:      "Total storage used on a drive",
		Type:      gaugeMetric,
	}
}

func getNodeDriveFreeBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      freeBytes,
		Help:      "Total storage available on a drive",
		Type:      gaugeMetric,
	}
}

func getClusterDrivesOfflineTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      offlineTotal,
		Help:      "Total drives offline",
		Type:      gaugeMetric,
	}
}

func getClusterDrivesOnlineTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      onlineTotal,
		Help:      "Total drives online",
		Type:      gaugeMetric,
	}
}

func getClusterDrivesTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      total,
		Help:      "Total drives",
		Type:      gaugeMetric,
	}
}

func getNodeDrivesOfflineTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      offlineTotal,
		Help:      "Total drives offline",
		Type:      gaugeMetric,
	}
}

func getNodeDrivesOnlineTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      onlineTotal,
		Help:      "Total drives online",
		Type:      gaugeMetric,
	}
}

func getNodeDrivesTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      total,
		Help:      "Total drives",
		Type:      gaugeMetric,
	}
}

func getNodeDrivesFreeInodes() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      freeInodes,
		Help:      "Total free inodes",
		Type:      gaugeMetric,
	}
}

func getNodeDriveTotalBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      totalBytes,
		Help:      "Total storage on a drive",
		Type:      gaugeMetric,
	}
}

func getUsageLastScanActivityMD() MetricDescription {
	return MetricDescription{
		Namespace: minioMetricNamespace,
		Subsystem: usageSubsystem,
		Name:      lastActivityTime,
		Help:      "Time elapsed (in nano seconds) since last scan activity. This is set to 0 until first scan cycle",
		Type:      gaugeMetric,
	}
}

func getBucketUsageQuotaTotalBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: quotaSubsystem,
		Name:      totalBytes,
		Help:      "Total bucket quota size in bytes",
		Type:      gaugeMetric,
	}
}

func getBucketTrafficReceivedBytes() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: trafficSubsystem,
		Name:      receivedBytes,
		Help:      "Total number of S3 bytes received for this bucket",
		Type:      gaugeMetric,
	}
}

func getBucketTrafficSentBytes() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: trafficSubsystem,
		Name:      sentBytes,
		Help:      "Total number of S3 bytes sent for this bucket",
		Type:      gaugeMetric,
	}
}

func getBucketUsageTotalBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: usageSubsystem,
		Name:      totalBytes,
		Help:      "Total bucket size in bytes",
		Type:      gaugeMetric,
	}
}

func getBucketUsageObjectsTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: usageSubsystem,
		Name:      objectTotal,
		Help:      "Total number of objects",
		Type:      gaugeMetric,
	}
}

func getBucketRepLatencyMD() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: replicationSubsystem,
		Name:      latencyMilliSec,
		Help:      "Replication latency in milliseconds",
		Type:      histogramMetric,
	}
}

func getBucketRepFailedBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: replicationSubsystem,
		Name:      failedBytes,
		Help:      "Total number of bytes failed at least once to replicate",
		Type:      gaugeMetric,
	}
}

func getBucketRepSentBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: replicationSubsystem,
		Name:      sentBytes,
		Help:      "Total number of bytes replicated to the target bucket",
		Type:      gaugeMetric,
	}
}

func getBucketRepReceivedBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: replicationSubsystem,
		Name:      receivedBytes,
		Help:      "Total number of bytes replicated to this bucket from another source bucket",
		Type:      gaugeMetric,
	}
}

func getBucketRepFailedOperationsMD() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: replicationSubsystem,
		Name:      failedCount,
		Help:      "Total number of objects which failed replication",
		Type:      gaugeMetric,
	}
}

func getBucketObjectDistributionMD() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: objectsSubsystem,
		Name:      sizeDistribution,
		Help:      "Distribution of object sizes in the bucket, includes label for the bucket name",
		Type:      histogramMetric,
	}
}

func getInternodeFailedRequests() MetricDescription {
	return MetricDescription{
		Namespace: interNodeMetricNamespace,
		Subsystem: trafficSubsystem,
		Name:      errorsTotal,
		Help:      "Total number of failed internode calls",
		Type:      counterMetric,
	}
}

func getInternodeTCPDialTimeout() MetricDescription {
	return MetricDescription{
		Namespace: interNodeMetricNamespace,
		Subsystem: trafficSubsystem,
		Name:      "dial_errors",
		Help:      "Total number of internode TCP dial timeouts and errors",
		Type:      counterMetric,
	}
}

func getInternodeTCPAvgDuration() MetricDescription {
	return MetricDescription{
		Namespace: interNodeMetricNamespace,
		Subsystem: trafficSubsystem,
		Name:      "dial_avg_time",
		Help:      "Average time of internodes TCP dial calls",
		Type:      gaugeMetric,
	}
}

func getInterNodeSentBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: interNodeMetricNamespace,
		Subsystem: trafficSubsystem,
		Name:      sentBytes,
		Help:      "Total number of bytes sent to the other peer nodes",
		Type:      counterMetric,
	}
}

func getInterNodeReceivedBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: interNodeMetricNamespace,
		Subsystem: trafficSubsystem,
		Name:      receivedBytes,
		Help:      "Total number of bytes received from other peer nodes",
		Type:      counterMetric,
	}
}

func getS3SentBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: trafficSubsystem,
		Name:      sentBytes,
		Help:      "Total number of s3 bytes sent",
		Type:      counterMetric,
	}
}

func getS3ReceivedBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: trafficSubsystem,
		Name:      receivedBytes,
		Help:      "Total number of s3 bytes received",
		Type:      counterMetric,
	}
}

func getS3RequestsInFlightMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsSubsystem,
		Name:      inflightTotal,
		Help:      "Total number of S3 requests currently in flight",
		Type:      gaugeMetric,
	}
}

func getS3RequestsInQueueMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsSubsystem,
		Name:      waitingTotal,
		Help:      "Number of S3 requests in the waiting queue",
		Type:      gaugeMetric,
	}
}

func getIncomingS3RequestsMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsSubsystem,
		Name:      incomingTotal,
		Help:      "Volatile number of total incoming S3 requests",
		Type:      gaugeMetric,
	}
}

func getS3RequestsTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsSubsystem,
		Name:      total,
		Help:      "Total number S3 requests",
		Type:      counterMetric,
	}
}

func getS3RequestsErrorsMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsSubsystem,
		Name:      errorsTotal,
		Help:      "Total number S3 requests with (4xx and 5xx) errors",
		Type:      counterMetric,
	}
}

func getS3Requests4xxErrorsMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsSubsystem,
		Name:      "4xx_" + errorsTotal,
		Help:      "Total number S3 requests with (4xx) errors",
		Type:      counterMetric,
	}
}

func getS3Requests5xxErrorsMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsSubsystem,
		Name:      "5xx_" + errorsTotal,
		Help:      "Total number S3 requests with (5xx) errors",
		Type:      counterMetric,
	}
}

func getS3RequestsCanceledMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsSubsystem,
		Name:      canceledTotal,
		Help:      "Total number S3 requests that were canceled from the client while processing",
		Type:      counterMetric,
	}
}

func getS3RejectedAuthRequestsTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsRejectedSubsystem,
		Name:      authTotal,
		Help:      "Total number S3 requests rejected for auth failure",
		Type:      counterMetric,
	}
}

func getS3RejectedHeaderRequestsTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsRejectedSubsystem,
		Name:      headerTotal,
		Help:      "Total number S3 requests rejected for invalid header",
		Type:      counterMetric,
	}
}

func getS3RejectedTimestampRequestsTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsRejectedSubsystem,
		Name:      timestampTotal,
		Help:      "Total number S3 requests rejected for invalid timestamp",
		Type:      counterMetric,
	}
}

func getS3RejectedInvalidRequestsTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsRejectedSubsystem,
		Name:      invalidTotal,
		Help:      "Total number S3 invalid requests",
		Type:      counterMetric,
	}
}

func getCacheHitsTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: cacheSubsystem,
		Name:      hitsTotal,
		Help:      "Total number of drive cache hits",
		Type:      counterMetric,
	}
}

func getCacheHitsMissedTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: cacheSubsystem,
		Name:      missedTotal,
		Help:      "Total number of drive cache misses",
		Type:      counterMetric,
	}
}

func getCacheUsagePercentMD() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: minioNamespace,
		Name:      usagePercent,
		Help:      "Total percentage cache usage",
		Type:      gaugeMetric,
	}
}

func getCacheUsageInfoMD() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: cacheSubsystem,
		Name:      usageInfo,
		Help:      "Total percentage cache usage, value of 1 indicates high and 0 low, label level is set as well",
		Type:      gaugeMetric,
	}
}

func getCacheUsedBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: cacheSubsystem,
		Name:      usedBytes,
		Help:      "Current cache usage in bytes",
		Type:      gaugeMetric,
	}
}

func getCacheTotalBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: cacheSubsystem,
		Name:      totalBytes,
		Help:      "Total size of cache drive in bytes",
		Type:      gaugeMetric,
	}
}

func getCacheSentBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: cacheSubsystem,
		Name:      sentBytes,
		Help:      "Total number of bytes served from cache",
		Type:      counterMetric,
	}
}

func getHealObjectsTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: healMetricNamespace,
		Subsystem: objectsSubsystem,
		Name:      total,
		Help:      "Objects scanned in current self healing run",
		Type:      gaugeMetric,
	}
}

func getHealObjectsHealTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: healMetricNamespace,
		Subsystem: objectsSubsystem,
		Name:      healTotal,
		Help:      "Objects healed in current self healing run",
		Type:      gaugeMetric,
	}
}

func getHealObjectsFailTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: healMetricNamespace,
		Subsystem: objectsSubsystem,
		Name:      errorsTotal,
		Help:      "Objects for which healing failed in current self healing run",
		Type:      gaugeMetric,
	}
}

func getHealLastActivityTimeMD() MetricDescription {
	return MetricDescription{
		Namespace: healMetricNamespace,
		Subsystem: timeSubsystem,
		Name:      lastActivityTime,
		Help:      "Time elapsed (in nano seconds) since last self healing activity. This is set to -1 until initial self heal activity",
		Type:      gaugeMetric,
	}
}

func getNodeOnlineTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: nodesSubsystem,
		Name:      onlineTotal,
		Help:      "Total number of MinIO nodes online",
		Type:      gaugeMetric,
	}
}

func getNodeOfflineTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: nodesSubsystem,
		Name:      offlineTotal,
		Help:      "Total number of MinIO nodes offline",
		Type:      gaugeMetric,
	}
}

func getMinIOVersionMD() MetricDescription {
	return MetricDescription{
		Namespace: minioMetricNamespace,
		Subsystem: softwareSubsystem,
		Name:      versionInfo,
		Help:      "MinIO Release tag for the server",
		Type:      gaugeMetric,
	}
}

func getMinIOCommitMD() MetricDescription {
	return MetricDescription{
		Namespace: minioMetricNamespace,
		Subsystem: softwareSubsystem,
		Name:      commitInfo,
		Help:      "Git commit hash for the MinIO release",
		Type:      gaugeMetric,
	}
}

func getS3TTFBDistributionMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: timeSubsystem,
		Name:      ttfbDistribution,
		Help:      "Distribution of the time to first byte across API calls",
		Type:      gaugeMetric,
	}
}

func getMinioFDOpenMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: fileDescriptorSubsystem,
		Name:      openTotal,
		Help:      "Total number of open file descriptors by the MinIO Server process",
		Type:      gaugeMetric,
	}
}

func getMinioFDLimitMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: fileDescriptorSubsystem,
		Name:      limitTotal,
		Help:      "Limit on total number of open file descriptors for the MinIO Server process",
		Type:      gaugeMetric,
	}
}

func getMinioProcessIOWriteBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: ioSubsystem,
		Name:      writeBytes,
		Help:      "Total bytes written by the process to the underlying storage system, /proc/[pid]/io write_bytes",
		Type:      counterMetric,
	}
}

func getMinioProcessIOReadBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: ioSubsystem,
		Name:      readBytes,
		Help:      "Total bytes read by the process from the underlying storage system, /proc/[pid]/io read_bytes",
		Type:      counterMetric,
	}
}

func getMinioProcessIOWriteCachedBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: ioSubsystem,
		Name:      wcharBytes,
		Help:      "Total bytes written by the process to the underlying storage system including page cache, /proc/[pid]/io wchar",
		Type:      counterMetric,
	}
}

func getMinioProcessIOReadCachedBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: ioSubsystem,
		Name:      rcharBytes,
		Help:      "Total bytes read by the process from the underlying storage system including cache, /proc/[pid]/io rchar",
		Type:      counterMetric,
	}
}

func getMinIOProcessSysCallRMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: sysCallSubsystem,
		Name:      readTotal,
		Help:      "Total read SysCalls to the kernel. /proc/[pid]/io syscr",
		Type:      counterMetric,
	}
}

func getMinIOProcessSysCallWMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: sysCallSubsystem,
		Name:      writeTotal,
		Help:      "Total write SysCalls to the kernel. /proc/[pid]/io syscw",
		Type:      counterMetric,
	}
}

func getMinIOGORoutineCountMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: goRoutines,
		Name:      total,
		Help:      "Total number of go routines running",
		Type:      gaugeMetric,
	}
}

func getMinIOProcessStartTimeMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: processSubsystem,
		Name:      startTime,
		Help:      "Start time for MinIO process per node, time in seconds since Unix epoc",
		Type:      gaugeMetric,
	}
}

func getMinIOProcessUptimeMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: processSubsystem,
		Name:      upTime,
		Help:      "Uptime for MinIO process per node in seconds",
		Type:      gaugeMetric,
	}
}

func getMinIOProcessResidentMemory() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: processSubsystem,
		Name:      memory,
		Help:      "Resident memory size in bytes",
		Type:      gaugeMetric,
	}
}

func getMinIOProcessCPUTime() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: processSubsystem,
		Name:      cpu,
		Help:      "Total user and system CPU time spent in seconds",
		Type:      counterMetric,
	}
}

func getTransitionPendingTasksMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: ilmSubsystem,
		Name:      transitionPendingTasks,
		Help:      "Number of pending ILM transition tasks in the queue",
		Type:      gaugeMetric,
	}
}

func getTransitionActiveTasksMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: ilmSubsystem,
		Name:      transitionActiveTasks,
		Help:      "Number of active ILM transition tasks",
		Type:      gaugeMetric,
	}
}

func getExpiryPendingTasksMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: ilmSubsystem,
		Name:      expiryPendingTasks,
		Help:      "Number of pending ILM expiry tasks in the queue",
		Type:      gaugeMetric,
	}
}

func getNodeScannerObjectsScanned() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: scannerSubsystem,
		Name:      "objects_scanned",
		Help:      "Total number of unique objects scanned since server start",
		Type:      counterMetric,
	}
}

func getNodeScannerVersionsScanned() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: scannerSubsystem,
		Name:      "versions_scanned",
		Help:      "Total number of object versions scanned since server start",
		Type:      counterMetric,
	}
}

func getNodeScannerDirectoriesScanned() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: scannerSubsystem,
		Name:      "directories_scanned",
		Help:      "Total number of directories scanned since server start",
		Type:      counterMetric,
	}
}

func getNodeScannerBucketScansStarted() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: scannerSubsystem,
		Name:      "bucket_scans_started",
		Help:      "Total number of bucket scans started since server start",
		Type:      counterMetric,
	}
}

func getNodeScannerBucketScansFinished() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: scannerSubsystem,
		Name:      "bucket_scans_finished",
		Help:      "Total number of bucket scans finished since server start",
		Type:      counterMetric,
	}
}

func getNodeILMVersionsScanned() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: ilmSubsystem,
		Name:      "versions_scanned",
		Help:      "Total number of object versions checked for ilm actions since server start",
		Type:      counterMetric,
	}
}

func getNodeIAMLastSyncDuration() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: iamSubsystem,
		Name:      "last_sync_duration_millis",
		Help:      "Last successful IAM data sync duration in milliseconds",
		Type:      gaugeMetric,
	}
}

func getNodeIAMSinceLastSync() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: iamSubsystem,
		Name:      "since_last_sync_millis",
		Help:      "Time (in milliseconds) since last successful IAM data sync. This is set to 0 until the first sync after server start.",
		Type:      gaugeMetric,
	}
}

func getNodeIAMSyncSuccesses() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: iamSubsystem,
		Name:      "sync_successes",
		Help:      "Number of successful IAM data syncs since server start.",
		Type:      counterMetric,
	}
}

func getNodeIAMSyncFailures() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: iamSubsystem,
		Name:      "sync_failures",
		Help:      "Number of failed IAM data syncs since server start.",
		Type:      counterMetric,
	}
}

func getMinIONotifyCurrentSendInProgress() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: notifySubsystem,
		Name:      "current_send_in_progress",
		Help:      "Number of concurrent async Send calls active to all targets",
		Type:      gaugeMetric,
	}
}

func getMinIONotifyTargetQueueLength() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: notifySubsystem,
		Name:      "target_queue_length",
		Help:      "Number of unsent notifications in queue for target",
		Type:      gaugeMetric,
	}
}

func getMinIOAuditTargetQueueLength() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: auditSubsystem,
		Name:      "target_queue_length",
		Help:      "Number of unsent messages in queue for target",
		Type:      gaugeMetric,
	}
}

func getMinIOAuditTotalMessages() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: auditSubsystem,
		Name:      "total_messages",
		Help:      "Total number of messages sent since start",
		Type:      counterMetric,
	}
}

func getMinIOAuditFailedMessages() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: auditSubsystem,
		Name:      "failed_messages",
		Help:      "Total number of messages that failed to send since start",
		Type:      counterMetric,
	}
}

func getClusterTransitionedBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: ilmSubsystem,
		Name:      transitionedBytes,
		Help:      "Total bytes transitioned to a tier",
		Type:      gaugeMetric,
	}
}

func getClusterTransitionedObjectsMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: ilmSubsystem,
		Name:      transitionedObjects,
		Help:      "Total number of objects transitioned to a tier",
		Type:      gaugeMetric,
	}
}

func getClusterTransitionedVersionsMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: ilmSubsystem,
		Name:      transitionedVersions,
		Help:      "Total number of versions transitioned to a tier",
		Type:      gaugeMetric,
	}
}

func getClusterKMSOnline() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: kmsSubsystem,
		Name:      kmsOnline,
		Help:      "Reports whether the KMS is online (1) or offline (0)",
		Type:      gaugeMetric,
	}
}

func getClusterKMSRequestsSuccess() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: kmsSubsystem,
		Name:      kmsRequestsSuccess,
		Help:      "Number of KMS requests that succeeded",
		Type:      counterMetric,
	}
}

func getClusterKMSRequestsError() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: kmsSubsystem,
		Name:      kmsRequestsError,
		Help:      "Number of KMS requests that failed due to some error. (HTTP 4xx status code)",
		Type:      counterMetric,
	}
}

func getClusterKMSRequestsFail() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: kmsSubsystem,
		Name:      kmsRequestsFail,
		Help:      "Number of KMS requests that failed due to some internal failure. (HTTP 5xx status code)",
		Type:      counterMetric,
	}
}

func getClusterKMSUptime() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: kmsSubsystem,
		Name:      kmsUptime,
		Help:      "The time the KMS has been up and running in seconds.",
		Type:      counterMetric,
	}
}
