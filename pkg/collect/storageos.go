package main

import (
	"context"
	"encoding/json"
	"fmt"

	apiclient "github.com/replicatedhq/troubleshoot/pkg/storageos"
	api "github.com/storageos/go-api/v2"
)

const (
	DefaultStorageOSNamespace = "storageos"
)

func main() {
	// ctx := context.TODO()
	username := "storageos"
	password := "storageos"
	apiEndpoint := "127.0.0.1:5705"

	ctx, storageos, err := apiclient.New(username, password, apiEndpoint)
	if err != nil {
		fmt.Print(err)
	}

	namespaces, err := getNamespaces(storageos, ctx)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(namespaces))

	cluster, err := getCluster(storageos, ctx)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(cluster))

	volumes, err := getVolumes(storageos, ctx)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(volumes))

}

func getNamespaces(storageos *api.APIClient, ctx context.Context) ([]byte, error) {
	namespaces, resp, err := storageos.DefaultApi.ListNamespaces(ctx)
	if err != nil {
		return nil, api.MapAPIError(err, resp)
	}

	b, err := json.Marshal(namespaces)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func getCluster(storageos *api.APIClient, ctx context.Context) ([]byte, error) {
	cluster, resp, err := storageos.DefaultApi.GetCluster(ctx)
	if err != nil {
		return nil, api.MapAPIError(err, resp)
	}

	b, err := json.Marshal(cluster)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func getVolumes(storageos *api.APIClient, ctx context.Context) ([]byte, error) {
	var volumes []api.Volume
	namespaces, resp, err := storageos.DefaultApi.ListNamespaces(ctx)
	if err != nil {
		return nil, api.MapAPIError(err, resp)
	}
	for _, ns := range namespaces {
		fmt.Print(ns.GetID())
		nsVolumes, resp, err := storageos.DefaultApi.ListVolumes(ctx, ns.GetID())
		if err != nil {
			return nil, api.MapAPIError(err, resp)
		}
		volumes = append(volumes, nsVolumes...)
	}

	b, err := json.Marshal(volumes)
	if err != nil {
		return nil, err
	}

	return b, nil
}
