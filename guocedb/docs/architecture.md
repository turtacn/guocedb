# guocedb 架构设计文档

## 概述
guocedb 是一个现代化的 MySQL 兼容关系型数据库，采用分层架构设计，旨在解决传统数据库在云原生环境中面临的挑战。项目基于 go-mysql-server 的存储无关查询引擎 和 BadgerDB 的高性能键值存储 构建，提供企业级的数据库解决方案。

### DFX 问题全景分析

#### 设计目标（Design）
- **可扩展性**: 支持水平扩展，从单节点到分布式集群的平滑演进
- **兼容性**: 100% MySQL 线路协议兼容，支持现有 MySQL 生态系统
- **性能**: 高吞吐量、低延迟的查询处理能力
- **可靠性**: ACID 事务保证，数据一致性和持久化

#### 功能特性（Features）
- **多存储引擎**: 可插拔的存储后端架构，支持 BadgerDB 及未来扩展
- **分布式事务**: 跨节点的事务一致性保证
- **查询优化**: 基于成本的查询优化器和向量化执行引擎
- **安全机制**: 全面的身份验证、授权和数据加密

#### 体验优化（eXperience）
- **简化部署**: 单二进制文件部署，配置文件驱动
- **监控友好**: 丰富的指标暴露和健康检查端点
- **运维工具**: 完整的 CLI 工具集，支持备份、恢复、诊断等操作

## 解决方案全景

### 技术选型决策
- **主要编程语言**: Go
  - **选择理由**：类型安全、并发支持、生态兼容性
  - **替代方案**：Python（被否决，性能考虑）
- **核心依赖技术**:
  - **go-mysql-server**: 提供 MySQL 协议兼容和查询引擎基础
  - **BadgerDB**: 提供高性能键值存储后端
  - **gRPC**: 内部服务通信和管理 API
  - **Prometheus**: 指标收集和监控

### 架构分层设计
```mermaid
graph TD
    %% 系统架构总览
    subgraph IL[接口层（Interface Layer）]
        A1[SQL API（MySQL Protocol）] --> A2[管理 API（Management API）]
        A3[REST API（RESTful Interface）] --> A2
        A4[CLI 接口（Command Line）] --> A2
    end

    subgraph CL[计算层（Compute Layer）]
        B1[查询解析器（Parser）] --> B2[查询分析器（Analyzer）]
        B2 --> B3[查询优化器（Optimizer）]
        B3 --> B4[执行计划（Plan）]
        B4 --> B5[执行引擎（Executor）]
        B6[事务管理器（Transaction Manager）]
        B7[分布式调度器（Scheduler）]
    end

    subgraph SL[存储层（Storage Layer）]
        C1[存储抽象层（SAL）]
        C2[BadgerDB 引擎（BadgerDB Engine）]
        C3[KVD 引擎（KVD Engine）]
        C4[MDD 引擎（MDD Engine）]
        C5[MDI 引擎（MDI Engine）]
        C1 --> C2
        C1 --> C3
        C1 --> C4
        C1 --> C5
    end

    subgraph ML[维护层（Maintenance Layer）]
        D1[性能指标（Metrics）]
        D2[健康检查（Health Check）]
        D3[诊断工具（Diagnostics）]
        D4[状态报告（Status Reporting）]
    end

    subgraph SL2[安全层（Security Layer）]
        E1[身份认证（Authentication）]
        E2[访问控制（Authorization）]
        E3[数据加密（Encryption）]
        E4[审计日志（Audit Logging）]
    end

    subgraph NL[网络层（Network Layer）]
        F1[MySQL 协议服务器（MySQL Protocol Server）]
        F2[gRPC 服务器（gRPC Server）]
        F3[服务网格集成（Service Mesh）]
    end

    IL --> CL
    CL --> SL
    ML --> CL
    SL2 --> IL
    NL --> IL
    CL --> ML
```

