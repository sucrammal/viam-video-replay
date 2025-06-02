# Viam Video Replay Module

A Viam module that provides a camera component capable of replaying video files or Viam dataset images at a specified frame rate.

## Features

-   **Local Mode**: Replay video files (MP4, AVI, etc.) using OpenCV
-   **Dataset Mode**: Replay images from Viam datasets using the Viam data client
-   Configurable frame rate (FPS)
-   Loop playback support for local videos
-   Seamless integration with Viam camera API

## Configuration

### Local Mode (Video Files)

```json
{
	"mode": "local",
	"video_path": "/path/to/your/video.mp4",
	"fps": 10,
	"loop_video": true
}
```

### Dataset Mode (Viam Datasets)

```json
{
	"mode": "dataset",
	"api_key": "your-viam-api-key",
	"api_key_id": "your-api-key-id",
	"organization_id": "your-org-id",
	"dataset_id": "your-dataset-id",
	"fps": 10
}
```

### Configuration Parameters

-   `mode`: Operating mode - `"local"` (default) or `"dataset"`
-   `video_path`: Path to video file (required for local mode)
-   `fps`: Frames per second for playback (default: 10)
-   `loop_video`: Whether to loop video playback (local mode only)
-   `api_key`: Viam API key (required for dataset mode)
-   `api_key_id`: Viam API key ID (required for dataset mode)
-   `organization_id`: Viam organization ID (required for dataset mode)
-   `dataset_id`: ID of the dataset to replay (required for dataset mode)

## Adding to Viam Machine Configuration

To use this video replay module in your Viam machine, you need to add both the module registration and camera component to your machine configuration JSON.

### Step 1: Register the Module

Add the module to your `modules` section:

```json
{
	"modules": [
		{
			"type": "local",
			"name": "bill_video-replay",
			"executable_path": "/path/to/your/viam-video-replay/bin/video-replay"
		}
	]
}
```

**Note**: Update the `executable_path` to match the actual location where you built the module binary.

### Step 2: Add the Camera Component

Add a camera component to your `components` section that uses the video replay module:

#### Local Mode Example

```json
{
	"components": [
		{
			"name": "replayCamera-1",
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

#### Dataset Mode Example

```json
{
	"components": [
		{
			"name": "replayCamera-dataset",
			"api": "rdk:component:camera",
			"model": "bill:camera:video-replay",
			"attributes": {
				"mode": "dataset",
				"api_key": "your-viam-api-key",
				"api_key_id": "your-api-key-id",
				"organization_id": "your-org-id",
				"dataset_id": "your-dataset-id",
				"fps": 5
			}
		}
	]
}
```

### Complete Configuration Example

Here's a complete machine configuration with the video replay module:

```json
{
	"components": [
		{
			"name": "my-replay-camera",
			"api": "rdk:component:camera",
			"model": "bill:camera:video-replay",
			"attributes": {
				"mode": "local",
				"video_path": "/Users/username/videos/cooking_demo.mp4",
				"fps": 15,
				"loop_video": true
			}
		}
	],
	"modules": [
		{
			"type": "local",
			"name": "bill_video-replay",
			"executable_path": "/Users/username/viam-video-replay/bin/video-replay"
		}
	]
}
```

### Important Notes

1. **Executable Path**: The `executable_path` must point to the compiled binary. Build it first with:

    ```bash
    go build -o bin/video-replay models/module.go
    ```

2. **Video File Access**: For local mode, ensure the video file path is accessible from where the Viam server is running.

3. **API Credentials**: For dataset mode, obtain your credentials from the Viam app:

    - **API Key & API Key ID**: Settings → API Keys
    - **Organization ID**: Organization settings
    - **Dataset ID**: Create/find datasets in the Viam app

4. **Camera Name**: Use a unique name for your camera component that doesn't conflict with other components.

5. **Model Registration**: The module registers itself as `bill:camera:video-replay`, which is what you reference in the `model` field.

## Implementation Status

✅ **Local Mode**: Complete and functional
✅ **Dataset Mode**: Complete with real Viam data client integration

## Getting Viam API Credentials

To use dataset mode, you'll need:

1. **API Key & API Key ID**: Generate these in the Viam app under Settings → API Keys
2. **Organization ID**: Found in your Viam organization settings
3. **Dataset ID**: Create a dataset in the Viam app and note its ID

## Building

```bash
go build -o bin/video-replay models/module.go
```

## Prerequisites

-   Go 1.19+
-   OpenCV (for video processing)
-   pkg-config (for OpenCV integration)

On macOS:

```bash
brew install opencv pkg-config
```

## Usage

The module registers as model `bill:camera:video-replay` and implements the standard Viam camera interface, supporting:

-   `Image()`: Get current frame as JPEG
-   `Images()`: Get current frame in multiple formats
-   `Properties()`: Get camera properties
-   `Stream()`: Video streaming support

Both local video files and dataset images are processed through the same camera API, allowing seamless switching between live video replay and recorded dataset replay for testing and simulation purposes.

## Development

### Building

```bash
make video-replay
```

### Testing Local Mode

1. Place a video file in an accessible location
2. Update the configuration with the correct `video_path`
3. Set `mode` to `"local"`

### Testing Dataset Mode

1. Obtain your Viam API credentials
2. Create or identify a dataset with image data
3. Update the configuration with your credentials and dataset ID
4. Set `mode` to `"dataset"`

## API Compatibility

This module implements the standard Viam camera component interface:

-   `Image()`: Returns the current frame as JPEG
-   `Images()`: Returns the current frame (single image)
-   `Stream()`: Provides video stream access
-   `Properties()`: Returns camera properties

## Limitations

-   Dataset mode currently uses placeholder implementation with test data
-   Real Viam data client integration requires additional API authentication setup
-   Real-time streaming (`SubscribeRTP`) is not yet implemented
-   Dataset images are cached locally for performance

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

[Add your license information here]
