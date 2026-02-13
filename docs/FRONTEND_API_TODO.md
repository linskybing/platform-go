# Frontend Integration TODO List — platform-go API

> Date: 2026-02-13
> Backend provides 82 API Endpoints. The following is a list of APIs to be integrated, categorized by frontend page/function.

---

## Authentication Module (Auth)

| # | API | Method | Path | Description | Frontend Page |
|---|-----|--------|------|-------------|---------------|
| 1 | Login | POST | `/login` | Login with credentials, returns JWT Token | Login Page |
| 2 | Register | POST | `/register` | Register new user | Register Page |
| 3 | Logout | POST | `/logout` | Logout and clear session | Global |
| 4 | Auth Status | GET | `/auth/status` | Check Token validity, returns user info | App Init |

**Frontend Requirements**:
- [ ] Login/Register Form + Validation
- [ ] JWT Token Storage (Cookie / localStorage)
- [ ] Axios interceptor for Authorization header auto-attach
- [ ] Auto redirect to login on Token expiry
- [ ] Global Auth Context / Store

---

## User Management (Users)

| # | API | Method | Path | Description | Frontend Page |
|---|-----|--------|------|-------------|---------------|
| 5 | User List | GET | `/users/` | Get all users | Admin User Mgmt |
| 6 | Paging Query | GET | `/users/paging` | Paging query `?page=&size=` | Admin User Mgmt |
| 7 | User Details | GET | `/users/:id` | Get single user info | Profile Page |
| 8 | Update User | PUT | `/users/:id` | Update user info | Profile Edit |
| 9 | Delete User | DELETE | `/users/:id` | Delete user | Admin User Mgmt |
| 10 | Get Settings | GET | `/users/:id/settings` | Get user preferences | Settings Page |
| 11 | Update Settings | PUT | `/users/:id/settings` | Update user preferences | Settings Page |

**Frontend Requirements**:
- [ ] User List + Pagination Component
- [ ] User Details / Profile Page
- [ ] User Search/Filter
- [ ] User Edit Modal / Form
- [ ] User Delete Confirmation Dialog
- [ ] User Settings Page (Theme/Language/Notification Prefs)

---

## Group Management (Groups)

| # | API | Method | Path | Description | Frontend Page |
|---|-----|--------|------|-------------|---------------|
| 12 | Group List | GET | `/groups` | Get all groups | Group Mgmt Page |
| 13 | Group Details | GET | `/groups/:id` | Get group details (Member only) | Group Detail Page |
| 14 | Create Group | POST | `/groups` | Create new group (Admin) | Admin Group Mgmt |
| 15 | Update Group | PUT | `/groups/:id` | Update group info (Group Admin) | Group Settings |
| 16 | Delete Group | DELETE | `/groups/:id` | Delete group (Group Admin) | Admin Group Mgmt |

**Frontend Requirements**:
- [ ] Group List Card/Table View
- [ ] Create Group Modal (Name + Description)
- [ ] Group Detail Page (Member List + Project List)
- [ ] Group Edit Form
- [ ] Group Delete Confirmation Dialog

---

## Group Member Management (User Groups)

| # | API | Method | Path | Description | Frontend Page |
|---|-----|--------|------|-------------|---------------|
| 17 | Add to Group | POST | `/user-groups` | Add user to group (Group Admin) | Group Member Mgmt |
| 18 | Update Role | PUT | `/user-groups` | Update member role (Group Admin) | Group Member Mgmt |
| 19 | Remove from Group | DELETE | `/user-groups` | Remove member (Group Admin) | Group Member Mgmt |
| 20 | All Relations | GET | `/user-groups` | Get all user-group relations (Admin) | Admin Panel |
| 21 | Group Members | GET | `/user-groups/:group_id/members` | List all members of a group | Group Detail Page |
| 22 | By Group | GET | `/user-groups/by-group` | Query members of specific group | Group Page |
| 23 | By User | GET | `/user-groups/by-user` | Query groups of a user | User Dashboard |

