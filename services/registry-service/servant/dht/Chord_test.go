// package dht

// import (
// 	"fmt"
// 	"testing"
// 	"unsafe"

// 	metaffiRuntime "github.com/MetaFFI/lang-plugin-go/go-runtime"
// )

// // Mock MetaFFI functions
// func mockMetaFFIFunctions() {
// 	// Initialize MetaFFIHandle based on the inferred definition
// 	mockHandle := metaffiRuntime.MetaFFIHandle{
// 		Val:       metaffiRuntime.Handle(unsafe.Pointer(uintptr(1))),
// 		RuntimeID: 1010,
// 	}

// 	newChord = func(...interface{}) ([]interface{}, error) {
// 		return []interface{}{mockHandle}, nil
// 	}
// 	joinChord = func(...interface{}) ([]interface{}, error) {
// 		return []interface{}{mockHandle}, nil
// 	}
// 	set = func(...interface{}) ([]interface{}, error) {
// 		return nil, nil
// 	}
// 	get = func(...interface{}) ([]interface{}, error) {
// 		return []interface{}{"value"}, nil
// 	}
// 	pdelete = func(...interface{}) ([]interface{}, error) {
// 		return nil, nil
// 	}
// 	getAllKeys = func(...interface{}) ([]interface{}, error) {
// 		return []interface{}{[]string{"key1", "key2"}}, nil
// 	}
// 	isFirst = func(...interface{}) ([]interface{}, error) {
// 		return []interface{}{true}, nil
// 	}
// }

