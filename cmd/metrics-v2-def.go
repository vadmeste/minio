package cmd

//prom:doc minio_cluster_capacity_raw_total_bytes, Total capacity online in the cluster
func getClusterCapacityTotalBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: capacityRawSubsystem,
		Name:      totalBytes,
		Help:      "Total capacity online in the cluster",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_cluster_capacity_raw_free_bytes, Total free capacity online in the cluster
func getClusterCapacityFreeBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: capacityRawSubsystem,
		Name:      freeBytes,
		Help:      "Total free capacity online in the cluster",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_cluster_capacity_usable_total_bytes, Total usable capacity online in the cluster
func getClusterCapacityUsageBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: capacityUsableSubsystem,
		Name:      totalBytes,
		Help:      "Total usable capacity online in the cluster",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_cluster_capacity_usable_free_bytes, Total free usable capacity online in the cluster
func getClusterCapacityUsageFreeBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: capacityUsableSubsystem,
		Name:      freeBytes,
		Help:      "Total free usable capacity online in the cluster",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_node_disk_latency_us, Average last minute latency in µs for drive API storage operations
func getNodeDriveAPILatencyMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      latencyMicroSec,
		Help:      "Average last minute latency in µs for drive API storage operations",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_node_disk_used_bytes, Total storage used on a drive
func getNodeDriveUsedBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      usedBytes,
		Help:      "Total storage used on a drive",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_node_disk_free_bytes, Total storage available on a drive
func getNodeDriveFreeBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      freeBytes,
		Help:      "Total storage available on a drive",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_cluster_disk_offline_total, Total drives offline
func getClusterDrivesOfflineTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      offlineTotal,
		Help:      "Total drives offline",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_cluster_disk_online_total, Total drives online
func getClusterDrivesOnlineTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      onlineTotal,
		Help:      "Total drives online",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_cluster_disk_total, Total drives
func getClusterDrivesTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      total,
		Help:      "Total drives",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_node_disk_offline_total, Total drives offline
func getNodeDrivesOfflineTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      offlineTotal,
		Help:      "Total drives offline",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_node_disk_online_total, Total drives online
func getNodeDrivesOnlineTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      onlineTotal,
		Help:      "Total drives online",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_node_disk_total, Total drives
func getNodeDrivesTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      total,
		Help:      "Total drives",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_node_disk_free_inodes, Total free inodes
func getNodeDrivesFreeInodes() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      freeInodes,
		Help:      "Total free inodes",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_node_disk_total_bytes, Total storage on a drive
func getNodeDriveTotalBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: diskSubsystem,
		Name:      totalBytes,
		Help:      "Total storage on a drive",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_usage_last_activity_nano_seconds, Time elapsed (in nano seconds) since last scan activity. This is set to 0 until first scan cycle
func getUsageLastScanActivityMD() MetricDescription {
	return MetricDescription{
		Namespace: minioMetricNamespace,
		Subsystem: usageSubsystem,
		Name:      lastActivityTime,
		Help:      "Time elapsed (in nano seconds) since last scan activity. This is set to 0 until first scan cycle",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_bucket_quota_total_bytes, Total bucket quota size in bytes
func getBucketUsageQuotaTotalBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: quotaSubsystem,
		Name:      totalBytes,
		Help:      "Total bucket quota size in bytes",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_bucket_traffic_received_bytes, Total number of S3 bytes received for this bucket
func getBucketTrafficReceivedBytes() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: trafficSubsystem,
		Name:      receivedBytes,
		Help:      "Total number of S3 bytes received for this bucket",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_bucket_traffic_sent_bytes, Total number of S3 bytes sent for this bucket
func getBucketTrafficSentBytes() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: trafficSubsystem,
		Name:      sentBytes,
		Help:      "Total number of S3 bytes sent for this bucket",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_bucket_usage_total_bytes, Total bucket size in bytes
func getBucketUsageTotalBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: usageSubsystem,
		Name:      totalBytes,
		Help:      "Total bucket size in bytes",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_bucket_usage_object_total, Total number of objects
func getBucketUsageObjectsTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: usageSubsystem,
		Name:      objectTotal,
		Help:      "Total number of objects",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_bucket_replication_latency_ms, Replication latency in milliseconds
func getBucketRepLatencyMD() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: replicationSubsystem,
		Name:      latencyMilliSec,
		Help:      "Replication latency in milliseconds",
		Type:      histogramMetric,
	}
}

