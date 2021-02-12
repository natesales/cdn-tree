# Packetframe Control Plane v3

Version 3 of the Packetframe CDN control plane. This is an incomplete experiment - for the real, currently-deployed code, see https://github.com/packetframe/cdn

### Development

Set `CDNV3_DEVELOPMENT=true` to enable local development mode of the API. Dev mode disables MongoDB from expecting a replica set, and enables more verbose logging.

### Architecture

The cdnv3 control plane is separated into a few services and daemons.

- the API is written in Go with [GoFiber](https://github.com/gofiber/fiber) and is used for internal and externally facing interactions
- the Database is a MongoDB replica set over a full mesh over WireGuard that runs on all the controllers in the control plane
- the client code runs on ECAs (edge nodes) and communicates with the control plane over gRPC with protobuf
