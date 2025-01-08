# vmware VCenter prometheus metrics exporter

[![Go Report Card](https://goreportcard.com/badge/github.com/intrinsec/govc_exporter)](https://goreportcard.com/report/github.com/intrinsec/govc_exporter)

Prometheus stand alone exporter for VCenter metrics. Metrics are fetched by govmomi api.

Works also with stand alone ESX without VCenter.

| Collectors          | Description |
| ------------------- | ----------- |
| `collector.ds`      | Datastore metrics collector |
| `collector.esx`     | ESX (HostSystem) metrics collector |
| `collector.respool` | ResourcePool metrics collector |
| `collector.spod`    | Datastore Cluster (StoragePod) metrics collector |
| `collector.vm`      | VirtualMachine metrics Collector |

## Building and running

### Build

```shell
make
```

### Running

```shell
export VC_URL=FIXME
export VC_USERNAME=FIXME
export VC_PASSWORD=FIXME
./govc_exporter <flags>
```

### Usage

```shell
./govc_exporter --help
usage: govc_exporter --collector.vc.password=COLLECTOR.VC.PASSWORD --collector.vc.username=COLLECTOR.VC.USERNAME --collector.vc.url=COLLECTOR.VC.URL [<flags>]

Flags:
  -h, --[no-]help               Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=":9752"  
                                Address on which to expose metrics and web interface.
      --web.telemetry-path="/metrics"  
                                Path under which to expose metrics.
      --[no-]web.disable-exporter-metrics  
                                Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).
      --web.max-requests=40     Maximum number of parallel scrape requests. Use 0 to disable.
      --collector.vc.url=COLLECTOR.VC.URL  
                                vc api username ($VC_URL)
      --collector.vc.username=COLLECTOR.VC.USERNAME  
                                vc api username ($VC_USERNAME)
      --collector.vc.password=COLLECTOR.VC.PASSWORD  
                                vc api password ($VC_PASSWORD)
      --scraper.ds="True"       Enable datastore metrics
      --scraper.repool="True"   Enable datastore metrics
      --scraper.spod="True"     Enable datastore metrics
      --scraper.vm="True"       Enable datastore metrics
      --scraper.cluster="True"  Enable datastore metrics
      --scraper.host_max_age=120  
                                time in seconds host metrics are cached
      --scraper.host_refresh_interval=60  
                                interval host metrics are refreshed
      --scraper.cluster_max_age=300  
                                time in seconds cluster metrics are cached
      --scraper.cluster_refresh_interval=30  
                                interval cluster metrics are refreshed
      --scraper.virtual_machine_max_age=120  
                                time in seconds vm metrics are cached
      --scraper.virtual_machine_refresh_interval=60  
                                interval vm metrics are refreshed
      --scraper.datastore_max_age=120  
                                time in seconds datastore metrics are cached
      --scraper.datastore_refresh_interval=60  
                                interval datastore metrics are refreshed
      --scraper.spod_max_age=120  
                                time in seconds spod metrics are cached
      --scraper.storagepod_refresh_interval=60  
                                interval spod metrics are refreshed
      --scraper.repool_max_age=120  
                                time in seconds resource pool metrics are cached
      --scraper.repool_refresh_interval=60  
                                interval resource pool metrics are refreshed
      --scraper.on_demand_cache_max_age=300  
                                time in seconds all other metrics are in cache. Used to get parent objects
      --scraper.clean_cache_interval=5  
                                interval the cache cleanup runs
      --scraper.client_pool_size=5  
                                number of simultanious requests to vCenter api
      --[no-]collector.intrinsec  
                                Enable intrinsec specific features
      --[no-]collector.vm.disk  Collect vm disk metrics
      --[no-]collector.vm.network  
                                Collect vm network metrics
      --log.level=info          Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt       Output format of log messages. One of: [logfmt, json]
      --[no-]version            Show application version.
```
