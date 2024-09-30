package TestService

import (
	"log"
	"testing"

	RegistryServiceClient "github.com/TAULargeScaleWorkshop/AAG/services/registry-service/client"
)

func startTestService() ([]string, *RegistryServiceClient.RegistryServiceClient) {
	registryAddressesList := []string{"127.0.0.1:8502", "127.0.0.1:8503", "127.0.0.1:8504"}
	newRegistryClient := RegistryServiceClient.NewRegistryServiceClient(registryAddressesList)

	serviceName := "TestService"
	discoveredAddresses, err := newRegistryClient.Discover(serviceName)

	if err != nil {
		log.Fatal("could not Discover addresses")
		return nil, nil
	}
	return discoveredAddresses, newRegistryClient
}

func TestHelloWorld(t *testing.T) {
	addresses, registryClient := startTestService()
	c := NewTestServiceClient(addresses, registryClient)
	r, err := c.HelloWorld()
	if err != nil {
		t.Fatalf("TSCT, could not call HelloWorld: %v", err)
		return
	}
	t.Logf("Response: %v", r)
}

func TestHelloToUser(t *testing.T) {
	addresses, registryClient := startTestService()
	c := NewTestServiceClient(addresses, registryClient)
	username := "AAG"
	r, err := c.HelloToUser(username)
	if err != nil {
		t.Fatalf("could not call HelloToUser: %v", err)
	}

	expected := "Hello " + username
	if r != expected {
		t.Errorf("HelloToUser(%s) = %s; want %s", username, r, expected)
	}
	t.Logf("Response: %v", r)

}

func TestStoreAndGet(t *testing.T) {
	addresses, registryClient := startTestService()

	c := NewTestServiceClient(addresses, registryClient)

	key := "key1"
	value := "value1"

	err := c.Store(key, value)
	if err != nil {
		t.Fatalf("could not store key '%s': %v", key, err)
	}
	retValue, err := c.Get(key)
	if err != nil {
		t.Fatalf("could not get value for key '%s': %v", key, err)
	}
	if retValue != value {
		t.Errorf("unexpected value for key '%s', got: %s, want: %s", key, retValue, value)
	}

	value2 := "value2"
	err = c.Store(key, value2)
	if err != nil {
		t.Fatalf("could not store key '%s': %v", key, err)
	}
	retValue, err = c.Get(key)
	if err != nil {
		t.Fatalf("could not get value for key '%s': %v", key, err)
	}

	if retValue != value2 {
		t.Errorf("unexpected value for key '%s', got: %s, want: %s", key, retValue, value)
	}
}
func TestStore(t *testing.T) {
	addresses, registryClient := startTestService()
	c := NewTestServiceClient(addresses, registryClient)
	key, value := "key1", "value1"

	err := c.Store(key, value)
	if err != nil {
		t.Fatalf("could not store key-value pair: %v", err)
	}

	t.Logf("Stored key-value pair: %s -> %s", key, value)
}

func TestGet(t *testing.T) {
	addresses, registryClient := startTestService()
	c := NewTestServiceClient(addresses, registryClient)
	key, value := "key1", "value1"

	// Store the key-value pair to ensure it exists before retrieving it
	err := c.Store(key, value)
	if err != nil {
		t.Fatalf("could not store key key-value pair: %v", err)
	}

	// Retrieve the value for the key
	r, err := c.Get(key)
	if err != nil {
		t.Fatalf("could not get value for key '%s': %v", key, err)
	}

	if r != value {
		t.Errorf("Get(%s) = %s; want %s", key, r, value)
	}

	t.Logf("Retrieved key-value pair: %s -> %s", key, r)
}

func TestWaitAndRand(t *testing.T) {
	addresses, registryClient := startTestService()
	c := NewTestServiceClient(addresses, registryClient)
	resPromise, err := c.WaitAndRand(3)
	if err != nil {
		t.Fatalf("Calling WaitAndRand failed: %v", err)
		return
	}
	res, err := resPromise()
	if err != nil {
		t.Fatalf("WaitAndRand failed: %v", err)
		return
	}
	t.Logf("Returned random number: %v\n", res)
}

func TestIsAlive(t *testing.T) {
	addresses, registryClient := startTestService()
	c := NewTestServiceClient(addresses, registryClient)

	// Call IsAlive
	aliveStatus, err := c.IsAlive()
	if err != nil {
		t.Fatalf("could not check IsAlive status: %v", err)
	}

	if !aliveStatus.GetValue() {
		t.Fatalf("expected IsAlive to be true, got false")
	}

	t.Logf("Service is alive: %v", aliveStatus.GetValue())
}

func TestExtractLinksFromURL(t *testing.T) {
	addresses, registryClient := startTestService()
	c := NewTestServiceClient(addresses, registryClient)
	url := "http://example.com"
	depth := int32(3) // Adjust depth to reach deeper levels

	links, err := c.ExtractLinksFromURL(url, depth)
	if err != nil {
		t.Fatalf("could not call ExtractLinksFromURL: %v", err)
	}
	expectedLink := "https://www.icann.org/privacy/cookies"
	if !contains(links, expectedLink) {
		t.Errorf("expected link %q not found in response: %v", expectedLink, links)
	}
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func TestHelloWorldAsync(t *testing.T) {
	addresses, registryClient := startTestService()
	c := NewTestServiceClient(addresses, registryClient)
	r, err := c.HelloWorldAsync()
	if err != nil {
		t.Fatalf("could not call HelloWorld Async: %v", err)
		return
	}
	res, err := r()
	if err != nil {
		t.Fatalf("HelloWorld Async returned error: %v", err)
		return
	}
	t.Logf("Async Response: %v", res)
}

func TestExtractLinksFromURLAsync(t *testing.T) {
	addresses, registryClient := startTestService()
	c := NewTestServiceClient(addresses, registryClient)
	url := "http://example.com"
	depth := int32(3) // Adjust depth to reach deeper levels

	// Asynchronously extract links from URL
	linksPromise, err := c.ExtractLinksFromURLAsync(url, depth)
	if err != nil {
		t.Fatalf("could not call ExtractLinksFromURLAsync: %v", err)
		return
	}

	// Use the promise function to get the links
	links, err := linksPromise()
	if err != nil {
		t.Fatalf("ExtractLinksFromURLAsync returned error: %v", err)
		return
	}

	// Check if a specific link is present in the response
	expectedLink := "https://www.icann.org/privacy/cookies"
	if !contains(links, expectedLink) {
		t.Errorf("expected link %q not found in response: %v", expectedLink, links)
	}

	t.Logf("Async Extracted Links seccessfully")
}
