import cv2
import os
import time
from typing import Any, ClassVar, Dict, Mapping, Optional, Sequence, Tuple, List
from viam.media.video import CameraMimeType, NamedImage, ViamImage
from typing_extensions import Self
from viam.proto.common import ResponseMetadata
from viam.components.camera import Camera
from viam.logging import getLogger
from viam.proto.app.robot import ComponentConfig
from viam.proto.common import ResourceName
from viam.resource.base import ResourceBase
from viam.resource.types import Model, ModelFamily
from viam.utils import struct_to_dict
from viam.errors import NoCaptureToStoreError

LOGGER = getLogger(__name__)

class MyCamera(Camera):
    MODEL: ClassVar[Model] = Model(ModelFamily("bill", "video"), "replay")
    REQUIRED_ATTRIBUTES = ["video_path"]
    
    def __init__(self, name: str):
        super().__init__(name)
        self.last_call_time = time.time()
        self.current_frame = 0
        self.video_cap = None

    @classmethod
    def new(
        cls, config: ComponentConfig, dependencies: Mapping[ResourceName, ResourceBase]
    ) -> Self:
        camera = cls(config.name)
        camera.reconfigure(config, dependencies)
        return camera

    @classmethod
    def validate_config(cls, config: Dict[str, Any]) -> Sequence[str]:
        missing_attrs = [attr for attr in cls.REQUIRED_ATTRIBUTES if attr not in config]
        if missing_attrs:
            raise ValueError(f"Missing required attributes in Video Replay Configuration: {', '.join(missing_attrs)}")
        return []

    def reconfigure(self, config: ComponentConfig, dependencies: Mapping[ResourceName, ResourceBase]):
        LOGGER.info("Reconfiguring " + self.name)
        self.config = config
        self.dependencies = dependencies
        config_dict = struct_to_dict(config.attributes)
        self.video_path = config_dict["video_path"]

        # Initialize the video capture object
        if self.video_cap:
            self.video_cap.release()
        self.video_cap = cv2.VideoCapture(self.video_path)

    async def get_image(self, mime_type: str = "", *, extra: Optional[Dict[str, Any]] = None, timeout: Optional[float] = None, **__,) -> ViamImage:
        current_time = time.time()
        time_since_last_call = current_time - self.last_call_time

        # Get frame rate and calculate the frame to capture
        fps = self.video_cap.get(cv2.CAP_PROP_FPS)
        frame_interval = int(fps * time_since_last_call)

        # Set the frame position
        self.current_frame += frame_interval
        self.video_cap.set(cv2.CAP_PROP_POS_FRAMES, self.current_frame)

        # Read the frame
        ret, frame = self.video_cap.read()
        if not ret:
            # Restart from beginning if end is reached
            self.video_cap.set(cv2.CAP_PROP_POS_FRAMES, 0)
            self.current_frame = 0
            ret, frame = self.video_cap.read()
            if not ret:
                raise RuntimeError("Failed to capture any frame from the video.")

        self.last_call_time = current_time

        # Encode the frame as a JPEG image
        ret, jpeg = cv2.imencode('.jpg', frame)
        if not ret:
            raise RuntimeError("Failed to encode frame as JPEG.")

        image_data = jpeg.tobytes()
        return ViamImage(image_data, CameraMimeType.JPEG)

    async def get_images(self, *, timeout: Optional[float] = None, **kwargs) -> Tuple[List[NamedImage], ResponseMetadata]:
        raise NotImplementedError()

    async def get_point_cloud(
        self, *, extra: Optional[Dict[str, Any]] = None, timeout: Optional[float] = None, **kwargs
    ) -> Tuple[bytes, str]:
        raise NotImplementedError()

    async def get_properties(self, *, timeout: Optional[float] = None, **kwargs) -> 'Camera.Properties':
        return Camera.Properties(
            supports_pcd=True,
            intrinsic_parameters={
                "fx": 600.0,
                "fy": 600.0,
                "cx": 320.0,
                "cy": 240.0,
                "k1": 0.1,
                "k2": 0.01,
                "p1": 0.001,
                "p2": 0.0001,
                "k3": 0.00001
            },
            distortion_model="radial_tangential",
            distortion_coefficients=[0.1, 0.01, 0.001, 0.0001, 0.00001]
        )
