---
title: Workflow Frontend Integration Guide
description: Frontend implementation guide for Argo Workflow, Volcano Scheduler, and K8s job submission
date: 2024-01-15
---

# Workflow & Scheduler Frontend Integration Guide

Complete guide for frontend developers to integrate with the platform-go workflow and scheduler backend APIs.

## Architecture Overview

```
┌─────────────────────────────────┐
│    Frontend Application         │
│  (React/Vue/Angular)            │
│  ┌─────────────────────────────┐│
│  │ - Workflow Editor           ││
│  │ - Job Submission Form       ││
│  │ - Progress Dashboard        ││
│  │ - Queue Monitoring          ││
│  └─────────────────────────────┘│
└────────────┬────────────────────┘
             │ REST API + WebSocket
┌────────────▼────────────────────┐
│   Platform-Go Backend           │
│  - Workflow Service             │
│  - Scheduler Dispatcher         │
│  - Queue Manager                │
└────────────┬────────────────────┘
             │ K8s Client
    ┌────────┼────────┐
    │        │        │
    ▼        ▼        ▼
  Argo    Volcano   K8s Job
```

## 1. Authentication & Setup

### Initialize API Client

**TypeScript/JavaScript:**
```typescript
// api-client.ts
import axios, { AxiosInstance } from 'axios';

class WorkflowAPIClient {
    private client: AxiosInstance;
    
    constructor(baseURL: string, token: string) {
        this.client = axios.create({
            baseURL,
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        });
    }
    
    // Workflow operations
    async submitWorkflow(payload: SubmitWorkflowRequest): Promise<Workflow> {
        const response = await this.client.post('/workflows', payload);
        return response.data.data;
    }
    
    async getWorkflow(workflowId: string): Promise<Workflow> {
        const response = await this.client.get(`/workflows/${workflowId}`);
        return response.data.data;
    }
    
    async listWorkflows(filters: WorkflowFilters): Promise<WorkflowListResponse> {
        const response = await this.client.get('/workflows', { params: filters });
        return response.data;
    }
    
    async terminateWorkflow(workflowId: string): Promise<void> {
        await this.client.delete(`/workflows/${workflowId}`);
    }
    
    // Queue operations
    async listQueues(scheduler?: string): Promise<Queue[]> {
        const response = await this.client.get('/queues', {
            params: { scheduler }
        });
        return response.data.data;
    }
    
    async getQueueStatus(queueId: string): Promise<QueueStatus> {
        const response = await this.client.get(`/queues/${queueId}/status`);
        return response.data.data;
    }
    
    // WebSocket watch
    watchWorkflow(workflowId: string, onUpdate: (update: WorkflowStatusUpdate) => void): WebSocket {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const ws = new WebSocket(
            `${protocol}//${window.location.host}/api/v1/workflows/${workflowId}/watch`
        );
        
        ws.onmessage = (event) => {
            const update = JSON.parse(event.data);
            onUpdate(update);
        };
        
        return ws;
    }
}

// Export singleton
export const workflowAPI = new WorkflowAPIClient(
    process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1',
    localStorage.getItem('authToken') || ''
);
```

### Type Definitions

```typescript
// types.ts
export interface Workflow {
    id: string;
    name: string;
    description: string;
    project_id: number;
    status: 'pending' | 'running' | 'succeeded' | 'failed' | 'error';
    scheduler: 'volcano' | 'k8s-native' | 'argo-default';
    queue_id?: string;
    progress: number;
    start_time?: string;
    end_time?: string;
    estimated_duration?: number;
    cpu_used?: string;
    memory_used?: string;
    created_at: string;
    created_by: number;
    nodes?: WorkflowNode[];
}

export interface WorkflowNode {
    id: string;
    name: string;
    type: 'container' | 'script' | 'resource';
    status: 'pending' | 'running' | 'succeeded' | 'failed';
    phase?: string;
    cpu?: string;
    memory?: string;
    start_time?: string;
    end_time?: string;
    output?: string;
    dependencies?: string[];
    progress?: number;
}

