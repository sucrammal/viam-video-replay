# viam-video-replay

## Description

This project demonstrates the implementation of a custom Viam Camera component that replays video files. The camera component reads a video file and simulates live camera behavior by returning frames as images upon request. The component utilizes OpenCV to handle video capture and encoding, allowing it to deliver frames in real-time according to the frame rate of the original video. This approach is ideal for scenarios where testing or development requires consistent, repeatable video input without relying on a physical camera.

## Viam Module

A module is a package with streamlined deployment to a Viam server. Modules can run alongside viam-server as separate processes, communicating with viam-server over UNIX sockets. A Viam Module can deploy and manage components such as a Viam Sensor.

## Viam Camera

The Viam Camera component in this project reads frames from a specified video file and returns them as images when the get_image method is called. The camera simulates live video feed behavior, ensuring consistent frame delivery based on the original video’s frame rate. This setup is particularly useful for testing machine vision applications where consistent, predictable video input is necessary.

## Configuration

To use this custom Viam Camera component, the following configuration is required:

Generalized Attribute Guide
```json
{
  "video_path": "/path/to/video_file"
}
```

Generic Example
```json
{
  "video_path": "/tmp/sample_video.mp4"
}
```

This configuration specifies the path to the video file that the camera will replay. Ensure the video file is accessible from the path provided.


## Key Methods

- get_image: Returns the current frame from the video as a JPEG-encoded image. If the video reaches its end, it restarts from the beginning.
- validate_config: Validates that the necessary attributes (such as video_path) are present in the configuration.
- reconfigure: Reconfigures the camera with new settings and initializes the video capture process.

This component offers a versatile solution for scenarios where a simulated video feed is necessary, such as in testing environments or controlled experiments.
