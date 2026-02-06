---
name: Workflow & Job Scheduler
description: Argo Workflow integration, Volcano scheduler, and K8s native scheduler for platform-go
version: "1.0"
date: "2024-01-15"
---

# Workflow & Job Scheduler Architecture

Complete guide for implementing Argo Workflow with Volcano and K8s native schedulers. Includes backend system design, API specifications, and frontend integration patterns.

## System Architecture

### Scheduling Options Comparison

| Feature | Volcano | K8s Native | Argo |
|---------|---------|-----------|------|
| **Purpose** | Batch job scheduling | General workloads | Workflow orchestration |
| **Best For** | ML/HPC jobs | Standard jobs | Complex multi-step workflows |
| **Queue Support** | ✓ Job queues | ✗ Pod priority | ✓ Built-in |
| **Fairness** | ✓ Gang scheduling | ✗ Per-pod | ✓ Per-workflow |
| **Resource Guarantee** | ✓ Reservation | ✓ Requests/Limits | ✓ Both |

### Architecture Diagram

```
┌─────────────────────────────────────────────────────┐
│         Frontend (React/Vue)                        │
│  - Workflow editor                                  │
│  - Job submission                                   │
│  - Monitoring dashboard                             │
└────────────────┬────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────┐
│    Platform-Go Backend API                          │
│  ┌────────────────────────────────────────────────┐ │
│  │ Workflow Service                               │ │
│  │ - Create/Submit workflows                      │ │
│  │ - List/Get workflow status                     │ │
│  │ - Delete/Terminate workflows                   │ │
│  └────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────┐ │
│  │ Scheduler Service                              │ │
│  │ - Dispatch to Volcano/K8s                      │ │
│  │ - Monitor job queue                            │ │
│  │ - Handle resource allocation                   │ │
│  └────────────────────────────────────────────────┘ │
└────────────────┬────────────────────────────────────┘
                 │
      ┌──────────┼──────────┐
      │          │          │
      ▼          ▼          ▼
   Argo      Volcano      K8s
 Workflows  Scheduler    Native
                       (Job/Pod)
```

## Backend System Design

### 1. Workflow Service

#### Domain Model
```go
// Workflow represents an Argo workflow
type Workflow struct {
    ID          string    `gorm:"primaryKey"`
    ProjectID   int       `gorm:"index"`
    Name        string
    Description string
    
    // Workflow specification
    WorkflowYAML string // Stored Argo Workflow definition
    Scheduler    string // "volcano" | "k8s-native" | "argo-default"
    
    // Execution info
    Status      string    // pending, running, succeeded, failed, error
    StartTime   *time.Time
    EndTime     *time.Time
    Duration    int64 // seconds
    
    // Resource tracking
    CpuUsed     string // "1000m", "2"
    MemoryUsed  string // "512Mi", "1Gi"
    
    // Metadata
    CreatedBy   int
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// WorkflowNode represents a step in the workflow
type WorkflowNode struct {
    ID          string
    WorkflowID  string
    Name        string
    Type        string // "container", "script", "resource"
    
    Status      string // pending, running, succeeded, failed
    Phase       string // additional status info
    
    // Resource request
    Cpu         string // "100m", "1"
    Memory      string // "128Mi", "256Mi"
    
    // Execution
    StartTime   *time.Time
    EndTime     *time.Time
    Output      string // stdout/stderr
    
    Dependencies []string // node IDs this depends on
}

// WorkflowQueue for batch job scheduling
type WorkflowQueue struct {
    ID          string    `gorm:"primaryKey"`
    Name        string    `gorm:"uniqueIndex"`
    Description string
    
    Scheduler   string    // "volcano" | "k8s-native"
    Priority    int       // Higher number = higher priority
    
    // Resource quota
    QuotaCpu    string
    QuotaMemory string
    
    // Settings
    MaxParallel int       // Max concurrent workflows in queue
    IsActive    bool
    
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

#### Service Interface
```go
type WorkflowService interface {
    // Create and submit workflow
    SubmitWorkflow(ctx context.Context, req *SubmitWorkflowRequest) (*Workflow, error)
    
    // Query workflows
    GetWorkflow(ctx context.Context, workflowID string) (*Workflow, error)
    ListWorkflows(ctx context.Context, projectID int, opts *ListOptions) ([]*Workflow, int64, error)
    
    // Workflow lifecycle
    TerminateWorkflow(ctx context.Context, workflowID string) error
    DeleteWorkflow(ctx context.Context, workflowID string) error
    
    // Workflow details
    GetWorkflowNodes(ctx context.Context, workflowID string) ([]*WorkflowNode, error)
    GetNodeOutput(ctx context.Context, workflowID, nodeID string) (string, error)
    
    // Status tracking
    WatchWorkflow(ctx context.Context, workflowID string) (<-chan WorkflowStatusUpdate, error)
}