export interface SubmitWorkflowRequest {
    name: string;
    description?: string;
    project_id: number;
    scheduler: 'volcano' | 'k8s-native' | 'argo-default';
    queue_id?: string;
    workflow_yaml: string;
    parameters?: Record<string, string>;
    estimated_cpu?: string;
    estimated_memory?: string;
}

export interface WorkflowFilters {
    project_id?: number;
    status?: string;
    scheduler?: string;
    page?: number;
    limit?: number;
}

export interface WorkflowStatusUpdate {
    type: string;
    workflow_id: string;
    status: string;
    progress: number;
    current_node?: string;
    message?: string;
    nodes?: WorkflowNode[];
    timestamp: string;
}

export interface Queue {
    id: string;
    name: string;
    description: string;
    scheduler: string;
    priority: number;
    quota_cpu: string;
    quota_memory: string;
    max_parallel: number;
    is_active: boolean;
}

export interface QueueStatus {
    id: string;
    name: string;
    quota: ResourceQuota;
    allocated: ResourceQuota;
    available: ResourceQuota;
    stats: {
        waiting: number;
        running: number;
        succeeded: number;
        failed: number;
    };
}

export interface ResourceQuota {
    cpu: string;
    memory: string;
}
```

## 2. Workflow Editor Component

### Visual Workflow Builder

```typescript
// WorkflowEditor.tsx
import React, { useState, useCallback } from 'react';
import ReactFlow, {
    Node,
    Edge,
    addEdge,
    MiniMap,
    Controls,
    Background,
    Connection
} from 'reactflow';

interface WorkflowEditorProps {
    onSubmit: (yaml: string) => Promise<void>;
}

export const WorkflowEditor: React.FC<WorkflowEditorProps> = ({ onSubmit }) => {
    const [nodes, setNodes] = useState<Node[]>([]);
    const [edges, setEdges] = useState<Edge[]>([]);
    const [selectedScheduler, setSelectedScheduler] = useState('volcano');
    const [loading, setLoading] = useState(false);
    
    const onConnect = useCallback((connection: Connection) => {
        setEdges((eds) => addEdge(connection, eds));
    }, []);
    
    const addNode = useCallback((type: 'container' | 'script') => {
        const newNode: Node = {
            id: `node-${Date.now()}`,
            data: { label: `${type}-${Date.now()}` },
            position: { x: 250, y: 25 },
            type: 'default'
        };
        setNodes((nds) => nds.concat(newNode));
    }, []);
    
    const generateYAML = useCallback((): string => {
        const templates = nodes.map(node => ({
            name: node.id,
            container: {
                image: node.data.image || 'python:3.9',
                command: node.data.command?.split(' ') || [],
                resources: {
                    requests: {
                        cpu: node.data.cpu || '100m',
                        memory: node.data.memory || '128Mi'
                    }
                }
            },
            dependencies: edges
                .filter(e => e.target === node.id)
                .map(e => e.source)
        }));
        
        const dagTasks = nodes.map(node => ({
            name: node.id,
            template: node.id,
            dependencies: edges
                .filter(e => e.target === node.id)
                .map(e => e.source)
        }));
        
        return JSON.stringify({
            apiVersion: 'argoproj.io/v1alpha1',
            kind: 'Workflow',
            metadata: {
                generateName: 'workflow-'
            },
            spec: {
                entrypoint: 'main',
                serviceName: 'workflow-service',
                schedulerName: selectedScheduler === 'volcano' ? 'volcano' : 'default',
                templates: [
                    ...templates,
                    {
                        name: 'main',
                        dag: {
                            tasks: dagTasks
                        }
                    }
                ]
            }
        }, null, 2);
    }, [nodes, edges, selectedScheduler]);
    
    const handleSubmit = async () => {
        setLoading(true);
        try {
            const yaml = generateYAML();
            await onSubmit(yaml);
        } finally {
            setLoading(false);
        }
    };
    
    return (
        <div style={{ display: 'flex', height: '100vh' }}>
            {/* Sidebar Controls */}
            <div style={{ width: '250px', padding: '20px', borderRight: '1px solid #ccc' }}>
                <h3>Workflow Builder</h3>
                
                <div style={{ marginBottom: '20px' }}>
                    <label>Scheduler:</label>
                    <select
                        value={selectedScheduler}
                        onChange={(e) => setSelectedScheduler(e.target.value)}
                    >
                        <option value="volcano">Volcano Scheduler</option>
                        <option value="k8s-native">K8s Native</option>
                        <option value="argo-default">Argo Default</option>
                    </select>
                </div>
                
                <button onClick={() => addNode('container')}>
                    + Add Container
                </button>
                <button onClick={() => addNode('script')} style={{ marginLeft: '10px' }}>
                    + Add Script
                </button>
                
                <div style={{ marginTop: '30px' }}>
                    <button
                        onClick={handleSubmit}
                        disabled={loading || nodes.length === 0}
                        style={{
                            width: '100%',
                            padding: '10px',
                            background: 'blue',
                            color: 'white',
                            border: 'none',
                            cursor: 'pointer'
                        }}
                    >
                        {loading ? 'Submitting...' : 'Submit Workflow'}
                    </button>
                </div>
                
                {/* YAML Preview */}
                <div style={{ marginTop: '30px', maxHeight: '300px', overflow: 'auto' }}>
                    <h4>YAML Preview:</h4>
                    <pre style={{ fontSize: '10px', background: '#f5f5f5', padding: '10px' }}>
                        {generateYAML()}
                    </pre>
                </div>
            </div>
            
            {/* Canvas */}
            <div style={{ flex: 1 }}>
                <ReactFlow
                    nodes={nodes}
                    edges={edges}
                    onConnect={onConnect}
                >
                    <Background />
                    <Controls />
                    <MiniMap />
                </ReactFlow>
            </div>
        </div>
    );
};
```

## 3. Workflow Submission

### Form Component

```typescript
// WorkflowSubmissionForm.tsx
import React, { useState, useEffect } from 'react';
import { workflowAPI } from './api-client';
import { SubmitWorkflowRequest, Queue } from './types';

