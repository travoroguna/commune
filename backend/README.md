# Backend - Community Marketplace API

This is the Go backend for the Community Marketplace application, built with GORM ORM and SQLite database.

## Database Schema

The application uses a comprehensive database schema designed to support:
- **Multi-community support**: Users can join and participate in multiple communities
- **Domain isolation**: Each community can have its own subdomain or custom domain
- **User roles**: Super admins, admins, moderators, service providers, and regular users
- **Content posting**: Users can post content within their communities
- **Service marketplace**: Request and offer services within communities
- **Comments and engagement**: Comment on posts and services with nested replies
- **Rating system**: Rate and review service providers

For detailed database documentation, see [DATABASE.md](./DATABASE.md).

## Database Models

### Core Entities

1. **User**: Authentication and user profile with role-based access
2. **Community**: Physical communities with domain isolation support (subdomain, custom domain, slug)
3. **UserCommunity**: Many-to-many relationship with community-specific roles
4. **Post**: Content posted within communities
5. **ServiceRequest**: Service requests from community members
6. **ServiceOffer**: Service provider offers for requests
7. **Comment**: Comments on posts, requests, and offers (supports nesting)
8. **Rating**: Ratings and reviews for service providers

### Entity Relationships

```
┌──────────┐       ┌─────────────────┐       ┌─────────────┐
│   User   │◄─────►│ UserCommunity   │◄─────►│  Community  │
└────┬─────┘       └─────────────────┘       └──────┬──────┘
     │                                               │
     │ creates                                creates│
     ▼                                               ▼
┌──────────┐                                   ┌──────────┐
│   Post   │                                   │ Service  │
│          │◄──────────────────────────────────│ Request  │
└────┬─────┘                                   └────┬─────┘
     │                                              │
     │ has                                     has  │
     ▼                                              ▼
┌──────────┐                                  ┌──────────┐
│ Comment  │                                  │ Service  │
└──────────┘                                  │  Offer   │
                                              └────┬─────┘
                                                   │
                                              has  │
                                                   ▼
                                              ┌──────────┐
                                              │  Rating  │
                                              └──────────┘
```

## Community Domain Isolation

Each community lives in its own space with three access methods:

### 1. Subdomain Routing
Communities can be accessed via subdomains:
- `sunset.commune.com`
- `tower-plaza.commune.com`
- `lakeside-villas.commune.com`

### 2. Custom Domain
Communities can have their own custom domains:
- `sunsetapts.com`
- `towerplaza.com`
- `lakesidevillas.com`

### 3. Slug-based URLs
For shared domain access:
- `commune.com/c/sunset-apartments`
- `commune.com/c/tower-plaza`
- `commune.com/c/lakeside-villas`

### Implementation
Use the `GetCommunityByDomain()` helper function in middleware to route requests to the correct community:

```go
func CommunityMiddleware(db *gorm.DB) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            domain := r.Host
            community, err := GetCommunityByDomain(db, domain)
            if err != nil {
                // Handle error or fallback to slug-based routing
            }
            // Add community to context
            ctx := context.WithValue(r.Context(), "community", community)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

## User Roles

### Global Roles (User.Role)
- **super_admin**: Can create communities and manage all communities
- **admin**: General administrative privileges
- **moderator**: Content moderation capabilities
- **service_provider**: Can offer services to the community
- **user**: Regular community member

### Community-Specific Roles (UserCommunity.Role)
Users can have different roles in different communities:
- **admin**: Administrator of a specific community
- **moderator**: Moderator of a specific community
- **service_provider**: Service provider within a specific community
- **user**: Regular member of a specific community

## Migrations

The application uses [gormigrate](https://github.com/go-gormigrate/gormigrate) for database migrations.

### Current Migrations

- **202402041300**: Initial User table (legacy)
- **202402041301**: Complete community marketplace schema

### Running Migrations

Migrations run automatically when the application starts. The database is created if it doesn't exist.

### Adding New Migrations

To add a new migration, edit `main.go` and add a new migration to the `runMigrations` function:

```go
{
    ID: "202402041302",
    Migrate: func(tx *gorm.DB) error {
        // Your migration code here
        return nil
    },
    Rollback: func(tx *gorm.DB) error {
        // Your rollback code here
        return nil
    },
}
```

## API Endpoints

### Current Endpoints

- `GET /api/health` - Health check endpoint
- `GET /api/users` - Get all users

### Future Endpoints

The following endpoints should be implemented:

#### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login user
- `POST /api/auth/logout` - Logout user
- `GET /api/auth/me` - Get current user

#### Communities
- `GET /api/communities` - List all communities
- `POST /api/communities` - Create community (super_admin only)
- `GET /api/communities/:id` - Get community details
- `PUT /api/communities/:id` - Update community
- `DELETE /api/communities/:id` - Delete community
- `POST /api/communities/:id/join` - Join community
- `POST /api/communities/:id/leave` - Leave community

#### Posts
- `GET /api/communities/:id/posts` - List posts in community
- `POST /api/communities/:id/posts` - Create post
- `GET /api/posts/:id` - Get post details
- `PUT /api/posts/:id` - Update post
- `DELETE /api/posts/:id` - Delete post

#### Service Requests
- `GET /api/communities/:id/service-requests` - List service requests
- `POST /api/communities/:id/service-requests` - Create service request
- `GET /api/service-requests/:id` - Get service request details
- `PUT /api/service-requests/:id` - Update service request
- `DELETE /api/service-requests/:id` - Delete service request

#### Service Offers
- `GET /api/service-requests/:id/offers` - List offers for request
- `POST /api/service-requests/:id/offers` - Create offer
- `PUT /api/service-offers/:id` - Update offer
- `DELETE /api/service-offers/:id` - Delete offer
- `POST /api/service-offers/:id/accept` - Accept offer

#### Comments
- `GET /api/posts/:id/comments` - Get comments for post
- `GET /api/service-requests/:id/comments` - Get comments for service request
- `POST /api/posts/:id/comments` - Add comment to post
- `POST /api/service-requests/:id/comments` - Add comment to service request
- `PUT /api/comments/:id` - Update comment
- `DELETE /api/comments/:id` - Delete comment

#### Ratings
- `GET /api/users/:id/ratings` - Get ratings for user (as provider)
- `POST /api/service-requests/:id/rating` - Rate completed service
- `PUT /api/ratings/:id` - Update rating
- `DELETE /api/ratings/:id` - Delete rating

## Development

### Prerequisites

- Go 1.24 or higher
- SQLite3

### Setup

1. Install dependencies:
```bash
go mod download
```

2. Run the application:
```bash
go run .
```

The database will be created automatically at `backend/commune.db`.

### Database Operations

#### View Database Schema
```bash
sqlite3 commune.db ".schema"
```

#### View All Tables
```bash
sqlite3 commune.db "SELECT name FROM sqlite_master WHERE type='table';"
```

#### Query Data
```bash
sqlite3 commune.db "SELECT * FROM users;"
sqlite3 commune.db "SELECT * FROM communities;"
```

#### Reset Database
To reset the database, simply delete the file and restart the application:
```bash
rm commune.db
go run .
```

## Testing

To test the database schema and relationships, you can create a simple test script or use the Go testing framework to verify:

- User creation and authentication
- Community creation and user membership
- Post creation and retrieval
- Service request workflow (request → offer → acceptance → rating)
- Comment creation and nesting
- Rating aggregation

## Security Considerations

### Implemented
- Password hashing (PasswordHash field in User model)
- Soft deletes for data recovery
- Role-based access control structure

### To Implement
- JWT or session-based authentication
- Authorization middleware for role checking
- Input validation and sanitization
- Rate limiting
- HTTPS/TLS
- CORS configuration
- Password strength requirements
- Account lockout after failed attempts

## Performance Considerations

### Current Optimizations
- Strategic indexes on foreign keys and frequently queried fields
- Indexes on status and category fields for filtering

### Future Optimizations
- Pagination for list endpoints
- Caching frequently accessed data
- Database connection pooling
- Query optimization for complex joins
- Consider PostgreSQL for production (better concurrent performance)

## Contributing

When adding new features:

1. Update models in `models.go`
2. Create migration in `main.go`
3. Add API handlers
4. Update this README
5. Update DATABASE.md if schema changes

## License

MIT