type SubmitWorkflowRequest struct {
    Name        string `json:"name" binding:"required"`
    Description string `json:"description"`
    ProjectID   int    `json:"project_id" binding:"required"`
    
    // Workflow definition (YAML)
    WorkflowYAML string `json:"workflow_yaml" binding:"required"`
    
    // Scheduling options
    Scheduler    string `json:"scheduler" binding:"required"` // "volcano" | "k8s-native"
    QueueID      string `json:"queue_id"`                      // Optional queue assignment
    
    // Parameters
    Parameters   map[string]string `json:"parameters"`
    
    // Resource hints
    EstimatedCpu    string `json:"estimated_cpu"`    // "1000m", "2"
    EstimatedMemory string `json:"estimated_memory"` // "512Mi", "1Gi"
}

type WorkflowStatusUpdate struct {
    WorkflowID string
    Status     string
    Message    string
    Nodes      []*WorkflowNode
    Progress   float64 // 0-100
    Timestamp  time.Time
}
```

### 2. Scheduler Service

#### Dispatcher Pattern
```go
type SchedulerDispatcher interface {
    // Send workflow to specific scheduler
    Dispatch(ctx context.Context, workflow *Workflow) error
    
    // Get scheduler-specific status
    GetStatus(ctx context.Context, workflowID string) (string, error)
    
    // Terminate workflow on scheduler
    Terminate(ctx context.Context, workflowID string) error
}

// Factory pattern for scheduler selection
func NewSchedulerDispatcher(schedulerType string) (SchedulerDispatcher, error) {
    switch schedulerType {
    case "volcano":
        return NewVolcanoDispatcher(), nil
    case "k8s-native":
        return NewK8sNativeDispatcher(), nil
    case "argo-default":
        return NewArgoDefaultDispatcher(), nil
    default:
        return nil, fmt.Errorf("unknown scheduler: %s", schedulerType)
    }
}
```

### 3. Volcano Scheduler Integration

#### Implementation
```go
type VolcanoDispatcher struct {
    clientset kubernetes.Interface
    namespace string
}

// Submit job to Volcano PodGroup
func (v *VolcanoDispatcher) Dispatch(ctx context.Context, workflow *Workflow) error {
    // Parse Argo Workflow YAML
    argoWf := &v1alpha1.Workflow{}
    if err := yaml.Unmarshal([]byte(workflow.WorkflowYAML), argoWf); err != nil {
        return fmt.Errorf("parse workflow failed: %w", err)
    }
    
    // Create PodGroup for gang scheduling
    podGroup := &batch.PodGroup{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "pg-" + workflow.ID,
            Namespace: v.namespace,
            Labels: map[string]string{
                "workflow-id": workflow.ID,
            },
        },
        Spec: batch.PodGroupSpec{
            Queue: workflow.QueueID,
            MinMember: 1,
            Tolerations: []corev1.Toleration{
                {
                    Operator: corev1.TolerationOpExists,
                },
            },
            SchedulingPolicy: &batch.SchedulingPolicy{
                TimeUnit: &metav1.Duration{Duration: 1 * time.Second},
            },
        },
    }
    
    // Inject PodGroup label into workflow pods
    for i, template := range argoWf.Spec.Templates {
        if template.Container != nil {
            if argoWf.Spec.Templates[i].Metadata.Labels == nil {
                argoWf.Spec.Templates[i].Metadata.Labels = make(map[string]string)
            }
            argoWf.Spec.Templates[i].Metadata.Labels["volcano.sh/group-name"] = podGroup.Name
        }
    }
    
    // Create PodGroup
    _, err := volcanoClient.BatchV1alpha1().PodGroups(v.namespace).Create(ctx, podGroup, metav1.CreateOptions{})
    if err != nil {
        return fmt.Errorf("create podgroup failed: %w", err)
    }
    
    // Submit Argo workflow
    updatedYAML, _ := yaml.Marshal(argoWf)
    workflow.WorkflowYAML = string(updatedYAML)
    
    return submitArgoWorkflow(ctx, argoWf)
}
```

#### Queue Management
```go
type QueueManager interface {
    // Create Volcano queue
    CreateQueue(ctx context.Context, queue *WorkflowQueue) error
    
    // Update queue quota
    UpdateQuota(ctx context.Context, queueID string, cpu, memory string) error
    
    // Get queue status
    GetQueueStatus(ctx context.Context, queueID string) (*QueueStatus, error)
}

