# ARCHITECTURE.md

## Overview

This document outlines the technical architecture for the **Projects & Notes** MVP prototype, built as part of a 3-hour technical assessment. The prototype demonstrates core functionality including user authentication via Google OAuth, project management, and note-taking capabilities.

**Note on Scope**: Due to the time constraint and OAuth debugging challenges encountered during development, the notes CRUD functionality is partially implemented. The authentication system and project management features are fully functional.

---

## 1. Tech Stack & Rationale

### Backend
- **Language**: Go 1.21+
- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL 15
- **Cache**: Redis
- **Authentication**: JWT + Google OAuth 2.0
- **Observability**: OpenTelemetry

**Why Go?**
- **Personal Expertise**: I have existing Go tooling (CLI generator) that accelerated development significantly, allowing me to scaffold the entire backend structure in under 30 minutes
- **Performance**: Go's concurrency model and low latency make it ideal for API services that need to scale
- **Simplicity**: Straightforward deployment (single binary), easy to maintain, and excellent standard library
- **Industry Adoption**: Widely used in backend services, making it easier to find experienced developers when scaling

**Why Gin?**
- Lightweight, fast HTTP framework with minimal overhead
- Extensive middleware ecosystem (CORS, JWT, logging)
- Well-documented and production-proven

**Why PostgreSQL?**
- ACID compliance for data integrity (critical for project/note relationships)
- JSON support (JSONB) for flexible schema evolution
- Mature ecosystem with excellent tooling for backup, replication, and monitoring
- Strong support for complex queries and transactions

**Why Redis?**
- Session management and JWT token caching
- Rate limiting for API endpoints
- Future use: caching frequently accessed projects/notes to reduce database load

### Frontend
- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite
- **Styling**: Tailwind CSS
- **HTTP Client**: Axios
- **State Management**: Context API (sufficient for MVP scope)

**Why React?**
- **Hiring Pool**: React is the most popular frontend framework, making it significantly easier to find qualified frontend engineers when scaling the team
- **Ecosystem**: Massive library ecosystem for any future feature needs
- **TypeScript**: Type safety reduces bugs and improves developer experience
- **Component Reusability**: Easy to build and maintain UI component library

**Why Tailwind CSS?**
- Rapid prototyping without writing custom CSS
- Consistent design system out of the box
- Small bundle size with purging unused styles
- Easy to customize and extend

---

## 2. High-Level Architecture

### Current MVP Architecture (1,000 users)

```
┌─────────────────┐
│   User Browser  │
└────────┬────────┘
         │ HTTPS
         ▼
┌─────────────────────────────────────┐
│   React Frontend (Vite Dev Server)  │
│   - Google OAuth UI                 │
│   - Project Management              │
│   - Project Detail                  │
└────────┬────────────────────────────┘
         │ REST API (JSON)
         ▼
┌─────────────────────────────────────┐
│   Go Backend API (Gin)              │
│   ┌───────────────────────────────┐ │
│   │  Middleware Layer             │ │
│   │  - CORS                       │ │
│   │  - JWT Auth                   │ │
│   │  - Request Logging            │ │
│   │  - Rate Limiting              │ │
│   └───────────────────────────────┘ │
│   ┌───────────────────────────────┐ │
│   │  Handler Layer                │ │
│   │  - Auth (Google OAuth)        │ │
│   │  - Projects CRUD              │ │
│   │  - Notes CRUD                 │ │
│   └───────────────────────────────┘ │
│   ┌───────────────────────────────┐ │
│   │  Usecase                      │ │
│   └───────────────────────────────┘ │
│   ┌───────────────────────────────┐ │
│   │  Domain/Business Logic        │ │
│   └───────────────────────────────┘ │
└────────┬────────────────────────────┘
         │
    ┌────┴─────┬─────────────┐
    ▼          ▼             ▼
┌──────────┐ ┌──────────┐ ┌──────────────┐
│PostgreSQL│ │  Redis   │ │Google OAuth  │
│          │ │  Cache   │ │     API      │
│ - users  │ │          │ └──────────────┘
│ - projects│ │- Sessions│
│ - notes  │ │- Tokens  │
└──────────┘ └──────────┘
```

### Request Flow

**Authentication Flow (Google OAuth):**
1. User clicks "Sign in with Google" in React app
2. Frontend receives Google credential token
3. POST `/api/v1/auth/google/callback` with credential
4. Backend verifies token with Google's tokeninfo API
5. Backend creates/updates user in PostgreSQL
6. Backend generates JWT token
7. Frontend stores JWT in localStorage
8. Subsequent requests include JWT in `Authorization: Bearer <token>` header

