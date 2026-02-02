# pkg/k8s - 代码分解完成报告

## 概述
成功将原始的 `client.go` (1233 行，28 个函数) 分解为 4 个专有模块，保持包名为 `k8s`，并验证编译通过。

## 文件统计

| 文件 | 行数 | 函数数 | 类型数 | 说明 |
|------|------|---------|---------|------|
| init.go | 104 | 2 | 0 | 初始化和全局变量定义 |
| websocket.go | 242 | 8 | 2 | WebSocket 相关功能 |
| watch.go | 599 | 11 | 0 | Watch 资源相关功能 |
| client.go | 317 | 7 | 2 | 核心 Client 功能 |
| **总计** | **1262** | **28** | **4** | **+ 原始编译验证** |

---

## 1. init.go - 初始化模块 (104 行)

**用途**: Kubernetes 集群初始化和全局客户端配置

### 全局变量
```go
var (
    Config        *rest.Config
    Clientset     kubernetes.Interface
    Dc            *discovery.DiscoveryClient
    Resources     []*restmapper.APIGroupResources
    Mapper        meta.RESTMapper
    DynamicClient *dynamic.DynamicClient
)
```

### 导出函数
| 函数 | 说明 |
|------|------|
| `InitTestCluster()` | 测试环境初始化（使用 fake client） |
| `Init()` | 生产环境初始化（支持多个配置来源） |

### 导入
```go
- k8s.io/apimachinery/pkg/api/meta
- k8s.io/client-go/discovery
- k8s.io/client-go/dynamic
- k8s.io/client-go/kubernetes
- k8s.io/client-go/kubernetes/fake
- k8s.io/client-go/rest
- k8s.io/client-go/restmapper
- k8s.io/client-go/tools/clientcmd
- k8s.io/client-go/util/homedir
```

---

## 2. websocket.go - WebSocket 模块 (242 行)

**用途**: WebSocket 与 Kubernetes Pod 执行命令的交互

### 类型定义
```go
type WebSocketIO struct {}       // WebSocket I/O 处理器
type TerminalMessage struct {}   // 终端消息协议
```

### 导出函数
| 函数 | 说明 |
|------|------|
| `NewWebSocketIO(conn)` | 创建 WebSocket 处理器并启动循环 |
| `ExecToPodViaWebSocket(...)` | 通过 WebSocket 执行 Pod 命令 |

### 方法（WebSocketIO）
| 方法 | 说明 |
|------|------|
| `Read(p)` | 实现 io.Reader |
| `Write(p)` | 实现 io.Writer（标准输出） |
| `Next()` | 实现 remotecommand.TerminalSizeQueue |
| `Close()` | 资源清理 |
| `readLoop()` (私有) | 主读循环 |
| `pingLoop()` (私有) | 心跳循环 |

### 导入
```go
- github.com/gorilla/websocket
- k8s.io/api/core/v1
- k8s.io/client-go/kubernetes
- k8s.io/client-go/kubernetes/scheme
- k8s.io/client-go/rest
- k8s.io/client-go/tools/remotecommand
```

---

## 3. watch.go - Watch 资源模块 (599 行)

**用途**: 监听 Kubernetes 资源变化并推送到客户端

### 导出函数
| 函数 | 说明 |
|------|------|
| `WatchNamespaceResources(ctx, ch, ns)` | 监听指定命名空间的资源 |
| `WatchUserNamespaceResources(ctx, ns, ch)` | 用户视角的命名空间资源监听 |

### 私有函数
| 函数 | 说明 |
|------|------|
| `watchUserAndSend()` | 用户资源监听和推送 |
| `watchAndSend()` | 通用资源监听和推送 |
| `buildDataMap()` | 提取 K8s 资源详情 |
| `extractStatusFields()` | 提取资源状态字段 |
| `extractServicePorts()` | 提取 Service 端口 |
| `extractServiceExternalIPs()` | 提取外部 IP |
| `extractServiceNodePorts()` | 提取 NodePort |
| `statusSnapshotString()` | 生成状态快照用于变化检测 |
| `fetchPodEvents()` | 获取 Pod 相关事件 |
| `isService()` | 判断是否为 Service |

### 支持的资源类型
- Pods (含 CrashLoopBackOff 检测、事件、容器信息)
- Services (IP、端口、外部 IP)
- Deployments (副本数)
- ReplicaSets
- Ingress
- Jobs

### 导入
```go
- k8s.io/apimachinery/pkg/apis/meta/v1
- k8s.io/apimachinery/pkg/apis/meta/v1/unstructured
- k8s.io/apimachinery/pkg/runtime/schema
- k8s.io/client-go/dynamic
```

---

## 4. client.go - 核心 Client 模块 (317 行)

**用途**: Kubernetes 资源操作（Job、FileBrowser、命名空间查询）

### 类型定义
```go
type JobSpec struct {}     // Job 创建规范
type VolumeSpec struct {}  // 卷挂载规范
```

### 导出函数
| 函数 | 说明 |
|------|------|
| `GetFilteredNamespaces(filter)` | 按名称过滤命名空间 |
| `CreateJob(ctx, spec)` | 创建 Kubernetes Job |
| `DeleteJob(ctx, ns, name)` | 删除 Job 及其 Pod |
| `CreateFileBrowserPod(ctx, ns, pvcs, ro, url)` | 创建 FileBrowser Pod |
| `CreateFileBrowserService(ctx, ns)` | 创建 FileBrowser Service |
| `DeleteFileBrowserResources(ctx, ns)` | 删除 FileBrowser 资源 |

### 特性
- Job 支持 GPU（共享和独占）、CPU/内存请求、环境变量、注解
- FileBrowser 支持多 PVC 挂载、只读模式、自定义 baseURL
- Pod/Service 级联删除管理

### 导入
```go
- github.com/linskybing/platform-go/internal/config
- k8s.io/api/batch/v1
- k8s.io/api/core/v1
- k8s.io/apimachinery/pkg/api/errors
- k8s.io/apimachinery/pkg/api/resource
- k8s.io/apimachinery/pkg/apis/meta/v1
- k8s.io/apimachinery/pkg/util/intstr
- k8s.io/apimachinery/pkg/util/wait
```

---

## 编译验证

```bash
$ cd /home/master/platform-go && go build ./pkg/k8s/...
# 编译成功，无错误
```

## 模块间依赖关系

```
    init.go (全局变量)
       ↓
  ┌────┴────┐
  ↓         ↓
watch.go  client.go
  ↓         ↓
websocket.go (ExecToPodViaWebSocket 使用 rest.Config)
```

---

## 迁移指南

### 对于导入该包的代码
无需更改导入语句，所有公共函数和类型仍可通过 `k8s` 包访问：

```go
import "github.com/linskybing/platform-go/pkg/k8s"

k8s.Init()
k8s.WatchNamespaceResources(...)
k8s.CreateJob(...)
```

### 新增 internal_test.go（如需）
```go
package k8s

import (
    "testing"
)

func TestInit(t *testing.T) {
    InitTestCluster()
    // test code
}
```

---

## 总结

✅ **分解完成** - 原始 1233 行代码按功能分为 4 个模块  
✅ **编译通过** - `go build ./pkg/k8s/...` 无误  
✅ **保持兼容** - 公共 API 完全保持一致  
✅ **增强可维护性** - 每个文件专注单一职责  

| 方面 | 改进 |
|------|------|
| 代码组织 | 相关功能聚集，职责明确 |
| 可读性 | 每个文件 < 600 行，易于理解 |
| 可维护性 | WebSocket、Watch、Job 操作独立，便于维护 |
| 测试 | 可针对各模块单独编写测试 |
