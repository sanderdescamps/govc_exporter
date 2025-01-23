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
make build
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
usage: govc_exporter --collector.vc.url=COLLECTOR.VC.URL --collector.vc.username=COLLECTOR.VC.USERNAME --collector.vc.password=COLLECTOR.VC.PASSWORD [<flags>]


Flags:
  -h, --[no-]help                Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=":9752"  
                                 Address on which to expose metrics and web interface.
      --web.telemetry-path="/metrics"  
                                 Path under which to expose metrics.
      --[no-]web.disable-exporter-metrics  
                                 Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).
      --web.max-requests=40      Maximum number of parallel scrape requests. Use 0 to disable.
      --collector.vc.url=COLLECTOR.VC.URL  
                                 vc api username ($VC_URL)
      --collector.vc.username=COLLECTOR.VC.USERNAME  
                                 vc api username ($VC_USERNAME)
      --collector.vc.password=COLLECTOR.VC.PASSWORD  
                                 vc api password ($VC_PASSWORD)
      --scraper.ds="True"        Enable datastore metrics
      --scraper.repool="True"    Enable resource pool metrics
      --scraper.spod="True"      Enable datastore cluster metrics
      --scraper.vm="True"        Enable virtualmachine metrics
      --scraper.cluster="True"   Enable cluster metrics
      --scraper.tags="True"      Collect tags
      --scraper.host.max_age=60  time in seconds hosts are cached
      --scraper.host.refresh_interval=25  
                                 interval hosts are refreshed
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
                                 time in seconds the scraper keeps all non-cache data. Used to get parent objects
      --scraper.clean_interval=5  
                                 interval the scraper cleans up old data
      --scraper.client_pool_size=5  
                                 number of simultanious requests to vCenter api
      --[no-]collector.intrinsec  
                                 Enable intrinsec specific features
      --[no-]collector.vm.disk   Collect extra vm disk metrics
      --[no-]collector.vm.network  
                                 Collect extra vm network metrics
      --[no-]collector.host.storage  
                                 Collect host storage metrics
      --collector.cluster.tag_label=COLLECTOR.CLUSTER.TAG_LABEL ...  
                                 List of vmware tag categories to collect which will be added as label in metrics
      --collector.datastore.tag_label=COLLECTOR.DATASTORE.TAG_LABEL ...  
                                 List of vmware tag categories to collect which will be added as label in metrics
      --collector.host.tag_label=COLLECTOR.HOST.TAG_LABEL ...  
                                 List of vmware tag categories which will be added as label in metrics
      --collector.repool.tag_label=COLLECTOR.REPOOL.TAG_LABEL ...  
                                 List of tag categories which will be added as label in metrics
      --collector.spod.tag_label=COLLECTOR.SPOD.TAG_LABEL ...  
                                 List of vmware tag categories to collect which will be added as label in metrics
      --collector.vm.tag_label=COLLECTOR.VM.TAG_LABEL ...  
                                 List of vmware tag categories to collect which will be added as label in metrics
      --log.level=info           Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt        Output format of log messages. One of: [logfmt, json]
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