//prom:doc minio_bucket_replication_failed_bytes, Total number of bytes failed at least once to replicate
func getBucketRepFailedBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: replicationSubsystem,
		Name:      failedBytes,
		Help:      "Total number of bytes failed at least once to replicate",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_bucket_replication_sent_bytes, Total number of bytes replicated to the target bucket
func getBucketRepSentBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: replicationSubsystem,
		Name:      sentBytes,
		Help:      "Total number of bytes replicated to the target bucket",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_bucket_replication_received_bytes, Total number of bytes replicated to this bucket from another source bucket
func getBucketRepReceivedBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: replicationSubsystem,
		Name:      receivedBytes,
		Help:      "Total number of bytes replicated to this bucket from another source bucket",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_bucket_replication_failed_count, Total number of objects which failed replication
func getBucketRepFailedOperationsMD() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: replicationSubsystem,
		Name:      failedCount,
		Help:      "Total number of objects which failed replication",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_bucket_objects_size_distribution, Distribution of object sizes in the bucket, includes label for the bucket name
func getBucketObjectDistributionMD() MetricDescription {
	return MetricDescription{
		Namespace: bucketMetricNamespace,
		Subsystem: objectsSubsystem,
		Name:      sizeDistribution,
		Help:      "Distribution of object sizes in the bucket, includes label for the bucket name",
		Type:      histogramMetric,
	}
}

//prom:doc minio_inter_node_traffic_errors_total, Total number of failed internode calls
func getInternodeFailedRequests() MetricDescription {
	return MetricDescription{
		Namespace: interNodeMetricNamespace,
		Subsystem: trafficSubsystem,
		Name:      errorsTotal,
		Help:      "Total number of failed internode calls",
		Type:      counterMetric,
	}
}

//prom:doc minio_inter_node_traffic_dial_errors, Total number of internode TCP dial timeouts and errors
func getInternodeTCPDialTimeout() MetricDescription {
	return MetricDescription{
		Namespace: interNodeMetricNamespace,
		Subsystem: trafficSubsystem,
		Name:      "dial_errors",
		Help:      "Total number of internode TCP dial timeouts and errors",
		Type:      counterMetric,
	}
}

//prom:doc minio_inter_node_traffic_dial_avg_time, Average time of internodes TCP dial calls
func getInternodeTCPAvgDuration() MetricDescription {
	return MetricDescription{
		Namespace: interNodeMetricNamespace,
		Subsystem: trafficSubsystem,
		Name:      "dial_avg_time",
		Help:      "Average time of internodes TCP dial calls",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_inter_node_traffic_sent_bytes, Total number of bytes sent to the other peer nodes
func getInterNodeSentBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: interNodeMetricNamespace,
		Subsystem: trafficSubsystem,
		Name:      sentBytes,
		Help:      "Total number of bytes sent to the other peer nodes",
		Type:      counterMetric,
	}
}

//prom:doc minio_inter_node_traffic_received_bytes, Total number of bytes received from other peer nodes
func getInterNodeReceivedBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: interNodeMetricNamespace,
		Subsystem: trafficSubsystem,
		Name:      receivedBytes,
		Help:      "Total number of bytes received from other peer nodes",
		Type:      counterMetric,
	}
}

//prom:doc minio_s3_traffic_sent_bytes, Total number of s3 bytes sent
func getS3SentBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: trafficSubsystem,
		Name:      sentBytes,
		Help:      "Total number of s3 bytes sent",
		Type:      counterMetric,
	}
}

