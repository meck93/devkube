package hetzner

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/hetznercloud/hcloud-go/hcloud"

	"github.com/adrianliechti/devkube/provider"
)

type Provider struct {
	client *hcloud.Client
}

func New(token string) provider.Provider {
	c := hcloud.NewClient(hcloud.WithToken(token))

	return &Provider{client: c}
}

func NewFromEnvironment() (provider.Provider, error) {
	token := os.Getenv("HETZNER_TOKEN")

	if token == "" {
		return nil, fmt.Errorf("HETZNER_TOKEN is not set")
	}

	return New(token), nil
}

func (p *Provider) List(ctx context.Context) ([]string, error) {
	var list []string

	servers, response, err := p.client.Server.List(ctx, hcloud.ServerListOpts{})
	println("response", response.Status)

	if err != nil {
		return list, err
	}

	for _, c := range servers {
		list = append(list, c.Name)
		println("server", c.ID, c.Created.String(), c.Name, c.Labels)
	}

	networks, _, err := p.client.Network.List(ctx, hcloud.NetworkListOpts{})

	if err != nil {
		return list, err
	}

	for _, c := range networks {
		list = append(list, c.Name)
		println("network", c.ID, c.Created.String(), c.Name, c.Labels)
	}

	return list, nil
}

func (p *Provider) Create(ctx context.Context, name string, kubeconfig string) error {
	// create the default network 10.0.0.0/16
	network, response, err := p.client.Network.Create(ctx, hcloud.NetworkCreateOpts{
		Name:    "default",
		IPRange: &net.IPNet{IP: net.IPv4(10, 0, 0, 0), Mask: net.IPv4Mask(255, 255, 0, 0)},
		Subnets: []hcloud.NetworkSubnet{
			{
				Type:        hcloud.NetworkSubnetTypeCloud,
				IPRange:     &net.IPNet{IP: net.IPv4(10, 0, 0, 0), Mask: net.IPv4Mask(255, 255, 255, 0)},
				NetworkZone: hcloud.NetworkZoneEUCentral,
			},
		},
	},
	)

	if err != nil {
		return err
	}

	println("response", response.Status)
	println("network", network.ID, network.Name, network.Created.String())

	// TODO: we need to create an SSH key for the server access

	// create the servers
	serverCreateResult, response, err := p.client.Server.Create(ctx, hcloud.ServerCreateOpts{
		Name: "server-01",
		// ServerType: &hcloud.ServerType{

		// },
		// Image: &hcloud.Image{

		// },
		// Datacenter: &hcloud.Datacenter{

		// },
		StartAfterCreate: hcloud.Bool(false),
	})
	println("serverCreateResult", serverCreateResult.Server.ID, serverCreateResult.Server.Name)

	// TODO: attach servers to network

	return writeKubeconfig(kubeconfig)
}

func (p *Provider) Delete(ctx context.Context, name string) error {
	return errors.New("Delete is not yet implemented.")
}

func (p *Provider) Export(ctx context.Context, name, kubeconfig string) error {
	return errors.New("Export is not yet implemented.")
}

func (p *Provider) clusterID(ctx context.Context, name string) (int, error) {
	return 0, errors.New("clusterId is not yet implemented.")
}

func writeKubeconfig(kubeconfig string) error {
	if kubeconfig == "" {
		home, err := os.UserHomeDir()

		if err != nil {
			return err
		}

		dir := filepath.Join(home, ".kube")

		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}

		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	// TODO: figure out what to pass as default kubeconfig data
	return os.WriteFile(kubeconfig, []byte{}, 0600)
}
