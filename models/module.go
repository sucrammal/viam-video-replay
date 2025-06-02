package models

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.viam.com/rdk/components/camera"
	"go.viam.com/rdk/components/camera/rtppassthrough"
	"go.viam.com/rdk/gostream"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/pointcloud"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/rimage/transform"
	"go.viam.com/utils/rpc"
	"gocv.io/x/gocv"

	// Viam app client for data access
	"go.viam.com/rdk/app"
)

// Our camera model
var Video = resource.NewModel("bill", "camera", "video-replay")

func init() {
	fmt.Println("[video-replay] init() called")
	resource.RegisterComponent(
		camera.API,
		Video,
		resource.Registration[camera.Camera, *Config]{
			Constructor: newVideoReplayVideo,
		},
	)
	fmt.Println("[video-replay] Registering", Video.String())
}

// Config holds the JSON attributes
type Config struct {
	// Local mode fields
	VideoPath *string `json:"video_path,omitempty"`
	FPS       *int    `json:"fps,omitempty"`
	LoopVideo *bool   `json:"loop_video,omitempty"`
	Height    *int    `json:"height,omitempty"`
	Width     *int    `json:"width,omitempty"`

	// Core dataset mode fields (simplified)
	Mode           *string `json:"mode,omitempty"`            // "local" or "dataset"
	APIKey         *string `json:"api_key,omitempty"`         // Viam API key
	APIKeyID       *string `json:"api_key_id,omitempty"`      // Viam API key ID
	OrganizationID *string `json:"organization_id,omitempty"` // Organization ID
	DatasetID      *string `json:"dataset_id,omitempty"`      // Dataset ID to replay from
}

// Validate ensures required fields are set based on mode
func (c *Config) Validate(path string) ([]string, []string, error) {
	// Determine mode (default to local if not specified)
	mode := "local"
	if c.Mode != nil {
		mode = *c.Mode
	}

	switch mode {
	case "local":
		if c.VideoPath == nil || *c.VideoPath == "" {
			return nil, nil, fmt.Errorf("video_path is required for local mode video replay camera")
		}
	case "dataset":
		if c.APIKey == nil || *c.APIKey == "" {
			return nil, nil, fmt.Errorf("api_key is required for dataset mode")
		}
		if c.APIKeyID == nil || *c.APIKeyID == "" {
			return nil, nil, fmt.Errorf("api_key_id is required for dataset mode")
		}
		if c.OrganizationID == nil || *c.OrganizationID == "" {
			return nil, nil, fmt.Errorf("organization_id is required for dataset mode")
		}
		if c.DatasetID == nil || *c.DatasetID == "" {
			return nil, nil, fmt.Errorf("dataset_id is required for dataset mode")
		}
	default:
		return nil, nil, fmt.Errorf("invalid mode '%s': must be 'local' or 'dataset'", mode)
	}

	return nil, nil, nil
}

// DatasetImage represents a cached image from a dataset
type DatasetImage struct {
	Data      []byte
	Timestamp time.Time
	Filename  string
}

// DatasetReplay handles fetching and replaying images from Viam datasets
type DatasetReplay struct {
	logger         logging.Logger
	apiKey         string
	apiKeyID       string
	organizationID string
	datasetID      string

	images       []DatasetImage
	currentIndex int
	mu           sync.RWMutex
}

// videoReplayVideo implements camera.Camera + resource.Reconfigurable
type videoReplayVideo struct {
	name       resource.Name
	logger     logging.Logger
	cfg        *Config
	cancelFunc context.CancelFunc // For main resource context

	// We store the camera's "lifetime" context in mainCtx.
	// This is created in newVideoReplayVideo and ends when the resource is closed.
	mainCtx context.Context

	// The loop context gets created (and canceled) each time we open a video,
	// e.g. in Reconfigure or initially.
	loopCtx    context.Context
	loopCancel context.CancelFunc

	// OpenCV capture (for local video mode)
	videoCapture *gocv.VideoCapture
	fps          float64

	// Current frame updated by background loop
	frameMutex       sync.RWMutex
	currentFrame     gocv.Mat
	currentFrameTime time.Time

	// Dataset replay fields
	mode          string
	datasetReplay *DatasetReplay
}

