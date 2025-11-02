# Project Context for AI Assistants

> **Purpose:** This document provides comprehensive context about the pm-tool codebase architecture, conventions, and patterns to help AI assistants provide accurate, context-aware assistance.

## Project Overview

- **Name:** pm-tool
- **Module:** github.com/zuhrulumam/pm-tool
- **Type:** REST API with Clean Architecture + Domain-Driven Design
- **Database:** postgres
- **Port:** 8080
- **Generated:** 2025-11-02 08:40:13
- **Schema File:** schema.yaml
- **Template Version:** 2.0.0 (Fixed Architecture)

## Architecture Overview

This project follows **Clean Architecture** with a strong **Domain Layer**:

```
pm-tool/
├── business/
│   ├── entity/         # Entities + Repository interfaces
│   ├── domain/         # ⭐ Business logic + Infrastructure calls
│   └── usecase/        # Orchestration only
├── handler/api/        # HTTP layer (DTO conversion)
├── infra/              # Infrastructure implementations
└── pkg/                # Utilities
```

## ⭐ Key Architecture Principles

### Request Flow:
```
HTTP Request
    ↓
Handler (Convert DTO → Entity)
    ↓
Usecase (Delegate/Orchestrate)
    ↓
Domain (Business Logic + Cache + Queue + DB + HTTP)
    ↓
Repository (SQL Only)
```

### Layer Rules:
- **Handler:** DTO ↔ Entity conversion, call usecase
- **Usecase:** Delegate to domain OR orchestrate cross-domain
- **Domain:** ALL business logic, cache, queue, DB, HTTP calls
- **Repository:** SQL queries only

## Generated Entities

### User
- **Table:** `users`
- **Files:**
  - `business/entity/user.go` - Entity + Repository interface
  - `business/domain/user_domain.go` - Business logic
  - `business/usecase/user_usecase.go` - Orchestration
  - `handler/api/user_handler.go` - HTTP handlers
  - `infra/postgres/user_repository.go` - DB implementation
- **Columns:** id (PK), email (unique) (required), password_hash, name, google_id (unique), avatar_url, email_verified (required), is_active (required), last_login_at, created_at (required), updated_at (required)
- **Primary Key:** id

### Project
- **Table:** `projects`
- **Files:**
  - `business/entity/project.go` - Entity + Repository interface
  - `business/domain/project_domain.go` - Business logic
  - `business/usecase/project_usecase.go` - Orchestration
  - `handler/api/project_handler.go` - HTTP handlers
  - `infra/postgres/project_repository.go` - DB implementation
- **Columns:** id (PK), user_id (required), name (required), description, color, is_archived (required), created_at (required), updated_at (required)
- **Primary Key:** id

### Note
- **Table:** `notes`
- **Files:**
  - `business/entity/note.go` - Entity + Repository interface
  - `business/domain/note_domain.go` - Business logic
  - `business/usecase/note_usecase.go` - Orchestration
  - `handler/api/note_handler.go` - HTTP handlers
  - `infra/postgres/note_repository.go` - DB implementation
- **Columns:** id (PK), project_id (required), title (required), content, content_type (required), audio_url, is_pinned (required), created_at (required), updated_at (required)
- **Primary Key:** id



## Critical Notes for AI Assistants

**ALWAYS follow these rules:**

1. **Domain Layer is King**
   - ALL business logic in domain
   - ALL infrastructure calls (cache, queue, HTTP) in domain
   - Handlers/usecases NEVER call infrastructure directly

2. **DTO ↔ Entity Conversion**
   - Handler receives DTO
   - Handler converts: `entity := req.ToEntity()`
   - Handler calls: `usecase.Create(ctx, entity)`
   - NEVER pass DTOs to usecase

3. **Repository Interface**
   - `Update(ctx, entity)` - Entity has ID
   - `FindAll(ctx, filters, page, pageSize)` - Always paginated
   - No separate List() method

4. **Entity Structure**
   - Use BaseEntity embedding
   - Don't repeat ID, timestamps, deleted_at

5. **Dependency Injection**
   - Domains get: Repo + Redis + Queue + HTTP + Tracer
   - Usecases get: Domains (NOT repos) + TxMgr + Tracer
   - Handlers get: Usecases only

## Redis (Enabled)
- Used by **domain layer** for: caching, rate_limit
- Cache methods in domain: cacheGet, cacheSet, cacheDelete




## Relationships
- **notes** → **projects** (many-to-one)
- **projects** → **users** (many-to-one)



---

**This architecture ensures clean separation: handlers handle HTTP, usecases orchestrate, domains contain ALL business logic and infrastructure calls, repositories do SQL only.**

Generated with ❤️ by go-template-gen v2.0 (Fixed Architecture)
