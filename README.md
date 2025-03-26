# VMware VCenter prometheus metrics exporter

## About 

govc_exporter is a Prometheus exporter that extracts and exposes all available metrics from a VMware vCenter. It is specifically designed to be able to function in a large-scale environments and handle the performance limitations of the vCenter API.

Originally started as a fork, govc_exporter has evolved with numerous breaking changes and improvements. Due to these changes, the project now continues as an independent repository.

## How it works internally

### Scraper

The exporter had an internal scraper which pulls all the data from vCenter. The scraper has multiple sensors. Each sensor pulls the data for a certain object (Hosts, Clusters, VMs,...). The sensor is periodically refreshed always keeping the latest version of the data. A sensor also has a periodic cleanup job which removes old data. 
Every sensor can be configured by the cli. Check the `--help` and look for `--scraper.[sensor].[option]` for more information. 

### Performance Metrics

The exporter also allows to query performance metrics for certain objects (Host, VM,...). Performance metrics are pulled from the vCenter performance endpoint and are usually more acurate. 

The perfmetrics are collected by a perf sensor, which is simular to the normal sensor except that it stores all metrics in a timed queue. When the collector is triggered, all the latest metrics will be popped from the queue and returned by the exporter.

Every performance sensor can be configured by the cli. Check the `--help` and look for `--scraper.[sensor].perf.[option]` for more information. 

## Building and running

### Build

```shell

# Build with GoReleaser
make build

# Build without GoReleaser
make build-simple
```

### Running

```shell
export VC_URL=FIXME
export VC_USERNAME=FIXME
export VC_PASSWORD=FIXME
./govc_exporter <flags>
```

### Usage