// newVideoReplayVideo is called once when camera is created
func newVideoReplayVideo(
	ctx context.Context,
	deps resource.Dependencies,
	rawConf resource.Config,
	logger logging.Logger,
) (camera.Camera, error) {
	logger.Infof("[newVideoReplayVideo] Called")

	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	// Determine mode (default to local if not specified)
	mode := "local"
	if conf.Mode != nil {
		mode = *conf.Mode
	}

	// Create a context for the camera's lifetime
	ctx, cancelFunc := context.WithCancel(context.Background())

	cam := &videoReplayVideo{
		name:       rawConf.ResourceName(),
		logger:     logger,
		cfg:        conf,
		cancelFunc: cancelFunc,
		mainCtx:    ctx,
		mode:       mode,

		// We'll open the video + start the loop below
		currentFrame: gocv.NewMat(),
	}

	// Initialize based on mode
	switch mode {
	case "local":
		if err := cam.openAndStartLoop(*conf.VideoPath); err != nil {
			// If we fail to open, do cleanup
			cancelFunc()
			return nil, fmt.Errorf("failed to open camera at creation: %w", err)
		}
	case "dataset":
		datasetReplay, err := newDatasetReplay(conf, logger)
		if err != nil {
			cancelFunc()
			return nil, fmt.Errorf("failed to initialize dataset replay: %w", err)
		}
		cam.datasetReplay = datasetReplay

		if err := cam.initDatasetReplay(); err != nil {
			cancelFunc()
			return nil, fmt.Errorf("failed to initialize dataset replay: %w", err)
		}
	}

	logger.Warnf("Camera %q: real-time streaming not implemented; SubscribeRTP calls will fail", cam.name)
	logger.Infof("[newVideoReplayVideo] Camera constructed successfully: %q", cam.name)
	return cam, nil
}

// openAndStartLoop is used by constructor + Reconfigure
func (s *videoReplayVideo) openAndStartLoop(videoPath string) error {
	// If a loop is running, cancel it
	if s.loopCancel != nil {
		s.loopCancel()
		s.loopCancel = nil
	}
	// Close existing capture if any
	if s.videoCapture != nil {
		s.videoCapture.Close()
		s.videoCapture = nil
	}

	// Open new file
	cap, err := gocv.VideoCaptureFile(videoPath)
	if err != nil {
		return fmt.Errorf("failed to open %q: %w", videoPath, err)
	}
	fps := cap.Get(gocv.VideoCaptureFPS)
	if fps <= 0 {
		fps = 30
	}

	// Read initial frame
	firstFrame := gocv.NewMat()
	if ok := cap.Read(&firstFrame); !ok || firstFrame.Empty() {
		firstFrame.Close()
		cap.Close()
		return fmt.Errorf("failed to read initial frame from %q", videoPath)
	}

	// Store in struct
	s.frameMutex.Lock()
	if !s.currentFrame.Empty() {
		s.currentFrame.Close()
	}
	s.currentFrame = firstFrame
	s.currentFrameTime = time.Now()
	s.frameMutex.Unlock()

	s.videoCapture = cap
	s.fps = fps

	// Start background loop with a fresh context from mainCtx
	loopCtx, loopCancel := context.WithCancel(s.mainCtx)
	s.loopCtx = loopCtx
	s.loopCancel = loopCancel

	s.logger.Infof("[openAndStartLoop] Opened %q (FPS=%.2f), starting loop...", videoPath, fps)
	go s.frameUpdateLoop(loopCtx, fps)

	return nil
}