**Frontend Requirements**:
- [ ] Member List Component (Show Role: admin/manager/user)
- [ ] Add Member Modal (User Search + Role Select)
- [ ] Role Switch Dropdown
- [ ] Remove Member Confirmation
- [ ] My Groups Sidebar/Dashboard

---

## Project Management (Projects)

| # | API | Method | Path | Description | Frontend Page |
|---|-----|--------|------|-------------|---------------|
| 24 | Project List | GET | `/projects` | Get all projects | Project List Page |
| 25 | User Projects | GET | `/projects/by-user` | Get current user's projects | Dashboard |
| 26 | Project Details | GET | `/projects/:id` | Get project details | Project Detail Page |
| 27 | Project Config | GET | `/projects/:id/config-files` | List project ConfigFiles | Project Detail Page |
| 28 | Create Project | POST | `/projects` | Create new project (Admin) | Admin Project Mgmt |
| 29 | Update Project | PUT | `/projects/:id` | Update project (Manager) | Project Settings |
| 30 | Delete Project | DELETE | `/projects/:id` | Delete project (Manager) | Admin Project Mgmt |

**Frontend Requirements**:
- [ ] Project List (Card View + Table View)
- [ ] Create Project Modal (Name + Description + Group)
- [ ] Project Detail Page (with ConfigFile List, Image List)
- [ ] Project Edit Form
- [ ] Delete Confirmation Dialog

---

## Project Image Management (Project Images)

| # | API | Method | Path | Description | Frontend Page |
|---|-----|--------|------|-------------|---------------|
| 31 | Allowed Images | GET | `/projects/:id/images` | List allowed images | Project Images Page |
| 32 | Add Image | POST | `/projects/:id/images` | Add allowed image (Manager) | Project Images Page |
| 33 | Remove Image | DELETE | `/projects/:id/images/:image_id` | Remove allowed image (Manager) | Project Images Page |
| 34 | Image Requests | GET | `/projects/:id/image-requests` | List project image requests | Project Images Page |

**Frontend Requirements**:
- [ ] Project Image Allow List
- [ ] Add/Search Image Modal
- [ ] Image Remove Confirmation
- [ ] Image Request List (Status Tags)

---

## Image Request Review (Image Requests — Admin)

| # | API | Method | Path | Description | Frontend Page |
|---|-----|--------|------|-------------|---------------|
| 35 | All Requests | GET | `/image-requests` | List all image requests (Admin) | Admin Image Review |
| 36 | Approve Request | PUT | `/image-requests/:id/approve` | Approve image request (Admin) | Admin Image Review |
| 37 | Reject Request | PUT | `/image-requests/:id/reject` | Reject image request (Admin) | Admin Image Review |

**Frontend Requirements**:
- [ ] Image Request Review List (Filter: pending/approved/rejected)
- [ ] Approve/Reject Buttons + Review Note Modal
- [ ] Request Detail Card (Registry + Image + Tag)

---

## ConfigFile Management

| # | API | Method | Path | Description | Frontend Page |
|---|-----|--------|------|-------------|---------------|
| 38 | List All | GET | `/configfiles` | List all ConfigFiles (Admin) | Admin ConfigFile |
| 39 | Get Single | GET | `/configfiles/:id` | Get ConfigFile details | ConfigFile Detail |
| 40 | Create | POST | `/configfiles` | Create ConfigFile (Manager) | ConfigFile Editor |
| 41 | Update | PUT | `/configfiles/:id` | Update ConfigFile (Manager) | ConfigFile Editor |
| 42 | Delete | DELETE | `/configfiles/:id` | Delete ConfigFile (Manager) | ConfigFile List |
| 43 | Project Query | GET | `/configfiles/project/:project_id` | List ConfigFiles by Project | Project Page |
| 44 | Create Instance | POST | `/configfiles/:id/instance` | Deploy Instance | ConfigFile Detail |
| 45 | Destroy Instance | DELETE | `/configfiles/:id/instance` | Destroy Instance | ConfigFile Detail |
| 46 | (Alias) Create | POST | `/instance/:id` | Create Instance (Alias) | Same as above |
| 47 | (Alias) Destroy | DELETE | `/instance/:id` | Destroy Instance (Alias) | Same as above |

