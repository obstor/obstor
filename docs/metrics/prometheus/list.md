# List of metrics reported cluster wide

Each metric includes a label for the server that calculated the metric.
Each metric has a label for the server that generated the metric.

These metrics can be from any Obstor server once per collection.

| Name                                         | Description                                                                                                         |
|:---------------------------------------------|:--------------------------------------------------------------------------------------------------------------------|
| `obstor_bucket_objects_size_distribution`     | Distribution of object sizes in the bucket, includes label for the bucket name.                                     |
| `obstor_bucket_replication_failed_bytes`      | Total number of bytes failed at least once to replicate.                                                            |
| `obstor_bucket_replication_pending_bytes`     | Total bytes pending to replicate.                                                                                   |
| `obstor_bucket_replication_received_bytes`    | Total number of bytes replicated to this bucket from another source bucket.                                         |
| `obstor_bucket_replication_sent_bytes`        | Total number of bytes replicated to the target bucket.                                                              |
| `obstor_bucket_replication_pending_count`     | Total number of replication operations pending for this bucket.                                                     |
| `obstor_bucket_replication_failed_count`      | Total number of replication foperations failed for this bucket.                                                     |
| `obstor_bucket_usage_object_total`            | Total number of objects                                                                                             |
| `obstor_bucket_usage_total_bytes`             | Total bucket size in bytes                                                                                          |
| `obstor_cache_hits_total`                     | Total number of disk cache hits                                                                                     |
| `obstor_cache_missed_total`                   | Total number of disk cache misses                                                                                   |
| `obstor_cache_sent_bytes`                     | Total number of bytes served from cache                                                                             |
| `obstor_cache_total_bytes`                    | Total size of cache disk in bytes                                                                                   |
| `obstor_cache_usage_info`                     | Total percentage cache usage, value of 1 indicates high and 0 low, label level is set as well                       |
| `obstor_cache_used_bytes`                     | Current cache usage in bytes                                                                                        |
| `obstor_cluster_capacity_raw_free_bytes`      | Total free capacity online in the cluster.                                                                          |
| `obstor_cluster_capacity_raw_total_bytes`     | Total capacity online in the cluster.                                                                               |
| `obstor_cluster_capacity_usable_free_bytes`   | Total free usable capacity online in the cluster.                                                                   |
| `obstor_cluster_capacity_usable_total_bytes`  | Total usable capacity online in the cluster.                                                                        |
| `obstor_cluster_nodes_offline_total`          | Total number of Obstor nodes offline.                                                                                |
| `obstor_cluster_nodes_online_total`           | Total number of Obstor nodes online.                                                                                 |
| `obstor_heal_objects_error_total`             | Objects for which healing failed in current self healing run                                                        |
| `obstor_heal_objects_heal_total`              | Objects healed in current self healing run                                                                          |
| `obstor_heal_objects_total`                   | Objects scanned in current self healing run                                                                         |
| `obstor_heal_time_last_activity_nano_seconds` | Time elapsed (in nano seconds) since last self healing activity. This is set to -1 until initial self heal activity |
| `obstor_inter_node_traffic_received_bytes`    | Total number of bytes received from other peer nodes.                                                               |
| `obstor_inter_node_traffic_sent_bytes`        | Total number of bytes sent to the other peer nodes.                                                                 |
| `obstor_node_disk_free_bytes`                 | Total storage available on a disk.                                                                                  |
| `obstor_node_disk_total_bytes`                | Total storage on a disk.                                                                                            |
| `obstor_node_disk_used_bytes`                 | Total storage used on a disk.                                                                                       |
| `obstor_node_file_descriptor_limit_total`     | Limit on total number of open file descriptors for the Obstor Server process.                                        |
| `obstor_node_file_descriptor_open_total`      | Total number of open file descriptors by the Obstor Server process.                                                  |
| `obstor_node_io_rchar_bytes`                  | Total bytes read by the process from the underlying storage system including cache, /proc/[pid]/io rchar            |
| `obstor_node_io_read_bytes`                   | Total bytes read by the process from the underlying storage system, /proc/[pid]/io read_bytes                       |
| `obstor_node_io_wchar_bytes`                  | Total bytes written by the process to the underlying storage system including page cache, /proc/[pid]/io wchar      |
| `obstor_node_io_write_bytes`                  | Total bytes written by the process to the underlying storage system, /proc/[pid]/io write_bytes                     |
| `obstor_node_process_starttime_seconds`       | Start time for Obstor process per node, time in seconds since Unix epoc.                                             |
| `obstor_node_process_uptime_seconds`          | Uptime for Obstor process per node in seconds.                                                                       |
| `obstor_node_syscall_read_total`              | Total read SysCalls to the kernel. /proc/[pid]/io syscr                                                             |
| `obstor_node_syscall_write_total`             | Total write SysCalls to the kernel. /proc/[pid]/io syscw                                                            |
| `obstor_s3_requests_error_total`              | Total number S3 requests with errors                                                                                |
| `obstor_s3_requests_inflight_total`           | Total number of S3 requests currently in flight                                                                     |
| `obstor_s3_requests_total`                    | Total number S3 requests                                                                                            |
| `obstor_s3_time_ttbf_seconds_distribution`    | Distribution of the time to first byte across API calls.                                                            |
| `obstor_s3_traffic_received_bytes`            | Total number of s3 bytes received.                                                                                  |
| `obstor_s3_traffic_sent_bytes`                | Total number of s3 bytes sent                                                                                       |
| `obstor_software_commit_info`                 | Git commit hash for the Obstor release.                                                                              |
| `obstor_software_version_info`                | Obstor Release tag for the server                                                                                    |