interface SubmissionFormProps {
    projectId: number;
    onSuccess: (workflowId: string) => void;
}

export const WorkflowSubmissionForm: React.FC<SubmissionFormProps> = ({
    projectId,
    onSuccess
}) => {
    const [formData, setFormData] = useState<SubmitWorkflowRequest>({
        name: '',
        description: '',
        project_id: projectId,
        scheduler: 'volcano',
        workflow_yaml: '',
        parameters: {}
    });
    
    const [queues, setQueues] = useState<Queue[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    
    useEffect(() => {
        const loadQueues = async () => {
            try {
                const result = await workflowAPI.listQueues(formData.scheduler);
                setQueues(result);
            } catch (err) {
                console.error('Failed to load queues:', err);
            }
        };
        
        loadQueues();
    }, [formData.scheduler]);
    
    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError(null);
        
        try {
            const workflow = await workflowAPI.submitWorkflow(formData);
            onSuccess(workflow.id);
        } catch (err: any) {
            setError(err.response?.data?.error?.message || 'Failed to submit workflow');
        } finally {
            setLoading(false);
        }
    };
    
    return (
        <form onSubmit={handleSubmit} style={{ maxWidth: '600px' }}>
            <div style={{ marginBottom: '15px' }}>
                <label>Workflow Name:</label>
                <input
                    type="text"
                    value={formData.name}
                    onChange={(e) =>
                        setFormData({ ...formData, name: e.target.value })
                    }
                    required
                    style={{ width: '100%', padding: '8px' }}
                />
            </div>
            
            <div style={{ marginBottom: '15px' }}>
                <label>Description:</label>
                <textarea
                    value={formData.description}
                    onChange={(e) =>
                        setFormData({ ...formData, description: e.target.value })
                    }
                    style={{ width: '100%', padding: '8px', minHeight: '80px' }}
                />
            </div>
            
            <div style={{ marginBottom: '15px' }}>
                <label>Scheduler:</label>
                <select
                    value={formData.scheduler}
                    onChange={(e) =>
                        setFormData({ ...formData, scheduler: e.target.value as any })
                    }
                    style={{ width: '100%', padding: '8px' }}
                >
                    <option value="volcano">Volcano Scheduler</option>
                    <option value="k8s-native">K8s Native</option>
                    <option value="argo-default">Argo Default</option>
                </select>
            </div>
            
            {queues.length > 0 && (
                <div style={{ marginBottom: '15px' }}>
                    <label>Queue (Optional):</label>
                    <select
                        value={formData.queue_id || ''}
                        onChange={(e) =>
                            setFormData({ ...formData, queue_id: e.target.value || undefined })
                        }
                        style={{ width: '100%', padding: '8px' }}
                    >
                        <option value="">-- No Queue --</option>
                        {queues.map(q => (
                            <option key={q.id} value={q.id}>
                                {q.name} (Priority: {q.priority})
                            </option>
                        ))}
                    </select>
                </div>
            )}
            
            <div style={{ marginBottom: '15px' }}>
                <label>Workflow YAML:</label>
                <textarea
                    value={formData.workflow_yaml}
                    onChange={(e) =>
                        setFormData({ ...formData, workflow_yaml: e.target.value })
                    }
                    required
                    style={{
                        width: '100%',
                        padding: '8px',
                        minHeight: '300px',
                        fontFamily: 'monospace',
                        fontSize: '12px'
                    }}
                />
            </div>
            
            {error && (
                <div style={{ color: 'red', marginBottom: '15px' }}>
                    Error: {error}
                </div>
            )}
            
            <button
                type="submit"
                disabled={loading}
                style={{
                    padding: '10px 20px',
                    background: 'blue',
                    color: 'white',
                    border: 'none',
                    cursor: 'pointer'
                }}
            >
                {loading ? 'Submitting...' : 'Submit Workflow'}
            </button>
        </form>
    );
};
```

## 4. Real-time Progress Monitoring

### Progress Dashboard

```typescript
// WorkflowDashboard.tsx
import React, { useState, useEffect } from 'react';
import { workflowAPI } from './api-client';
import { Workflow, WorkflowStatusUpdate } from './types';