type QueueStatus struct {
    Name        string
    Reserved    ResourceQuota
    Allocated   ResourceQuota
    Available   ResourceQuota
    Waiting     int // Jobs waiting in queue
    Running     int // Jobs running
    Succeeded   int
    Failed      int
}

type ResourceQuota struct {
    Cpu    string
    Memory string
}

// Implementation
func (q *QueueManager) CreateQueue(ctx context.Context, wfQueue *WorkflowQueue) error {
    // Parse resource strings
    cpuVal := parseResource(wfQueue.QuotaCpu)
    memVal := parseResource(wfQueue.QuotaMemory)
    
    queue := &batch.Queue{
        ObjectMeta: metav1.ObjectMeta{
            Name: wfQueue.ID,
        },
        Spec: batch.QueueSpec{
            Weight:    int32(wfQueue.Priority),
            Reclaimable: &boolTrue,
            Guarantee: corev1.ResourceList{
                "cpu":    *cpuVal,
                "memory": *memVal,
            },
        },
    }
    
    _, err := volcanoClient.BatchV1alpha1().Queues().Create(ctx, queue, metav1.CreateOptions{})
    return err
}
```

### 4. K8s Native Scheduler Integration

#### Job Queue Pattern
```go
// K8s Job for simple batch workloads
type K8sJobDispatcher struct {
    clientset kubernetes.Interface
    namespace string
}

