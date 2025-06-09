module github.com/turtacn/guocedb

go 1.23.0

toolchain go1.23.6

require (
	github.com/dgraph-io/badger/v4 v4.7.0
	github.com/dolthub/go-mysql-server v0.18.0
	github.com/dolthub/vitess v0.0.0-20240228192915-d55088cef56a
	github.com/natefinch/lumberjack v2.0.0+incompatible
	github.com/stretchr/testify v1.10.0
	go.uber.org/zap v1.27.0
	google.golang.org/protobuf v1.36.6
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgraph-io/ristretto/v2 v2.2.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/flatbuffers v25.2.10+incompatible // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/tools v0.22.0 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
	google.golang.org/grpc v1.53.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/src-d/go-errors.v1 v1.0.0 // indirect
)

// 兼容 Go 1.20，解决 golang.org 访问失败的问题
replace (
	github.com/golang/protobuf => github.com/golang/protobuf v1.5.4 // 兼容 protobuf 旧接口
	golang.org/x/exp => github.com/golang/exp v0.0.0-20230522175609-2e198f4a06a1
	golang.org/x/net => github.com/golang/net v0.17.0 // 网络支持
	golang.org/x/sync => github.com/golang/sync v0.1.0 // 兼容 Go 1.20
	golang.org/x/sys => github.com/golang/sys v0.9.0 // 广泛用于 1.20
	golang.org/x/tools => github.com/golang/tools v0.10.0 // 保守稳定
	google.golang.org/grpc => github.com/grpc/grpc-go v1.54.0 // 兼容性强版本

	google.golang.org/protobuf => github.com/protocolbuffers/protobuf-go v1.30.0
)
