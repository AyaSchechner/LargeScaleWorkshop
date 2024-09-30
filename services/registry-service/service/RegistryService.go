package registryservice

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	CacheServicePb "github.com/TAULargeScaleWorkshop/AAG/services/cache-service/common"
	pb "github.com/TAULargeScaleWorkshop/AAG/services/registry-service/common"
	dht "github.com/TAULargeScaleWorkshop/AAG/services/registry-service/servant/dht"
	TestServicePb "github.com/TAULargeScaleWorkshop/AAG/services/test-service/common"

	// "github.com/TAULargeScaleWorkshop/AAG/utils"
	"io/ioutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"gopkg.in/yaml.v2"
)

var mut sync.Mutex
var mut2 sync.Mutex

type Config struct {
	Type                 string `yaml:"type"`
	Port                 int    `yaml:"port"` // root node port
	IsAliveCheckInterval int    `yaml:"isAliveCheckInterval"`
	ChordPort            int    `yaml:"chordPort"`
	ChordNodeName        string `yaml:"chordNodeName"`
}

func LoadConfig(configFile string) (*Config, error) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

type RegistryServiceServer struct {
	pb.UnimplementedRegistryServiceServer
	mutex        sync.Mutex
	isAliveCheck time.Duration
	Chord        *dht.Chord
}

func Start(configFile string) error {
	config, err := LoadConfig(configFile)
	if err != nil {
		return err
	}
	return startRegistryService(config)
}

func findAvailablePort(startPort int) (net.Listener, int, error) {
	mut2.Lock()
	defer mut2.Unlock()
	port := startPort
	for {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			return lis, port, nil
		}
		if opErr, ok := err.(*net.OpError); ok && strings.Contains(opErr.Error(), "address already in use") {
			port++
			continue
		}
		return nil, 0, fmt.Errorf("failed to check port %d: %v", port, err)
	}
}

func startRegistryService(config *Config) error {
	lis, newPort, err := findAvailablePort(config.Port)
	if err != nil {
		return err
	}
	// Initialize or join Chord based on IsRoot
	var chord *dht.Chord
	mut.Lock()
	if config.Port == newPort {
		chord, err = dht.NewChord(config.ChordNodeName, int32(config.ChordPort))
		if err != nil {
			log.Printf("Failed to initialize Chord: %v", err)
			return err
		}
		log.Printf("Chord initialized successfully")

	} else {
		chord, err = dht.JoinChord("ChordNotRoot"+strconv.Itoa(newPort), config.ChordNodeName, int32(config.ChordPort))
		if err != nil {
			log.Printf("Failed to join Chord: %v", err)
			return err
		}
		log.Println("Chord joined successfully")
	}
	s := grpc.NewServer()
	server := &RegistryServiceServer{
		Chord:        chord,
		isAliveCheck: time.Duration(config.IsAliveCheckInterval) * time.Second,
	}

	mut.Unlock()
	first, err := chord.IsFirst()
	log.Printf("is first: %v", first)

	if err != nil {
		log.Printf("Failed to call IsFirst Chord: %v", err)
		return err
	}
	if first {
		go server.IsAliveCheck()
	}
	pb.RegisterRegistryServiceServer(s, server)
	log.Printf("RegistryService listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Printf("Failed to serve: %v", err)
		return err
	}
	return nil
}

func appendAddress(existingAddresses, newAddress string) string {
	if existingAddresses == "" {
		return newAddress
	}
	return existingAddresses + "," + newAddress
}

func removeAddress(addressList, addressToRemove string) string {
	addresses := strings.Split(addressList, ",")
	filteredAddresses := []string{}
	for _, addr := range addresses {
		if addr != addressToRemove {
			filteredAddresses = append(filteredAddresses, addr)
		}
	}
	return strings.Join(filteredAddresses, ",")
}

