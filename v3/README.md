# Packetframe Control Plane v3

Version 3 of the Packetframe CDN control plane. This is an incomplete experiment - for the real, currently-deployed code, see https://github.com/packetframe/cdn

### Development

Set `CDNV3_DEVELOPMENT=true` to enable local development mode of the API. Dev mode disables MongoDB from expecting a replica set, and enables more verbose logging.

### Architecture

The cdnv3 control plane is separated into a few services and daemons.

- the Public API is written in Python with [FastAPI](https://github.com/tiangolo/fastapi) + [Pydantic](https://github.com/samuelcolvin/pydantic) for user facing API operations
- the Internal API is written in Go with [GoFiber](https://github.com/gofiber/fiber) and is used for internal ops such as managing cache nodes
- the Database is a MongoDB replica set over a full mesh over WireGuard that runs on all the controllers in the control plane
