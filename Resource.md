### Resource Usage

```
curl --unix-socket /var/run/docker.sock -X  GET http://localhost/v1.40/containers/<CONTAINER_ID>/stats
```

```json
{
    "read": "2020-12-05T02:29:05.755726235Z",
    "preread": "0001-01-01T00:00:00Z",
    "pids_stats": {
        "current": 5
    },
    "blkio_stats": {
        "io_service_bytes_recursive": [],
        "io_serviced_recursive": [],
        "io_queue_recursive": [],
        "io_service_time_recursive": [],
        "io_wait_time_recursive": [],
        "io_merged_recursive": [],
        "io_time_recursive": [],
        "sectors_recursive": []
    },
    "num_procs": 0,
    "storage_stats": {},
    "cpu_stats": {
        "cpu_usage": {
            "total_usage": 61005527,
            "percpu_usage": [
                43314272,
                2099402,
                3378757,
                4218370,
                7045851,
                948875
            ],
            "usage_in_kernelmode": 20000000,
            "usage_in_usermode": 30000000
        },
        "system_cpu_usage": 26021320000000,
        "online_cpus": 6,
        "throttling_data": {
            "periods": 0,
            "throttled_periods": 0,
            "throttled_time": 0
        }
    },
    "precpu_stats": {
        "cpu_usage": {
            "total_usage": 0,
            "usage_in_kernelmode": 0,
            "usage_in_usermode": 0
        },
        "throttling_data": {
            "periods": 0,
            "throttled_periods": 0,
            "throttled_time": 0
        }
    },
    "memory_stats": {
        "usage": 2748416,
        "max_usage": 3051520,
        "stats": {
            "active_anon": 2195456,
            "active_file": 0,
            "cache": 0,
            "dirty": 0,
            "hierarchical_memory_limit": 9223372036854771712,
            "hierarchical_memsw_limit": 9223372036854771712,
            "inactive_anon": 0,
            "inactive_file": 0,
            "mapped_file": 0,
            "pgfault": 825,
            "pgmajfault": 0,
            "pgpgin": 561,
            "pgpgout": 530,
            "rss": 2195456,
            "rss_huge": 2097152,
            "total_active_anon": 2195456,
            "total_active_file": 0,
            "total_cache": 0,
            "total_dirty": 0,
            "total_inactive_anon": 0,
            "total_inactive_file": 0,
            "total_mapped_file": 0,
            "total_pgfault": 825,
            "total_pgmajfault": 0,
            "total_pgpgin": 561,
            "total_pgpgout": 530,
            "total_rss": 2195456,
            "total_rss_huge": 2097152,
            "total_unevictable": 0,
            "total_writeback": 0,
            "unevictable": 0,
            "writeback": 0
        },
        "limit": 5177552896
    },
    "name": "/service-A",
    "id": "74c3f873d0743ff278da30448f55537d4aa26a98af1564e8c0f63334480460dd",
    "networks": {
        "eth0": {
            "rx_bytes": 2230,
            "rx_packets": 30,
            "rx_errors": 0,
            "rx_dropped": 0,
            "tx_bytes": 3008,
            "tx_packets": 12,
            "tx_errors": 0,
            "tx_dropped": 0
        }
    }
}
```