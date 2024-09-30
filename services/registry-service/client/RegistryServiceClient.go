package RegistryServiceClient

import (
	"context"
	"fmt"
	"math/rand"

	"time"

	pb "github.com/TAULargeScaleWorkshop/AAG/services/registry-service/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RegistryServiceClient struct {
	conn   *grpc.ClientConn
	client pb.RegistryServiceClient
}

func NewRegistryServiceClient(addresses []string) *RegistryServiceClient {
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(addresses))
	randomAddress := addresses[randomIndex]
	conn, err := grpc.Dial(randomAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("Failed to connect to %s: %v\n", randomAddress, err)
		return nil
	}

	client := pb.NewRegistryServiceClient(conn)
	return &RegistryServiceClient{
		conn:   conn,
		client: client,
	}
}

func (obj *RegistryServiceClient) Close() {
	if obj.conn != nil {
		obj.conn.Close()
	}
}

func (obj *RegistryServiceClient) Register(serviceName, nodeAddress string) error {
	_, err := obj.client.Register(context.Background(), &pb.RegisterRequest{
		ServiceName: serviceName,
		NodeAddress: nodeAddress,
	})
	if err != nil {
		return fmt.Errorf("could not call Register: %v", err)
	}
	return nil
}

func (obj *RegistryServiceClient) Unregister(serviceName, nodeAddress string) error {
	_, err := obj.client.Unregister(context.Background(), &pb.UnregisterRequest{
		ServiceName: serviceName,
		NodeAddress: nodeAddress,
	})
	if err != nil {
		return fmt.Errorf("could not call Unregister: %v", err)
	}
	return nil
}

func (obj *RegistryServiceClient) Discover(serviceName string) ([]string, error) {
	resp, err := obj.client.Discover(context.Background(), &pb.DiscoverRequest{
		ServiceName: serviceName,
	})
	if err != nil {
		return nil, fmt.Errorf("could not call Discover: %v", err)
	}

	return resp.NodeAddresses, nil
}

func (obj *RegistryServiceClient) IsAlive() (bool, error) {
	resp, err := obj.client.IsAlive(context.Background(), &emptypb.Empty{})
	if err != nil {
		return false, fmt.Errorf("could not call IsAlive: %v", err)
	}
	return resp.GetValue(), nil
}
