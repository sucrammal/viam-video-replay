#!/bin/bash

# Video Replay Module Test Runner
# This script helps you run various tests for the video replay module

set -e  # Exit on any error

echo "=== Video Replay Module Test Runner ==="
echo

# Load .env file if it exists
if [[ -f ".env" ]]; then
    echo "Loading environment variables from .env file..."
    export $(grep -v '^#' .env | xargs)
    echo "✅ Environment variables loaded"
else
    echo "No .env file found. Use test_scripts/env_template.txt as a guide."
fi
echo

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
echo "Checking prerequisites..."

if ! command_exists go; then
    echo "❌ Go is not installed or not in PATH"
    exit 1
fi

if ! command_exists pkg-config; then
    echo "⚠️  pkg-config not found - OpenCV might not work"
fi

echo "✅ Go found: $(go version)"
echo

# Function to run a test
run_test() {
    local test_name="$1"
    local test_dir="$2"
    local description="$3"
    
    echo "=== $test_name ==="
    echo "$description"
    echo
    
    if [[ ! -d "$test_dir" ]]; then
        echo "❌ Test directory not found: $test_dir"
        return 1
    fi
    
    if [[ ! -f "$test_dir/main.go" ]]; then
        echo "❌ Test file not found: $test_dir/main.go"
        return 1
    fi
    
    echo "Running: go run $test_dir/main.go"
    echo "Press Ctrl+C to stop, or wait for completion..."
    echo
    
    if (cd "$test_dir" && go run main.go); then
        echo "✅ $test_name completed successfully"
    else
        echo "❌ $test_name failed"
        return 1
    fi
    echo
}

# Show menu
show_menu() {
    echo "Available tests:"
    echo "1. Simple Local Test - Test basic OpenCV video reading"
    echo "2. Simple Dataset Test - Test Viam Data API connectivity"  
    echo "3. Local Module Integration Test - Test full local mode module"
    echo "4. Dataset Module Integration Test - Test full dataset mode module"
    echo "5. Run all local tests"
    echo "6. Run all dataset tests"
    echo "7. Run all tests"
    echo "8. Show help"
    echo "q. Quit"
    echo
}

# Show help
show_help() {
    echo "=== Test Help ==="
    echo
    echo "Local Mode Tests:"
    echo "  - Require a local video file (MP4, MOV, AVI, etc.)"
    echo "  - Update video path in test_scripts/simple_local_test/main.go before running"
    echo "  - Test OpenCV video reading and module functionality"
    echo
    echo "Dataset Mode Tests:" 
    echo "  - Require Viam credentials (API key, org ID, dataset ID)"
    echo "  - Option 1: Create a .env file in project root with:"
    echo "    VIAM_API_KEY=your-api-key"
    echo "    VIAM_API_KEY_ID=your-api-key-id"
    echo "    VIAM_ORG_ID=your-org-id"
    echo "    VIAM_DATASET_ID=your-dataset-id"
    echo "  - Option 2: Set environment variables manually:"
    echo "    export VIAM_API_KEY=\"your-api-key\""
    echo "    export VIAM_API_KEY_ID=\"your-api-key-id\""
    echo "    export VIAM_ORG_ID=\"your-org-id\""
    echo "    export VIAM_DATASET_ID=\"your-dataset-id\""
    echo "  - Option 3: Source the template file:"
    echo "    cp test_scripts/env_template.txt .env"
    echo "    # Edit .env with your credentials, then run this script"
    echo "  - Get credentials from https://app.viam.com"
    echo
    echo "Manual test commands:"
    echo "  go run test_scripts/simple_local_test/main.go"
    echo "  go run test_scripts/simple_dataset_test/main.go"
    echo
    echo "For more details, see test_scripts/README.md"
    echo
}

# Check dataset environment variables
check_dataset_env() {
    if [[ -z "$VIAM_API_KEY" || -z "$VIAM_API_KEY_ID" || -z "$VIAM_ORG_ID" || -z "$VIAM_DATASET_ID" ]]; then
        echo "❌ Dataset tests require environment variables:"
        echo "   VIAM_API_KEY, VIAM_API_KEY_ID, VIAM_ORG_ID, VIAM_DATASET_ID"
        echo "   Run option 8 for setup instructions"
        echo "   Or create a .env file using test_scripts/env_template.txt as a guide"
        return 1
    fi
    echo "✅ Dataset environment variables found"
    return 0
}

# Main loop
while true; do
    show_menu
    read -p "Select an option: " choice
    echo
    
    case $choice in
        1)
            run_test "Simple Local Test" "test_scripts/simple_local_test" \
                "Tests basic OpenCV video reading functionality"
            ;;
        2)
            if check_dataset_env; then
                run_test "Simple Dataset Test" "test_scripts/simple_dataset_test" \
                    "Tests basic Viam Data API connectivity"
            fi
            ;;
        3)
            echo "❌ Local Module Integration Test not yet implemented"
            echo "   You can test the full module using a config file:"
            echo "   go run main.go -config test_scripts/config_local_example.json"
            ;;
        4)
            echo "❌ Dataset Module Integration Test not yet implemented"
            echo "   You can test the full module using a config file with dataset mode:"
            echo "   go run main.go -config test_scripts/config_dataset_example.json"
            ;;
        5)
            echo "=== Running All Local Tests ==="
            run_test "Simple Local Test" "test_scripts/simple_local_test" \
                "Tests basic OpenCV video reading functionality"
            # Add more local tests here when implemented
            ;;
        6)
            echo "=== Running All Dataset Tests ==="
            if check_dataset_env; then
                run_test "Simple Dataset Test" "test_scripts/simple_dataset_test" \
                    "Tests basic Viam Data API connectivity"
                # Add more dataset tests here when implemented
            fi
            ;;
        7)
            echo "=== Running All Tests ==="
            run_test "Simple Local Test" "test_scripts/simple_local_test" \
                "Tests basic OpenCV video reading functionality"
            if check_dataset_env; then
                run_test "Simple Dataset Test" "test_scripts/simple_dataset_test" \
                    "Tests basic Viam Data API connectivity"
            fi
            ;;
        8)
            show_help
            ;;
        q|Q)
            echo "Goodbye!"
            exit 0
            ;;
        *)
            echo "❌ Invalid option: $choice"
            ;;
    esac
    
    read -p "Press Enter to continue..."
    echo
done 