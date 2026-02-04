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