**Frontend Requirements**:
- [ ] ConfigFile List (Group by Project)
- [ ] YAML Editor (Monaco Editor / CodeMirror)
- [ ] ConfigFile Create/Edit Page
- [ ] Instance Deploy/Destroy Buttons + Status Indicator
- [ ] Resource Limits Settings (CPU/Memory/GPU)

---

## Job Management (Plugin)

| # | API | Method | Path | Description | Frontend Page |
|---|-----|--------|------|-------------|---------------|
| 48 | Job Templates | GET | `/jobs/templates` | List Job templates | Job Submit Page |
| 49 | Submit Job | POST | `/jobs/submit` | Submit new Job | Job Submit Page |
| 50 | Job List | GET | `/jobs` | List all Jobs | Job List Page |
| 51 | Job Details | GET | `/jobs/:id` | Get Job details | Job Detail Page |
| 52 | Cancel Job | POST | `/jobs/:id/cancel` | Cancel running Job | Job Detail Page |
| 53 | GPU Usage | GET | `/jobs/:id/gpu-usage` | Job GPU usage history | Job Monitor Page |
| 54 | GPU Summary | GET | `/jobs/:id/gpu-summary` | Job GPU usage stats | Job Monitor Page |

**Frontend Requirements**:
- [ ] Job List Page (Status Filter: Pending/Running/Completed/Failed)
- [ ] Job Submit Form (Select Template + Set Params + Select Image)
- [ ] Job Detail Page (Status, Submit Time, Start/End Time)
- [ ] Job Real-time Logs (WebSocket Stream)
- [ ] GPU Usage Line Chart (Chart.js / Recharts)
- [ ] GPU Summary Cards (Avg/Peak/Duration)
- [ ] Job Cancel Confirmation Dialog

---

## Cluster Monitoring (Cluster)

| # | API | Method | Path | Description | Frontend Page |
|---|-----|--------|------|-------------|---------------|
| 55 | Cluster Summary | GET | `/api/cluster/summary` | Cluster resource summary | Dashboard |
| 56 | Node List | GET | `/api/cluster/nodes` | List all nodes | Cluster Monitor Page |
| 57 | Node Details | GET | `/api/cluster/nodes/:name` | Get node resource details | Node Detail Page |
| 58 | GPU Usage | GET | `/api/cluster/gpu-usage` | List Pod GPU usage | GPU Monitor Page |

**Frontend Requirements**:
- [ ] Cluster Dashboard (CPU/Memory/GPU Usage Pie Charts)
- [ ] Node List (Name, Status, Resource Usage)
- [ ] Node Detail Page (CPU/Memory/GPU Allocation Breakdown)
- [ ] GPU Usage Global Heatmap/Table
- [ ] Auto Refresh (polling / SSE)

---

## Form System (Forms)

| # | API | Method | Path | Description | Frontend Page |
|---|-----|--------|------|-------------|---------------|
| 59 | Create Form | POST | `/forms` | Create application form | Form Submit Page |
| 60 | My Forms | GET | `/forms/my` | Query my forms | My Forms Page |
| 61 | All Forms | GET | `/forms` | List all forms | Admin Form Mgmt |
| 62 | Update Status | PUT | `/forms/:id/status` | Update form status | Admin Form Mgmt |
| 63 | Add Comment | POST | `/forms/:id/messages` | Add comment to form | Form Detail Page |
| 64 | List Comments | GET | `/forms/:id/messages` | Get form comments | Form Detail Page |

