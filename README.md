# sgortsp
Simple Go RTSP Implementation

# Usage

## convert mp4 to mjpg

- `ffmpeg -i src.mp4 dest.mjpg`

- use the util/converter.go to turn this into a 5 prefixed length mjpg: `go run converter.go src.mjpg dst.mjpg`

 - put the file into `routes/routes.go`
