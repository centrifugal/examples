#!/bin/bash

set -e

echo "replacing tags of structs..."

gomodifytags -file proxyproto/proxy.pb.go -field User -struct RefreshRequest -all -w -remove-options json=omitempty >/dev/null
gomodifytags -file proxyproto/proxy.pb.go -field User -struct SubscribeRequest -all -w -remove-options json=omitempty >/dev/null
gomodifytags -file proxyproto/proxy.pb.go -field User -struct PublishRequest -all -w -remove-options json=omitempty >/dev/null
gomodifytags -file proxyproto/proxy.pb.go -field User -struct RPCRequest -all -w -remove-options json=omitempty >/dev/null
