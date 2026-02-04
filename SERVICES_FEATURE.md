# Services Feature Documentation

## Overview
The Services feature provides a main page where users can search and view available service requests within their community. This includes filtering by category, status, and search terms.

## Backend API Endpoints

### 1. List Services
**Endpoint:** `GET /api/services`

**Query Parameters:**
- `search` - Search in title and description
- `category` - Filter by service category
- `status` - Filter by status (open, in_progress, completed, cancelled)
- `community_id` - Filter by community

**Response:** Array of ServiceRequest objects with preloaded Requester, Community, and ServiceOffers

**Example:**
```bash
curl http://localhost:3000/api/services?status=open&community_id=1
```

### 2. Get Service by ID
**Endpoint:** `GET /api/services/{id}`

**Response:** Single ServiceRequest object with all relationships loaded

**Example:**
```bash
curl http://localhost:3000/api/services/1
```

### 3. Create Service Request
**Endpoint:** `POST /api/services`

**Authentication:** Required (X-User-ID header)

**Body:**
```json
{
  "title": "Need plumber for kitchen sink",
  "description": "Kitchen sink is leaking...",
  "category": "Plumbing",
  "community_id": 1,
  "budget": 150.00
}
```

### 4. Update Service Request
**Endpoint:** `PUT /api/services/{id}`

**Authentication:** Required (Requester or Admin)

**Body:** Partial updates supported for title, description, category, status, budget

### 5. Delete Service Request
**Endpoint:** `DELETE /api/services/{id}`

**Authentication:** Required (Requester or Admin)

## Frontend Components

### Services Page
**Location:** `/frontend/src/routes/_authenticated/services.tsx`

**Features:**
- Real-time search across title and description
- Category filter dropdown
- Status filter dropdown  
- Responsive grid layout
- Service cards with:
  - Title and status badge
  - Category
  - Description (truncated to 3 lines)
  - Community name with icon
  - Requester name with icon
  - Budget formatted as currency
  - Posted date
  - Number of offers received

**Navigation:**
- Added to main navigation menu
- Quick action card on dashboard

## TypeScript Types

### ServiceRequest
```typescript
interface ServiceRequest {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  Title: string;
  Description: string;
  Category?: string;
  RequesterID: number;
  CommunityID: number;
  Status: 'open' | 'in_progress' | 'completed' | 'cancelled';
  Budget?: number;
  Requester?: User;
  Community?: Community;
  ServiceOffers?: ServiceOffer[];
}
```

### ServiceOffer
```typescript
interface ServiceOffer {
  ID: number;
  ServiceRequestID: number;
  ProviderID: number;
  Description: string;
  ProposedPrice?: number;
  EstimatedDuration?: string;
  Status: 'pending' | 'accepted' | 'rejected' | 'withdrawn';
  Provider?: User;
}
```

## Database Models

The feature uses these existing database models:
- **ServiceRequest** - Main service request entity
- **ServiceOffer** - Offers from service providers
- **User** - Requester and Provider information
- **Community** - Community association
- **Comment** - Comments on service requests (future use)

## Test Data

A test data script is provided in `backend/testdata.go` that creates:
- 1 Community (Sunset Apartments)
- 8 Service Requests across various categories:
  - Plumbing
  - Electrical
  - Cleaning
  - HVAC
  - Painting
  - Security
  - Appliance Repair
  - Pest Control

**Run test data:**
```bash
cd backend
go run testdata.go models.go helpers.go
```

## API Test Results

All API endpoints verified and working:

✅ List all services: 8 results  
✅ Filter by status (open): 6 results  
✅ Filter by category (Plumbing): 1 result  
✅ Search for "electrician": 1 result  
✅ Filter by community: 8 results  
✅ Get single service: Full object with relationships  
✅ Combined filters (open + Electrical): 1 result  

## Status Colors

The UI uses color-coded badges for service status:
- **Open** - Green (bg-green-100 text-green-800)
- **In Progress** - Blue (bg-blue-100 text-blue-800)
- **Completed** - Gray (bg-gray-100 text-gray-800)
- **Cancelled** - Red (bg-red-100 text-red-800)

## Future Enhancements

Potential improvements for the services feature:
1. Service detail page with full information
2. Create/edit service request forms
3. Service offer creation and management
4. Real-time updates for new services
5. Pagination for large result sets
6. Advanced filtering (date ranges, budget ranges)
7. Sorting options (newest, budget, popularity)
8. Map view for location-based services
9. Service request templates
10. Notification system for new offers

## Files Modified/Created

**Backend:**
- `backend/services.go` (New) - Service API handlers
- `backend/main.go` (Modified) - Added service routes
- `backend/testdata.go` (New) - Test data generator
- `backend/demo.html` (New) - Standalone demo page

**Frontend:**
- `frontend/src/routes/_authenticated/services.tsx` (New) - Services page component
- `frontend/src/routes/_authenticated.tsx` (Modified) - Added Services nav link
- `frontend/src/routes/_authenticated/dashboard.tsx` (Modified) - Added Browse Services card
- `frontend/src/types/index.ts` (Modified) - Added ServiceRequest, ServiceOffer, Comment types
