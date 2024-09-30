package cacheservice

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

	. "github.com/TAULargeScaleWorkshop/AAG/services/cache-service/common"
	dht "github.com/TAULargeScaleWorkshop/AAG/services/registry-service/servant/dht"

	services "github.com/TAULargeScaleWorkshop/AAG/services/common"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"gopkg.in/yaml.v2"
)

var mut sync.Mutex

type cacheServiceImplementation struct {
	UnimplementedCacheServiceServer
	mutex sync.Mutex
	Chord *dht.Chord
}

type Config struct {
	Type            string `yaml:"type"`
	RegistryAddress string `yaml:"registryAddress"`
	RegNum          int    `yaml:"regNum"`
	Port            int    `yaml:"port"` // root node port
	ChordPort       int    `yaml:"chordPort"`
	ChordNodeName   string `yaml:"chordNodeName"`
}

func loadConfigFromData(configData []byte) (*Config, error) {
	var config Config
	err := yaml.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
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

func findAvailablePort(startPort int) (net.Listener, int, error) {
	mut.Lock()
	defer mut.Unlock()
	port := startPort
	for {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			_ = lis.Close() // Close the listener as we only need to check availability
			return lis, port, nil
		}
		if opErr, ok := err.(*net.OpError); ok && strings.Contains(opErr.Error(), "address already in use") {
			port++
			continue
		}
		return nil, 0, fmt.Errorf("failed to check port %d: %v", port, err)
	}
}

func Start(configData []byte) error {
	config, err := loadConfigFromData(configData)
	if err != nil {
		log.Printf("Failed to unmarshal config: %v", err)
		return err
	}

	serviceName := config.Type
	registryAddresses, err := generateRegAddresses(config.RegistryAddress, config.RegNum)
	if err != nil {
		log.Printf("Failed to generate registry addresses: %v", err)
		return err
	}

	var chord *dht.Chord
	_, newPort, err := findAvailablePort(config.Port)
	if err != nil {
		log.Printf("Failed to find available port: %v", err)
		return err
	}
	mut.Lock()
	if config.Port == newPort {
		chord, err = dht.NewChord(config.ChordNodeName, int32(config.ChordPort))
		if err != nil {
			log.Printf("Failed to initialize Chord: %v", err)
			return err
		}
		log.Printf("Chord initialized successfully")

	} else {
		chord, err = dht.JoinChord("ChordNode"+strconv.Itoa(newPort), config.ChordNodeName, int32(config.ChordPort))
		if err != nil {
			log.Printf("Failed to join Chord: %v", err)
			return err
		}
		log.Println("Chord joined successfully")
	}
	mut.Unlock()

	bindgRPCToService := func(s grpc.ServiceRegistrar) {
		RegisterCacheServiceServer(s, &cacheServiceImplementation{Chord: chord})
	}
	grpcServer := grpc.NewServer()
	RegisterCacheServiceServer(grpcServer, &cacheServiceImplementation{Chord: chord})

	newAddress := services.Start(serviceName, newPort, bindgRPCToService)
	unregister := services.RegisterAddress(serviceName, registryAddresses, newAddress)

	if unregister == nil {
		log.Fatalf("Failed to register the service\n")
	}
	log.Printf("CacheService registered at %v", newAddress)

	return nil
}

func (c *cacheServiceImplementation) Set(ctx context.Context, req *StoreKeyValue) (*emptypb.Empty, error) {
	c.mutex.Lock()
	err := c.Chord.Set(req.Key, req.Value)
	if err != nil {
		return nil, err
	}
	c.mutex.Unlock()
	return &emptypb.Empty{}, nil
}

func (c *cacheServiceImplementation) Get(ctx context.Context, req *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
	c.mutex.Lock()
	value, err := c.Chord.Get(req.Value)
	if err != nil {
		return nil, err
	}
	c.mutex.Unlock()
	return wrapperspb.String(value), nil
}

func (c *cacheServiceImplementation) Delete(ctx context.Context, req *wrapperspb.StringValue) (*emptypb.Empty, error) {
	c.mutex.Lock()
	err := c.Chord.Delete(req.Value)
	if err != nil {
		return nil, err
	}
	c.mutex.Unlock()
	return &emptypb.Empty{}, nil
}

func (c *cacheServiceImplementation) IsAlive(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	_, err := c.Chord.IsFirst()
	if err != nil {
		return wrapperspb.Bool(false), nil
	}
	return wrapperspb.Bool(true), nil
}
