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
