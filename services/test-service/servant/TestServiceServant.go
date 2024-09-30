package TestServiceServant

import (
	"fmt"
	"math/rand"
	"time"

	CacheServiceClient "github.com/TAULargeScaleWorkshop/AAG/services/cache-service/client"

	metaffi "github.com/MetaFFI/lang-plugin-go/api"
	"github.com/MetaFFI/plugin-sdk/compiler/go/IDL"
	"github.com/TAULargeScaleWorkshop/AAG/utils"
)

var pythonRuntime *metaffi.MetaFFIRuntime
var crawlerModule *metaffi.MetaFFIModule
var extract_links_from_url func(...interface{}) ([]interface{}, error)

var cacheClient *CacheServiceClient.CacheServiceClient

func init() {
	// Initialize local cacheClient
	var err error
	// cacheClient = CacheServiceClient.NewCacheServiceClient([]string{"127.0.0.1:1000"}, ) // Adjust the address as needed
	// if err != nil {
	// 	msg := fmt.Sprintf("Failed to initialize CacheServiceClient: %v", err)
	// 	utils.Logger.Fatalf(msg)
	// 	panic(msg)
	// }

	pythonRuntime = metaffi.NewMetaFFIRuntime("python311")
	err = pythonRuntime.LoadRuntimePlugin()
	if err != nil {
		msg := fmt.Sprintf("Failed to load runtime plugin: %v", err)
		utils.Logger.Fatalf(msg)
		panic(msg)
	}
	// Load the Crawler module
	crawlerModule, err = pythonRuntime.LoadModule("crawler.py")
	if err != nil {
		msg := fmt.Sprintf("Failed to load ./crawler/crawler.py module: %v", err)
		utils.Logger.Fatalf(msg)
		panic(msg)
	}
	// Load the crawler function
	extract_links_from_url, err = crawlerModule.Load("callable=extract_links_from_url",
		[]IDL.MetaFFIType{IDL.STRING8, IDL.INT64}, // parameters types
		[]IDL.MetaFFIType{IDL.STRING8_ARRAY})      // return type

	if err != nil {
		msg := fmt.Sprintf("Failed to load extract_links_from_url function: %v", err)
		utils.Logger.Fatalf(msg)
		panic(msg)
	}
}

func HelloWorld() string {
	return "Hello World"
}

func HelloToUser(username string) string {
	return "Hello " + username
}

func Store(key string, value string) {
	// cacheMap[key] = value
	err := cacheClient.Set(key, value)
	if err != nil {
		utils.Logger.Printf("Failed to store key-value pair: %v", err)
	}

}

func Get(key string) string {
	// return cacheMap[key]
	value, err := cacheClient.Get(key)
	if err != nil {
		utils.Logger.Printf("Failed to get value for key: %v", err)
		return ""
	}
	return value
}

func WaitAndRand(seconds int32, sendToClient func(x int32) error) error {
	time.Sleep(time.Duration(seconds) * time.Second)
	return sendToClient(int32(rand.Intn(10)))
}

func IsAlive() bool {
	return true
}

func ExtractLinksFromURL(url string, depth int32) ([]string, error) {
	utils.Logger.Printf("ExtractLinksFromURL called with URL: %s and depth: %d", url, depth)

	// Call Python's extract_links_from_url.
	res, err := extract_links_from_url(url, int64(depth))
	if err != nil {
		utils.Logger.Printf("Error calling extract_links_from_url: %v", err)
		return nil, err
	}

	// utils.Logger.Printf("Received result from extract_links_from_url: %v", res)

	links, ok := res[0].([]string)
	if !ok {
		utils.Logger.Printf("Unexpected result type: %T", res[0])
		return nil, fmt.Errorf("unexpected result type: %T", res[0])
	}

	return links, nil
}
