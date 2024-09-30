#!/bin/bash

# Set the output directory
OUTPUT_DIR="./output"

# Create the output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

echo "Building the code..."

# Build the Go code
go build -o "$OUTPUT_DIR/large-scale-workshop" /workspaces/AAG/main.go

if [ $? -ne 0 ]; then
  echo "Build failed. Exiting."
  exit 1
fi

echo "Build successful."

exit 0


# # #!/bin/bash

# # go build -o ./output/large-scale-workshop
# # # Define the output directory
# # OUTPUT_DIR="/workspaces/AAG/output"
# # DHT_DIR="/workspaces/AAG/services/registry-service/servant/dht"

# # # Create the output directory if it doesn't exist
# # mkdir -p "$OUTPUT_DIR"

# # echo "Building the code..."

# # # Build the Go code
# # go build -o "$OUTPUT_DIR/large-scale-workshop" /workspaces/AAG/main.go

# # if [ $? -ne 0 ]; then
# #   echo "Build failed. Exiting."
# #   exit 1
# # fi

# # echo "Build successful."

# # echo "Copying configuration files and additional files to the output directory..."

# # # Copy configuration files
# # cp -r /workspaces/AAG/config "$OUTPUT_DIR/"

# # # Copy Python scripts if any
# # cp -r /workspaces/AAG/python "$OUTPUT_DIR/"

# # # Copy .class files for Chord
# # cp -r $DHT_DIR/*.class "$OUTPUT_DIR/"

# # echo "All files copied successfully."

# # echo "Build and setup completed."

# # exit 0

# #!/bin/bash

# # Define the output directory
# OUTPUT_DIR="/workspaces/AAG/output"
# DHT_DIR="/workspaces/AAG/services/registry-service/servant/dht"

# # Create the output directory if it doesn't exist
# mkdir -p "$OUTPUT_DIR"

# echo "Building the code..."

# # Build the Go code
# go build -o "$OUTPUT_DIR/large-scale-workshop" /workspaces/AAG/main.go

# if [ $? -ne 0 ]; then
#   echo "Build failed. Exiting."
#   exit 1
# fi

# echo "Build successful."

# echo "Copying configuration files and additional files to the output directory..."

# # Copy configuration files
# if [ -d "/workspaces/AAG/config" ]; then
#   cp -r /workspaces/AAG/config "$OUTPUT_DIR/"
# else
#   echo "No config directory found, skipping..."
# fi

# # Copy .class files for Chord
# if [ -d "$DHT_DIR" ]; then
#   cp -r $DHT_DIR/*.class "$OUTPUT_DIR/"
# else
#   echo "No .class files found in DHT directory, skipping..."
# fi

# echo "All files copied successfully."

# echo "Build and setup completed."

# exit 0
