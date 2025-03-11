# VMware VCenter prometheus metrics exporter

## About 

govc_exporter is a Prometheus exporter that extracts and exposes all available metrics from a VMware vCenter. It is specifically designed for large-scale environments and optimized to handle the performance limitations of the vCenter API.

Originally started as a fork, govc_exporter has evolved with numerous breaking changes and improvements. Due to these changes, the project now continues as an independent repository.


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
./govc_exporter --help
usage: exporter --collector.vc.url=COLLECTOR.VC.URL --collector.vc.username=COLLECTOR.VC.USERNAME --collector.vc.password=COLLECTOR.VC.PASSWORD [<flags>]

Flags:
  -h, --[no-]help                Show context-sensitive help (also try
                                 --help-long and --help-man).
      --web.listen-address=":9752"  
                                 Address on which to expose metrics and web
                                 interface.
      --web.telemetry-path="/metrics"  
                                 Path under which to expose metrics.
      --[no-]web.disable-exporter-metrics  
                                 Exclude metrics about the exporter itself
                                 (promhttp_*, process_*, go_*).
      --web.max-requests=40      Maximum number of parallel scrape requests.
                                 Use 0 to disable.
      --[no-]web.manual-refresh  Enable /refresh/{sensor} path to trigger a
                                 refresh of a sensor.
      --[no-]web.allow-dumps     Enable /dump path to trigger a dump of the
                                 cache data in ./dumps folder on server side.
                                 Only enable for debugging.
      --collector.vc.url=COLLECTOR.VC.URL  
                                 vc api username ($VC_URL)
      --collector.vc.username=COLLECTOR.VC.USERNAME  
                                 vc api username ($VC_USERNAME)
      --collector.vc.password=COLLECTOR.VC.PASSWORD  
                                 vc api password ($VC_PASSWORD)
      --[no-]scraper.esx         Enable host scraper
      --[no-]scraper.ds          Enable datastore scraper
      --[no-]scraper.repool      Enable resource pool scraper
      --[no-]scraper.spod        Enable datastore cluster scraper
      --[no-]scraper.vm          Enable virtualmachine scraper
      --[no-]scraper.cluster     Enable cluster scraper
      --[no-]scraper.compute_resource  
                                 Enable compute_resource scraper
      --[no-]scraper.tags        Collect tags
      --scraper.host.max_age=60  time in seconds hosts are cached
      --scraper.host.refresh_interval=25  
                                 interval hosts are refreshed
      --scraper.compute_resource.max_age=300  
                                 time in seconds clusters are cached
      --scraper.compute_resource.refresh_interval=25  
                                 interval clusters are refreshed
      --scraper.cluster.max_age=300  
                                 time in seconds clusters are cached
      --scraper.cluster.refresh_interval=25  
                                 interval clusters are refreshed
      --scraper.vm.max_age=120   time in seconds vm's are cached
      --scraper.vm.refresh_interval=55  
                                 interval vm's are refreshed
      --scraper.datastore.max_age=120  
                                 time in seconds datastores are cached
      --scraper.datastore.refresh_interval=55  
                                 interval datastores are refreshed
      --scraper.spod.max_age=120  
                                 time in seconds spods are cached
      --scraper.spod.refresh_interval=55  
                                 interval spods are refreshed
      --scraper.repool.max_age=120  
                                 time in seconds resource pools are cached
      --scraper.repool.refresh_interval=55  
                                 interval resource pools are refreshed
      --scraper.tags.max_age=600  
                                 time in seconds tags are cached
      --scraper.tags.refresh_interval=290  
                                 interval tags are refreshed
      --scraper.on_demand_cache.max_age=300  
                                 time in seconds the scraper keeps all
                                 non-cache data. Used to get parent objects
      --scraper.clean_interval=5  
                                 interval the scraper cleans up old data
      --scraper.client_pool_size=5  
                                 number of simultanious requests to vCenter
                                 api
      --[no-]collector.intrinsec  
                                 Enable intrinsec specific features
      --[no-]collector.vm.disk   Collect extra vm disk metrics
      --[no-]collector.vm.network  
                                 Collect extra vm network metrics
      --[no-]collector.host.storage  
                                 Collect host storage metrics
      --collector.cluster.tag_label=COLLECTOR.CLUSTER.TAG_LABEL ...  
                                 List of vmware tag categories to collect
                                 which will be added as label in metrics
      --collector.datastore.tag_label=COLLECTOR.DATASTORE.TAG_LABEL ...  
                                 List of vmware tag categories to collect
                                 which will be added as label in metrics
      --collector.host.tag_label=COLLECTOR.HOST.TAG_LABEL ...  
                                 List of vmware tag categories which will be
                                 added as label in metrics
      --collector.repool.tag_label=COLLECTOR.REPOOL.TAG_LABEL ...  
                                 List of tag categories which will be added as
                                 label in metrics
      --collector.spod.tag_label=COLLECTOR.SPOD.TAG_LABEL ...  
                                 List of vmware tag categories to collect
                                 which will be added as label in metrics
      --collector.vm.tag_label=COLLECTOR.VM.TAG_LABEL ...  
                                 List of vmware tag categories to collect
                                 which will be added as label in metrics
      --log.level=info           Only log messages with the given severity or
                                 above. One of: [debug, info, warn, error]
      --log.format=logfmt        Output format of log messages. One of:
                                 [logfmt, json]
      --[no-]version             Show application version.
```

# Get metrics

```
curl -s "localhost:9752/metrics"
curl -s "localhost:9752/metrics?collect=all"
curl -s "localhost:9752/metrics?collect=datastore&collect=vm&collect=spod&collect=cluster"
curl -s "localhost:9752/metrics?exclude=datastore"
curl -s "localhost:9752/metrics?collect=all&exclude=exporter_metrics"
```