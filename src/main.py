# wifi-sensor/src/main.py
import asyncio

from viam.module.module import Module
from viam.components.camera import Camera
from viam.resource.registry import Registry, ResourceCreatorRegistration

from myCamera import MyCamera 


async def main():
    """This function creates and starts a new module, after adding all desired resources.
    Resources must be pre-registered. For an example, see the `__init__.py` file.
    """
    Registry.register_resource_creator(Camera.SUBTYPE, MyCamera.MODEL, ResourceCreatorRegistration(MyCamera.new))
    module = Module.from_args()
    module.add_model_from_registry(Camera.SUBTYPE, MyCamera.MODEL)
    await module.start()


if __name__ == "__main__":
    asyncio.run(main())