func (k *K8sJobDispatcher) Dispatch(ctx context.Context, workflow *Workflow) error {
    // For simple workflows without dependencies, use K8s Job
    job := &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "job-" + workflow.ID,
            Namespace: k.namespace,
            Labels: map[string]string{
                "workflow-id": workflow.ID,
                "scheduler": "k8s-native",
            },
        },
        Spec: batchv1.JobSpec{
            Parallelism:             int32Ptr(1),
            Completions:             int32Ptr(1),
            BackoffLimit:            int32Ptr(3),
            ActiveDeadlineSeconds:   int64Ptr(3600), // 1 hour timeout
            TTLSecondsAfterFinished: int32Ptr(3600), // Cleanup after 1 hour
            
            Template: corev1.PodTemplateSpec{
                Spec: corev1.PodSpec{
                    RestartPolicy: corev1.RestartPolicyNever,
                    Containers: []corev1.Container{
                        {
                            Name:  "worker",
                            Image: "your-worker:latest",
                            Env: []corev1.EnvVar{
                                {
                                    Name:  "WORKFLOW_ID",
                                    Value: workflow.ID,
                                },
                            },
                            Resources: corev1.ResourceRequirements{
                                Requests: corev1.ResourceList{
                                    "cpu":    resource.MustParse(workflow.CpuUsed),
                                    "memory": resource.MustParse(workflow.MemoryUsed),
                                },
                            },
                        },
                    },
                },
            },
        },
    }
    
    _, err := k.clientset.BatchV1().Jobs(k.namespace).Create(ctx, job, metav1.CreateOptions{})
    return err
}
```

#### Priority Class for Job Ordering
```go
// Create priority class for queue ordering
func setupPriorityClasses(ctx context.Context, clientset kubernetes.Interface) error {
    priorityClasses := []struct {
        name  string
        value int32
    }{
        {"critical", 1000},
        {"high", 100},
        {"normal", 10},
        {"low", 1},
    }
    
    for _, pc := range priorityClasses {
        priorityClass := &scheduling.PriorityClass{
            ObjectMeta: metav1.ObjectMeta{
                Name: pc.name,
            },
            Value:       pc.value,
            GlobalDefault: pc.name == "normal",
            Description: pc.name + " priority for workflows",
        }
        
        _, err := clientset.SchedulingV1().PriorityClasses().Create(ctx, priorityClass, metav1.CreateOptions{})
        if err != nil && !errors.IsAlreadyExists(err) {
            return err
        }
    }
    
    return nil
}
```

## API Specifications

### 1. Workflow Management Endpoints

#### Submit Workflow
```
POST /api/v1/workflows
```

**Request:**
```json
{
  "name": "data-processing-pipeline",
  "description": "Process daily data",
  "project_id": 1,
  "scheduler": "volcano",
  "queue_id": "high-priority-queue",
  "workflow_yaml": "apiVersion: argoproj.io/v1alpha1\nkind: Workflow\n...",
  "parameters": {
    "input_path": "/data/input",
    "output_path": "/data/output"
  },
  "estimated_cpu": "4",
  "estimated_memory": "8Gi"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "wf-abc123",
    "name": "data-processing-pipeline",
    "status": "pending",
    "scheduler": "volcano",
    "queue_id": "high-priority-queue",
    "created_at": "2024-01-15T10:30:00Z",
    "estimated_duration": 600
  }
}
```

#### Get Workflow
```
GET /api/v1/workflows/{workflow_id}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "wf-abc123",
    "name": "data-processing-pipeline",
    "status": "running",
    "progress": 45.5,
    "start_time": "2024-01-15T10:30:00Z",
    "estimated_end_time": "2024-01-15T11:10:00Z",
    "nodes": [
      {
        "id": "step-1",
        "name": "read-data",
        "status": "succeeded",
        "duration": 120
      },
      {
        "id": "step-2",
        "name": "process-data",
        "status": "running",
        "progress": 60
      },
      {
        "id": "step-3",
        "name": "write-results",
        "status": "pending"
      }
    ]
  }
}
```

#### List Workflows
```
GET /api/v1/workflows?project_id=1&status=running&scheduler=volcano&page=1&limit=20
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "wf-abc123",
      "name": "pipeline-1",
      "status": "running",
      "scheduler": "volcano",
      "progress": 45.5,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "pagination": {
    "total": 42,
    "page": 1,
    "limit": 20
  }
}
```

#### Terminate Workflow
```
DELETE /api/v1/workflows/{workflow_id}
```

**Response:**
```json
{
  "success": true,
  "message": "Workflow terminated successfully"
}
```

#### Get Workflow Node Output
```
GET /api/v1/workflows/{workflow_id}/nodes/{node_id}/output
```

**Response:**
```json
{
  "success": true,
  "data": {
    "stdout": "Processing complete: 1000 records\n",
    "stderr": "Warning: low memory\n",
    "exit_code": 0
  }
}
```

### 2. Queue Management Endpoints

#### Create Queue
```
POST /api/v1/queues
```

**Request:**
```json
{
  "name": "gpu-training-queue",
  "description": "For GPU training jobs",
  "scheduler": "volcano",
  "priority": 100,
  "quota_cpu": "16",
  "quota_memory": "32Gi",
  "max_parallel": 4
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "q-gpu123",
    "name": "gpu-training-queue",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

#### Get Queue Status
```
GET /api/v1/queues/{queue_id}/status
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "q-gpu123",
    "name": "gpu-training-queue",
    "quota": {
      "cpu": "16",
      "memory": "32Gi"
    },
    "allocated": {
      "cpu": "12",
      "memory": "24Gi"
    },
    "available": {
      "cpu": "4",
      "memory": "8Gi"
    },
    "stats": {
      "waiting": 2,
      "running": 3,
      "succeeded": 15,
      "failed": 1
    }
  }
}
```

#### List Queues
```
GET /api/v1/queues?scheduler=volcano
```

### 3. WebSocket Streaming

#### Watch Workflow Progress
```
WebSocket: /api/v1/workflows/{workflow_id}/watch
```

**Message Format:**
```json
{
  "type": "status_update",
  "workflow_id": "wf-abc123",
  "status": "running",
  "progress": 45.5,
  "current_node": "process-data",
  "nodes": [
    {
      "id": "step-2",
      "name": "process-data",
      "status": "running",
      "progress": 60
    }
  ],
  "timestamp": "2024-01-15T10:35:00Z"
}
```

## Frontend Integration Guide

### 1. Workflow Submission

```javascript
// Frontend code example
async function submitWorkflow(workflowData) {
    const response = await fetch('/api/v1/workflows', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
            name: workflowData.name,
            description: workflowData.description,
            project_id: workflowData.projectId,
            scheduler: 'volcano', // or 'k8s-native'
            queue_id: workflowData.queueId,
            workflow_yaml: generateArgoYAML(workflowData),
            parameters: workflowData.parameters,
            estimated_cpu: workflowData.estimatedCpu,
            estimated_memory: workflowData.estimatedMemory
        })
    });
    
    const result = await response.json();
    return result.data.id; // workflow_id
}
```

### 2. Real-time Monitoring

```javascript
// WebSocket monitoring
function watchWorkflow(workflowId) {
    const ws = new WebSocket(`/api/v1/workflows/${workflowId}/watch`);
    
    ws.onmessage = (event) => {
        const update = JSON.parse(event.data);
        
        // Update UI with progress
        updateProgressBar(update.progress);
        updateNodeStatus(update.nodes);
        
        if (update.status === 'succeeded' || update.status === 'failed') {
            ws.close();
            showCompletionMessage(update.status);
        }
    };
    
    ws.onerror = (error) => {
        console.error('WebSocket error:', error);
    };
}
```

### 3. Workflow Editor

```javascript
// Visual workflow editor integration
function buildWorkflowYAML(nodes, edges) {
    const templates = nodes.map(node => ({
        name: node.id,
        container: {
            image: node.image,
            command: node.command,
            resources: {
                requests: {
                    cpu: node.cpu,
                    memory: node.memory
                }
            }
        },
        dependencies: edges
            .filter(e => e.target === node.id)
            .map(e => e.source)
    }));
    
    return {
        apiVersion: 'argoproj.io/v1alpha1',
        kind: 'Workflow',
        metadata: { generateName: 'workflow-' },
        spec: {
            entrypoint: 'main',
            templates: [
                ...templates,
                {
                    name: 'main',
                    dag: {
                        tasks: buildDAGTasks(nodes, edges)
                    }
                }
            ]
        }
    };
}
```

## Configuration & Deployment

### 1. Argo Workflow Installation
```bash
kubectl create namespace argo
kubectl apply -n argo -f https://github.com/argoproj/argo-workflows/releases/download/v3.5.0/install.yaml
```

### 2. Volcano Scheduler Installation
```bash
helm repo add volcano https://volcano-sh.github.io/helm-charts
helm install volcano volcano/volcano -n volcano-system --create-namespace
```

### 3. Enable Schedulers in Workflows
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  name: example-workflow
spec:
  serviceAccountName: argo-workflow
  schedulerName: volcano  # Use Volcano scheduler
  entrypoint: main
  templates:
    - name: main
      container:
        image: python:3.9
        command: [python]
        args: ["script.py"]
```

## Monitoring & Observability

### Metrics to Track
```go
// Prometheus metrics
var (
    workflowSubmitted = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "workflow_submitted_total",
            Help: "Total workflows submitted",
        },
        []string{"scheduler", "queue"},
    )
    
    workflowDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "workflow_duration_seconds",
            Buckets: prometheus.ExponentialBuckets(1, 2, 10),
        },
        []string{"scheduler", "status"},
    )
    
    queueWaitTime = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "queue_wait_seconds",
            Buckets: []float64{10, 30, 60, 300, 600, 1800},
        },
        []string{"queue_id"},
    )
)
```

## References
- [Argo Workflows Documentation](https://argoproj.github.io/argo-workflows/)
- [Volcano Scheduler](https://volcano.sh/)
- [Kubernetes Job Documentation](https://kubernetes.io/docs/concepts/workloads/controllers/job/)
- [Kubernetes Scheduler Framework](https://kubernetes.io/docs/concepts/scheduling-eviction/scheduling-framework/)