//prom:doc minio_s3_traffic_received_bytes, Total number of s3 bytes received
func getS3ReceivedBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: trafficSubsystem,
		Name:      receivedBytes,
		Help:      "Total number of s3 bytes received",
		Type:      counterMetric,
	}
}

//prom:doc minio_s3_requests_inflight_total, Total number of S3 requests currently in flight
func getS3RequestsInFlightMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsSubsystem,
		Name:      inflightTotal,
		Help:      "Total number of S3 requests currently in flight",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_s3_requests_waiting_total, Number of S3 requests in the waiting queue
func getS3RequestsInQueueMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsSubsystem,
		Name:      waitingTotal,
		Help:      "Number of S3 requests in the waiting queue",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_s3_requests_incoming_total, Volatile number of total incoming S3 requests
func getIncomingS3RequestsMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsSubsystem,
		Name:      incomingTotal,
		Help:      "Volatile number of total incoming S3 requests",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_s3_requests_total, Total number S3 requests
func getS3RequestsTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsSubsystem,
		Name:      total,
		Help:      "Total number S3 requests",
		Type:      counterMetric,
	}
}

//prom:doc minio_s3_requests_errors_total, Total number S3 requests with (4xx and 5xx) errors
func getS3RequestsErrorsMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsSubsystem,
		Name:      errorsTotal,
		Help:      "Total number S3 requests with (4xx and 5xx) errors",
		Type:      counterMetric,
	}
}

//prom:doc minio_s3_requests_4xx_errors_total, Total number S3 requests with (4xx) errors
func getS3Requests4xxErrorsMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsSubsystem,
		Name:      "4xx_" + errorsTotal,
		Help:      "Total number S3 requests with (4xx) errors",
		Type:      counterMetric,
	}
}

//prom:doc minio_s3_requests_5xx_errors_total, Total number S3 requests with (5xx) errors
func getS3Requests5xxErrorsMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsSubsystem,
		Name:      "5xx_" + errorsTotal,
		Help:      "Total number S3 requests with (5xx) errors",
		Type:      counterMetric,
	}
}

//prom:doc minio_s3_requests_canceled_total, Total number S3 requests that were canceled from the client while processing
func getS3RequestsCanceledMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsSubsystem,
		Name:      canceledTotal,
		Help:      "Total number S3 requests that were canceled from the client while processing",
		Type:      counterMetric,
	}
}

//prom:doc minio_s3_requests_rejected_auth_total, Total number S3 requests rejected for auth failure
func getS3RejectedAuthRequestsTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsRejectedSubsystem,
		Name:      authTotal,
		Help:      "Total number S3 requests rejected for auth failure",
		Type:      counterMetric,
	}
}

//prom:doc minio_s3_requests_rejected_header_total, Total number S3 requests rejected for invalid header
func getS3RejectedHeaderRequestsTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsRejectedSubsystem,
		Name:      headerTotal,
		Help:      "Total number S3 requests rejected for invalid header",
		Type:      counterMetric,
	}
}

//prom:doc minio_s3_requests_rejected_timestamp_total, Total number S3 requests rejected for invalid timestamp
func getS3RejectedTimestampRequestsTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsRejectedSubsystem,
		Name:      timestampTotal,
		Help:      "Total number S3 requests rejected for invalid timestamp",
		Type:      counterMetric,
	}
}

//prom:doc minio_s3_requests_rejected_invalid_total, Total number S3 invalid requests
func getS3RejectedInvalidRequestsTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: requestsRejectedSubsystem,
		Name:      invalidTotal,
		Help:      "Total number S3 invalid requests",
		Type:      counterMetric,
	}
}

//prom:doc minio_cache_hits_total, Total number of drive cache hits
func getCacheHitsTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: cacheSubsystem,
		Name:      hitsTotal,
		Help:      "Total number of drive cache hits",
		Type:      counterMetric,
	}
}

