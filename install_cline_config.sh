#!/bin/bash

# Get the absolute path of the script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Define the target directory for Cline MCP settings
TARGET_DIR="$HOME/Library/Application Support/Code/User/globalStorage/saoudrizwan.claude-dev/settings"

# Create the target directory if it doesn't exist
mkdir -p "$TARGET_DIR"

# Read the existing cline_mcp_settings.json file if it exists
if [ -f "$TARGET_DIR/cline_mcp_settings.json" ]; then
    echo "Existing Cline MCP settings found. Merging with new settings..."
    
    # Create a temporary file for the merged settings
    TEMP_FILE=$(mktemp)
    
    # Use jq to merge the existing settings with the new settings
    # If jq is not available, we'll just overwrite the file
    if command -v jq &> /dev/null; then
        jq -s '.[0] * .[1]' "$TARGET_DIR/cline_mcp_settings.json" "$SCRIPT_DIR/updated_cline_mcp_settings.json" > "$TEMP_FILE"
        
        # Check if the merge was successful
        if [ $? -eq 0 ]; then
            # Replace the existing file with the merged file
            mv "$TEMP_FILE" "$TARGET_DIR/cline_mcp_settings.json"
            echo "Settings merged successfully."
        else
            echo "Error merging settings. Overwriting with new settings..."
            cp "$SCRIPT_DIR/updated_cline_mcp_settings.json" "$TARGET_DIR/cline_mcp_settings.json"
        fi
    else
        echo "jq not found. Overwriting with new settings..."
        cp "$SCRIPT_DIR/updated_cline_mcp_settings.json" "$TARGET_DIR/cline_mcp_settings.json"
    fi
else
    echo "No existing Cline MCP settings found. Installing new settings..."
    cp "$SCRIPT_DIR/updated_cline_mcp_settings.json" "$TARGET_DIR/cline_mcp_settings.json"
fi

# Update the command path in the settings file
BINARY_PATH="$SCRIPT_DIR/build/restv2-api-server-go"
if command -v jq &> /dev/null; then
    # Use jq to update the command path
    TEMP_FILE=$(mktemp)
    jq --arg path "$BINARY_PATH" '.mcpServers["Solace REST V2 MCP Go"].command = $path' "$TARGET_DIR/cline_mcp_settings.json" > "$TEMP_FILE"
    
    # Check if the update was successful
    if [ $? -eq 0 ]; then
        # Replace the existing file with the updated file
        mv "$TEMP_FILE" "$TARGET_DIR/cline_mcp_settings.json"
        echo "Command path updated successfully."
    else
        echo "Error updating command path."
    fi
else
    echo "jq not found. Unable to update command path automatically."
    echo "Please update the command path manually in $TARGET_DIR/cline_mcp_settings.json"
fi

# Update the cwd path in the settings file
if command -v jq &> /dev/null; then
    # Use jq to update the cwd path
    TEMP_FILE=$(mktemp)
    jq --arg path "$SCRIPT_DIR" '.mcpServers["Solace REST V2 MCP Go"].cwd = $path' "$TARGET_DIR/cline_mcp_settings.json" > "$TEMP_FILE"
    
    # Check if the update was successful
    if [ $? -eq 0 ]; then
        # Replace the existing file with the updated file
        mv "$TEMP_FILE" "$TARGET_DIR/cline_mcp_settings.json"
        echo "Working directory path updated successfully."
    else
        echo "Error updating working directory path."
    fi
else
    echo "jq not found. Unable to update working directory path automatically."
    echo "Please update the working directory path manually in $TARGET_DIR/cline_mcp_settings.json"
fi

echo "Cline MCP settings installed successfully."
echo "Please restart VS Code for the changes to take effect."
