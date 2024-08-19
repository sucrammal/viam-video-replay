import cv2
import os
import time

# Path to the .mov video file
video_path = 'path/to/your/video.mov'
output_folder = 'output_frames'
interval = 1  # Time interval in seconds

# Create output directory if it doesn't exist
if not os.path.exists(output_folder):
    os.makedirs(output_folder)

# Capture the video from the file
cap = cv2.VideoCapture(video_path)

# Frame rate of the video
fps = cap.get(cv2.CAP_PROP_FPS)

# Calculate frame interval
frame_interval = int(fps * interval)

frame_count = 0

while cap.isOpened():
    # Set the frame position
    cap.set(cv2.CAP_PROP_POS_FRAMES, frame_count * frame_interval)

    # Read the frame
    ret, frame = cap.read()
    if not ret:
        break

    # Save the frame as a .jpeg image
    output_path = os.path.join(output_folder, f'frame_{frame_count}.jpg')
    cv2.imwrite(output_path, frame)

    print(f'Saved {output_path}')

    frame_count += 1
    time.sleep(interval)

cap.release()
print('Processing complete.')
