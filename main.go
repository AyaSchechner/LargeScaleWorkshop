package main

import (
	"log"
	"os"

	"github.com/TAULargeScaleWorkshop/AAG/config"
	CacheService "github.com/TAULargeScaleWorkshop/AAG/services/cache-service/service"       // Import CacheService package
	RegistryService "github.com/TAULargeScaleWorkshop/AAG/services/registry-service/service" // import RegistryService package
	TestService "github.com/TAULargeScaleWorkshop/AAG/services/test-service/service"         // Import TestService package
	"github.com/TAULargeScaleWorkshop/AAG/utils"
	"gopkg.in/yaml.v2"
)

func main() {
	// read configuration file from command line argument
	if len(os.Args) != 2 {
		utils.Logger.Fatal("Expecting exactly one configuration file")
		os.Exit(1)
	}
	configFile := os.Args[1]
	configData, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("error reading file: %v", err)
		os.Exit(2)
	}
	var config config.ConfigBase
	err = yaml.Unmarshal(configData, &config) // parses YAML
	if err != nil {
		log.Fatalf("error unmarshaling data: %v", err)
		os.Exit(3)
	}

	switch config.Type {
	case "TestService":
		utils.Logger.Printf("Loading service type: %v\n", config.Type)
		TestService.Start(configData)

	case "RegistryService":
		utils.Logger.Printf("Loading service type: %v\n", config.Type)
		err := RegistryService.Start(configFile)
		if err != nil {
			utils.Logger.Fatalf("Failed to start RegistryService: %v", err)
			os.Exit(5)

		}

	case "CacheService":
		utils.Logger.Printf("Loading service type: %v\n", config.Type)
		err := CacheService.Start(configData)
		if err != nil {
			utils.Logger.Fatalf("Failed to start CacheService")
			os.Exit(5)
		}

	default:
		utils.Logger.Fatalf("Unknown configuration type: %v", config.Type)
		os.Exit(4)
	}
	select {}
}
