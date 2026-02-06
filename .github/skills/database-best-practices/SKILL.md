---
name: database-best-practices
description: Database design, ORM usage with GORM, query optimization, transactions, and migrations for PostgreSQL in platform-go
license: Proprietary
metadata:
  author: platform-go
  version: "1.0"
---


# Database Best Practices

This skill provides guidelines for database operations, schema design, and ORM patterns in platform-go.

## When to Use

Apply this skill when:
- Creating or modifying database tables and schemas
- Writing database queries or using GORM
- Implementing database transactions
- Creating database migrations
- Optimizing slow database queries
- Setting up database indexes
- Implementing pagination in queries
- Handling database connection pools

## Quick Start: Using Database Scripts

This skill includes ready-to-use migration validation scripts:

```bash
# Check migration safety and versioning
bash .github/skills/database-best-practices/scripts/migration-check.sh
```

## Database Design Principles

### 1. Schema Design

```go
// Define models with appropriate types and constraints
type User struct {
    ID        uint           `gorm:"primaryKey"`
    Username  string         `gorm:"index;unique;size:32"`
    Email     string         `gorm:"index;unique;size:255"`
    Password  string         `gorm:"size:255;not null"`
    CreatedAt time.Time      `gorm:"autoCreateTime"`
    UpdatedAt time.Time      `gorm:"autoUpdateTime"`
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Use appropriate indexes for frequent queries
type Job struct {
    ID        uint
    UserID    uint           `gorm:"index"`
    ProjectID uint           `gorm:"index"`
    Status    string         `gorm:"index"`
    CreatedAt time.Time      `gorm:"index"`
    UpdatedAt time.Time
    
    // Composite index for common queries
}

// Create composite indexes for WHERE + ORDER BY combinations
func (Job) TableName() string {
    return "jobs"
}
```

### 2. GORM Query Patterns

```go
// Always use prepared statements to prevent SQL injection
func (r *JobRepository) GetJobByID(ctx context.Context, jobID uint) (*Job, error) {
    var job Job
    if err := r.db.WithContext(ctx).First(&job, "id = ?", jobID).Error; err != nil {
        return nil, fmt.Errorf("failed to get job: %w", err)
    }
    return &job, nil
}

// Use eager loading to prevent N+1 queries
func (r *UserRepository) GetUserWithJobs(ctx context.Context, userID uint) (*User, error) {
    var user User
    if err := r.db.WithContext(ctx).
        Preload("Jobs").
        First(&user, "id = ?", userID).Error; err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return &user, nil
}

// Use select to retrieve only needed columns
func (r *UserRepository) GetUserEmails(ctx context.Context) ([]string, error) {
    var emails []string
    if err := r.db.WithContext(ctx).
        Table("users").
        Where("deleted_at IS NULL").
        Select("email").
        Scan(&emails).Error; err != nil {
        return nil, fmt.Errorf("failed to get emails: %w", err)
    }
    return emails, nil
}

// Use builder pattern for complex queries
func (r *JobRepository) ListJobs(ctx context.Context, filter *JobFilter) ([]Job, int64, error) {
    query := r.db.WithContext(ctx)
    
    if filter.UserID > 0 {
        query = query.Where("user_id = ?", filter.UserID)
    }
    if filter.Status != "" {
        query = query.Where("status = ?", filter.Status)
    }
    if filter.ProjectID > 0 {
        query = query.Where("project_id = ?", filter.ProjectID)
    }
    
    var total int64
    if err := query.Model(&Job{}).Count(&total).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to count jobs: %w", err)
    }
    
    var jobs []Job
    if err := query.
        Order("created_at DESC").
        Offset((filter.Page - 1) * filter.Limit).
        Limit(filter.Limit).
        Find(&jobs).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to list jobs: %w", err)
    }
    
    return jobs, total, nil
}
```

### 3. Database Transactions

```go
// Always use transactions for multi-step operations
func (s *JobService) CreateJobWithSteps(ctx context.Context, jobReq *CreateJobRequest) (*Job, error) {
    // Start transaction
    tx := s.db.BeginTx(ctx, nil)
    if tx.Error != nil {
        return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
    }
    
    // Create job
    job := &Job{
        UserID:    jobReq.UserID,
        ProjectID: jobReq.ProjectID,
        Status:    "pending",
    }
    if err := tx.Create(job).Error; err != nil {
        tx.Rollback()
        return nil, fmt.Errorf("failed to create job: %w", err)
    }
    
    // Update related data
    if err := tx.Model(&Project{}).Where("id = ?", jobReq.ProjectID).
        Update("last_job_id", job.ID).Error; err != nil {
        tx.Rollback()
        return nil, fmt.Errorf("failed to update project: %w", err)
    }
    
    // Commit transaction
    if err := tx.Commit().Error; err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }
    
    return job, nil
}

// Use savepoint for nested transactions
func (s *Service) ComplexOperation(ctx context.Context) error {
    tx := s.db.BeginTx(ctx, nil)
    
    if err := tx.Create(&Entity1{}).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    // Create savepoint
    sp := tx.SavePoint("before_entity2")
    
    if err := tx.Create(&Entity2{}).Error; err != nil {
        tx.RollbackTo(sp)
        // Continue with alternative path
    }
    
    return tx.Commit().Error
}
```