**Frontend Requirements**:
- [ ] Form Submit Page (Title + Description + Tag + Related Project)
- [ ] My Forms List (Status: Pending/Approved/Rejected)
- [ ] Admin Form Review List
- [ ] Form Details + Conversational Comment Section
- [ ] Status Update Dropdown

---

## Storage Management (Storage)

### Group Storage

| # | API | Method | Path | Description | Frontend Page |
|---|-----|--------|------|-------------|---------------|
| 65 | Group Storage | GET | `/storage/group/:id` | List group storage | Group Storage Page |
| 66 | My Storage | GET | `/storage/my-storages` | Query my group storage | Dashboard |
| 67 | Create Storage | POST | `/storage/:id/storage` | Create group storage (Group Admin) | Group Storage Mgmt |
| 68 | Delete Storage | DELETE | `/storage/:id/storage/:pvcId` | Delete group storage | Group Storage Mgmt |
| 69 | Start Browse | POST | `/storage/:id/storage/:pvcId/start` | Start FileBrowser | Group Storage Page |
| 70 | Stop Browse | DELETE | `/storage/:id/storage/:pvcId/stop` | Stop FileBrowser | Group Storage Page |

### Admin User Storage

| # | API | Method | Path | Description | Frontend Page |
|---|-----|--------|------|-------------|---------------|
| 71 | Query Status | GET | `/admin/user-storage/:username/status` | Query user storage status | Admin Storage Mgmt |
| 72 | Init | POST | `/admin/user-storage/:username/init` | Initialize user storage | Admin Storage Mgmt |
| 73 | Expand | PUT | `/admin/user-storage/:username/expand` | Expand user storage | Admin Storage Mgmt |
| 74 | Delete | DELETE | `/admin/user-storage/:username` | Delete user storage | Admin Storage Mgmt |

### Storage Permissions

| # | API | Method | Path | Description | Frontend Page |
|---|-----|--------|------|-------------|---------------|
| 75 | Set Permission | POST | `/storage/permissions` | Set storage permission | Storage Permission Mgmt |
| 76 | Batch Set | POST | `/storage/permissions/batch` | Batch set permissions | Storage Permission Mgmt |
| 77 | Query Permission | GET | `/storage/permissions/group/:gid/pvc/:pvc_id` | Query user permission | Storage Permission Page |
| 78 | Permission List | GET | `/storage/permissions/group/:gid/pvc/:pvc_id/list` | List PVC permissions | Storage Permission Page |
| 79 | Access Policy | POST | `/storage/policies` | Set access policy | Storage Policy Mgmt |

### PVC Binding

| # | API | Method | Path | Description | Frontend Page |
|---|-----|--------|------|-------------|---------------|
| 80 | Create Binding | POST | `/k8s/pvc-binding` | Create PVC Binding | Project Storage Settings |
| 81 | List Bindings | GET | `/k8s/pvc-binding/project/:project_id` | List Project Bindings | Project Storage Page |
| 82 | Delete (ID) | DELETE | `/k8s/pvc-binding/:binding_id` | Delete binding by ID | Project Storage Page |
| 83 | Delete (Name) | DELETE | `/k8s/pvc-binding/project/:pid/:pvc_name` | Delete by name | Project Storage Page |

### FileBrowser

| # | API | Method | Path | Description | Frontend Page |
|---|-----|--------|------|-------------|---------------|
| 84 | Get Access | POST | `/k8s/filebrowser/access` | Get FileBrowser access | Storage Browser |

**Frontend Requirements**:
- [ ] Group Storage List (Capacity, Usage, Status)
- [ ] Create Storage Modal (Name + Capacity + StorageClass)
- [ ] FileBrowser Start/Stop Buttons
- [ ] FileBrowser iframe embed or new tab open
- [ ] Admin User Storage Mgmt Panel (Init/Expand/Delete)
- [ ] Storage Permission Matrix (User x PVC x Permission Level)
- [ ] PVC Binding to Project Settings UI
- [ ] Storage Capacity Progress Bar