// func TestNewChord(t *testing.T) {
// 	mockMetaFFIFunctions()
// 	chord, err := NewChord("node1", 8080)
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	expectedHandle := metaffiRuntime.MetaFFIHandle{
// 		Val:       metaffiRuntime.Handle(unsafe.Pointer(uintptr(1))),
// 		RuntimeID: 1010,
// 	}
// 	if chord.handle.Val != expectedHandle.Val || chord.handle.RuntimeID != expectedHandle.RuntimeID {
// 		t.Fatalf("expected handle to be %v, got %v", expectedHandle, chord.handle)
// 	}
// 	fmt.Println("TestNewChord done")
// }

// func TestJoinChord(t *testing.T) {
// 	mockMetaFFIFunctions()
// 	chord, err := JoinChord("existingNode", "node1", 8080)
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	expectedHandle := metaffiRuntime.MetaFFIHandle{
// 		Val:       metaffiRuntime.Handle(unsafe.Pointer(uintptr(1))),
// 		RuntimeID: 1010,
// 	}
// 	if chord.handle.Val != expectedHandle.Val || chord.handle.RuntimeID != expectedHandle.RuntimeID {
// 		t.Fatalf("expected handle to be %v, got %v", expectedHandle, chord.handle)
// 	}
// }

// func TestIsFirst(t *testing.T) {
// 	mockMetaFFIFunctions()
// 	chord := &Chord{handle: metaffiRuntime.MetaFFIHandle{
// 		Val:       metaffiRuntime.Handle(unsafe.Pointer(uintptr(1))),
// 		RuntimeID: 1010,
// 	}}
// 	isFirst, err := chord.IsFirst()
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if !isFirst {
// 		t.Fatalf("expected isFirst to be true, got %v", isFirst)
// 	}
// }

// func TestSet(t *testing.T) {
// 	mockMetaFFIFunctions()
// 	chord := &Chord{handle: metaffiRuntime.MetaFFIHandle{
// 		Val:       metaffiRuntime.Handle(unsafe.Pointer(uintptr(1))),
// 		RuntimeID: 1010,
// 	}}
// 	err := chord.Set("key", "value")
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// }

// func TestGet(t *testing.T) {
// 	mockMetaFFIFunctions()
// 	chord := &Chord{handle: metaffiRuntime.MetaFFIHandle{
// 		Val:       metaffiRuntime.Handle(unsafe.Pointer(uintptr(1))),
// 		RuntimeID: 1010,
// 	}}
// 	value, err := chord.Get("key")
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if value != "value" {
// 		t.Fatalf("expected value to be 'value', got %v", value)
// 	}
// }

// func TestDelete(t *testing.T) {
// 	mockMetaFFIFunctions()
// 	chord := &Chord{handle: metaffiRuntime.MetaFFIHandle{
// 		Val:       metaffiRuntime.Handle(unsafe.Pointer(uintptr(1))),
// 		RuntimeID: 1010,
// 	}}
// 	err := chord.Delete("key")
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// }

// func TestGetAllKeys(t *testing.T) {
// 	mockMetaFFIFunctions()
// 	chord := &Chord{handle: metaffiRuntime.MetaFFIHandle{
// 		Val:       metaffiRuntime.Handle(unsafe.Pointer(uintptr(1))),
// 		RuntimeID: 1010,
// 	}}
// 	keys, err := chord.GetAllKeys()
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	expectedKeys := []string{"key1", "key2"}
// 	for i, key := range keys {
// 		if key != expectedKeys[i] {
// 			t.Fatalf("expected key to be %v, got %v", expectedKeys[i], key)
// 		}
// 	}
// }

package dht

import (
	"fmt"
	"testing"
)

func TestChordOperations(t *testing.T) {
	fmt.Println("test begin")
	// Create a new Chord instance
	chord, err := NewChord("Node1", 1099)
	if err != nil {
		t.Fatalf("Failed to create Chord instance: %v", err)
	}

	// Test Set and Get operations
	fmt.Println("Test Set and Get operation")
	key := "key1"
	val := "value1"
	err = chord.Set(key, val)
	if err != nil {
		t.Errorf("Set operation failed: %v", err)
	}

	retrievedVal, err := chord.Get(key)
	if err != nil {
		t.Errorf("Get operation failed: %v", err)
	}
	if retrievedVal != val {
		t.Errorf("Expected value %s, got %s", val, retrievedVal)
	}

	// Test Delete operation
	fmt.Println("Test Delete operation")
	err = chord.Delete(key)
	if err != nil {
		t.Errorf("Delete operation failed: %v", err)
	}

	// Test GetAllKeys operation
	fmt.Println("Test GetAllKeys operation")
	keys, err := chord.GetAllKeys()
	if err != nil {
		t.Errorf("GetAllKeys operation failed: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys, got %d", len(keys))
	}

	// Test IsFirst operation
	fmt.Println("Test IsFirst operation")
	isFirst, err := chord.IsFirst()
	if err != nil {
		t.Errorf("IsFirst operation failed: %v", err)
	}
	if !isFirst {
		t.Errorf("Expected isFirst to be true, got false")
	}

}

/*
package dht

import (
	"fmt"
	"testing"
)

// Helper function to create a Chord instance and handle errors
func createChordInstance(t *testing.T) *Chord {
	chord, err := NewChord("Node1", 1099)
	if err != nil {
		t.Fatalf("Failed to create Chord instance: %v", err)
	}
	return chord
}

func TestSetOperation(t *testing.T) {
	fmt.Println("Test Set operation")
	chord := createChordInstance(t)

	key := "key1"
	val := "value1"
	err := chord.Set(key, val)
	if err != nil {
		t.Errorf("Set operation failed: %v", err)
	}

	// Clean up (delete the key after the test)
	fmt.Println("delete key in Set operation")
	defer chord.Delete(key)
}

func TestGetOperation(t *testing.T) {
	fmt.Println("Test Get operation")
	chord := createChordInstance(t)

	key := "key1"
	expectedVal := "value1"
	err := chord.Set(key, expectedVal)
	if err != nil {
		t.Fatalf("Set operation failed: %v", err)
	}

	retrievedVal, err := chord.Get(key)
	if err != nil {
		t.Errorf("Get operation failed: %v", err)
	}
	if retrievedVal != expectedVal {
		t.Errorf("Expected value %s, got %s", expectedVal, retrievedVal)
	}

	// Clean up (delete the key after the test)
	defer chord.Delete(key)
}

func TestDeleteOperation(t *testing.T) {
	fmt.Println("Test Delete operation")
	chord := createChordInstance(t)

	key := "key1"
	val := "value1"
	err := chord.Set(key, val)
	if err != nil {
		t.Fatalf("Set operation failed: %v", err)
	}

	err = chord.Delete(key)
	if err != nil {
		t.Errorf("Delete operation failed: %v", err)
	}

	// Check that the key no longer exists
	_, err = chord.Get(key)
	if err == nil {
		t.Errorf("Expected key %s to be deleted, but it still exists", key)
	}
}

func TestGetAllKeysOperation(t *testing.T) {
	fmt.Println("Test GetAllKeys operation")
	chord := createChordInstance(t)

	// Ensure no keys initially exist
	keys, err := chord.GetAllKeys()
	if err != nil {
		t.Fatalf("GetAllKeys operation failed: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys initially, got %d", len(keys))
	}

	// Set a key for testing
	key := "key1"
	val := "value1"
	err = chord.Set(key, val)
	if err != nil {
		t.Fatalf("Set operation failed: %v", err)
	}

	// Test GetAllKeys after setting a key
	keys, err = chord.GetAllKeys()
	if err != nil {
		t.Errorf("GetAllKeys operation failed: %v", err)
	}
	if len(keys) != 1 {
		t.Errorf("Expected 1 key, got %d", len(keys))
	}

	// Clean up (delete the key after the test)
	defer chord.Delete(key)
}

func TestIsFirstOperation(t *testing.T) {
	fmt.Println("Test IsFirst operation")
	chord := createChordInstance(t)

	isFirst, err := chord.IsFirst()
	if err != nil {
		t.Errorf("IsFirst operation failed: %v", err)
	}
	if !isFirst {
		t.Errorf("Expected isFirst to be true, got false")
	}
}
*/