// frameUpdateLoop updates currentFrame ~fps
func (s *videoReplayVideo) frameUpdateLoop(ctx context.Context, fps float64) {
	s.logger.Infof("[frameUpdateLoop] Starting for camera %q at FPS=%.2f", s.name, fps)
	ticker := time.NewTicker(time.Second / time.Duration(fps))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Infof("[frameUpdateLoop] canceled for %q", s.name)
			return
		case <-ticker.C:
			newFrame := gocv.NewMat()
			if ok := s.videoCapture.Read(&newFrame); !ok || newFrame.Empty() {
				// Check if looping is enabled
				shouldLoop := true // default to true for backward compatibility
				if s.cfg.LoopVideo != nil {
					shouldLoop = *s.cfg.LoopVideo
				}
				
				if shouldLoop {
					s.logger.Infof("[frameUpdateLoop] End of file => reset to 0 for %q (loop enabled)", s.name)
					s.videoCapture.Set(gocv.VideoCapturePosFrames, 0)
					newFrame.Close()
					newFrame = gocv.NewMat()
					s.videoCapture.Read(&newFrame)
				} else {
					s.logger.Infof("[frameUpdateLoop] End of file => stopping playback for %q (loop disabled)", s.name)
					newFrame.Close()
					// Keep the last frame frozen instead of stopping completely
					return
				}
			}
			now := time.Now()
			s.frameMutex.Lock()
			s.currentFrame.Close()
			s.currentFrame = newFrame
			s.currentFrameTime = now
			s.frameMutex.Unlock()
		}
	}
}

// Reconfigure changes the video by always stopping and restarting the loop.
func (s *videoReplayVideo) Reconfigure(
	ctx context.Context,
	deps resource.Dependencies,
	rawConf resource.Config,
) error {
	s.logger.Infof("[Reconfigure] Called for %q", s.name)

	newConf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return err
	}

	// Determine new mode
	newMode := "local"
	if newConf.Mode != nil {
		newMode = *newConf.Mode
	}

	// If mode changed, we need to reinitialize everything
	modeChanged := s.mode != newMode

	// Always stop the running loop first
	if s.loopCancel != nil {
		s.loopCancel()
		s.loopCancel = nil
	}

	// Clean up based on current mode
	if s.mode == "local" && s.videoCapture != nil {
		s.videoCapture.Close()
		s.videoCapture = nil
	}

	// Update configuration and mode
	s.cfg = newConf
	s.mode = newMode

	// Initialize based on new mode
	switch newMode {
	case "local":
		if err := s.openAndStartLoop(*newConf.VideoPath); err != nil {
			return fmt.Errorf("reconfigure local mode: %w", err)
		}
	case "dataset":
		if modeChanged || s.datasetReplay == nil {
			datasetReplay, err := newDatasetReplay(newConf, s.logger)
			if err != nil {
				return fmt.Errorf("reconfigure dataset mode: failed to initialize dataset replay: %w", err)
			}
			s.datasetReplay = datasetReplay
		}

		if err := s.initDatasetReplay(); err != nil {
			return fmt.Errorf("reconfigure dataset mode: failed to initialize dataset replay: %w", err)
		}
	}

	s.logger.Infof("[Reconfigure] Successfully reconfigured to mode '%s'", newMode)
	return nil
}

// Name returns the camera name
func (s *videoReplayVideo) Name() resource.Name {
	return s.name
}

// Image returns the latest frame as JPEG
func (s *videoReplayVideo) Image(
	ctx context.Context,
	mimeType string,
	extra map[string]interface{},
) ([]byte, camera.ImageMetadata, error) {
	s.logger.Infof("[Image] Called for camera %q, mimeType=%q", s.name, mimeType)

	s.frameMutex.RLock()
	defer s.frameMutex.RUnlock()

	if s.currentFrame.Empty() {
		return nil, camera.ImageMetadata{}, fmt.Errorf("no frame available")
	}
	buf, err := gocv.IMEncode(".jpg", s.currentFrame)
	if err != nil {
		return nil, camera.ImageMetadata{}, fmt.Errorf("encode fail: %w", err)
	}

	meta := camera.ImageMetadata{
		MimeType: "image/jpeg",
	}
	return buf.GetBytes(), meta, nil
}

// Images returns one NamedImage
func (s *videoReplayVideo) Images(ctx context.Context) ([]camera.NamedImage, resource.ResponseMetadata, error) {
	b, _, err := s.Image(ctx, "image/jpeg", nil)
	if err != nil {
		return nil, resource.ResponseMetadata{}, err
	}
	mat, err := gocv.IMDecode(b, gocv.IMReadColor)
	if err != nil {
		return nil, resource.ResponseMetadata{}, fmt.Errorf("imdecode fail: %w", err)
	}
	defer mat.Close()

	goImg, err := mat.ToImage()
	if err != nil {
		return nil, resource.ResponseMetadata{}, fmt.Errorf("mat.ToImage fail: %w", err)
	}

	named := []camera.NamedImage{{
		Image:      goImg,
		SourceName: "color",
	}}
	return named, resource.ResponseMetadata{}, nil
}

