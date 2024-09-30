package common

import (
	"fmt"
	"log"
	"net"
	"time"

	RegistryServiceClient "github.com/TAULargeScaleWorkshop/AAG/services/registry-service/client"
	"github.com/TAULargeScaleWorkshop/AAG/utils"
	"github.com/pebbe/zmq4"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func startgRPC(listenPort int) (listeningAddress string, grpcServer *grpc.Server, startListening func()) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", listenPort))
	if err != nil {
		utils.Logger.Fatalf("failed to listen: %v", err)
	}
	listeningAddress = lis.Addr().String()
	grpcServer = grpc.NewServer()

	startListening = func() {
		if err := grpcServer.Serve(lis); err != nil {
			utils.Logger.Fatalf("failed to serve: %v", err)
		}
	}
	return
}

func Start(serviceName string, grpcListenPort int, bindgRPCToService func(s grpc.ServiceRegistrar)) string {
	listeningAddress, grpcServer, startListening := startgRPC(grpcListenPort)
	// Extract the port number from the listening address
	_, port, err := net.SplitHostPort(listeningAddress)
	if err != nil {
		panic(fmt.Sprintf("Failed to split host and port: %v", err))
	}
	fmt.Printf("Service %s is listening at port %s\n", serviceName, port) // Optional: For logging purposes

	bindgRPCToService(grpcServer)
	go startListening()
	return listeningAddress
}

func RegisterAddress(serviceName string, registryAddresses []string, listeningAddress string) (unregister func()) {
	registryClient := RegistryServiceClient.NewRegistryServiceClient(registryAddresses)

	err := registryClient.Register(serviceName, listeningAddress)
	if err != nil {
		utils.Logger.Fatalf("Failed to register to registry service: %v", err)
	}

	return func() {
		registryClient.Unregister(serviceName, listeningAddress)
	}
}

func BindMQToService(listenPort int, messageHandler func(method string, parameters []byte) (response proto.Message, err error)) (startMQ func(), listeningAddress string) {
	socket, err := zmq4.NewSocket(zmq4.REP)
	if err != nil {
		log.Fatalf("Failed to create a new zmq socket: %v", err)
	}

	if listenPort == 0 {
		listeningAddress = "tcp://127.0.0.1:*"
	} else {
		listeningAddress = fmt.Sprintf("tcp://127.0.0.1:%v", listenPort)
	}

	err = socket.Bind(listeningAddress)
	if err != nil {
		log.Fatalf("Failed to bind a zmq socket: %v", err)
	}

	listeningAddress, err = socket.GetLastEndpoint()
	if err != nil {
		log.Fatalf("Failed to get listening address of zmq socket: %v", err)
	}
	startMQ = func() {
		for {
			time.Sleep(10 * time.Second)
			data, readErr := socket.RecvBytes(0)
			if readErr != nil {
				log.Printf("Failed to receive bytes from MQ socket: %v\n", readErr)
				continue
			}
			if len(data) == 0 {
				continue
			}
			utils.Logger.Printf("data len: %v\n", len(data))

			go func(data []byte) {

				var parameters CallParameters
				if err := proto.Unmarshal(data, &parameters); err != nil {
					log.Printf("Failed to unmarshal data: %v", err)
					return
				}

				response, err := messageHandler(parameters.Method, parameters.Data)
				if err != nil {
					log.Printf("Message handler error: %v\n", err)
					return
				}

				returnData, err := proto.Marshal(response)
				if err != nil {
					log.Printf("Failed to marshal ReturnValue: %v\n", err)
					return
				}
				_, sendErr := socket.SendBytes(returnData, 0)
				if sendErr != nil {
					log.Printf("Failed to send response: %v\n", sendErr)
				}
			}(data)
		}
	}

	return startMQ, listeningAddress
}

func NewMarshaledCallParameter(method string, msg proto.Message) ([]byte, error) {
	params := &CallParameters{
		Method: method,
		Data:   nil,
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	params.Data = data
	return proto.Marshal(params)
}
