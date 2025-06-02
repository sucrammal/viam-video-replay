package main

import (
	"fmt"
	"os"
	"time"

	"gocv.io/x/gocv"
)

func main() {
	fmt.Println("=== Simple Local Video Test ===")

	// Test video path - update this to your video file
	testVideoPath := "/Users/marcuslam/Desktop/Gambit/data/out.mp4"

	// Check if video file exists
	if _, err := os.Stat(testVideoPath); os.IsNotExist(err) {
		fmt.Printf("Video file not found: %s\n", testVideoPath)
		fmt.Println("Please update testVideoPath in the script to point to a valid video file")
		return
	}

	fmt.Printf("Testing with video: %s\n", testVideoPath)

	// Open video using OpenCV
	cap, err := gocv.VideoCaptureFile(testVideoPath)
	if err != nil {
		fmt.Printf("Failed to open video: %v\n", err)
		return
	}
	defer cap.Close()

	fps := cap.Get(gocv.VideoCaptureFPS)
	frameCount := cap.Get(gocv.VideoCapturePosFrames)
	fmt.Printf("Video FPS: %.2f\n", fps)
	fmt.Printf("Current frame position: %.0f\n", frameCount)

	// Test reading frames
	fmt.Println("\n=== Reading frames from video ===")
	for i := 0; i < 5; i++ {
		frame := gocv.NewMat()
		if ok := cap.Read(&frame); !ok || frame.Empty() {
			fmt.Printf("Failed to read frame %d\n", i+1)
			frame.Close()
			break
		}

		fmt.Printf("Frame %d: %dx%d pixels\n", i+1, frame.Cols(), frame.Rows())

		// Encode frame as JPEG
		buf, err := gocv.IMEncode(".jpg", frame)
		if err != nil {
			fmt.Printf("Failed to encode frame %d: %v\n", i+1, err)
			frame.Close()
			continue
		}

		// Save frame to file
		filename := fmt.Sprintf("simple_test_frame_%d.jpg", i+1)
		if err := os.WriteFile(filename, buf.GetBytes(), 0644); err != nil {
			fmt.Printf("Failed to save frame %d: %v\n", i+1, err)
		} else {
			fmt.Printf("Saved frame to: %s\n", filename)
		}

		frame.Close()
		buf.Close()

		// Simulate frame rate timing
		time.Sleep(time.Duration(1000/fps) * time.Millisecond)
	}

	fmt.Println("\n=== Simple Local Video Test Complete ===")
	fmt.Println("This test verifies that OpenCV can read your video file.")
	fmt.Println("If this works, the local mode should work too.")
}
