# guocedb
0-1 通用数据库 Demo，面向guoce

## 项目目录结构 (Project Directory Structure)

```plaintext
guocedb/
├── api/                      # 外部 API 定义 (External API definitions)
│   ├── protobuf/             # gRPC 服务定义 (gRPC service definitions)
│   │   └── mgmt/             # 管理 API (Management API)
│   │       └── v1/
│   │           └── management.proto # 管理服务 proto 文件 (Management service proto file)
│   └── rest/                 # RESTful API (未来扩展 / Future extension)
│       └── rest.go           # REST API 占位符 (REST API placeholder)
├── cmd/                      # 项目可执行文件入口 (Project executables entry points)
│   ├── guocedb-cli/          # 命令行客户端 (Command-line client)
│   │   └── main.go
│   └── guocedb-server/       # 数据库服务端 (Database server)
│       └── main.go
├── common/                   # 通用基础包 (Common base packages)
│   ├── config/               # 配置加载与管理 (Configuration loading and management)
│   │   └── config.go
│   ├── constants/            # 全局常量定义 (Global constants definition)
│   │   └── constants.go
│   ├── errors/               # 统一错误定义与处理 (Unified error definition and handling)
│   │   └── errors.go
│   ├── log/                  # 统一日志接口与实现 (Unified logging interface and implementation)
│   │   └── logger.go
│   └── types/                # 基础数据类型 (Basic data types)
│       ├── enum/             # 全局枚举类型 (Global enumeration types)
│       │   └── enum.go
│       └── value/            # SQL 值类型 (SQL value types)
│           └── value.go
├── compute/                  # 计算层 (Compute Layer)
│   ├── analyzer/             # SQL 查询分析器 (SQL query analyzer)
│   │   └── analyzer.go       # 分析器实现/包装 (Analyzer implementation/wrapper)
│   ├── catalog/              # 元数据目录管理 (Metadata catalog management)
│   │   ├── catalog.go        # Catalog 接口定义 (Catalog interface definition)
│   │   ├── memory/           # 内存 Catalog 实现 (In-memory Catalog implementation)
│   │   │   └── memory_catalog.go
│   │   └── persistent/       # 持久化 Catalog 实现 (Persistent Catalog implementation)
│   │       └── persistent_catalog.go
│   ├── executor/             # 查询执行引擎 (Query execution engine)
│   │   ├── engine.go         # 执行引擎接口与基础实现 (Engine interface and base implementation)
│   │   └── vector/           # 向量化执行相关 (Vectorized execution related)
│   │       └── vector.go     # 向量化占位符 (Vectorization placeholder)
│   ├── optimizer/            # 查询优化器 (Query optimizer)
│   │   └── optimizer.go      # 优化器实现/包装 (Optimizer implementation/wrapper)
│   ├── parser/               # SQL 解析器 (SQL parser)
│   │   └── parser.go         # 解析器实现/包装 (Parser implementation/wrapper)
│   ├── plan/                 # 查询计划节点 (Query plan nodes)
│   │   └── plan.go           # 计划节点定义 (Plan node definitions)
│   ├── scheduler/            # 分布式调度器 (Distributed scheduler)
│   │   └── scheduler.go      # 调度器占位符 (Scheduler placeholder)
│   └── transaction/          # 事务管理器 (Transaction manager)
│       └── manager.go        # 事务管理器接口与实现 (Txn manager interface and implementation)
├── interfaces/               # 核心抽象接口定义 (Core abstraction interface definitions)
│   └── storage.go            # 存储抽象层接口 (Storage Abstraction Layer interface)
├── internal/                 # 项目内部实现细节 (Internal implementation details)
│   ├── encoding/             # 内部数据编码/解码 (Internal data encoding/decoding)
│   │   └── encoding.go       # 内部编码工具 (Internal encoding utilities)
│   └── utils/                # 内部工具函数 (Internal utility functions)
│       └── utils.go          # 内部辅助函数 (Internal helper functions)
├── maintenance/              # 维护层 (Maintenance Layer)
│   ├── diagnostic/           # 诊断工具 (Diagnostic tools)
│   │   └── diagnostic.go     # 诊断功能实现 (Diagnostic features implementation)
│   ├── metrics/              # 性能指标收集 (Performance metrics collection)
│   │   └── metrics.go        # 指标收集实现 (Metrics collection implementation)
│   └── status/               # 状态报告 (Status reporting)
│       └── status.go         # 状态报告实现 (Status reporting implementation)
├── network/                  # 网络层 (Networking Layer)
│   ├── mesh/                 # 服务网格集成 (Service mesh integration)
│   │   └── mesh.go           # 服务网格占位符 (Service mesh placeholder)
│   └── server/               # 基础网络服务 (Basic network server)
│       └── server.go         # 网络服务基础结构 (Network server infrastructure)
├── protocol/                 # 数据库协议实现 (Database protocol implementations)
│   └── mysql/                # MySQL 协议处理 (MySQL protocol handling)
│       ├── auth.go           # MySQL 认证处理 (MySQL authentication handling)
│       ├── connection.go     # 连接处理 (Connection handling)
│       ├── handler.go        # GMS Handler 实现 (GMS Handler implementation)
│       └── server.go         # MySQL 协议服务启动器 (MySQL protocol server starter)
├── security/                 # 安全层 (Security Layer)
│   ├── audit/                # 审计日志 (Audit logging)
│   │   └── audit.go          # 审计实现 (Audit implementation)
│   ├── authn/                # 身份认证 (Authentication)
│   │   └── authn.go          # 认证实现 (Authentication implementation)
│   ├── authz/                # 访问控制 (Authorization)
│   │   └── authz.go          # 授权实现 (Authorization implementation)
│   └── crypto/               # 数据加解密 (Data encryption/decryption)
│       └── crypto.go         # 加密实现 (Encryption implementation)
├── storage/                  # 存储层 (Storage Layer)
│   ├── engines/              # 存储引擎实现 (Storage engine implementations)
│   │   ├── badger/           # Badger KV 存储引擎 (Badger KV storage engine)
│   │   │   ├── badger.go     # Badger 引擎适配器 (Badger engine adapter)
│   │   │   ├── database.go   # 数据库级别操作 (Database level operations)
│   │   │   ├── encoding.go   # Badger 特定的 K/V 编码 (Badger specific K/V encoding)
│   │   │   ├── iterator.go   # 数据迭代器 (Data iterator)
│   │   │   ├── table.go      # 表级别操作 (Table level operations)
│   │   │   └── transaction.go # 事务适配 (Transaction adaptation)
│   │   ├── kvd/              # KVD 引擎占位符 (KVD engine placeholder)
│   │   │   └── kvd.go
│   │   ├── mdd/              # MDD 引擎占位符 (MDD engine placeholder)
│   │   │   └── mdd.go
│   │   └── mdi/              # MDI 引擎占位符 (MDI engine placeholder)
│   │       └── mdi.go
│   └── sal/                  # 存储抽象层实现 (Storage Abstraction Layer implementation)
│       └── adapter.go        # 适配器实现 (Adapter implementation)
├── docs/                     # 文档 (Documentation)
│   └── architecture.md       # 架构设计文档 (Architecture design document)
├── scripts/                  # 构建、部署、测试脚本 (Build, deploy, test scripts)
│   ├── build.sh
│   └── test.sh
├── configs/                  # 配置文件示例 (Configuration examples)
│   └── config.yaml.example
├── test/                     # 测试代码 (Test code)
│   ├── integration/          # 集成测试 (Integration tests)
│   │   └── integration_test.go
│   └── unit/                 # 单元测试 (Unit tests)
│       └── unit_test.go
├── .gitignore
├── go.mod
├── go.sum
├── LICENSE                   # 选择合适的开源许可证 (Choose an appropriate open-source license)
└── README.md
```