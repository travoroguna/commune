# Database Schema Documentation

This document describes the database schema for the Community Marketplace application.

## Overview

The database is designed to support a multi-community marketplace where users can:
- Join multiple communities
- Post content within communities
- Request and offer services
- Comment on posts and services
- Rate service providers

## Entity Relationship Diagram (Text Description)

```
Users (1) ←→ (M) UserCommunities (M) ←→ (1) Communities
  ↓ (1:M)                                    ↓ (1:M)
  Posts                                      Posts
  ServiceRequests                            ServiceRequests
  ServiceOffers
  Comments
  Ratings (as rater)
  Ratings (as provider)
```

## Tables

### Users
Stores all user accounts with authentication and role information.

**Columns:**
- `id` (PK): Auto-incrementing primary key
- `created_at`: Timestamp of creation
- `updated_at`: Timestamp of last update
- `deleted_at`: Soft delete timestamp
- `name`: User's full name (required)
- `email`: User's email address (required, unique)
- `password_hash`: Hashed password (required)
- `role`: Global user role (required, default: 'user')
  - Possible values: `super_admin`, `admin`, `moderator`, `service_provider`, `user`
- `is_active`: Whether the user account is active (required, default: true)

**Relationships:**
- Has many Communities (through UserCommunities)
- Has many Posts
- Has many ServiceRequests (as requester)
- Has many ServiceOffers (as provider)
- Has many Comments
- Has many Ratings (as rater)
- Has many Ratings (as provider being rated)

**Indexes:**
- Unique index on `email`
- Index on `deleted_at` (for soft deletes)

### Communities
Represents physical communities like apartment complexes or estates.

**Columns:**
- `id` (PK): Auto-incrementing primary key
- `created_at`: Timestamp of creation
- `updated_at`: Timestamp of last update
- `deleted_at`: Soft delete timestamp
- `name`: Community name (required)
- `description`: Detailed description (text)
- `address`: Street address
- `city`: City name
- `state`: State/Province
- `country`: Country
- `zip_code`: Postal code
- `is_active`: Whether the community is active (required, default: true)

**Relationships:**
- Has many Users (through UserCommunities)
- Has many Posts
- Has many ServiceRequests

**Indexes:**
- Index on `deleted_at` (for soft deletes)

### UserCommunities
Junction table for many-to-many relationship between Users and Communities with additional metadata.

**Columns:**
- `user_id` (PK, FK): Reference to Users table
- `community_id` (PK, FK): Reference to Communities table
- `role`: User's role within this specific community (required, default: 'user')
  - Possible values: `admin`, `moderator`, `service_provider`, `user`
- `joined_at`: Timestamp when user joined the community
- `is_active`: Whether the membership is active (required, default: true)

**Relationships:**
- Belongs to User
- Belongs to Community

### Posts
Content posted by users within a community.

**Columns:**
- `id` (PK): Auto-incrementing primary key
- `created_at`: Timestamp of creation
- `updated_at`: Timestamp of last update
- `deleted_at`: Soft delete timestamp
- `title`: Post title (required)
- `content`: Post content (required, text)
- `author_id` (FK): Reference to Users table (required)
- `community_id` (FK): Reference to Communities table (required)
- `is_published`: Whether the post is published (required, default: true)
- `view_count`: Number of views (default: 0)

**Relationships:**
- Belongs to User (author)
- Belongs to Community
- Has many Comments

**Indexes:**
- Index on `author_id`
- Index on `community_id`
- Index on `deleted_at` (for soft deletes)

### ServiceRequests
Requests for services posted by community members.

**Columns:**
- `id` (PK): Auto-incrementing primary key
- `created_at`: Timestamp of creation
- `updated_at`: Timestamp of last update
- `deleted_at`: Soft delete timestamp
- `title`: Request title (required)
- `description`: Detailed description (required, text)
- `category`: Service category (indexed)
- `requester_id` (FK): Reference to Users table (required)
- `community_id` (FK): Reference to Communities table (required)
- `status`: Current status (required, default: 'open')
  - Possible values: `open`, `in_progress`, `completed`, `cancelled`
- `budget`: Budget for the service
- `accepted_offer_id` (FK): Reference to accepted ServiceOffer
- `completed_at`: Timestamp of completion

**Relationships:**
- Belongs to User (requester)
- Belongs to Community
- Has many ServiceOffers
- Has many Comments
- Has one AcceptedOffer (ServiceOffer)