//prom:doc minio_cache_missed_total, Total number of drive cache misses
func getCacheHitsMissedTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: cacheSubsystem,
		Name:      missedTotal,
		Help:      "Total number of drive cache misses",
		Type:      counterMetric,
	}
}

//prom:doc minio_minio_update_percent, Total percentage cache usage
func getCacheUsagePercentMD() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: minioNamespace,
		Name:      usagePercent,
		Help:      "Total percentage cache usage",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_cache_usage_info, Total percentage cache usage, value of 1 indicates high and 0 low, label level is set as well
func getCacheUsageInfoMD() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: cacheSubsystem,
		Name:      usageInfo,
		Help:      "Total percentage cache usage, value of 1 indicates high and 0 low, label level is set as well",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_cache_used_bytes, Current cache usage in bytes
func getCacheUsedBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: cacheSubsystem,
		Name:      usedBytes,
		Help:      "Current cache usage in bytes",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_cache_total_bytes, Total size of cache drive in bytes
func getCacheTotalBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: cacheSubsystem,
		Name:      totalBytes,
		Help:      "Total size of cache drive in bytes",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_cache_sent_bytes, Total number of bytes served from cache
func getCacheSentBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: cacheSubsystem,
		Name:      sentBytes,
		Help:      "Total number of bytes served from cache",
		Type:      counterMetric,
	}
}

//prom:doc minio_heal_objects_total, Objects scanned in current self healing run
func getHealObjectsTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: healMetricNamespace,
		Subsystem: objectsSubsystem,
		Name:      total,
		Help:      "Objects scanned in current self healing run",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_heal_objects_heal_total, Objects healed in current self healing run
func getHealObjectsHealTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: healMetricNamespace,
		Subsystem: objectsSubsystem,
		Name:      healTotal,
		Help:      "Objects healed in current self healing run",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_heal_objects_errors_total, Objects for which healing failed in current self healing run
func getHealObjectsFailTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: healMetricNamespace,
		Subsystem: objectsSubsystem,
		Name:      errorsTotal,
		Help:      "Objects for which healing failed in current self healing run",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_heal_time_last_activity_nano_seconds, Time elapsed (in nano seconds) since last self healing activity. This is set to -1 until initial self heal activity