```
govc_exporter --help
usage: exporter --scraper.vc.url=SCRAPER.VC.URL --scraper.vc.username=SCRAPER.VC.USERNAME --scraper.vc.password=SCRAPER.VC.PASSWORD [<flags>]

Prometheus vCenter exporter


Flags:
  -h, --[no-]help                Show context-sensitive help (also try --help-long and --help-man).
      --log.level=info           Only log messages with the given severity or above. One of: [debug,
                                 info, warn, error]
      --log.format=logfmt        Output format of log messages. One of: [logfmt, json]
      --[no-]version             Show application version.
      --web.listen-address=":9752"  
                                 Address on which to expose metrics and web interface.
      --web.telemetry-path="/metrics"  
                                 Path under which to expose metrics.
      --web.max-requests=40      Maximum number of parallel scrape requests. Use 0 to disable.
      --[no-]web.manual-refresh  Enable /refresh/{sensor} path to trigger a refresh of a sensor.
      --[no-]web.allow-dumps     Enable /dump path to trigger a dump of the cache data in ./dumps
                                 folder on server side. Only enable for debugging.
      --[no-]web.disable-exporter-metrics  
                                 Exclude metrics about the exporter itself (promhttp_*, process_*,
                                 go_*).
      --[no-]collector.intrinsec  
                                 Enable intrinsec specific features
      --collector.cluster.tag_label=COLLECTOR.CLUSTER.TAG_LABEL ...  
                                 List of vmware tag categories to collect which will be added as label
                                 in metrics
      --collector.datastore.tag_label=COLLECTOR.DATASTORE.TAG_LABEL ...  
                                 List of vmware tag categories to collect which will be added as label
                                 in metrics
      --[no-]collector.host.storage  
                                 Collect host storage metrics
      --collector.host.tag_label=COLLECTOR.HOST.TAG_LABEL ...  
                                 List of vmware tag categories which will be added as label in metrics
      --collector.repool.tag_label=COLLECTOR.REPOOL.TAG_LABEL ...  
                                 List of tag categories which will be added as label in metrics
      --collector.spod.tag_label=COLLECTOR.SPOD.TAG_LABEL ...  
                                 List of vmware tag categories to collect which will be added as label
                                 in metrics
      --collector.vm.tag_label=COLLECTOR.VM.TAG_LABEL ...  
                                 List of vmware tag categories to collect which will be added as label
                                 in metrics
      --scraper.vc.url=SCRAPER.VC.URL  
                                 vc api username ($VC_URL)
      --scraper.vc.username=SCRAPER.VC.USERNAME  
                                 vc api username ($VC_USERNAME)
      --scraper.vc.password=SCRAPER.VC.PASSWORD  
                                 vc api password ($VC_PASSWORD)
      --scraper.client_pool_size=5  
                                 number of simultanious requests to vCenter api
      --[no-]scraper.cluster     Enable cluster sensor
      --scraper.cluster.max_age=5m  
                                 time in seconds clusters are cached
      --scraper.cluster.refresh_interval=25s  
                                 interval clusters are refreshed
      --scraper.cluster.clean_interval=5s  
                                 interval to clean up old metrics
      --[no-]scraper.compute_resource  
                                 Enable compute_resource sensor
      --scraper.compute_resource.max_age=5m  
                                 time in seconds clusters are cached
      --scraper.compute_resource.refresh_interval=25s  
                                 interval clusters are refreshed
      --scraper.compute_resource.clean_interval=5s  
                                 interval to clean up old metrics
      --[no-]scraper.datastore   Enable datastore sensor
      --scraper.datastore.max_age=2m  
                                 time in seconds datastores are cached
      --scraper.datastore.refresh_interval=55s  
                                 interval datastores are refreshed
      --scraper.datastore.clean_interval=5s  
                                 interval to clean up old metrics
      --[no-]scraper.host        Enable host sensor
      --scraper.host.max_age=1m  time in seconds hosts are cached
      --scraper.host.refresh_interval=25s  
                                 interval hosts are refreshed
      --scraper.host.clean_interval=5s  
                                 interval to clean up old metrics
      --[no-]scraper.host.perf   Enable host performance metrics
      --scraper.host.perf.max_age=10m  
                                 time in seconds performance metrics are cached
      --scraper.host.perf.refresh_interval=55s  
                                 perf metrics refresh interval
      --scraper.host.perf.clean_interval=5s  
                                 interval to clean up old metrics
      --scraper.host.perf.max_sample_window=5m  
                                 max window metrics are collected
      --scraper.host.perf.sample_interval=20s  
                                 time between metrics
      --[no-]scraper.host.perf.default_metrics  
                                 Collect default host perf metrics
      --scraper.host.perf.extra_metric=SCRAPER.HOST.PERF.EXTRA_METRIC ...  
                                 Collect additional host perf metrics
      --[no-]scraper.repool      Enable resource pool sensor
      --scraper.repool.max_age=2m  
                                 time in seconds resource pools are cached
      --scraper.repool.refresh_interval=55s  
                                 interval resource pools are refreshed
      --scraper.repool.clean_interval=5s  
                                 interval to clean up old metrics
      --[no-]scraper.spod        Enable datastore cluster sensor
      --scraper.spod.max_age=2m  time in seconds spods are cached
      --scraper.spod.refresh_interval=55s  
                                 interval spods are refreshed
      --scraper.spod.clean_interval=5s  
                                 interval to clean up old metrics
      --[no-]scraper.tags        Collect tags
      --scraper.tags.max_age=10m  
                                 time in seconds tags are cached
      --scraper.tags.refresh_interval=290s  
                                 interval tags are refreshed
      --scraper.tags.clean_interval=5s  
                                 interval to clean up old metrics
      --[no-]scraper.vm          Enable virtualmachine sensor
      --scraper.vm.max_age=2m    time in seconds vm's are cached
      --scraper.vm.refresh_interval=55s  
                                 interval vm's are refreshed
      --scraper.vm.clean_interval=5s  
                                 interval to clean up old metrics
      --[no-]collector.vm.disk   Collect extra vm disk metrics
      --[no-]collector.vm.network  
                                 Collect extra vm network metrics
      --[no-]scraper.vm.perf     Enable vm performance metrics
      --scraper.vm.perf.max_age=10m  
                                 time in seconds perf metrics are cached
      --scraper.vm.perf.refresh_interval=55s  
                                 perf metrics refresh interval
      --scraper.vm.perf.clean_interval=5s  
                                 interval to clean up old metrics
      --scraper.vm.perf.max_sample_window=5m  
                                 max window metrics are collected
      --scraper.vm.perf.sample_interval=20s  
                                 time between metrics
      --[no-]scraper.vm.perf.default_metrics  
                                 Collect default vm perf metrics
      --scraper.vm.perf.extra_metric=SCRAPER.VM.PERF.EXTRA_METRIC ...  
                                 Collect additional vm perf metrics
      --scraper.on_demand.max_age=5m  
                                 Time in seconds the scraper keeps all non-cache data. Used when no
                                 other sensor is available
      --scraper.on_demand.clean_interval=5s  
                                 interval to clean up old metrics
```

# Get metrics

```
curl -s "localhost:9752/metrics"
curl -s "localhost:9752/metrics?collect=all"
curl -s "localhost:9752/metrics?collect=datastore&collect=vm&collect=spod&collect=cluster"
curl -s "localhost:9752/metrics?exclude=datastore"
curl -s "localhost:9752/metrics?collect=all&exclude=exporter_metrics"
```

# Debug

## Manual refresh

When enabled with `--web.manual-refresh`, the sensors can be refreshed manually. This feature is only intended for debugging purposes. 

    curl -s "localhost:9752/refresh/<sensor_name>"
    curl -s "localhost:9752/refresh/vm"

## Sensor data dump

When enabled with `--web.allow-dumps`, the exporter can take a datadump of all the sensor data. All data will be stored in `./dumps` on the exporter server. This feature is only intended for debugging purposes. 

    curl -s "localhost:9752/dump/<sensor_name>"
    curl -s "localhost:9752/dump/vm"
    curl -s "localhost:9752/dump?collect=vm&collect=perfvm"