### 数据流架构
```mermaid
sequenceDiagram
    participant CLI as 客户端（Client）
    participant PS as 协议服务器（Protocol Server）
    participant QE as 查询引擎（Query Engine）
    participant TM as 事务管理器（Txn Manager）
    participant SAL as 存储抽象层（SAL）
    participant SE as 存储引擎（Storage Engine）

    CLI->>+PS: 1. SQL 查询请求
    PS->>+QE: 2. 解析查询语句
    QE->>QE: 3. 语法分析与优化
    QE->>+TM: 4. 获取事务上下文
    TM-->>-QE: 5. 返回事务句柄
    QE->>+SAL: 6. 执行存储操作
    SAL->>+SE: 7. 调用具体存储引擎
    SE-->>-SAL: 8. 返回数据结果
    SAL-->>-QE: 9. 返回处理结果
    QE-->>-PS: 10. 返回查询结果
    PS-->>-CLI: 11. 响应客户端
```

## 核心组件详细设计

### 接口层（Interface Layer）
- **MySQL 协议服务器**：
  - 实现完整的 MySQL 线路协议（版本 8.0 兼容）
  - 支持认证握手、查询执行、结果集返回
  - 连接池管理和会话状态维护
- **管理 API**：
  - gRPC 服务提供数据库管理功能
  - RESTful API 支持 HTTP 客户端集成
  - CLI 工具通过 gRPC 与服务器通信

### 计算层（Compute Layer）
- **查询处理流水线**：
```mermaid
flowchart LR
    A[SQL 输入] --> B[词法分析<br/>Lexical Analysis]
    B --> C[语法分析<br/>Syntax Analysis]
    C --> D[语义分析<br/>Semantic Analysis]
    D --> E[查询优化<br/>Query Optimization]
    E --> F[执行计划<br/>Execution Plan]
    F --> G[向量化执行<br/>Vectorized Execution]
    G --> H[结果集<br/>Result Set]
```
- **事务管理**：
  - MVCC（多版本并发控制）实现
  - 分布式事务协调（2PC 协议）
  - 死锁检测和恢复机制

### 存储层（Storage Layer）
- **存储抽象层（SAL）设计**：
```mermaid
classDiagram
    class StorageInterface {
        <<interface>>
        +Get(key []byte) ([]byte, error)
        +Set(key, value []byte) error
        +Delete(key []byte) error
        +Iterator(prefix []byte) Iterator
        +NewTransaction() Transaction
    }

    class BadgerStorage {
        +db *badger.DB
        +Get(key []byte) ([]byte, error)
        +Set(key, value []byte) error
        +Delete(key []byte) error
        +Iterator(prefix []byte) Iterator
        +NewTransaction() Transaction
    }

    class KVDStorage {
        +cluster *kvd.Cluster
        +Get(key []byte) ([]byte, error)
        +Set(key, value []byte) error
        +Delete(key []byte) error
        +Iterator(prefix []byte) Iterator
        +NewTransaction() Transaction
    }

    StorageInterface <|-- BadgerStorage
    StorageInterface <|-- KVDStorage
```

## 预期效果与展望

### 性能目标
- **吞吐量**：单节点支持 50,000+ QPS
- **延迟**：P99 延迟 < 5ms
- **扩展性**：支持 100+ 节点集群
- **可用性**：99.99% 服务可用性

### 部署架构

#### 单节点部署
```mermaid
graph TB
    subgraph "guocedb 单节点"
        A[guocedb-server]
        B[BadgerDB]
        C[配置文件]
        A --> B
        A --> C
    end

    D[MySQL 客户端] --> A
    E[guocedb-cli] --> A
    F[应用程序] --> A
```

#### 集群部署
```mermaid
graph TB
    subgraph "负载均衡层"
        LB[负载均衡器<br/>Load Balancer]
    end

    subgraph "guocedb 集群"
        N1[节点 1<br/>guocedb-server]
        N2[节点 2<br/>guocedb-server]
        N3[节点 3<br/>guocedb-server]
    end

    subgraph "存储层"
        S1[存储节点 1]
        S2[存储节点 2]
        S3[存储节点 3]
    end

    subgraph "监控层"
        M1[Prometheus]
        M2[Grafana]
        M3[AlertManager]
    end

    LB --> N1
    LB --> N2
    LB --> N3

    N1 --> S1
    N2 --> S2
    N3 --> S3

    N1 --> M1
    N2 --> M1
    N3 --> M1
    M1 --> M2
    M1 --> M3
```
