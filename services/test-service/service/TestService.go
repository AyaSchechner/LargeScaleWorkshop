package TestService

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	CacheServiceClient "github.com/TAULargeScaleWorkshop/AAG/services/cache-service/client" //
	RegistryServiceClient "github.com/TAULargeScaleWorkshop/AAG/services/registry-service/client"

	services "github.com/TAULargeScaleWorkshop/AAG/services/common"
	pb "github.com/TAULargeScaleWorkshop/AAG/services/test-service/common"
	TestServiceServant "github.com/TAULargeScaleWorkshop/AAG/services/test-service/servant"
	"gopkg.in/yaml.v2"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var serviceInstance *testServiceImplementation

type Config struct {
	Type            string `yaml:"type"`
	RegistryAddress string `yaml:"registryAddress"`
	RegNum          int    `yaml:"regNum"`
}

type testServiceImplementation struct {
	pb.UnimplementedTestServiceServer
	CacheClient *CacheServiceClient.CacheServiceClient
}

func loadConfigFromData(configData []byte) (*Config, error) {
	var config Config
	err := yaml.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func ConnectCacheService(registryAddresses []string, registryClient *RegistryServiceClient.RegistryServiceClient) *testServiceImplementation {
	cacheClient := CacheServiceClient.NewCacheServiceClient(registryAddresses, registryClient)
	return &testServiceImplementation{CacheClient: cacheClient}
}

func generateRegAddresses(baseAddress string, numAddresses int) ([]string, error) {
	var addresses []string

	// Split the base address into host and port
	parts := strings.Split(baseAddress, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid base address format")
	}

	basePort, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid port number")
	}

	// Generate addresses
	for i := 0; i < numAddresses; i++ {
		newAddress := fmt.Sprintf("%s:%d", parts[0], basePort+i)
		addresses = append(addresses, newAddress)
	}
	return addresses, nil
}

func Start(configData []byte) string {
	// Unmarshal the configuration data
	config, err := loadConfigFromData(configData)
	if err != nil {
		log.Printf("Failed to unmarshal config: %v", err)
		return ""
	}

	serviceName := config.Type
	// Generate registry addresses based on the base address and regNum
	registryAddresses, err := generateRegAddresses(config.RegistryAddress, config.RegNum)
	registryClient := RegistryServiceClient.NewRegistryServiceClient(registryAddresses)
	if err != nil {
		log.Printf("Failed to generate registry addresses: %v", err)
		return ""
	}

	testServiceImp := ConnectCacheService(registryAddresses, registryClient)
	serviceInstance = testServiceImp
	bindgRPCToService := func(s grpc.ServiceRegistrar) {
		pb.RegisterTestServiceServer(s, testServiceImp)
	}
	grpcServer := grpc.NewServer()
	bindgRPCToService(grpcServer)

	newAddress := services.Start(serviceName, 0, bindgRPCToService)
	// MQ setup
	startMQ, mqAddress := services.BindMQToService(0, messageHandler)
	MQwithTestAddress := mqAddress + "@" + newAddress

	unregister := services.RegisterAddress(serviceName, registryAddresses, newAddress)

	if unregister == nil {
		log.Fatalf("Failed to register the service\n")
	}

	go startMQ()

	// Register MQ address
	registerMQAddress := services.RegisterAddress(serviceName+"MQ", registryAddresses, MQwithTestAddress)
	if registerMQAddress == nil {
		log.Fatalf("Failed to register MQ address")
	}

	return newAddress
}

func (obj *testServiceImplementation) HelloWorld(ctxt context.Context, _ *emptypb.Empty) (res *wrapperspb.StringValue, err error) {
	return wrapperspb.String(TestServiceServant.HelloWorld()), nil
}

func (obj *testServiceImplementation) HelloToUser(ctx context.Context, req *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
	// Call the servant's HelloToUser method with req's value
	message := TestServiceServant.HelloToUser(req.Value)
	return &wrapperspb.StringValue{Value: message}, nil
}

func (obj *testServiceImplementation) Store(ctx context.Context, req *pb.StoreKeyValue) (*emptypb.Empty, error) {
	err := obj.CacheClient.Set(req.Key, req.Value)
	if err != nil {
		return &emptypb.Empty{}, nil
	}
	return &emptypb.Empty{}, nil
}

func (obj *testServiceImplementation) Get(ctx context.Context, req *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
	value, err := obj.CacheClient.Get(req.Value)
	if err != nil {
		return nil, fmt.Errorf("could not get value for key: %v", err)
	}
	return wrapperspb.String(value), nil
}

func (obj *testServiceImplementation) WaitAndRand(seconds *wrapperspb.Int32Value, streamRet pb.TestService_WaitAndRandServer) error {
	streamClient := func(x int32) error {
		return streamRet.Send(wrapperspb.Int32(x))
	}
	return TestServiceServant.WaitAndRand(seconds.Value, streamClient)
}

func (obj *testServiceImplementation) IsAlive(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	return &wrapperspb.BoolValue{Value: true}, nil
}

func (obj *testServiceImplementation) ExtractLinksFromURL(ctx context.Context, req *pb.ExtractLinksFromURLParameters) (*pb.ExtractLinksFromURLReturnedValue, error) {
	if req == nil {
		return nil, fmt.Errorf("URL required")

	}
	links, err := TestServiceServant.ExtractLinksFromURL(req.Url, req.Depth)
	if err != nil {
		return nil, err
	}

	return &pb.ExtractLinksFromURLReturnedValue{Links: links}, nil
}

// messageHandler using reflection
func messageHandler(method string, parameters []byte) (proto.Message, error) {
	// Get the reflect.Value of the serviceInstance
	instanceValue := reflect.ValueOf(serviceInstance)

	// Find the method by name
	methodValue := instanceValue.MethodByName(method)
	if !methodValue.IsValid() {
		return nil, fmt.Errorf("MQ message called unknown method: %v", method)
	}

	// Get the method type and number of inputs
	methodType := methodValue.Type()
	if methodType.NumIn() != 2 { // Ensure the method has exactly 2 inputs: context and proto.Message
		return nil, fmt.Errorf("method %s has unexpected number of inputs", method)
	}

	// Determine the parameter type expected by the method
	paramType := methodType.In(1).Elem()

	// Create a new instance of the parameter type
	paramInstance := reflect.New(paramType).Interface()

	// Unmarshal parameters into the correct type
	err := proto.Unmarshal(parameters, paramInstance.(proto.Message))
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters: %v", err)
	}

	// Call the method with context and parameter instance
	result := methodValue.Call([]reflect.Value{
		reflect.ValueOf(context.Background()),
		reflect.ValueOf(paramInstance),
	})

	// Extract the response and error from the result
	response := result[0].Interface().(proto.Message)
	errInterface := result[1].Interface()

	if errInterface != nil {
		return nil, errInterface.(error)
	}

	return response, nil
}