func (s *RegistryServiceServer) Register(ctx context.Context, req *pb.RegisterRequest) (*emptypb.Empty, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	serviceName := req.GetServiceName()
	nodeAddress := req.GetNodeAddress()

	// Retrieve all keys from the DHT
	servicesList, err := s.Chord.GetAllKeys()
	if err != nil {
		log.Printf("Error retrieving all services: %v\n", err)
		return nil, err
	}

	var existingAddresses string

	// Check if the service already exists
	for _, service := range servicesList {
		if service == serviceName {
			existingAddresses, err = s.Chord.Get(serviceName)
			if err != nil {
				log.Printf("Error retrieving existing addresses for %s: %v\n", serviceName, err)
				return nil, err
			}
			break
		}
	}

	// Append the new address to the existing list
	addressList := appendAddress(existingAddresses, nodeAddress)

	err = s.Chord.Set(serviceName, addressList)
	if err != nil {
		return nil, err
	}

	log.Printf("Registered %s at %s\n", serviceName, nodeAddress)
	return &emptypb.Empty{}, nil
}

func (s *RegistryServiceServer) Unregister(ctx context.Context, req *pb.UnregisterRequest) (*emptypb.Empty, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	serviceName := req.GetServiceName()
	nodeAddress := req.GetNodeAddress()

	if !s.ContainsService(serviceName) {
		log.Printf("Service already deleted")
		return nil, nil
	}
	// Retrieve the current list of addresses for the service
	existingAddresses, err := s.Chord.Get(serviceName)
	if err != nil {
		log.Printf("Failed to get addresses for service %s: %v\n", serviceName, err)
		return nil, err
	}

	// Remove the specific node address from the list
	updatedAddresses := removeAddress(existingAddresses, nodeAddress)

	// If no addresses are left, delete the service entry completely
	if updatedAddresses == "" {
		err = s.Chord.Delete(serviceName)
		if err != nil {
			log.Printf("Failed to delete service entry for %s: %v\n", serviceName, err)
			return nil, err
		}
		log.Printf("Unregistered %s from %s\n", serviceName, nodeAddress)
	} else {
		// Update the registry with the modified list of addresses
		err = s.Chord.Set(serviceName, updatedAddresses)
		if err != nil {
			log.Printf("Failed to update addresses for %s: %v\n", serviceName, err)
			return nil, err
		}
		log.Printf("Updated addresses for %s after unregistration: %v\n", serviceName, updatedAddresses)
	}

	return &emptypb.Empty{}, nil
}

func (s *RegistryServiceServer) Discover(ctx context.Context, req *pb.DiscoverRequest) (*pb.DiscoverResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	serviceName := req.GetServiceName()

	if !s.ContainsService(serviceName) {
		log.Printf("Service already deleted")
		return nil, nil
	}
	// Get the serialized address list
	serializedAddresses, err := s.Chord.Get(serviceName)
	if err != nil {
		return nil, grpc.Errorf(codes.NotFound, "Service not found")
	}

	// Deserialize the address list
	nodeAddresses := strings.Split(serializedAddresses, ",")

	log.Printf("Discovered addresses for %s: %v\n", serviceName, nodeAddresses)
	return &pb.DiscoverResponse{NodeAddresses: nodeAddresses}, nil
}

func (s *RegistryServiceServer) IsAlive(ctx context.Context, req *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	return &wrapperspb.BoolValue{Value: true}, nil
}

func (s *RegistryServiceServer) IsAliveCheck() {
	for range time.Tick(s.isAliveCheck) {
		servicesList, err := s.Chord.GetAllKeys()
		if err != nil {
			log.Printf("Failed to get service keys from Chord: %v", err)
			return
		}
		log.Printf("GetAllKeys returned servicesList: %v", servicesList)

		for _, serviceName := range servicesList {
			nodeAddresses, err := s.Chord.Get(serviceName)
			if err != nil {
				log.Printf("Failed to get node address for service %v: %v", serviceName, err)
				return
			}
			if nodeAddresses == "" {
				log.Printf("No address found for service: %s", serviceName)
				return
			}
			// Check each address for health
			addresses := strings.Split(nodeAddresses, ",")
			for _, nodeAddress := range addresses {
				log.Printf("Checking health of %s at %s\n", serviceName, nodeAddress)
				go s.checkNodeHealth(serviceName, nodeAddress)
			}
		}
	}
}