---

## K8s Operations

| # | API | Method | Path | Description | Frontend Page |
|---|-----|--------|------|-------------|---------------|
| 85 | Pod Logs | GET | `/k8s/namespaces/:ns/pods/:name/logs` | Get Pod Logs | Pod Log Page |
| 86 | Storage Status | GET | `/k8s/user-storage/status` | Query my storage status | Personal Storage Page |
| 87 | Start Browse | POST | `/k8s/user-storage/browse` | Open Personal FileBrowser | Personal Storage Page |
| 88 | Stop Browse | DELETE | `/k8s/user-storage/browse` | Stop Personal FileBrowser | Personal Storage Page |
| 89 | Storage Proxy | Any | `/k8s/user-storage/proxy/*path` | FileBrowser Reverse Proxy | Storage Browser |

**Frontend Requirements**:
- [ ] Pod Log Viewer (Auto Scroll + Timestamps)
- [ ] Personal Storage Dashboard (Status/Capacity)
- [ ] FileBrowser Embed Page
- [ ] Proxy Request Auto Routing

---

## Notifications (Notifications)

| # | API | Method | Path | Description | Frontend Page |
|---|-----|--------|------|-------------|---------------|
| 90 | Read All | PUT | `/api/notifications/read-all` | Mark all read | Notification Panel |
| 91 | Clear All | DELETE | `/api/notifications/clear-all` | Clear all notifications | Notification Panel |
| 92 | Read Single | PUT | `/api/notifications/:id/read` | Mark single read | Notification Panel |

**Frontend Requirements**:
- [ ] Notification Bell Icon + Unread Count Badge
- [ ] Notification Dropdown Panel (Real-time update)
- [ ] Notification History Page
- [ ] Mark All Read / Clear All Buttons

---

## Audit Logs (Audit)

| # | API | Method | Path | Description | Frontend Page |
|---|-----|--------|------|-------------|---------------|
| 93 | Audit Logs | GET | `/audit/logs` | Get operation audit logs (Admin) | Admin Audit Page |

**Frontend Requirements**:
- [ ] Audit Log Table (Operator, Action, Resource, Time)
- [ ] Filter (By User, Action Type, Date Range)
- [ ] Pagination
- [ ] Operation Detail Modal (old_data / new_data JSON diff)

---

## WebSocket Real-time Communication

| # | API | Method | Path | Description | Frontend Page |
|---|-----|--------|------|-------------|---------------|
| 94 | Container Exec | WS | `/ws/exec` | Pod Terminal (xterm.js) | Terminal Page |
| 95 | Resource Watch | WS | `/ws/watch/:namespace` | Watch namespace changes | Dashboard |
| 96 | Pod Log Stream| WS | `/ws/pod-logs` | Real-time Pod Log Stream | Log Page |
| 97 | Job Status | WS | `/ws/job-status/:id` | Real-time Job Status Update | Job Details Page |

**Frontend Requirements**:
- [ ] Web Terminal (xterm.js + WebSocket)
- [ ] Real-time Resource Monitor Dashboard (Pod status updates)
- [ ] Real-time Log Stream (ANSI color support)
- [ ] Job Status Real-time Update (Progress bar + Status Badge)
- [ ] WebSocket Reconnection Mechanism
- [ ] Heartbeat (ping/pong) Keep-alive

---

## Frontend Tech Recommendations

### Recommended Stack

