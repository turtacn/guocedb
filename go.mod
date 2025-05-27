module github.com/turtacn/guocedb

go 1.20

require (
	github.com/dolthub/go-mysql-server v0.18.0
	github.com/dgraph-io/badger/v4 v4.3.0
)

// 兼容 Go 1.20，解决 golang.org 访问失败的问题
replace (
    golang.org/x/sync => github.com/golang/sync v0.1.0            // 兼容 Go 1.20
    golang.org/x/tools => github.com/golang/tools v0.10.0         // 保守稳定
    golang.org/x/sys => github.com/golang/sys v0.9.0              // 广泛用于 1.20
    golang.org/x/net => github.com/golang/net v0.17.0             // 网络支持
    golang.org/x/exp => github.com/golang/exp v0.0.0-20230522175609-2e198f4a06a1

    google.golang.org/protobuf => github.com/protocolbuffers/protobuf-go v1.30.0
    github.com/golang/protobuf => github.com/golang/protobuf v1.5.4             // 兼容 protobuf 旧接口
    google.golang.org/grpc => github.com/grpc/grpc-go v1.54.0                   // 兼容性强版本
)