func getHealLastActivityTimeMD() MetricDescription {
	return MetricDescription{
		Namespace: healMetricNamespace,
		Subsystem: timeSubsystem,
		Name:      lastActivityTime,
		Help:      "Time elapsed (in nano seconds) since last self healing activity. This is set to -1 until initial self heal activity",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_cluster_nodes_online_total, Total number of MinIO nodes online
func getNodeOnlineTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: nodesSubsystem,
		Name:      onlineTotal,
		Help:      "Total number of MinIO nodes online",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_cluster_nodes_offline_total, Total number of MinIO nodes offline
func getNodeOfflineTotalMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: nodesSubsystem,
		Name:      offlineTotal,
		Help:      "Total number of MinIO nodes offline",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_software_version_info, MinIO Release tag for the server
func getMinIOVersionMD() MetricDescription {
	return MetricDescription{
		Namespace: minioMetricNamespace,
		Subsystem: softwareSubsystem,
		Name:      versionInfo,
		Help:      "MinIO Release tag for the server",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_software_commit_info, Git commit hash for the MinIO release
func getMinIOCommitMD() MetricDescription {
	return MetricDescription{
		Namespace: minioMetricNamespace,
		Subsystem: softwareSubsystem,
		Name:      commitInfo,
		Help:      "Git commit hash for the MinIO release",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_s3_time_ttfb_seconds_distribution, Distribution of the time to first byte across API calls
func getS3TTFBDistributionMD() MetricDescription {
	return MetricDescription{
		Namespace: s3MetricNamespace,
		Subsystem: timeSubsystem,
		Name:      ttfbDistribution,
		Help:      "Distribution of the time to first byte across API calls",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_node_file_descriptor_open_total, Total number of open file descriptors by the MinIO Server process
func getMinioFDOpenMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: fileDescriptorSubsystem,
		Name:      openTotal,
		Help:      "Total number of open file descriptors by the MinIO Server process",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_node_file_descriptor_limit_total, Limit on total number of open file descriptors for the MinIO Server process
func getMinioFDLimitMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: fileDescriptorSubsystem,
		Name:      limitTotal,
		Help:      "Limit on total number of open file descriptors for the MinIO Server process",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_node_io_write_bytes, Total bytes written by the process to the underlying storage system, /proc/[pid]/io write_bytes
func getMinioProcessIOWriteBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: ioSubsystem,
		Name:      writeBytes,
		Help:      "Total bytes written by the process to the underlying storage system, /proc/[pid]/io write_bytes",
		Type:      counterMetric,
	}
}

//prom:doc minio_node_io_read_bytes, Total bytes read by the process from the underlying storage system, /proc/[pid]/io read_bytes
func getMinioProcessIOReadBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: ioSubsystem,
		Name:      readBytes,
		Help:      "Total bytes read by the process from the underlying storage system, /proc/[pid]/io read_bytes",
		Type:      counterMetric,
	}
}

//prom:doc minio_node_io_wchar_bytes, Total bytes written by the process to the underlying storage system including page cache, /proc/[pid]/io wchar
func getMinioProcessIOWriteCachedBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: ioSubsystem,
		Name:      wcharBytes,
		Help:      "Total bytes written by the process to the underlying storage system including page cache, /proc/[pid]/io wchar",
		Type:      counterMetric,
	}
}

//prom:doc minio_node_io_rchar_bytes, Total bytes read by the process from the underlying storage system including cache, /proc/[pid]/io rchar
func getMinioProcessIOReadCachedBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: ioSubsystem,
		Name:      rcharBytes,
		Help:      "Total bytes read by the process from the underlying storage system including cache, /proc/[pid]/io rchar",
		Type:      counterMetric,
	}
}

//prom:doc minio_node_syscall_read_total, Total read SysCalls to the kernel. /proc/[pid]/io syscr
func getMinIOProcessSysCallRMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: sysCallSubsystem,
		Name:      readTotal,
		Help:      "Total read SysCalls to the kernel. /proc/[pid]/io syscr",
		Type:      counterMetric,
	}
}

//prom:doc minio_node_syscall_write_total, Total write SysCalls to the kernel. /proc/[pid]/io syscw
func getMinIOProcessSysCallWMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: sysCallSubsystem,
		Name:      writeTotal,
		Help:      "Total write SysCalls to the kernel. /proc/[pid]/io syscw",
		Type:      counterMetric,
	}
}

//prom:doc minio_node_go_routine_total, Total number of go routines running
func getMinIOGORoutineCountMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: goRoutines,
		Name:      total,
		Help:      "Total number of go routines running",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_node_process_starttime_seconds, Start time for MinIO process per node, time in seconds since Unix epoc
func getMinIOProcessStartTimeMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: processSubsystem,
		Name:      startTime,
		Help:      "Start time for MinIO process per node, time in seconds since Unix epoc",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_node_process_uptime_seconds, Uptime for MinIO process per node in seconds
func getMinIOProcessUptimeMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: processSubsystem,
		Name:      upTime,
		Help:      "Uptime for MinIO process per node in seconds",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_node_process_resident_memory_bytes, Resident memory size in bytes
func getMinIOProcessResidentMemory() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: processSubsystem,
		Name:      memory,
		Help:      "Resident memory size in bytes",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_node_process_cpu_total_seconds, Total user and system CPU time spent in seconds
func getMinIOProcessCPUTime() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: processSubsystem,
		Name:      cpu,
		Help:      "Total user and system CPU time spent in seconds",
		Type:      counterMetric,
	}
}

