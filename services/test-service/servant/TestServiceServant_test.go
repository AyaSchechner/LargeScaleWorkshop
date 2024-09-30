package TestServiceServant

import (
	"testing"

	metaffi "github.com/MetaFFI/lang-plugin-go/api"
	"github.com/MetaFFI/plugin-sdk/compiler/go/IDL"
	"github.com/stretchr/testify/assert"
)

func TestExtractLinksFromURL(t *testing.T) {
	// Mock Python module path
	mockModulePath := "./mock_crawler.py"

	// Initialize Python runtime and load mock module
	pythonRuntime := metaffi.NewMetaFFIRuntime("python311")
	err := pythonRuntime.LoadRuntimePlugin()
	if err != nil {
		t.Fatalf("Failed to load Python runtime plugin: %v", err)
	}

	mockModule, err := pythonRuntime.LoadModule(mockModulePath)
	if err != nil {
		t.Fatalf("Failed to load mock module %s: %v", mockModulePath, err)
	}

	// Load the function from the mock module
	extractLinksFunc, err := mockModule.Load("callable=extract_links_from_url",
		[]IDL.MetaFFIType{IDL.STRING8, IDL.INT64},
		[]IDL.MetaFFIType{IDL.STRING8_ARRAY})
	if err != nil {
		t.Fatalf("Failed to load function from mock module: %v", err)
	}

	// Test case: Mock a URL and depth
	url := "https://example.com"
	depth := int32(2)

	// Call the mocked Python function
	res, err := extractLinksFunc(url, int64(depth))
	if err != nil {
		t.Errorf("Error calling extractLinksFunc: %v", err)
	}

	// Assert the results
	assert.NotNil(t, res, "Expected non-nil result")
	assert.Equal(t, []string{"http://example.com/page1", "http://example.com/page2"}, res[0].([]string), "Expected specific links from mock response")
}