**Project Creation Flow:**
1. Authenticated user sends POST `/api/v1/projects` with project data
2. JWT middleware extracts user_id from token
3. Handler validates request and creates project entity
4. Repository persists to PostgreSQL with user_id foreign key
5. Response returns created project with generated UUID

**Notes Flow (Designed, not yet implemented):**
1. GET `/api/v1/projects/:id/notes` - List all notes in a project
2. POST `/api/v1/notes` - Create note linked to project_id
3. DELETE `/api/v1/notes/:id` - Remove note

### Data Model

```sql
auth.users
  - id (uuid, PK)
  - email (unique)
  - google_id (unique)
  - name
  - avatar_url
  - created_at, updated_at

app.projects
  - id (uuid, PK)
  - user_id (FK -> users.id, CASCADE)
  - name
  - description
  - created_at, updated_at

app.notes
  - id (uuid, PK)
  - project_id (FK -> projects.id, CASCADE)
  - title
  - content (text)
  - content_type (enum: 'text', 'audio_transcription')
  - audio_url (nullable, for future audio feature)
  - created_at, updated_at
```

---

## 3. Audio-to-Text Transcription Extension

### Recommended Service: OpenAI Whisper API

**Why Whisper?**
- **Accuracy**: State-of-the-art speech recognition (supports 98 languages)
- **Cost-Effective**: $0.006 per minute of audio (~$0.36 per hour)
- **Simple Integration**: Single REST API call, no complex setup
- **Reliable**: Backed by OpenAI's infrastructure

**Alternative Considerations:**
- **Google Speech-to-Text**: Better for real-time streaming, more expensive ($0.016/min), better for live transcription
- **AWS Transcribe**: Good for AWS-native deployments, similar pricing to Google
- **Assembly AI**: Best accuracy for English, speaker detection included, but higher cost ($0.012/min)

### Integration Architecture

```
┌──────────────┐
│ User uploads │
│ audio file   │
└──────┬───────┘
       │
       ▼
┌─────────────────────┐
│ POST /api/v1/notes  │
│   + audio file      │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────────────────┐
│ Handler:                        │
│ 1. Validate file (format, size) │
│ 2. Generate unique filename     │
│ 3. Upload to S3/Cloud Storage   │
│ 4. Create note record (pending) │
│ 5. Publish transcription job    │
└──────┬──────────────────────────┘
       │
       ▼
┌────────────────────────┐
│ Message Queue          │
│ (RabbitMQ/AWS SQS)     │
│                        │
│ Job: {                 │
│   note_id: uuid        │
│   audio_url: s3://...  │
│   user_id: uuid        │
│ }                      │
└──────┬─────────────────┘
       │
       ▼
┌─────────────────────────────────┐
│ Transcription Worker (Go)       │
│ 1. Poll queue for jobs          │
│ 2. Download audio from S3       │
│ 3. Call Whisper API             │
│ 4. Receive text transcript      │
│ 5. Update note.content in DB    │
│ 6. Update content_type = 'audio'│
│ 7. Ack message (or retry)       │
└─────────────────────────────────┘
```

### Key Challenges & Solutions

**1. File Size Limitations**
- **Challenge**: Whisper API has a 25MB file limit
- **Solution**: 
  - Implement client-side compression before upload
  - For longer recordings, chunk audio into segments and transcribe separately
  - Merge transcripts with timestamps to maintain context

