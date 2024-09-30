package common

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	RegistryServiceClient "github.com/TAULargeScaleWorkshop/AAG/services/registry-service/client"

	zmq4 "github.com/pebbe/zmq4"
	"google.golang.org/grpc"
)

type ServiceClientBase[client_t any] struct {
	RegistryAddresses []string
	CreateClient      func(grpc.ClientConnInterface) client_t
	RegistryClient    *RegistryServiceClient.RegistryServiceClient
}

func NewServiceClientBase[client_t any](registryClient *RegistryServiceClient.RegistryServiceClient, addresses []string, createClient func(grpc.ClientConnInterface) client_t) *ServiceClientBase[client_t] {
	return &ServiceClientBase[client_t]{
		RegistryAddresses: addresses,
		CreateClient:      createClient,
		RegistryClient:    registryClient,
	}
}

func (obj *ServiceClientBase[client_t]) pickNode(serviceName string) (string, error) {
	nodes, err := obj.RegistryClient.Discover(serviceName)
	if err != nil {
		return "", fmt.Errorf("failed to discover nodes: %v", err)
	}
	if len(nodes) == 0 {
		return "", fmt.Errorf("no nodes available")
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	selectedNode := nodes[r.Intn(len(nodes))]

	return selectedNode, nil
}

func (obj *ServiceClientBase[client_t]) Connect(serviceName string) (res client_t, closeFunc func(), err error) {
	NodeAddress, err := obj.pickNode(serviceName)
	if err != nil {
		var empty client_t
		return empty, nil, fmt.Errorf("failed to pickNode: %v", err)
	}
	conn, err := grpc.Dial(NodeAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		var empty client_t
		return empty, nil, fmt.Errorf("failed to connect client to %v: %v", NodeAddress, err)
	}
	c := obj.CreateClient(conn)
	return c, func() { conn.Close() }, nil
}

// getMQNodes retrieves the list of MQ nodes from the registry
func (obj *ServiceClientBase[client_t]) getMQNodes() ([]string, error) {
	nodes, err := obj.RegistryClient.Discover("TestServiceMQ")
	if err != nil {
		return nil, fmt.Errorf("failed to discover MQ nodes: %v", err)
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no MQ nodes available")
	}

	// Use the helper function to extract only the MQ addresses
	mqAddresses := extractMQAddresses(nodes)
	return mqAddresses, nil
}

// ConnectMQ connects to all MQ nodes and returns a ZeroMQ socket
func (obj *ServiceClientBase[client_t]) ConnectMQ() (socket *zmq4.Socket, err error) {
	nodes, err := obj.getMQNodes()
	if err != nil {
		return nil, fmt.Errorf("failed to get MQ nodes: %v", err)
	}
	socket, err = zmq4.NewSocket(zmq4.REQ)
	if err != nil {
		return nil, fmt.Errorf("failed to create ZeroMQ socket: %v", err)
	}

	// Connect to all the nodes
	for _, node := range nodes {
		err := socket.Connect(node)
		if err != nil {
			socket.Close() // Clean up if there is an error
			return nil, fmt.Errorf("failed to connect to MQ node %v: %v", node, err)
		}
		log.Printf("Connected to MQ node: %s", node)
	}

	return socket, nil
}

// extractMQAddresses takes a list of nodes with appended addresses and returns only the MQ addresses.
func extractMQAddresses(nodes []string) []string {
	mqAddresses := make([]string, 0, len(nodes))
	for _, node := range nodes {
		parts := strings.Split(node, "@")
		if len(parts) > 0 {
			mqAddresses = append(mqAddresses, parts[0]) // Take only the MQ address part
		}
	}
	return mqAddresses
}
