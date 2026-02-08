// User roles matching backend
export type UserRole = 'super_admin' | 'admin' | 'moderator' | 'service_provider' | 'user';

// User type
export interface User {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt?: string;
  Name: string;
  Email: string;
  Role: UserRole;
  IsActive: boolean;
}

// Community type
export interface Community {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt?: string;
  Name: string;
  Slug: string;
  Description?: string;
  Subdomain?: string;
  CustomDomain?: string;
  Address?: string;
  City?: string;
  State?: string;
  Country?: string;
  ZipCode?: string;
  IsActive: boolean;
}

// UserCommunity type
export interface UserCommunity {
  UserID: number;
  CommunityID: number;
  Role: UserRole;
  JoinedAt: string;
  IsActive: boolean;
  Community?: Community;
}

// Join request type
export interface JoinRequest {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  UserID: number;
  CommunityID: number;
  Status: 'pending' | 'approved' | 'rejected';
  Message?: string;
  User?: User;
  Community?: Community;
}

// Service status type
export type ServiceStatus = 'open' | 'in_progress' | 'completed' | 'cancelled';

// Service request status type (alias for ServiceStatus)
export type ServiceRequestStatus = ServiceStatus;

// Service offer status type
export type ServiceOfferStatus = 'pending' | 'accepted' | 'rejected' | 'withdrawn';

// Common service categories
export type ServiceCategory = 
  | 'Plumbing' 
  | 'Electrical' 
  | 'Cleaning' 
  | 'HVAC' 
  | 'Painting' 
  | 'Security' 
  | 'Appliance Repair' 
  | 'Pest Control'
  | 'Landscaping'
  | 'Carpentry'
  | 'Other';

// Service Request type
export interface ServiceRequest {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt?: string;
  Title: string;
  Description: string;
  Category?: ServiceCategory | string;
  Category?: ServiceCategory;
  RequesterID: number;
  CommunityID: number;
  Status: ServiceStatus;
  Budget?: number;
  AcceptedOfferID?: number;
  CompletedAt?: string;
  Requester?: User;
  Community?: Community;
  ServiceOffers?: ServiceOffer[];
  AcceptedOffer?: ServiceOffer;
  Comments?: Comment[];
}

  Comments?: Comment[];
}

// Service status type
export type ServiceStatus = 'open' | 'in_progress' | 'completed' | 'cancelled';

// Common service categories
export type ServiceCategory = 
  | 'Plumbing' 
  | 'Electrical' 
  | 'Cleaning' 
  | 'HVAC' 
  | 'Painting' 
  | 'Security' 
  | 'Appliance Repair' 
  | 'Pest Control'
  | 'Landscaping'
  | 'Carpentry'
  | 'Other';

// Service Offer type
export interface ServiceOffer {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt?: string;
  ServiceRequestID: number;
  ProviderID: number;
  Description: string;
  ProposedPrice?: number;
  EstimatedDuration?: string;
  Status: ServiceOfferStatus;
  ServiceRequest?: ServiceRequest;
  Status: 'pending' | 'accepted' | 'rejected' | 'withdrawn';
  Provider?: User;
}

// Comment type
export interface Comment {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  Content: string;
  AuthorID: number;
  Author?: User;
  PostID?: number;
  ServiceRequestID?: number;
  ServiceOfferID?: number;
  ParentCommentID?: number;
}

// Auth context type
export interface AuthContextType {
  user: User | null;
  currentCommunity: Community | null;
  userCommunities: UserCommunity[];
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  switchCommunity: (communityId: number) => void;
  refreshUser: () => Promise<void>;
}