//prom:doc minio_node_ilm_transition_pending_tasks, Number of pending ILM transition tasks in the queue
func getTransitionPendingTasksMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: ilmSubsystem,
		Name:      transitionPendingTasks,
		Help:      "Number of pending ILM transition tasks in the queue",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_node_ilm_transition_active_tasks, Number of active ILM transition tasks
func getTransitionActiveTasksMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: ilmSubsystem,
		Name:      transitionActiveTasks,
		Help:      "Number of active ILM transition tasks",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_node_ilm_expiry_pending_tasks, Number of pending ILM expiry tasks in the queue
func getExpiryPendingTasksMD() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: ilmSubsystem,
		Name:      expiryPendingTasks,
		Help:      "Number of pending ILM expiry tasks in the queue",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_node_scanner_objects_scanned, Total number of unique objects scanned since server start
func getNodeScannerObjectsScanned() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: scannerSubsystem,
		Name:      "objects_scanned",
		Help:      "Total number of unique objects scanned since server start",
		Type:      counterMetric,
	}
}

//prom:doc minio_node_scanner_versions_scanned, Total number of object versions scanned since server start
func getNodeScannerVersionsScanned() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: scannerSubsystem,
		Name:      "versions_scanned",
		Help:      "Total number of object versions scanned since server start",
		Type:      counterMetric,
	}
}

//prom:doc minio_node_scanner_directories_scanned, Total number of directories scanned since server start
func getNodeScannerDirectoriesScanned() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: scannerSubsystem,
		Name:      "directories_scanned",
		Help:      "Total number of directories scanned since server start",
		Type:      counterMetric,
	}
}

//prom:doc minio_node_scanner_bucket_scans_started, Total number of bucket scans started since server start
func getNodeScannerBucketScansStarted() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: scannerSubsystem,
		Name:      "bucket_scans_started",
		Help:      "Total number of bucket scans started since server start",
		Type:      counterMetric,
	}
}

//prom:doc minio_node_scanner_bucket_scans_finished, Total number of bucket scans finished since server start
func getNodeScannerBucketScansFinished() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: scannerSubsystem,
		Name:      "bucket_scans_finished",
		Help:      "Total number of bucket scans finished since server start",
		Type:      counterMetric,
	}
}

//prom:doc minio_node_ilm_versions_scanned, Total number of object versions checked for ilm actions since server start
func getNodeILMVersionsScanned() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: ilmSubsystem,
		Name:      "versions_scanned",
		Help:      "Total number of object versions checked for ilm actions since server start",
		Type:      counterMetric,
	}
}

//prom:doc minio_node_iam_last_sync_duration_millis, Last successful IAM data sync duration in milliseconds
func getNodeIAMLastSyncDuration() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: iamSubsystem,
		Name:      "last_sync_duration_millis",
		Help:      "Last successful IAM data sync duration in milliseconds",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_node_iam_since_last_sync_millis, Time (in milliseconds) since last successful IAM data sync. This is set to 0 until the first sync after server start.
func getNodeIAMSinceLastSync() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: iamSubsystem,
		Name:      "since_last_sync_millis",
		Help:      "Time (in milliseconds) since last successful IAM data sync. This is set to 0 until the first sync after server start.",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_node_iam_sync_successes, Number of successful IAM data syncs since server start.
func getNodeIAMSyncSuccesses() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: iamSubsystem,
		Name:      "sync_successes",
		Help:      "Number of successful IAM data syncs since server start.",
		Type:      counterMetric,
	}
}

//prom:doc minio_node_iam_sync_failures, Number of failed IAM data syncs since server start.
func getNodeIAMSyncFailures() MetricDescription {
	return MetricDescription{
		Namespace: nodeMetricNamespace,
		Subsystem: iamSubsystem,
		Name:      "sync_failures",
		Help:      "Number of failed IAM data syncs since server start.",
		Type:      counterMetric,
	}
}