**Indexes:**
- Index on `requester_id`
- Index on `community_id`
- Index on `category`
- Index on `status`
- Index on `accepted_offer_id`
- Index on `deleted_at` (for soft deletes)

### ServiceOffers
Offers from service providers for service requests.

**Columns:**
- `id` (PK): Auto-incrementing primary key
- `created_at`: Timestamp of creation
- `updated_at`: Timestamp of last update
- `deleted_at`: Soft delete timestamp
- `service_request_id` (FK): Reference to ServiceRequests table (required)
- `provider_id` (FK): Reference to Users table (required)
- `description`: Offer description (required, text)
- `proposed_price`: Price proposed by provider
- `estimated_duration`: Estimated time to complete
- `status`: Current status (required, default: 'pending')
  - Possible values: `pending`, `accepted`, `rejected`, `withdrawn`

**Relationships:**
- Belongs to ServiceRequest
- Belongs to User (provider)
- Has many Comments

**Indexes:**
- Index on `service_request_id`
- Index on `provider_id`
- Index on `deleted_at` (for soft deletes)

### Comments
Comments on posts, service requests, or service offers. Supports nested comments.

**Columns:**
- `id` (PK): Auto-incrementing primary key
- `created_at`: Timestamp of creation
- `updated_at`: Timestamp of last update
- `deleted_at`: Soft delete timestamp
- `content`: Comment content (required, text)
- `author_id` (FK): Reference to Users table (required)
- `post_id` (FK): Reference to Posts table (nullable)
- `service_request_id` (FK): Reference to ServiceRequests table (nullable)
- `service_offer_id` (FK): Reference to ServiceOffers table (nullable)
- `parent_comment_id` (FK): Reference to Comments table for nested replies (nullable)

**Relationships:**
- Belongs to User (author)
- Belongs to Post (optional)
- Belongs to ServiceRequest (optional)
- Belongs to ServiceOffer (optional)
- Belongs to ParentComment (optional)
- Has many Replies (Comments)

**Indexes:**
- Index on `author_id`
- Index on `post_id`
- Index on `service_request_id`
- Index on `service_offer_id`
- Index on `parent_comment_id`
- Index on `deleted_at` (for soft deletes)

### Ratings
Ratings and reviews for service providers after service completion.

**Columns:**
- `id` (PK): Auto-incrementing primary key
- `created_at`: Timestamp of creation
- `updated_at`: Timestamp of last update
- `deleted_at`: Soft delete timestamp
- `provider_id` (FK): Reference to Users table (provider being rated) (required)
- `rater_id` (FK): Reference to Users table (user giving rating) (required)
- `service_request_id` (FK): Reference to ServiceRequests table (required)
- `score`: Rating score 1-5 stars (required)
- `review`: Written review (text, optional)

**Relationships:**
- Belongs to User (provider)
- Belongs to User (rater)
- Belongs to ServiceRequest

**Indexes:**
- Index on `provider_id`
- Index on `rater_id`
- Index on `service_request_id`
- Index on `deleted_at` (for soft deletes)

## User Roles

### Global Roles (on User table)
- **super_admin**: Can create communities and manage all communities
- **admin**: Administrative privileges
- **moderator**: Can moderate content
- **service_provider**: Can offer services
- **user**: Regular user

### Community-Specific Roles (on UserCommunity table)
- **admin**: Administrator of the specific community
- **moderator**: Moderator of the specific community
- **service_provider**: Service provider within the specific community
- **user**: Regular member of the specific community

## Key Features

### Multi-Community Support
Users can join multiple communities through the UserCommunities junction table. Each membership can have a different role specific to that community.

### Service Marketplace
- Users can post service requests within their communities
- Service providers can offer services for these requests
- Requesters can accept offers
- After completion, requesters can rate and review providers

### Content & Engagement
- Users can post content within communities
- Comments can be added to posts, service requests, and service offers
- Comments support nested replies through the parent_comment_id

### Rating System
Service providers accumulate ratings from completed service requests, helping build reputation within the community.

## Soft Deletes

All main tables use GORM's soft delete feature through the `deleted_at` column. Records are not physically deleted but marked as deleted, allowing for data recovery and audit trails.

## Indexes

Strategic indexes are placed on:
- Foreign keys for efficient joins
- Status fields for filtering
- Category fields for filtering
- deleted_at for soft delete queries

## Migrations

The database uses gormigrate for versioned migrations. Current migrations:
- **202402041300**: Initial User table (legacy)
- **202402041301**: Complete community marketplace schema
