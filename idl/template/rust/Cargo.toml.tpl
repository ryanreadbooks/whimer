[package]
name = "whimer-pb"
version = "0.1.0"
edition = "2021"

[features]
default = ["proto_full"]
## @@protoc_insertion_point(features)

[dependencies]
prost = "0.14"
prost-types = "0.14"
tonic = { version = "0.14", features = ["transport"] }
tonic-prost = "0.14"