| Category | Option | Usage |
|----------|--------|-------|
| Framework | React 18+ / Next.js 14+ | SPA / SSR |
| State Mgmt | Zustand / TanStack Query | Global State + Server State |
| UI Lib | Ant Design / Shadcn/ui | Table/Form/Modal |
| HTTP Client | Axios | API Requests + Interceptor |
| WebSocket | native WebSocket / Socket.io | Real-time Communication |
| Terminal | xterm.js | Web Terminal |
| Code Editor | Monaco Editor | YAML ConfigFile Edit |
| Charts | Recharts / Chart.js | GPU/Resource Monitor Charts |
| Form Validation | React Hook Form + Zod | Form Validation |

### Frontend Page Planning

```
/                          → Dashboard (Cluster Summary + My Projects + My Jobs)
/login                     → Login Page
/register                  → Register Page
/profile                   → Profile + Settings
/groups                    → Group List
/groups/:id                → Group Details (Members + Storage + Projects)
/projects                  → Project List
/projects/:id              → Project Details (ConfigFile + Images + Storage)
/projects/:id/configfiles  → ConfigFile List
/projects/:id/images       → Project Image Management
/jobs                      → Job List
/jobs/:id                  → Job Details (Status + Logs + GPU)
/storage                   → My Storage Overview
/forms                     → Form List
/forms/:id                 → Form Details + Comments
/terminal                  → Web Terminal
/admin/users               → Admin User Management
/admin/groups              → Admin Group Management
/admin/audit               → Admin Audit Logs
/admin/images              → Admin Image Review
/admin/storage             → Admin Storage Management
/cluster                   → Cluster Monitoring
/cluster/nodes/:name       → Node Details
/cluster/gpu               → GPU Usage Monitoring
/notifications             → Notification Center
```

### API Service File Structure Recommendation

```
src/
├── api/
│   ├── client.ts          # Axios instance + interceptors
│   ├── auth.ts            # Login/Register/Logout
│   ├── users.ts           # User CRUD
│   ├── groups.ts          # Group CRUD
│   ├── userGroups.ts      # Group Member Mgmt
│   ├── projects.ts        # Project CRUD
│   ├── configFiles.ts     # ConfigFile CRUD + Instance
│   ├── jobs.ts            # Job Submit/Query/Cancel
│   ├── images.ts          # Image Mgmt + Requests
│   ├── cluster.ts         # Cluster Monitoring
│   ├── storage.ts         # Storage Mgmt (Group + User + PVC)
│   ├── forms.ts           # Form System
│   ├── audit.ts           # Audit Logs
│   ├── notifications.ts   # Notifications
│   └── websocket.ts       # WebSocket Connection Mgmt
├── types/
│   ├── auth.ts
│   ├── user.ts
│   ├── group.ts
│   ├── project.ts
│   ├── configFile.ts
│   ├── job.ts
│   ├── image.ts
│   ├── storage.ts
│   ├── form.ts
│   ├── cluster.ts
│   ├── audit.ts
│   └── notification.ts
└── hooks/
    ├── useAuth.ts
    ├── useUsers.ts
    ├── useGroups.ts
    ├── useProjects.ts
    ├── useJobs.ts
    ├── useCluster.ts
    ├── useStorage.ts
    └── useWebSocket.ts
```

---

## Integration Priority

| Priority | Module | Reason |
|----------|--------|--------|
| P0 | Auth (Login/Register/Logout) | Foundation of all features |
| P0 | Users (Profile) | Immediate need after login |
| P1 | Groups + User Groups | Core organizational structure |
| P1 | Projects | Core workspace |
| P1 | ConfigFiles + Instance | Core operation flow |
| P1 | Jobs (Submit/List/Detail) | Core feature |
| P2 | Cluster (Monitoring) | Dashboard requirement |
| P2 | Storage (Group + Personal) | Data management |
| P2 | Images (Allow List + Request) | Project settings |
| P3 | Forms | Application process |
| P3 | Audit | Admin feature |
| P3 | Notifications | UX enhancement |
| P3 | WebSocket (Terminal + Logs) | Advanced feature |
| P3 | GPU Usage (Charts) | Monitoring enhancement |
