{
"components": [
{
"name": "core-dev",
"namespace": "rdk",
"type": "sensor",
"model": "gambit-robotics:chef:core",
"attributes": {
"rgb": "overhead-rgb",
"hc-burners": [
{
"rgb-center": {
"x": 191,
"y": 211.5,
"radius": 100.25
}
},
{
"rgb-center": {
"radius": 91,
"x": 443,
"y": 205
}
}
],
"heat": "overhead-heat-sensor",
"heat_height": 24,
"heat_width": 32
}
},
{
"name": "overlay-dev",
"api": "rdk:component:camera",
"model": "gambit-robotics:chef:overlay-full",
"attributes": {
"core": "core-dev"
}
},
{
"name": "burner-0-dev",
"namespace": "rdk",
"type": "camera",
"model": "gambit-robotics:chef:burner",
"attributes": {
"burner": 0,
"core": "core-dev",
"vision": "burner-classifier-vision"
},
"service_configs": [
{
"type": "data_manager",
"attributes": {
"capture_methods": [
{
"method": "ReadImage",
"capture_frequency_hz": 1,
"disabled": false,
"additional_params": {
"mime_type": "image/jpeg"
}
}
]
}
}
]
},
{
"name": "burner-1-dev",
"namespace": "rdk",
"type": "camera",
"model": "gambit-robotics:chef:burner",
"attributes": {
"core": "core-dev",
"vision": "burner-classifier-vision",
"burner": 1
},
"service_configs": [
{
"type": "data_manager",
"attributes": {
"capture_methods": [
{
"method": "ReadImage",
"capture_frequency_hz": 1,
"disabled": false,
"additional_params": {
"mime_type": "image/jpeg"
}
}
]
}
}
]
},
{
"name": "replayCamera-1",
"api": "rdk:component:camera",
"model": "bill:camera:video-replay",
"attributes": {
"video_path": "/Users/marcuslam/Desktop/Gambit/testVideo/test.MOV"
}
}
],
"modules": [
{
"type": "local",
"name": "bill_video-replay",
"executable_path": "/Users/marcuslam/Desktop/Gambit/viam-video-replay/video-replay"
}
],
"fragments": [
"45ab316c-4086-4858-bd77-ab01f415fed4"
],
"fragment_mods": [
{
"fragment_id": "45ab316c-4086-4858-bd77-ab01f415fed4",
"mods": [
{
"$set": {
"modules.local-module-1.executable_path": "/Users/marcuslam/Desktop/Gambit/chef/bin/chefmodule"
}
},
{
"$set": {
"remotes.remoteMachine.auth.entity": "5a066cdb-4671-4112-8280-2658090f8c2d",
"remotes.remoteMachine.address": "ur15-main.uk9vujugl1.viam.cloud",
"remotes.remoteMachine.auth.credentials.payload": "uvqao8m8wolmv43qtodl60n2ofynvrex"
}
}
]
}
]
}
