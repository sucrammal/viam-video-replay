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
	VideoPath string `json:"video_path"`
}

// Validate ensures VideoPath is set
func (c *Config) Validate(path string) ([]string, error) {
	if c.VideoPath == "" {
		return nil, fmt.Errorf("video_path is required for video replay camera")
	}
	return nil, nil
}

// videoReplayVideo implements camera.Camera + resource.Reconfigurable
type videoReplayVideo struct {
	name       resource.Name
	logger     logging.Logger
	cfg        *Config
	cancelFunc context.CancelFunc // For main resource context

	// We store the camera’s "lifetime" context in mainCtx.
	// This is created in newVideoReplayVideo and ends when the resource is closed.
	mainCtx context.Context

	// The loop context gets created (and canceled) each time we open a video,
	// e.g. in Reconfigure or initially.
	loopCtx    context.Context
	loopCancel context.CancelFunc

	// OpenCV capture
	videoCapture *gocv.VideoCapture
	fps          float64

	// Current frame updated by background loop
	frameMutex       sync.RWMutex
	currentFrame     gocv.Mat
	currentFrameTime time.Time
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

	// Create a context for the camera’s lifetime

	ctx, cancelFunc := context.WithCancel(context.Background())

	cam := &videoReplayVideo{
		name:       rawConf.ResourceName(),
		logger:     logger,
		cfg:        conf,
		cancelFunc: cancelFunc,
		mainCtx:    ctx,

		// We'll open the video + start the loop below
		currentFrame: gocv.NewMat(),
	}

	if err := cam.openAndStartLoop(conf.VideoPath); err != nil {
		// If we fail to open, do cleanup
		cancelFunc()
		return nil, fmt.Errorf("failed to open camera at creation: %w", err)
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
				s.logger.Infof("[frameUpdateLoop] End of file or invalid => reset to 0 for %q", s.name)
				s.videoCapture.Set(gocv.VideoCapturePosFrames, 0)
				newFrame.Close()
				newFrame = gocv.NewMat()
				s.videoCapture.Read(&newFrame)
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

	// Always stop the running loop and open a new one
	if err := s.openAndStartLoop(newConf.VideoPath); err != nil {
		return fmt.Errorf("reconfigure: %w", err)
	}
	s.cfg = newConf

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