interface DashboardProps {
    workflowId: string;
}

export const WorkflowDashboard: React.FC<DashboardProps> = ({ workflowId }) => {
    const [workflow, setWorkflow] = useState<Workflow | null>(null);
    const [loading, setLoading] = useState(true);
    const [ws, setWs] = useState<WebSocket | null>(null);
    
    useEffect(() => {
        // Initial load
        const loadWorkflow = async () => {
            try {
                const data = await workflowAPI.getWorkflow(workflowId);
                setWorkflow(data);
            } finally {
                setLoading(false);
            }
        };
        
        loadWorkflow();
        
        // Setup WebSocket watch
        const websocket = workflowAPI.watchWorkflow(
            workflowId,
            (update: WorkflowStatusUpdate) => {
                setWorkflow(prev => prev ? {
                    ...prev,
                    status: update.status,
                    progress: update.progress,
                    nodes: update.nodes || prev.nodes
                } : null);
            }
        );
        
        setWs(websocket);
        
        return () => {
            if (websocket) {
                websocket.close();
            }
        };
    }, [workflowId]);
    
    if (loading) return <div>Loading...</div>;
    if (!workflow) return <div>Workflow not found</div>;
    
    return (
        <div style={{ padding: '20px' }}>
            <h2>{workflow.name}</h2>
            
            {/* Status Badge */}
            <div style={{
                display: 'inline-block',
                padding: '5px 10px',
                borderRadius: '4px',
                background: getStatusColor(workflow.status),
                color: 'white',
                marginBottom: '20px'
            }}>
                {workflow.status.toUpperCase()}
            </div>
            
            {/* Progress Bar */}
            <div style={{ marginBottom: '20px' }}>
                <label>Progress: {workflow.progress.toFixed(1)}%</label>
                <div style={{
                    width: '100%',
                    height: '24px',
                    background: '#e0e0e0',
                    borderRadius: '4px',
                    overflow: 'hidden'
                }}>
                    <div style={{
                        height: '100%',
                        width: `${workflow.progress}%`,
                        background: 'blue',
                        transition: 'width 0.3s'
                    }} />
                </div>
            </div>
            
            {/* Metadata */}
            <div style={{
                marginBottom: '20px',
                padding: '15px',
                background: '#f5f5f5',
                borderRadius: '4px'
            }}>
                <p><strong>Scheduler:</strong> {workflow.scheduler}</p>
                <p><strong>Started:</strong> {workflow.start_time}</p>
                {workflow.end_time && (
                    <p><strong>Ended:</strong> {workflow.end_time}</p>
                )}
                {workflow.estimated_duration && (
                    <p><strong>Estimated Duration:</strong> {workflow.estimated_duration}s</p>
                )}
            </div>
            
            {/* Nodes/Steps */}
            {workflow.nodes && workflow.nodes.length > 0 && (
                <div>
                    <h3>Steps</h3>
                    <div style={{ display: 'grid', gap: '10px' }}>
                        {workflow.nodes.map(node => (
                            <NodeCard key={node.id} node={node} />
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
};

interface NodeCardProps {
    node: any;
}

const NodeCard: React.FC<NodeCardProps> = ({ node }) => (
    <div style={{
        padding: '15px',
        border: '1px solid #ddd',
        borderRadius: '4px'
    }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '10px' }}>
            <strong>{node.name}</strong>
            <span style={{
                padding: '4px 8px',
                borderRadius: '3px',
                background: getStatusColor(node.status),
                color: 'white',
                fontSize: '12px'
            }}>
                {node.status}
            </span>
        </div>
        
        {node.progress !== undefined && (
            <div style={{ marginBottom: '10px' }}>
                <div style={{
                    height: '16px',
                    background: '#e0e0e0',
                    borderRadius: '2px',
                    overflow: 'hidden'
                }}>
                    <div style={{
                        height: '100%',
                        width: `${node.progress}%`,
                        background: 'green'
                    }} />
                </div>
            </div>
        )}
        
        {node.output && (
            <div style={{
                marginTop: '10px',
                padding: '10px',
                background: '#f9f9f9',
                borderRadius: '2px',
                fontFamily: 'monospace',
                fontSize: '12px',
                maxHeight: '100px',
                overflow: 'auto'
            }}>
                <pre>{node.output}</pre>
            </div>
        )}
    </div>
);

function getStatusColor(status: string): string {
    switch (status) {
        case 'running': return '#2196F3';
        case 'succeeded': return '#4CAF50';
        case 'failed': return '#F44336';
        case 'pending': return '#FFC107';
        default: return '#9E9E9E';
    }
}
```

## 5. Queue Management Dashboard

```typescript
// QueueDashboard.tsx
import React, { useState, useEffect } from 'react';
import { workflowAPI } from './api-client';
import { Queue, QueueStatus } from './types';

export const QueueDashboard: React.FC = () => {
    const [queues, setQueues] = useState<Queue[]>([]);
    const [selectedQueue, setSelectedQueue] = useState<string | null>(null);
    const [queueStatus, setQueueStatus] = useState<QueueStatus | null>(null);
    const [loading, setLoading] = useState(true);
    
    useEffect(() => {
        const loadQueues = async () => {
            try {
                const result = await workflowAPI.listQueues();
                setQueues(result);
            } finally {
                setLoading(false);
            }
        };
        
        loadQueues();
    }, []);
    
    useEffect(() => {
        if (!selectedQueue) return;
        
        const loadStatus = async () => {
            try {
                const status = await workflowAPI.getQueueStatus(selectedQueue);
                setQueueStatus(status);
            } catch (err) {
                console.error('Failed to load queue status:', err);
            }
        };
        
        loadStatus();
        
        // Refresh every 5 seconds
        const interval = setInterval(loadStatus, 5000);
        return () => clearInterval(interval);
    }, [selectedQueue]);
    
    if (loading) return <div>Loading queues...</div>;
    
    return (
        <div style={{ padding: '20px' }}>
            <h2>Queue Dashboard</h2>
            
            {/* Queue Selector */}
            <div style={{ marginBottom: '30px' }}>
                <h3>Queues</h3>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(250px, 1fr))', gap: '15px' }}>
                    {queues.map(queue => (
                        <div
                            key={queue.id}
                            onClick={() => setSelectedQueue(queue.id)}
                            style={{
                                padding: '15px',
                                border: selectedQueue === queue.id ? '2px solid blue' : '1px solid #ddd',
                                borderRadius: '4px',
                                cursor: 'pointer',
                                background: selectedQueue === queue.id ? '#f0f0ff' : '#fff'
                            }}
                        >
                            <h4>{queue.name}</h4>
                            <p><small>Scheduler: {queue.scheduler}</small></p>
                            <p><small>Priority: {queue.priority}</small></p>
                        </div>
                    ))}
                </div>
            </div>
            
            {/* Selected Queue Status */}
            {queueStatus && (
                <div style={{ marginTop: '30px' }}>
                    <h3>{queueStatus.name} - Details</h3>
                    
                    {/* Resource Usage */}
                    <div style={{
                        display: 'grid',
                        gridTemplateColumns: 'repeat(2, 1fr)',
                        gap: '20px',
                        marginBottom: '30px'
                    }}>
                        <ResourceUsageCard
                            title="CPU"
                            quota={queueStatus.quota.cpu}
                            allocated={queueStatus.allocated.cpu}
                            available={queueStatus.available.cpu}
                        />
                        <ResourceUsageCard
                            title="Memory"
                            quota={queueStatus.quota.memory}
                            allocated={queueStatus.allocated.memory}
                            available={queueStatus.available.memory}
                        />
                    </div>
                    
                    {/* Job Stats */}
                    <div style={{
                        display: 'grid',
                        gridTemplateColumns: 'repeat(4, 1fr)',
                        gap: '15px'
                    }}>
                        <StatCard label="Waiting" value={queueStatus.stats.waiting} color="#FFC107" />
                        <StatCard label="Running" value={queueStatus.stats.running} color="#2196F3" />
                        <StatCard label="Succeeded" value={queueStatus.stats.succeeded} color="#4CAF50" />
                        <StatCard label="Failed" value={queueStatus.stats.failed} color="#F44336" />
                    </div>
                </div>
            )}
        </div>
    );
};

interface ResourceUsageCardProps {
    title: string;
    quota: string;
    allocated: string;
    available: string;
}

const ResourceUsageCard: React.FC<ResourceUsageCardProps> = ({
    title,
    quota,
    allocated,
    available
}) => (
    <div style={{
        padding: '15px',
        border: '1px solid #ddd',
        borderRadius: '4px'
    }}>
        <h4>{title}</h4>
        <p><strong>Quota:</strong> {quota}</p>
        <p><strong>Allocated:</strong> {allocated}</p>
        <p><strong>Available:</strong> {available}</p>
    </div>
);

interface StatCardProps {
    label: string;
    value: number;
    color: string;
}

const StatCard: React.FC<StatCardProps> = ({ label, value, color }) => (
    <div style={{
        padding: '15px',
        border: `2px solid ${color}`,
        borderRadius: '4px',
        textAlign: 'center'
    }}>
        <p style={{ color, fontSize: '24px', fontWeight: 'bold' }}>{value}</p>
        <p>{label}</p>
    </div>
);
```

## Best Practices

### 1. Error Handling
- Always wrap API calls in try-catch
- Display user-friendly error messages
- Implement retry logic for transient failures
- Log errors for debugging

### 2. Performance
- Use WebSocket for real-time updates (not polling)
- Implement pagination for large lists
- Cache queue data with TTL
- Debounce rapid status checks

### 3. UX/DX
- Show progress indicators during submission
- Provide clear error messages with actionable solutions
- Use consistent styling for status indicators
- Implement auto-refresh for dashboards

### 4. Testing
```typescript
// Example test
describe('WorkflowAPI', () => {
    it('should submit workflow and return workflow id', async () => {
        const result = await workflowAPI.submitWorkflow({
            name: 'test-workflow',
            project_id: 1,
            scheduler: 'volcano',
            workflow_yaml: validYAML
        });
        
        expect(result.id).toBeDefined();
        expect(result.status).toBe('pending');
    });
});
```

## References
- [Argo Workflows REST API](https://github.com/argoproj/argo-workflows/blob/master/docs/rest-api.md)
- [React Flow Documentation](https://reactflow.dev/)
- [WebSocket API](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket)
- [Axios Documentation](https://axios-http.com/)
