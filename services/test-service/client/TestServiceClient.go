package TestService

import (
	"context"
	"fmt"

	services "github.com/TAULargeScaleWorkshop/AAG/services/common"

	RegistryServiceClient "github.com/TAULargeScaleWorkshop/AAG/services/registry-service/client"
	service "github.com/TAULargeScaleWorkshop/AAG/services/test-service/common"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const serviceName = "TestService"

type TestServiceClient struct {
	services.ServiceClientBase[service.TestServiceClient]
}

func NewTestServiceClient(address []string, registryClient *RegistryServiceClient.RegistryServiceClient) *TestServiceClient {
	return &TestServiceClient{
		ServiceClientBase: services.ServiceClientBase[service.TestServiceClient]{
			RegistryAddresses: address,
			CreateClient:      service.NewTestServiceClient,
			RegistryClient:    registryClient,
		},
	}
}

func (obj *TestServiceClient) HelloWorld() (string, error) {
	c, closeFunc, err := obj.Connect(serviceName)
	defer closeFunc()
	if err != nil {
		return "", err
	}
	r, err := c.HelloWorld(context.Background(), &emptypb.Empty{})
	if err != nil {
		return "", fmt.Errorf("TSC could not call HelloWorld: %v", err)
	}
	return r.Value, nil
}

func (obj *TestServiceClient) HelloToUser(username string) (string, error) {
	c, closeFunc, err := obj.Connect(serviceName)
	if err != nil {
		return "", err
	}
	defer closeFunc()

	req := &wrapperspb.StringValue{Value: username}
	r, err := c.HelloToUser(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("could not call HelloToUser: %v", err)
	}
	return r.Value, nil
}

func (obj *TestServiceClient) Store(key, value string) error {
	c, closeFunc, err := obj.Connect(serviceName)

	if err != nil {
		return err
	}
	defer closeFunc()
	req := &service.StoreKeyValue{Key: key, Value: value}
	_, err = c.Store(context.Background(), req)
	if err != nil {
		return fmt.Errorf("could not store key-value pair: %v", err)
	}
	return nil
}

func (obj *TestServiceClient) Get(key string) (string, error) {
	c, closeFunc, err := obj.Connect(serviceName)
	if err != nil {
		return "", err
	}
	defer closeFunc()

	req := &wrapperspb.StringValue{Value: key}
	r, err := c.Get(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("could not get value for key: %v", err)
	}
	return r.Value, nil
}

func (obj *TestServiceClient) WaitAndRand(seconds int32) (func() (int32, error), error) {
	c, closeFunc, err := obj.Connect(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect %v. Error: %v", obj.RegistryAddresses, err)
	}
	r, err := c.WaitAndRand(context.Background(), &wrapperspb.Int32Value{Value: seconds})
	if err != nil {
		return nil, fmt.Errorf("could not call Get: %v", err)
	}
	res := func() (int32, error) {
		defer closeFunc()
		x, err := r.Recv()
		return x.Value, err
	}
	return res, nil
}

func (obj *TestServiceClient) IsAlive() (*wrapperspb.BoolValue, error) {
	c, closeFunc, err := obj.Connect(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect %v. Error: %v", obj.RegistryAddresses, err)
	}
	defer closeFunc()

	req := &emptypb.Empty{}
	r, err := c.IsAlive(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("could not call IsAlive: %v", err)
	}
	return r, nil
}

func (obj *TestServiceClient) ExtractLinksFromURL(url string, depth int32) ([]string, error) {
	c, closeFunc, err := obj.Connect(serviceName)
	if err != nil {
		return nil, err
	}
	defer closeFunc()

	req := &service.ExtractLinksFromURLParameters{
		Url:   url,
		Depth: depth,
	}
	r, err := c.ExtractLinksFromURL(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("could not call ExtractLinksFromURL: %v", err)
	}
	return r.Links, nil
}

func (obj *TestServiceClient) HelloWorldAsync() (func() (string, error), error) {
	mqsocket, err := obj.ConnectMQ()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MQ: %w", err)
	}

	msg, err := services.NewMarshaledCallParameter("HelloWorld", &emptypb.Empty{})
	if err != nil {
		mqsocket.Close()
		return nil, fmt.Errorf("failed to marshal call parameters: %w", err)
	}
	_, err = mqsocket.SendBytes(msg, 0)
	if err != nil {
		mqsocket.Close()
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	ret := func() (string, error) {
		defer mqsocket.Close()
		rv, err := mqsocket.RecvBytes(0)
		if err != nil {
			return "", fmt.Errorf("failed to receive response: %w", err)
		}
		str := &wrapperspb.StringValue{}
		err = proto.Unmarshal(rv, str)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal string value: %w", err)
		}
		return str.Value, nil
	}

	return ret, nil
}

func (obj *TestServiceClient) ExtractLinksFromURLAsync(url string, depth int32) (func() ([]string, error), error) {
	// Step 1: Connect to the MQ
	mqsocket, err := obj.ConnectMQ()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MQ: %w", err)
	}

	// Step 2: Prepare request parameters
	req := &service.ExtractLinksFromURLParameters{
		Url:   url,
		Depth: depth,
	}

	// Step 3: Marshal the call parameters
	msg, err := services.NewMarshaledCallParameter("ExtractLinksFromURL", req)
	if err != nil {
		mqsocket.Close()
		return nil, fmt.Errorf("failed to marshal call parameters: %w", err)
	}

	// Step 4: Send the marshaled message
	_, err = mqsocket.SendBytes(msg, 0)
	if err != nil {
		mqsocket.Close()
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Step 5: Return a function to receive the response asynchronously
	ret := func() ([]string, error) {
		defer mqsocket.Close()

		// Step 6: Receive the response
		rv, err := mqsocket.RecvBytes(0)
		if err != nil {
			return nil, fmt.Errorf("failed to receive response: %w", err)
		}

		// Step 7: Unmarshal the response
		resp := &service.ExtractLinksFromURLReturnedValue{}
		err = proto.Unmarshal(rv, resp)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		// Step 8: Return the extracted links
		return resp.Links, nil
	}

	return ret, nil
}