//prom:doc minio_notify_current_send_in_progress, Number of concurrent async Send calls active to all targets
func getMinIONotifyCurrentSendInProgress() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: notifySubsystem,
		Name:      "current_send_in_progress",
		Help:      "Number of concurrent async Send calls active to all targets",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_notify_target_queue_length, Number of unsent notifications in queue for target
func getMinIONotifyTargetQueueLength() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: notifySubsystem,
		Name:      "target_queue_length",
		Help:      "Number of unsent notifications in queue for target",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_audit_target_queue_length, Number of unsent messages in queue for target
func getMinIOAuditTargetQueueLength() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: auditSubsystem,
		Name:      "target_queue_length",
		Help:      "Number of unsent messages in queue for target",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_audit_total_messages, Total number of messages sent since start
func getMinIOAuditTotalMessages() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: auditSubsystem,
		Name:      "total_messages",
		Help:      "Total number of messages sent since start",
		Type:      counterMetric,
	}
}

//prom:doc minio_audit_failed_messages, Total number of messages that failed to send since start
func getMinIOAuditFailedMessages() MetricDescription {
	return MetricDescription{
		Namespace: minioNamespace,
		Subsystem: auditSubsystem,
		Name:      "failed_messages",
		Help:      "Total number of messages that failed to send since start",
		Type:      counterMetric,
	}
}

//prom:doc minio_cluster_ilm_transitioned_bytes, Total bytes transitioned to a tier
func getClusterTransitionedBytesMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: ilmSubsystem,
		Name:      transitionedBytes,
		Help:      "Total bytes transitioned to a tier",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_cluster_ilm_transitioned_objects, Total number of objects transitioned to a tier
func getClusterTransitionedObjectsMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: ilmSubsystem,
		Name:      transitionedObjects,
		Help:      "Total number of objects transitioned to a tier",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_cluster_ilm_transitioned_versions, Total number of versions transitioned to a tier
func getClusterTransitionedVersionsMD() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: ilmSubsystem,
		Name:      transitionedVersions,
		Help:      "Total number of versions transitioned to a tier",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_cluster_kms_online, Reports whether the KMS is online (1) or offline (0)
func getClusterKMSOnline() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: kmsSubsystem,
		Name:      kmsOnline,
		Help:      "Reports whether the KMS is online (1) or offline (0)",
		Type:      gaugeMetric,
	}
}

//prom:doc minio_cluster_kms_request_success, Number of KMS requests that succeeded
func getClusterKMSRequestsSuccess() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: kmsSubsystem,
		Name:      kmsRequestsSuccess,
		Help:      "Number of KMS requests that succeeded",
		Type:      counterMetric,
	}
}

//prom:doc minio_cluster_kms_request_error, Number of KMS requests that failed due to some error. (HTTP 4xx status code)
func getClusterKMSRequestsError() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: kmsSubsystem,
		Name:      kmsRequestsError,
		Help:      "Number of KMS requests that failed due to some error. (HTTP 4xx status code)",
		Type:      counterMetric,
	}
}

//prom:doc minio_cluster_kms_request_failure, Number of KMS requests that failed due to some internal failure. (HTTP 5xx status code)
func getClusterKMSRequestsFail() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: kmsSubsystem,
		Name:      kmsRequestsFail,
		Help:      "Number of KMS requests that failed due to some internal failure. (HTTP 5xx status code)",
		Type:      counterMetric,
	}
}

//prom:doc minio_cluster_kms_uptime, The time the KMS has been up and running in seconds.
func getClusterKMSUptime() MetricDescription {
	return MetricDescription{
		Namespace: clusterMetricNamespace,
		Subsystem: kmsSubsystem,
		Name:      kmsUptime,
		Help:      "The time the KMS has been up and running in seconds.",
		Type:      counterMetric,
	}
}
