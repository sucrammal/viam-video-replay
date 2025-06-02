# Video Replay Module Test Scripts

This directory contains test scripts to verify both **local mode** and **dataset mode** functionality of the video replay module.

## Prerequisites

1. **Go Environment**: Ensure Go is installed and this module is built:

    ```bash
    cd /path/to/viam-video-replay
    go mod tidy
    make server
    ```

2. **OpenCV**: Ensure OpenCV (gocv) is properly installed on your system.

## Test Structure

Each test is organized in its own subdirectory to avoid `main()` function conflicts:

```
test_scripts/
├── simple_local_test/main.go     # Local video test
├── simple_dataset_test/main.go   # Dataset connectivity test
├── config_local_example.json     # Local mode config
├── config_dataset_example.json   # Dataset mode config
├── run_tests.sh                  # Interactive test runner
└── README.md                     # This file
```

## Local Mode Tests

### Simple Local Test

Tests basic OpenCV video reading functionality:

```bash
# Update the video path in the script first
go run test_scripts/simple_local_test/main.go
```

**Prerequisites:**

-   A video file (MP4, MOV, AVI, etc.) accessible on your local system
-   Update `testVideoPath` variable in `test_scripts/simple_local_test/main.go`

**What it tests:**

-   OpenCV can open and read the video file
-   Frame extraction and JPEG encoding
-   Basic video properties (FPS, dimensions)

### Full Local Module Test

Tests the complete local mode module functionality:

```bash
# Build and test with a local video file
go run main.go -config test_scripts/config_local_example.json
```

**What it tests:**

-   Module initialization in local mode
-   Frame update loop
-   Camera API compliance (Image(), Properties(), etc.)
-   Thread-safe frame handling

## Dataset Mode Tests

### Simple Dataset Test

Tests basic Viam Data API connectivity:

```bash
# Set environment variables first
export VIAM_API_KEY="your-api-key"
export VIAM_API_KEY_ID="your-api-key-id"
export VIAM_ORG_ID="your-organization-id"
export VIAM_DATASET_ID="your-dataset-id"

go run test_scripts/simple_dataset_test/main.go
```

**Prerequisites:**

-   Viam account at https://app.viam.com
-   API credentials (see Getting Credentials section below)
-   A dataset with image data

**What it tests:**

-   Viam Data API connectivity
-   Dataset image fetching
-   Image decoding and validation
-   Metadata extraction

### Full Dataset Module Test

Tests the complete dataset mode module functionality:

```bash
# Set environment variables first (same as above)
go run main.go -config test_scripts/config_dataset_example.json
```

**What it tests:**

-   Module initialization in dataset mode
-   Dataset image caching
-   Dataset replay loop
-   Camera API compliance

## Interactive Test Runner

Use the interactive test runner for convenience:

```bash
./test_scripts/run_tests.sh
```

This provides a menu-driven interface to run all tests with proper environment checking.

## Getting Viam Credentials

To test dataset mode, you'll need credentials from your Viam account:

1. **Login to Viam**: Go to https://app.viam.com and log in
2. **Create API Key**:
    - Navigate to your organization settings
    - Go to "API Keys" section
    - Create a new API key
    - Save both the **API Key** and **API Key ID**
3. **Get Organization ID**:
    - Found in your organization settings URL or profile
4. **Get Dataset ID**:
    - Navigate to your dataset in the Data tab
    - The dataset ID is in the URL or dataset details

## Configuration Files

You can also test using configuration files:

### Local Mode Config

```json
{
	"components": [
		{
			"name": "test-camera",
			"api": "rdk:component:camera",
			"model": "bill:camera:video-replay",
			"attributes": {
				"mode": "local",
				"video_path": "/path/to/your/video.mp4",
				"fps": 10,
				"loop_video": true
			}
		}
	]
}
```

### Dataset Mode Config

```json
{
	"components": [
		{
			"name": "test-dataset-camera",
			"api": "rdk:component:camera",
			"model": "bill:camera:video-replay",
			"attributes": {
				"mode": "dataset",
				"api_key": "your-api-key",
				"api_key_id": "your-api-key-id",
				"organization_id": "your-org-id",
				"dataset_id": "your-dataset-id",
				"fps": 10
			}
		}
	]
}
```

Then test with:

```bash
go run main.go -config test_scripts/config_local_example.json
```

## Troubleshooting

### Common Issues

1. **"Video file not found"**:

    - Check the video file path in `test_scripts/simple_local_test/main.go`
    - Ensure the file is readable
    - Try with different video formats (MP4, MOV, AVI)

2. **"Failed to create Viam client"**:

    - Check your API credentials
    - Ensure network connectivity
    - Verify organization ID and API key permissions

3. **"No images found in dataset"**:

    - Verify the dataset ID is correct
    - Ensure the dataset contains image data
    - Check that your API key has access to the dataset

4. **OpenCV errors**:

    - Ensure OpenCV/gocv is properly installed
    - Check video codec support
    - Try with a different video file

5. **"main redeclared" errors**:
    - Each test is now in its own directory to avoid this
    - Run tests individually: `go run test_scripts/simple_local_test/main.go`

### Debug Mode

Add debug logging by setting:

```bash
export VIAM_LOG_LEVEL=debug
```

### Output Files

Test scripts generate output files:

-   `simple_test_frame_*.jpg` - Frames from simple local test
-   `dataset_frame_*.jpg` - Decoded frames from dataset test
-   `dataset_raw_*.bin` - Raw binary data (for debugging decode issues)

## Quick Start

1. **Test local mode** (requires a video file):

    ```bash
    # Edit video path in test_scripts/simple_local_test/main.go first
    go run test_scripts/simple_local_test/main.go
    ```

2. **Test dataset mode** (requires Viam credentials):

    ```bash
    export VIAM_API_KEY="your-key"
    export VIAM_API_KEY_ID="your-key-id"
    export VIAM_ORG_ID="your-org-id"
    export VIAM_DATASET_ID="your-dataset-id"

    go run test_scripts/simple_dataset_test/main.go
    ```

3. **Use interactive runner**:
    ```bash
    ./test_scripts/run_tests.sh
    ```

## Next Steps

After running these tests successfully:

1. **Integration Testing**: Test the module with a full Viam robot configuration
2. **Performance Testing**: Test with larger datasets and longer videos
3. **Edge Cases**: Test with corrupted files, network interruptions, etc.
4. **Production Deployment**: Use in your actual robot configuration

## Support

If you encounter issues:

1. Check the troubleshooting section above
2. Review the main module logs
3. Create an issue with details about your setup and error messages
