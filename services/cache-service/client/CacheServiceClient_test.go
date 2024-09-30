package CacheServiceClient

import (
	"context"
	"net"
	"testing"
	"time"

	common "github.com/TAULargeScaleWorkshop/AAG/services/cache-service/common" // Adjust this path as needed

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestCacheServiceClient(t *testing.T) {
	// Create a listener
	lis, err := net.Listen("tcp", ":0") // Use :0 to get an available port
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer lis.Close() // Ensure the listener is closed after tests

	// Create a mock server
	mockServer := grpc.NewServer()
	service := &MockCacheService{data: make(map[string]string)}
	common.RegisterCacheServiceServer(mockServer, service)

	// Start the mock server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- mockServer.Serve(lis)
	}()
	defer func() {
		mockServer.GracefulStop()
		if err := <-serverErr; err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait briefly to ensure the server starts
	time.Sleep(100 * time.Millisecond)

	// Connect to the server
	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := common.NewCacheServiceClient(conn)

	// Test Set method
	t.Run("Set", func(t *testing.T) {
		req := &common.StoreKeyValue{Key: "testKey", Value: "testValue"}
		_, err := client.Set(context.Background(), req)
		if err != nil {
			t.Fatalf("Set() failed: %v", err)
		}
	})

	// Test Get method
	t.Run("Get", func(t *testing.T) {
		req := &wrapperspb.StringValue{Value: "testKey"}
		res, err := client.Get(context.Background(), req)
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}
		if res.Value != "testValue" {
			t.Errorf("Get() returned wrong value: got %v, want %v", res.Value, "testValue")
		}
	})

	// Test Delete method
	t.Run("Delete", func(t *testing.T) {
		req := &wrapperspb.StringValue{Value: "testKey"}
		_, err := client.Delete(context.Background(), req)
		if err != nil {
			t.Fatalf("Delete() failed: %v", err)
		}
	})

	// Test IsAlive method
	t.Run("IsAlive", func(t *testing.T) {
		_, err := client.IsAlive(context.Background(), &emptypb.Empty{})
		if err != nil {
			t.Fatalf("IsAlive() failed: %v", err)
		}
	})
}

// Mock implementation of CacheServiceServer
type MockCacheService struct {
	common.UnimplementedCacheServiceServer
	data map[string]string
}

func (m *MockCacheService) Set(ctx context.Context, req *common.StoreKeyValue) (*emptypb.Empty, error) {
	m.data[req.Key] = req.Value
	return &emptypb.Empty{}, nil
}

func (m *MockCacheService) Get(ctx context.Context, req *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
	val, exists := m.data[req.Value]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "key not found")
	}
	return &wrapperspb.StringValue{Value: val}, nil
}

func (m *MockCacheService) Delete(ctx context.Context, req *wrapperspb.StringValue) (*emptypb.Empty, error) {
	delete(m.data, req.Value)
	return &emptypb.Empty{}, nil
}

func (m *MockCacheService) IsAlive(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	return &wrapperspb.BoolValue{Value: true}, nil
}
