# Protocol Buffers for Go with Gadgets

[![Build Status](https://drone.io/github.com/gogo/protobuf/status.png)](https://drone.io/github.com/gogo/protobuf/latest)

### Getting Started (Give me the speed I don't care about the rest)

Install the protoc-gen-gofast binary

    go get github.com/gogo/protobuf/protoc-gen-gofast

Use it to generate faster marshaling and unmarshaling go code for you protocol buffers.

    protoc -gofast_out=. myproto.proto

### Getting started (I have heard about fields without pointers and more code generation)

Other binaries are also included:
    
    protoc-gen-gogofast (same as gofast, but imports gogoprotobuf)
    protoc-gen-gogofaster (same as gogofast, without XXX_unrecognized, less pointer fields)
    protoc-gen-gogoslick (same as gogofaster, but with generated string, gostring and equal methods)

### Getting started (I want more customization power over fields, speed, other serialization formats and tests, etc.) 

Please visit the [homepage](http://gogo.github.io) for more documentation.

### Installation

To install it, you must first have Go (at least version 1.2.2) installed (see [http://golang.org/doc/install](http://golang.org/doc/install)).

Next, install the standard protocol buffer implementation from [https://github.com/google/protobuf](https://github.com/google/protobuf); you must be running version 2.3, 2.4.1, 2.5, 2.6 or 2.6.1.

Finally run:

    go get github.com/gogo/protobuf/proto
    go get github.com/gogo/protobuf/protoc-gen-gogo
    go get github.com/gogo/protobuf/gogoproto

### Proto3

Proto3 is supported, but since a specification has not been released yet, this support is limited.


