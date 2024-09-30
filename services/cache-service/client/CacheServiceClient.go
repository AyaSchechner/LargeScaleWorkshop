package CacheServiceClient

import (
	"context"
	"time"

	service "github.com/TAULargeScaleWorkshop/AAG/services/cache-service/common"
	services "github.com/TAULargeScaleWorkshop/AAG/services/common"

	RegistryServiceClient "github.com/TAULargeScaleWorkshop/AAG/services/registry-service/client"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const serviceName = "CacheService"

type CacheServiceClient struct {
	services.ServiceClientBase[service.CacheServiceClient]
}

func NewCacheServiceClient(addresses []string, registryClient *RegistryServiceClient.RegistryServiceClient) *CacheServiceClient {
	return &CacheServiceClient{
		ServiceClientBase: services.ServiceClientBase[service.CacheServiceClient]{
			RegistryAddresses: addresses,
			CreateClient:      service.NewCacheServiceClient,
			RegistryClient:    registryClient,
		},
	}
}

// conn, err := grpc.Dial(address, grpc.WithInsecure())
// if err != nil {
// 	return nil, err
// }
// return &CacheServiceClient{client: CacheService.NewCacheServiceClient(conn)}, nil
// }

func (obj *CacheServiceClient) Set(key, value string) error {
	c, closeFunc, err := obj.Connect(serviceName)
	if err != nil {
		return err
	}
	defer closeFunc()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &service.StoreKeyValue{Key: key, Value: value}
	_, err = c.Set(ctx, req)
	return err
}

func (obj *CacheServiceClient) Get(key string) (string, error) {
	c, closeFunc, err := obj.Connect(serviceName)
	if err != nil {
		return "", err
	}
	defer closeFunc()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &wrapperspb.StringValue{Value: key}
	resp, err := c.Get(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Value, nil
}

func (obj *CacheServiceClient) Delete(key string) error {
	c, closeFunc, err := obj.Connect(serviceName)
	if err != nil {
		return err
	}
	defer closeFunc()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &wrapperspb.StringValue{Value: key}
	_, err = c.Delete(ctx, req)
	return err
}

func (obj *CacheServiceClient) IsAlive() (bool, error) {
	c, closeFunc, err := obj.Connect(serviceName)
	if err != nil {
		return false, err
	}
	defer closeFunc()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := c.IsAlive(ctx, &emptypb.Empty{})
	if err != nil {
		return false, err
	}
	return resp.GetValue(), nil
}