### 4. Query Optimization

```go
// Use pagination for large result sets
func (r *Repository) GetPagedResults(ctx context.Context, page, limit int) ([]Item, int64, error) {
    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 20
    }
    
    var items []Item
    var total int64
    
    offset := (page - 1) * limit
    if err := r.db.WithContext(ctx).
        Model(&Item{}).
        Count(&total).
        Offset(offset).
        Limit(limit).
        Find(&items).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to get paged results: %w", err)
    }
    
    return items, total, nil
}

// Use batching for bulk operations
func (r *Repository) UpdateManyJobs(ctx context.Context, jobIDs []uint, status string) error {
    if len(jobIDs) == 0 {
        return nil
    }
    
    batchSize := 1000
    for i := 0; i < len(jobIDs); i += batchSize {
        end := i + batchSize
        if end > len(jobIDs) {
            end = len(jobIDs)
        }
        
        batch := jobIDs[i:end]
        if err := r.db.WithContext(ctx).
            Model(&Job{}).
            Where("id IN ?", batch).
            Update("status", status).Error; err != nil {
            return fmt.Errorf("failed to update jobs: %w", err)
        }
    }
    
    return nil
}

// Use explain to analyze slow queries
func (r *Repository) AnalyzeQuery() {
    var result map[string]interface{}
    r.db.Raw("EXPLAIN ANALYZE SELECT * FROM jobs WHERE status = ? LIMIT 100", "running").
        Scan(&result)
}
```

### 5. Database Migrations

```go
// Use migrations for schema changes
type Migration struct {
    ID        uint
    Version   string `gorm:"unique"`
    AppliedAt time.Time
}

// Migration file: 001_create_users_table.go
func (m *Migrator) Up() error {
    return m.db.Migrator().CreateTable(&User{})
}

func (m *Migrator) Down() error {
    return m.db.Migrator().DropTable(&User{})
}

// Apply migrations
func RunMigrations(db *gorm.DB) error {
    return db.AutoMigrate(
        &User{},
        &Project{},
        &Job{},
        &Resource{},
    )
}

// Add indexes through migrations
func AddIndexes(db *gorm.DB) error {
    return db.Migrator().CreateIndex(&Job{}, "user_id", "project_id", "status")
}

// Rename column safely
func RenameColumn(db *gorm.DB) error {
    return db.Migrator().RenameColumn(&Job{}, "old_name", "new_name")
}
```

### 6. Connection Pool Management

```go
// Configure connection pool for optimal performance
func InitDatabase(dsn string) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        PrepareStmt:                              true,
        DisableForeignKeyConstraintWhenMigrating: false,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect database: %w", err)
    }
    
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get database instance: %w", err)
    }
    
    // Configure connection pool
    sqlDB.SetMaxIdleConns(10)       // Idle connections
    sqlDB.SetMaxOpenConns(100)      // Max open connections
    sqlDB.SetConnMaxLifetime(time.Hour) // Connection lifetime
    
    return db, nil
}
```

### 7. Error Handling

```go
// Handle specific database errors
func (r *Repository) CreateUser(ctx context.Context, user *User) error {
    if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
        if errors.Is(err, gorm.ErrDuplicatedKey) {
            return fmt.Errorf("user with this email already exists: %w", err)
        }
        if errors.Is(err, gorm.ErrInvalidDB) {
            return fmt.Errorf("invalid database connection: %w", err)
        }
        return fmt.Errorf("failed to create user: %w", err)
    }
    return nil
}

// Timeout handling
func (r *Repository) QueryWithTimeout(timeout time.Duration, query string, args ...interface{}) error {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    return r.db.WithContext(ctx).Raw(query, args...).Error
}
```

## Database Checklist

- [ ] All database operations use context with timeout
- [ ] Prepared statements are used (GORM handles this by default)
- [ ] N+1 queries are eliminated with preloading
- [ ] Pagination is implemented for large result sets
- [ ] Indexes are created for frequently queried columns
- [ ] Transactions are used for multi-step operations
- [ ] Connection pool is properly configured
- [ ] Database errors are handled appropriately
- [ ] Slow queries are identified and optimized
- [ ] Migrations are versioned and tested

## Performance Guidelines

- Query <100ms for standard operations
- Pagination with limit 20-100 items per page
- Batch operations in chunks of 1000 or less
- Use connection pool max of 100 for high concurrency
- Index foreign key columns and frequently filtered fields
- Avoid SELECT * - retrieve only needed columns
- Use EXPLAIN ANALYZE for slow queries over 500ms
