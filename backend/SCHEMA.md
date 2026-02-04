# Database Schema Visualization

## Entity Relationship Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           COMMUNITY MARKETPLACE DATABASE                     │
└─────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────┐
│       USERS          │
├──────────────────────┤
│ id (PK)              │
│ name                 │
│ email (unique)       │
│ password_hash        │
│ role                 │◄─────────────────────┐
│ is_active            │                      │
└──────┬───────────────┘                      │
       │                                      │
       │ Many-to-Many                         │
       │                                      │
       ▼                                      │
┌──────────────────────┐                     │
│  USER_COMMUNITIES    │                     │
├──────────────────────┤                     │
│ user_id (PK, FK)     │                     │
│ community_id (PK, FK)│                     │
│ role                 │                     │
│ joined_at            │                     │
│ is_active            │                     │
└──────┬───────────────┘                     │
       │                                      │
       │                                      │
       ▼                                      │
┌──────────────────────┐                     │
│    COMMUNITIES       │                     │
├──────────────────────┤                     │
│ id (PK)              │                     │
│ name                 │                     │
│ description          │                     │
│ address              │                     │
│ city                 │                     │
│ state                │                     │
│ country              │                     │
│ zip_code             │                     │
│ is_active            │                     │
└──────┬───────────────┘                     │
       │                                      │
       │ has many                             │
       │                                      │
       ├──────────────────┐                  │
       │                  │                  │
       ▼                  ▼                  │
┌─────────────┐   ┌─────────────────┐       │
│    POSTS    │   │ SERVICE_REQUESTS│       │
├─────────────┤   ├─────────────────┤       │
│ id (PK)     │   │ id (PK)         │       │
│ title       │   │ title           │       │
│ content     │   │ description     │       │
│ author_id(FK)│  │ category        │       │
│ community_id│   │ requester_id(FK)│───────┘
│ is_published│   │ community_id(FK)│
│ view_count  │   │ status          │
└──────┬──────┘   │ budget          │
       │          │ accepted_offer  │
       │          │ completed_at    │
       │          └──────┬──────────┘
       │                 │
       │ has many        │ has many
       │                 │
       │                 ▼
       │          ┌─────────────────┐
       │          │ SERVICE_OFFERS  │
       │          ├─────────────────┤
       │          │ id (PK)         │
       │          │ service_req(FK) │
       │          │ provider_id (FK)│───────────┐
       │          │ description     │           │
       │          │ proposed_price  │           │
       │          │ estimated_dur   │           │
       │          │ status          │           │
       │          └──────┬──────────┘           │
       │                 │                      │
       │                 │ has many             │
       │                 │                      │
       └─────────┬───────┴──────────────────────┤
                 │                              │
                 ▼                              │
          ┌─────────────┐                      │
          │  COMMENTS   │                      │
          ├─────────────┤                      │
          │ id (PK)     │                      │
          │ content     │                      │
          │ author_id(FK)│──────────────────────┘
          │ post_id (FK)│
          │ service_req │
          │ service_off │
          │ parent_comm │
          └─────────────┘
                 ▲
                 │ nested replies
                 │
                 └──────────────┐
                                │
                                
                         ┌─────────────┐
                         │   RATINGS   │
                         ├─────────────┤
                         │ id (PK)     │
                         │ provider(FK)│
                         │ rater_id(FK)│
                         │ service_req │
                         │ score (1-5) │
                         │ review      │
                         └─────────────┘
```

## Key Relationships

### User Relationships
- Users ↔ Communities (Many-to-Many via UserCommunities)
- Users → Posts (One-to-Many)
- Users → ServiceRequests (One-to-Many as Requester)
- Users → ServiceOffers (One-to-Many as Provider)
- Users → Comments (One-to-Many)
- Users → Ratings (One-to-Many as Rater)
- Users ← Ratings (One-to-Many as Provider)

### Community Relationships
- Communities ↔ Users (Many-to-Many via UserCommunities)
- Communities → Posts (One-to-Many)
- Communities → ServiceRequests (One-to-Many)

### Service Workflow
1. User creates ServiceRequest in a Community
2. ServiceProviders create ServiceOffers for the Request
3. Requester accepts one ServiceOffer
4. After completion, Requester creates Rating for Provider

### Content Engagement
- Posts can have Comments
- ServiceRequests can have Comments
- ServiceOffers can have Comments
- Comments can have nested Comments (replies)

## Database Features

### ✅ Multi-tenancy
- Multiple communities supported
- Users can join multiple communities
- Community-specific roles

### ✅ Role-Based Access Control
- Global roles (super_admin, admin, moderator, service_provider, user)
- Community-specific roles
- Hierarchical permission structure

### ✅ Service Marketplace
- Request and offer workflow
- Status tracking (open → in_progress → completed)
- Budget management
- Offer acceptance logic

### ✅ Social Features
- Content posting
- Commenting with nesting
- Rating and review system

### ✅ Data Integrity
- Foreign key constraints
- Check constraints (e.g., rating score 1-5)
- Cascade rules for referential integrity
- Soft deletes for data recovery

### ✅ Performance
- Strategic indexes on foreign keys
- Indexes on frequently queried fields (status, category)
- Indexes on unique fields (email)

## Example Usage Flow

```
1. Super Admin creates Community
   ↓
2. Users join Community
   ↓
3. User posts ServiceRequest in Community
   ↓
4. ServiceProvider creates ServiceOffer
   ↓
5. User accepts ServiceOffer
   ↓
6. ServiceProvider completes service
   ↓
7. User creates Rating for ServiceProvider
   ↓
8. Other users view Provider's ratings
```

## Implementation Details

- **ORM**: GORM (Go)
- **Database**: SQLite (can be migrated to PostgreSQL)
- **Migrations**: gormigrate for version control
- **Constraints**: Database-level validation
- **Deletion**: Soft deletes on all models
- **Timestamps**: Automatic created_at, updated_at, deleted_at
