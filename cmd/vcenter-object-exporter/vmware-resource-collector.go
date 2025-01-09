package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/url"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"

	kingpin "github.com/alecthomas/kingpin/v2"
)

var logger *slog.Logger

func init() {
	logger = slog.Default()
}

func GetClient(endpoint string, username string, password string) *govmomi.Client {
	u, err := soap.ParseURL(endpoint)
	if err != nil {
		panic(err)
	}
	u.User = url.UserPassword(username, password)
	client, err := govmomi.NewClient(context.Background(), u, true)
	if err != nil {
		logger.Error("Client creation failed", "err", err)
	}
	return client
}

func getResourcePools(client govmomi.Client) []mo.ResourcePool {
	var items []mo.ResourcePool

	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"ResourcePool"},
		true,
	)
	if err != nil {
		logger.Error("Failed to create container view", "err", err)
		return nil
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(
		ctx,
		[]string{"ResourcePool"},
		[]string{
			// "parent",
			// "summary",
			// "name",
		},
		&items,
	)
	if err != nil {
		logger.Error("Failed to retrieve data", "err", err)
		return nil
	}
	return items
}

func getVMs(client govmomi.Client) []mo.VirtualMachine {
	var items []mo.VirtualMachine

	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"VirtualMachine"},
		true,
	)
	if err != nil {
		logger.Error("Failed to create container view", "err", err)
		return nil
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(
		ctx,
		[]string{"VirtualMachine"},
		[]string{
			// "parent",
			// "summary",
			// "name",
		},
		&items,
	)
	if err != nil {
		logger.Error("Failed to retrieve data", "err", err)
		return nil
	}
	return items
}

func getHostSystem(client govmomi.Client) []mo.HostSystem {

	var items []mo.HostSystem

	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"HostSystem"},
		true,
	)
	if err != nil {
		logger.Error("Failed to create container view", "err", err)
		return nil
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(
		ctx,
		[]string{"HostSystem"},
		[]string{
			// "parent",
			// "summary",
			// "name",
		},
		&items,
	)
	if err != nil {
		logger.Error("Failed to retrieve data", "err", err)
		return nil
	}

	return items

}

func getClusters(client govmomi.Client) []mo.ClusterComputeResource {
	var items []mo.ClusterComputeResource

	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"ClusterComputeResource"},
		true,
	)
	if err != nil {
		logger.Error("Failed to create container view", "err", err)
		return nil
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(
		ctx,
		[]string{"ClusterComputeResource"},
		[]string{
			// "parent",
			// "summary",
			// "name",
		},
		&items,
	)
	if err != nil {
		logger.Error("Failed to retrieve data", "err", err)
		return nil
	}
	return items
}

func getComputeResource(client govmomi.Client) []mo.ComputeResource {
	var items []mo.ComputeResource

	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"ComputeResource"},
		true,
	)
	if err != nil {
		logger.Error("Failed to create container view", "err", err)
		return nil
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(
		ctx,
		[]string{"ComputeResource"},
		[]string{
			// "parent",
			// "summary",
			// "name",
		},
		&items,
	)
	if err != nil {
		logger.Error("Failed to retrieve data", "err", err)
		return nil
	}
	return items
}

func getDatacenter(client govmomi.Client) []mo.Datacenter {

	var items []mo.Datacenter

	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"Datacenter"},
		true,
	)
	if err != nil {
		logger.Error("Failed to create container view", "err", err)
		return nil
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(
		ctx,
		[]string{"Datacenter"},
		[]string{
			// "parent",
			// "summary",
			// "name",
		},
		&items,
	)
	if err != nil {
		logger.Error("Failed to retrieve data", "err", err)
		return nil
	}

	return items

}

func getSpod(client govmomi.Client) []mo.StoragePod {

	var items []mo.StoragePod

	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"StoragePod"},
		true,
	)
	if err != nil {
		logger.Error("Failed to create container view", "err", err)
		return nil
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(
		ctx,
		[]string{"StoragePod"},
		[]string{
			// "parent",
			// "summary",
			// "name",
		},
		&items,
	)
	if err != nil {
		logger.Error("Failed to retrieve data", "err", err)
		return nil
	}

	return items

}

func getDatastore(client govmomi.Client) []mo.Datastore {

	var items []mo.Datastore

	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"Datastore"},
		true,
	)
	if err != nil {
		logger.Error("Failed to create container view", "err", err)
		return nil
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(
		ctx,
		[]string{"Datastore"},
		[]string{
			// "parent",
			// "summary",
			// "name",
		},
		&items,
	)
	if err != nil {
		logger.Error("Failed to retrieve data", "err", err)
		return nil
	}

	return items

}

func getFolder(client govmomi.Client) []mo.Folder {

	var items []mo.Folder

	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"Folder"},
		true,
	)
	if err != nil {
		logger.Error("Failed to create container view", "err", err)
		return nil
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(
		ctx,
		[]string{"Folder"},
		[]string{
			// "parent",
			// "summary",
			// "name",
		},
		&items,
	)
	if err != nil {
		logger.Error("Failed to retrieve data", "err", err)
		return nil
	}

	return items

}