**2. Processing Time**
- **Challenge**: Transcription can take 10-30 seconds for typical notes (2-5 minutes of audio)
- **Solution**:
  - Asynchronous processing via message queue (user doesn't wait)
  - WebSocket or polling mechanism to notify user when complete
  - Show "Transcribing..." status in UI with progress indicator
  - For this use case (short project notes), audio should be 30 seconds to 3 minutes max

**3. Audio Length Consideration for Notes**
- **Challenge**: Project notes should be concise summaries, not lengthy recordings
- **Recommendation**: 
  - **Enforce max audio length of 5 minutes** in the UI and API validation
  - Show warning at 3 minutes: "Keep notes brief for better organization"
  - This keeps transcription costs predictable (~$0.03 per note maximum)
  - Encourages users to create multiple focused notes rather than one long recording
  - Faster processing: 5-minute audio typically transcribes in 15-20 seconds

**4. Cost at Scale**
- **Challenge**: At 100K users with 10 audio notes/month = $60K/month in transcription costs
- **Solution**:
  - Rate limiting per user (e.g., 20 audio notes/month on free tier)
  - Consider tiered pricing for heavy audio users
  - Implement caching: if same audio uploaded, reuse transcript (hash-based deduplication)
  - Enforce maximum audio duration (3-5 minutes for project notes)

**5. Error Handling & Retry Logic**
- **Challenge**: API failures, network issues, corrupted audio files
- **Solution**:
  - **Outbox pattern**: Store transcription job in database before publishing to queue
  - Exponential backoff retry (3 attempts: immediate, +2s, +10s)
  - Dead letter queue for failed jobs requiring manual review
  - User-facing error states: "Transcription failed - audio may be corrupted. Please try again."

**6. Accuracy & Language Support**
- **Challenge**: Different accents, technical jargon, background noise
- **Solution**:
  - Allow users to specify language hint (auto-detect otherwise)
  - Show confidence scores in UI when available
  - Enable manual editing of transcripts
  - Consider preprocessing: noise reduction using ffmpeg before sending to API

### Storage Considerations

- **Audio Files**: Store in S3 or similar object storage
  - Original audio retained for 30 days (allow re-transcription)
  - Auto-delete after retention period to save costs
  - Use presigned URLs for secure access
- **Transcripts**: Store in `notes.content` field (PostgreSQL TEXT column)
  - Indexed for full-text search
  - Version history if user edits transcript

---

## 4. Scaling from 1,000 to 100,000 Users

### Key Scaling Considerations

#### **1. Microservices Architecture with Message Queues**

**Current State (Monolith - 1K users):**
- Single Go binary handles all requests
- Works well for low traffic, simple to deploy and debug

**Future State (Microservices - 100K users):**

```
                    ┌────────────────┐
                    │  Load Balancer │
                    │   (AWS ALB)    │
                    └────────┬───────┘
                             │
            ┌────────────────┼────────────────┐
            │                │                │
            ▼                ▼                ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │   Auth       │ │  Projects    │ │   Notes      │
    │   Service    │ │  Service     │ │   Service    │
    └──────┬───────┘ └──────┬───────┘ └──────┬───────┘
           │                │                │
           └────────────────┼────────────────┘
                            │
                    ┌───────▼────────┐
                    │  Message Queue │
                    │  RabbitMQ/SQS  │
                    └───────┬────────┘
                            │
                    ┌───────▼────────┐
                    │  Transcription │
                    │    Worker      │
                    └────────────────┘
```

**Why Microservices?**
- **Independent Scaling**: Scale transcription workers separately from auth service
- **Fault Isolation**: If notes service fails, projects and auth still work
- **Team Autonomy**: Different teams can own different services
- **Technology Flexibility**: Use Python for ML features while keeping Go for APIs

**Message Queue Strategy (Outbox Pattern):**
```go
// Outbox table schema
type OutboxEvent struct {
    ID        string    `json:"id"`
    EventType string    `json:"event_type"` // "note.created", "transcription.requested"
    Payload   json.Raw  `json:"payload"`
    Status    string    `json:"status"` // "pending", "published", "failed"
    CreatedAt time.Time `json:"created_at"`
    Retries   int       `json:"retries"`
}

// Transaction guarantees reliability
func (s *Service) CreateNoteWithAudio(note Note) error {
    tx := s.db.Begin()
    
    // 1. Insert note
    if err := tx.Create(&note).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    // 2. Insert outbox event
    event := OutboxEvent{
        ID: uuid.New(),
        EventType: "transcription.requested",
        Payload: json.Marshal(TranscriptionJob{NoteID: note.ID}),
        Status: "pending",
    }
    if err := tx.Create(&event).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    tx.Commit()
    
    // 3. Background worker publishes events from outbox to queue
    return nil
}
```

**Benefits of Outbox Pattern:**
- **Guaranteed Delivery**: Event stored in DB before publishing (no message loss)
- **Easy Retry**: Failed events automatically retried by worker
- **Monitoring**: Query outbox table to see stuck/failed events
- **Debugging**: Full audit trail of all events

**Queue Selection:**
- **Development/Small Scale (< 10K users)**: RabbitMQ or NSQ (self-hosted, free)
- **Production Scale (10K-100K users)**: AWS SQS or Google Cloud Pub/Sub
  - Fully managed, no operational overhead
  - Auto-scaling, built-in monitoring
  - Pay-per-use pricing (~$0.40 per million requests)

#### **2. Database Scaling Strategy**

**Phase 1: Optimization (1K - 10K users)**
- **Connection Pooling**: Limit concurrent DB connections (e.g., 20-50 max)
  ```go
  db.SetMaxOpenConns(25)
  db.SetMaxIdleConns(10)
  db.SetConnMaxLifetime(5 * time.Minute)
  ```
- **Indexing**: Add indexes on frequently queried columns
  - `projects.user_id` (for listing user's projects)
  - `notes.project_id` (for listing project's notes)
  - `notes.created_at` (for sorting by date)
- **Query Optimization**: Use `SELECT` specific columns, avoid `SELECT *`

**Phase 2: Vertical Scaling (10K - 30K users)**
- Upgrade to larger PostgreSQL instance (more CPU, RAM)
- Add Redis for caching:
  - Cache user sessions (reduce DB reads)
  - Cache frequently accessed projects (TTL: 5 minutes)
  - Cache note counts per project

**Phase 3: Read Replicas (30K - 100K users)**
```
┌──────────────────┐
│  Write Requests  │
│  (POST, PUT,     │
│   DELETE)        │
└────────┬─────────┘
         │
         ▼
┌────────────────────┐
│ PostgreSQL Primary │◄──────── Streaming Replication
│   (Read + Write)   │
└────────────────────┘
         │
         ├─────────────────┬─────────────────┐
         ▼                 ▼                 ▼
┌────────────────┐ ┌────────────────┐ ┌────────────────┐
│ Read Replica 1 │ │ Read Replica 2 │ │ Read Replica 3 │
│  (Read Only)   │ │  (Read Only)   │ │  (Read Only)   │
└────────────────┘ └────────────────┘ └────────────────┘
         ▲                 ▲                 ▲
         │                 │                 │
         └─────────────────┴─────────────────┘
                    Read Requests
              (GET /projects, GET /notes)
```

**Database Configuration:**
- Primary: All writes and critical reads
- Read Replicas: List operations, analytics, reports
- Connection routing in application layer:
  ```go
  // Write to primary
  primaryDB.Create(&project)
  
  // Read from replica
  replicaDB.Find(&projects, "user_id = ?", userID)
  ```

**Phase 4: Sharding (> 100K users, if needed)**
- Shard by `user_id` (all user's data on same shard)
- Use consistent hashing or database proxy (e.g., Vitess, Citus)

#### **3. Container Orchestration with Kubernetes**

**When to adopt**: At ~10K users or when managing multiple services becomes complex

**Why Kubernetes?**
- **Auto-scaling**: Horizontal Pod Autoscaler adjusts replicas based on CPU/memory
- **Self-healing**: Automatically restarts failed containers
- **Rolling Updates**: Zero-downtime deployments
- **Service Discovery**: Built-in DNS for service-to-service communication

**Example Deployment:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: projects-service
spec:
  replicas: 3  # Start with 3 pods
  template:
    spec:
      containers:
      - name: projects-api
        image: projects-service:v1.2.0
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: projects-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: projects-service
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### Additional Scaling Considerations

**CDN for Frontend Assets**
- Serve React build from CloudFront/Cloudflare
- Reduces latency for global users
- Caches static assets (JS, CSS, images)

**API Rate Limiting**
- Redis-based rate limiter: 100 requests/min per user
- Prevents abuse and ensures fair resource allocation

**Monitoring & Observability**
- **Metrics**: Prometheus + Grafana
  - Request latency (p50, p95, p99)
  - Error rates per endpoint
  - Database connection pool utilization
- **Tracing**: OpenTelemetry (already implemented)
  - Trace requests across microservices
  - Identify bottlenecks
- **Logging**: Centralized logging (ELK stack or CloudWatch)
- **Alerting**: PagerDuty for critical errors

**Security Enhancements**
- Rate limiting per IP
- DDoS protection (Cloudflare)
- Database encryption at rest
- Regular security audits and dependency updates

---

## Summary

This architecture balances **pragmatic MVP development** with **long-term scalability**. The current monolithic design serves 1,000 users efficiently while maintaining clean separation of concerns that enables future refactoring into microservices.

Key strengths:
- **Proven tech stack** with strong community support
- **Clear migration path** from monolith → microservices
- **Cost-effective** at current scale, with identified optimization points for growth
- **Developer-friendly** architecture that facilitates team expansion

The proposed audio transcription feature integrates cleanly with the existing architecture, and the scaling plan provides a clear roadmap from 1K to 100K+ users with specific implementation milestones.

---

**Development Time**: ~2h 45min (including OAuth debugging)  
**Status**: Authentication ✅ | Projects ✅ | Notes ⚠️ (not yet implemented)