// NextPointCloud is not supported
func (s *videoReplayVideo) NextPointCloud(ctx context.Context) (pointcloud.PointCloud, error) {
	return nil, fmt.Errorf("pointcloud not supported")
}

// Properties returns minimal info
func (s *videoReplayVideo) Properties(ctx context.Context) (camera.Properties, error) {
	return camera.Properties{
		SupportsPCD: false,
		ImageType:   camera.ColorStream,
		IntrinsicParams: &transform.PinholeCameraIntrinsics{
			Width:  640,
			Height: 480,
		},
		MimeTypes: []string{"image/jpeg"},
	}, nil
}

// DoCommand is not supported
func (s *videoReplayVideo) DoCommand(
	ctx context.Context,
	cmd map[string]interface{},
) (map[string]interface{}, error) {
	return nil, fmt.Errorf("do command not supported")
}

// Close cleans up on resource removal
func (s *videoReplayVideo) Close(ctx context.Context) error {
	s.logger.Infof("[Close] Called for %q", s.name)
	// stop loop
	if s.loopCancel != nil {
		s.loopCancel()
	}
	// close capture
	if s.videoCapture != nil {
		s.videoCapture.Close()
	}
	// free last frame
	s.frameMutex.Lock()
	s.currentFrame.Close()
	s.frameMutex.Unlock()
	// end main resource context
	s.cancelFunc()
	return nil
}

// The streaming methods
func (s *videoReplayVideo) SubscribeRTP(
	ctx context.Context,
	bufferSize int,
	packetsCB rtppassthrough.PacketCallback,
) (rtppassthrough.Subscription, error) {
	return rtppassthrough.Subscription{}, fmt.Errorf("streaming not implemented")
}

func (s *videoReplayVideo) Unsubscribe(
	ctx context.Context,
	id rtppassthrough.SubscriptionID,
) error {
	return fmt.Errorf("streaming not implemented")
}

// If you need a remote client approach:
func (s *videoReplayVideo) NewClientFromConn(
	ctx context.Context,
	conn rpc.ClientConn,
	remoteName string,
	name resource.Name,
	logger logging.Logger,
) (camera.Camera, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *videoReplayVideo) Stream(
	ctx context.Context,
	errHandlers ...gostream.ErrorHandler,
) (gostream.VideoStream, error) {
	return nil, fmt.Errorf("not implemented")
}

// newDatasetReplay creates a new DatasetReplay instance
func newDatasetReplay(conf *Config, logger logging.Logger) (*DatasetReplay, error) {
	dr := &DatasetReplay{
		logger:         logger,
		apiKey:         *conf.APIKey,
		apiKeyID:       *conf.APIKeyID,
		organizationID: *conf.OrganizationID,
		datasetID:      *conf.DatasetID,
	}

	return dr, nil
}

// initDatasetReplay initializes the dataset replay by fetching images
func (s *videoReplayVideo) initDatasetReplay() error {
	if err := s.datasetReplay.fetchImages(); err != nil {
		return fmt.Errorf("failed to fetch images from dataset: %w", err)
	}

	// Start the dataset replay loop
	loopCtx, loopCancel := context.WithCancel(s.mainCtx)
	s.loopCtx = loopCtx
	s.loopCancel = loopCancel

	// Calculate FPS based on dataset or use default
	fps := 30.0
	if s.cfg.FPS != nil {
		fps = float64(*s.cfg.FPS)
	}
	s.fps = fps

	s.logger.Infof("[initDatasetReplay] Starting dataset replay loop with %d images at FPS=%.2f",
		len(s.datasetReplay.images), fps)
	go s.datasetReplayLoop(loopCtx, fps)

	return nil
}

