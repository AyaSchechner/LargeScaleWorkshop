package RegistryServiceClient

import (
	"testing"
)

func TestRegistryServiceClient(t *testing.T) {
	addresses := []string{"127.0.0.1:8502"}
	registryClient := NewRegistryServiceClient(addresses)

	// Test Register
	serviceName := "TestService"
	nodeAddress := "127.0.0.1:50051"
	if err := registryClient.Register(serviceName, nodeAddress); err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}
	t.Logf("Registered service '%s' at '%s'", serviceName, nodeAddress)

	// Test Discover
	addresses, err := registryClient.Discover(serviceName)
	if err != nil {
		t.Fatalf("Failed to discover service addresses: %v", err)
	}
	t.Logf("Discovered addresses for service '%s': %v", serviceName, addresses)

	// Test Unregister
	if err := registryClient.Unregister(serviceName, nodeAddress); err != nil {
		t.Fatalf("Failed to unregister service: %v", err)
	}
	t.Logf("Unregistered service '%s' from '%s'", serviceName, nodeAddress)

	// Test IsAlive
	aliveStatus, err := registryClient.IsAlive()
	if err != nil {
		t.Fatalf("Failed to unregister service: %v", err)
	}
	t.Logf("Unregistered service '%s' from '%s'", serviceName, nodeAddress)

	if !aliveStatus {
		t.Fatalf("expected IsAlive to be true, got false")
	}

	t.Logf("Service is alive: %v", aliveStatus)
}
