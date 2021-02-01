# cryptod

cryptod is an auxiliary service for cryptographic operations used by the control plane.

#### Flags

```
Usage of cryptod:
  -l string
        addr:port to bind to (default ":8081")
```

#### Routes

| Route | Method | Usage |
| ------ | ------ | ------ |
| /dnssec/newkey | GET | Generate a new DNSSEC signing key |
| /healthcheck | GET | Return ok if server is running |