// datasetReplayLoop cycles through dataset images at the specified FPS
func (s *videoReplayVideo) datasetReplayLoop(ctx context.Context, fps float64) {
	s.logger.Infof("[datasetReplayLoop] Starting for camera %q at FPS=%.2f", s.name, fps)
	ticker := time.NewTicker(time.Second / time.Duration(fps))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Infof("[datasetReplayLoop] canceled for %q", s.name)
			return
		case <-ticker.C:
			if err := s.datasetReplay.loadNextFrame(s); err != nil {
				s.logger.Errorf("[datasetReplayLoop] Failed to load next frame: %v", err)
			}
		}
	}
}

// fetchImages retrieves images from the Viam dataset
func (dr *DatasetReplay) fetchImages() error {
	dr.logger.Info("Fetching images from Viam dataset...")

	ctx := context.Background()

	// Create Viam app client with API key
	viamClient, err := app.CreateViamClientWithAPIKey(ctx, app.Options{}, dr.apiKey, dr.apiKeyID, dr.logger)
	if err != nil {
		return fmt.Errorf("failed to create Viam client: %v", err)
	}
	defer viamClient.Close()

	dataClient := viamClient.DataClient()

	// Create filter for dataset
	filter := &app.Filter{
		DatasetID: dr.datasetID,
	}

	// Fetch binary data from dataset
	resp, err := dataClient.BinaryDataByFilter(ctx, true, &app.DataByFilterOptions{
		Filter: filter,
		Limit:  100, // Start with reasonable limit
	})
	if err != nil {
		return fmt.Errorf("failed to fetch dataset images: %v", err)
	}

	dr.logger.Infof("Found %d images in dataset", len(resp.BinaryData))

	// Convert binary data to DatasetImage objects
	dr.images = make([]DatasetImage, 0, len(resp.BinaryData))
	for i, binaryData := range resp.BinaryData {
		if binaryData.Binary == nil {
			dr.logger.Warnf("Skipping image %d with no binary data", i)
			continue
		}

		// Extract timestamp from metadata
		var timestamp time.Time
		if binaryData.Metadata != nil {
			timestamp = binaryData.Metadata.TimeRequested
		} else {
			timestamp = time.Now()
		}

		// Create filename from metadata or generate one
		filename := fmt.Sprintf("dataset_image_%d.jpg", i)
		if binaryData.Metadata != nil && binaryData.Metadata.FileName != "" {
			filename = binaryData.Metadata.FileName
		}

		datasetImage := DatasetImage{
			Data:      binaryData.Binary,
			Timestamp: timestamp,
			Filename:  filename,
		}

		dr.images = append(dr.images, datasetImage)
	}

	dr.logger.Infof("Successfully loaded %d images from dataset", len(dr.images))
	return nil
}

// loadNextFrame loads the next frame from the dataset into the camera
func (dr *DatasetReplay) loadNextFrame(cam *videoReplayVideo) error {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	if len(dr.images) == 0 {
		return fmt.Errorf("no images available")
	}

	// Get current image
	currentImage := dr.images[dr.currentIndex]

	// Convert image data to gocv.Mat
	// Decode the image bytes directly (JPEG/PNG/etc) into a proper image matrix
	newFrame, err := gocv.IMDecode(currentImage.Data, gocv.IMReadColor)
	if err != nil || newFrame.Empty() {
		// If decoding fails (e.g., with test data), create a colored placeholder frame
		dr.logger.Warnf("Failed to decode image data for %s, using placeholder: %v", currentImage.Filename, err)
		newFrame = gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)

		// Fill with a color based on frame index for visual distinction
		color := gocv.NewScalar(
			float64((dr.currentIndex*50)%255),  // Blue
			float64((dr.currentIndex*100)%255), // Green
			float64((dr.currentIndex*150)%255), // Red
			0,                                  // Alpha
		)
		newFrame.SetTo(color)
	}

	// Update camera's current frame
	cam.frameMutex.Lock()
	if !cam.currentFrame.Empty() {
		cam.currentFrame.Close()
	}
	cam.currentFrame = newFrame
	cam.currentFrameTime = currentImage.Timestamp
	cam.frameMutex.Unlock()

	// Move to next frame (loop back to start if at end)
	dr.currentIndex = (dr.currentIndex + 1) % len(dr.images)

	dr.logger.Debugf("Loaded frame %d: %s", dr.currentIndex, currentImage.Filename)
	return nil
}