func (s *RegistryServiceServer) checkNodeHealth(serviceName, nodeAddress string) {
	conn, err := grpc.Dial(nodeAddress, grpc.WithInsecure())
	if err != nil {
		log.Printf("Connection failed to %s: %v\n", nodeAddress, err)
		s.handleFailure(serviceName, nodeAddress)
		return
	}
	defer conn.Close()

	var isAliveResponse func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*wrapperspb.BoolValue, error)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	switch serviceName {
	case "TestService":
		client := TestServicePb.NewTestServiceClient(conn)
		isAliveResponse = client.IsAlive
	case "CacheService":
		client := CacheServicePb.NewCacheServiceClient(conn)
		isAliveResponse = client.IsAlive
	default:
		return
	}

	_, err = isAliveResponse(ctx, &emptypb.Empty{})

	if err != nil {
		log.Printf("IsAlive check failed for %s at %s: %v", serviceName, nodeAddress, err)
		// Retry logic for health check
		time.Sleep(8 * time.Second)
		log.Printf("calling isAliveResponse again")
		_, err := isAliveResponse(ctx, &emptypb.Empty{})
		if err != nil {
			log.Printf("IsAlive check failed again for %s at %s: %v\n", serviceName, nodeAddress, err)
			s.handleFailure(serviceName, nodeAddress)
		}
	} else {
		s.handleSuccess(serviceName, nodeAddress)
	}
}

func (s *RegistryServiceServer) handleFailure(serviceName, nodeAddress string) {
	log.Printf("handleFailure: %v, %v", serviceName, nodeAddress)
	var err error
	if serviceName == "TestService" {
		// Handle removal of the TestServiceMQ
		s.handleMQFailure(nodeAddress)
	}

	_, err = s.Unregister(context.Background(), &pb.UnregisterRequest{
		ServiceName: serviceName,
		NodeAddress: nodeAddress,
	})

	if err != nil {
		log.Printf("Failed to unregister failed node %s for service %s: %v\n", nodeAddress, serviceName, err)
		return
	}
	log.Printf("Unregistered failed node %s for service %s\n", nodeAddress, serviceName)
}

func (s *RegistryServiceServer) handleMQFailure(nodeAddress string) {
	mqServiceName := "TestServiceMQ"
	if !s.ContainsService(mqServiceName) {
		log.Printf("Service already deleted")
		return
	}
	mqAddresses, err := s.Chord.Get(mqServiceName)
	if err != nil {
		log.Printf("Failed to get MQ addresses for %s: %v\n", mqServiceName, err)
		return
	}
	mqAddress := GetMQAddress(mqAddresses, nodeAddress)
	_, err = s.Unregister(context.Background(), &pb.UnregisterRequest{
		ServiceName: mqServiceName,
		NodeAddress: mqAddress,
	})

	if err != nil {
		log.Printf("Failed to unregister failed node %s for service %s: %v\n", mqAddress, mqServiceName, err)
		return
	}
	log.Printf("Deleted failed MQ address for node %s from service %s\n", mqAddress, mqServiceName)
}

func GetMQAddress(mqAddresses string, nodeAddress string) string {
	addresses := strings.Split(mqAddresses, ",")
	for _, addr := range addresses {
		// Extract the TestService address from MQwithTestAddress
		parts := strings.Split(addr, "@")
		if len(parts) == 2 && parts[1] == nodeAddress {
			return strings.Join(parts, "@")
		}
	}
	return ""
}

func (s *RegistryServiceServer) ContainsService(serviceName string) bool {
	servicesList, err := s.Chord.GetAllKeys()
	if err != nil {
		log.Printf("Failed to get service keys from Chord: %v", err)
		return false
	}
	for _, service := range servicesList {
		if strings.Contains(service, serviceName) {
			return true
		}
	}
	return false
}

func (s *RegistryServiceServer) handleSuccess(serviceName, nodeAddress string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log.Println("handleSuccess:", serviceName, nodeAddress)
}
