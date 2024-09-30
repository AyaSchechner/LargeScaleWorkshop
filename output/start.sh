#!/bin/bash

# Path to the executable
EXECUTABLE="/workspaces/AAG/output/large-scale-workshop"

# Path to the configuration files
CONFIG_DIR="/workspaces/AAG/output"

# Function to start a service
start_service() {
  local service_name="$1"
  local config_file="$2"
  local log_file="$3"

  echo "Starting $service_name with config $config_file"
  nohup "$EXECUTABLE" "$config_file" > "$log_file" 2>&1 &
  echo "$service_name started, logs available at $log_file"
}

# Start 3 Registry Services
start_service "RegistryService 1" "$CONFIG_DIR/RegistryService.yaml" "registry1.log"
sleep 5
start_service "RegistryService 2" "$CONFIG_DIR/RegistryService.yaml" "registry2.log"
sleep 5
start_service "RegistryService 3" "$CONFIG_DIR/RegistryService.yaml" "registry3.log"
sleep 5

# Start 3 Cache Services
start_service "CacheService 1" "$CONFIG_DIR/CacheService.yaml" "cache1.log"
sleep 10
start_service "CacheService 2" "$CONFIG_DIR/CacheService.yaml" "cache2.log"
sleep 10
start_service "CacheService 3" "$CONFIG_DIR/CacheService.yaml" "cache3.log"
sleep 5

# Start 3 Test Services
start_service "TestService 1" "$CONFIG_DIR/TestService.yaml" "test1.log"
sleep 5
start_service "TestService 2" "$CONFIG_DIR/TestService.yaml" "test2.log"
sleep 5
start_service "TestService 3" "$CONFIG_DIR/TestService.yaml" "test3.log"
sleep 5


echo "All services started."

exit 0