func writeAsJSON(path string, content interface{}) error {
	b, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

func main() {
	endpoint := kingpin.Flag("url", "vc api username").Short('e').Envar("VC_URL").Required().String()
	username := kingpin.Flag("username", "vc api username").Short('u').Envar("VC_USERNAME").Required().String()
	password := kingpin.Flag("password", "vc api password").Short('p').Envar("VC_PASSWORD").Required().String()
	collect := kingpin.Flag("collect", "type of data to collect").Short('c').Default("*").String()
	directory := kingpin.Flag("directory", "directory where to save output files").Short('d').Required().ExistingDir()

	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	allowedCollect := []string{
		"*",
		"resourcepool",
		"clustercomputeresource",
		"cluster",
		"host",
		"vm",
		"virtualmachine",
		"datacenter",
		"computeresource",
		"spod",
		"storagepod",
		"folder",
		"datastore",
		"ds",
	}

	collectString := strings.ToLower(*collect)
	if !slices.Contains(allowedCollect, collectString) {
		logger.Error("Invalid collect option", "collect:", collectString)
		return
	}

	client := GetClient(*endpoint, *username, *password)
	if collectString == "*" || strings.EqualFold(collectString, "ResourcePool") {
		logger.Info("Collecting resource pools...")
		items := getResourcePools(*client)
		outputPath := path.Join(*directory, "ResourcePool.json")
		err := writeAsJSON(outputPath, items)
		if err != nil {
			logger.Error("Failed to write response to file", "err", err)
			return
		}
	}

	if collectString == "*" || strings.EqualFold(collectString, "Host") || strings.EqualFold(collectString, "HostSystem") {
		logger.Info("Collecting hosts...")
		items := getHostSystem(*client)

		outputPath := path.Join(*directory, "HostSystem.json")
		err := writeAsJSON(outputPath, items)
		if err != nil {
			logger.Error("Failed to write response to file", "err", err)
			return
		}

		if len(items) > 0 {
			outputPath = path.Join(*directory, "HostSystem-single.json")
			err = writeAsJSON(outputPath, items[0])
			if err != nil {
				logger.Error("Failed to write response to file", "err", err)
				return
			}
		}

	}

	if collectString == "*" || strings.EqualFold(collectString, "Cluster") || strings.EqualFold(collectString, "ClusterComputereSource") {
		logger.Info("Collecting clusters...")
		items := getClusters(*client)
		outputPath := path.Join(*directory, "ClusterComputeResource.json")
		err := writeAsJSON(outputPath, items)
		if err != nil {
			logger.Error("Failed to write response to file", "err", err)
			return
		}
	}

	if collectString == "*" || strings.EqualFold(collectString, "VirtualMachine") {
		logger.Info("Collecting virtual machines...")
		items := getVMs(*client)
		outputPath := path.Join(*directory, "VirtualMachine.json")
		err := writeAsJSON(outputPath, items)
		if err != nil {
			logger.Error("Failed to write response to file", "err", err)
			return
		}

		if len(items) > 0 {
			outputPath = path.Join(*directory, "VirtualMachine-single.json")
			err = writeAsJSON(outputPath, items[0])
			if err != nil {
				logger.Error("Failed to write response to file", "err", err)
				return
			}
		}
	}

	if collectString == "*" || strings.EqualFold(collectString, "Datacenter") {
		logger.Info("Collecting datacenters...")
		items := getDatacenter(*client)
		outputPath := path.Join(*directory, "Datacenter.json")
		err := writeAsJSON(outputPath, items)
		if err != nil {
			logger.Error("Failed to write response to file", "err", err)
			return
		}
	}
	if collectString == "*" || strings.EqualFold(collectString, "ComputeResource") {
		logger.Info("Collecting compute resources...")
		items := getComputeResource(*client)
		outputPath := path.Join(*directory, "ComputeResource.json")
		err := writeAsJSON(outputPath, items)
		if err != nil {
			logger.Error("Failed to write response to file", "err", err)
			return
		}
	}
	if collectString == "*" || strings.EqualFold(collectString, "Spod") || strings.EqualFold(collectString, "StoragePod") {
		logger.Info("Collecting storage pods...")
		items := getSpod(*client)
		outputPath := path.Join(*directory, "StoragePod.json")
		err := writeAsJSON(outputPath, items)
		if err != nil {
			logger.Error("Failed to write response to file", "err", err)
			return
		}
	}
	if collectString == "*" || strings.EqualFold(collectString, "Folder") {
		logger.Info("Collecting folders...")
		items := getFolder(*client)
		outputPath := path.Join(*directory, "Folder.json")
		err := writeAsJSON(outputPath, items)
		if err != nil {
			logger.Error("Failed to write response to file", "err", err)
			return
		}
	}
	if collectString == "*" || strings.EqualFold(collectString, "Datastore") {
		logger.Info("Collecting datastore...")
		items := getDatastore(*client)
		outputPath := path.Join(*directory, "Datastore.json")
		err := writeAsJSON(outputPath, items)
		if err != nil {
			logger.Error("Failed to write response to file", "err", err)
			return
		}
	}
}
