import type { User, UserRole, Community, UserCommunity, JoinRequest, ServiceRequest, ServiceOffer } from '@/types';

const API_BASE = '/api';

// Helper to handle API responses
async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'An error occurred' }));
    throw new Error(error.message || `HTTP ${response.status}`);
  }
  return response.json();
}

// Auth APIs
export const authApi = {
  async login(email: string, password: string): Promise<{ user: User; token: string }> {
    const response = await fetch(`${API_BASE}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async logout(): Promise<void> {
    const response = await fetch(`${API_BASE}/auth/logout`, {
      method: 'POST',
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async getCurrentUser(): Promise<User> {
    const response = await fetch(`${API_BASE}/auth/me`, {
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async checkFirstBoot(): Promise<{ needsSetup: boolean }> {
    const response = await fetch(`${API_BASE}/auth/first-boot`);
    return handleResponse(response);
  },

  async createSuperUser(data: {
    name: string;
    email: string;
    password: string;
  }): Promise<{ user: User; token: string }> {
    const response = await fetch(`${API_BASE}/auth/setup-super-user`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
      credentials: 'include',
    });
    return handleResponse(response);
  },
};

// User APIs
export const userApi = {
  async getAll(): Promise<User[]> {
    const response = await fetch(`${API_BASE}/users`, {
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async getById(id: number): Promise<User> {
    const response = await fetch(`${API_BASE}/users/${id}`, {
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async create(data: {
    name: string;
    email: string;
    password: string;
    role: UserRole;
  }): Promise<User> {
    const response = await fetch(`${API_BASE}/users`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async update(id: number, data: Partial<User>): Promise<User> {
    const response = await fetch(`${API_BASE}/users/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async delete(id: number): Promise<void> {
    const response = await fetch(`${API_BASE}/users/${id}`, {
      method: 'DELETE',
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async changePassword(oldPassword: string, newPassword: string): Promise<void> {
    const response = await fetch(`${API_BASE}/users/change-password`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ oldPassword, newPassword }),
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async getUserCommunities(userId: number): Promise<UserCommunity[]> {
    const response = await fetch(`${API_BASE}/users/${userId}/communities`, {
      credentials: 'include',
    });
    return handleResponse(response);
  },
};

// Community APIs
export const communityApi = {
  async getAll(): Promise<Community[]> {
    const response = await fetch(`${API_BASE}/communities`, {
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async getById(id: number): Promise<Community> {
    const response = await fetch(`${API_BASE}/communities/${id}`, {
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async create(data: Partial<Community>): Promise<Community> {
    const response = await fetch(`${API_BASE}/communities`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async update(id: number, data: Partial<Community>): Promise<Community> {
    const response = await fetch(`${API_BASE}/communities/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async delete(id: number): Promise<void> {
    const response = await fetch(`${API_BASE}/communities/${id}`, {
      method: 'DELETE',
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async getMembers(communityId: number): Promise<UserCommunity[]> {
    const response = await fetch(`${API_BASE}/communities/${communityId}/members`, {
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async addMember(
    communityId: number,
    userId: number,
    role: UserRole
  ): Promise<UserCommunity> {
    const response = await fetch(`${API_BASE}/communities/${communityId}/members`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ userId, role }),
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async removeMember(communityId: number, userId: number): Promise<void> {
    const response = await fetch(
      `${API_BASE}/communities/${communityId}/members/${userId}`,
      {
        method: 'DELETE',
        credentials: 'include',
      }
    );
    return handleResponse(response);
  },

  async updateMemberRole(
    communityId: number,
    userId: number,
    role: UserRole
  ): Promise<UserCommunity> {
    const response = await fetch(
      `${API_BASE}/communities/${communityId}/members/${userId}`,
      {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ role }),
        credentials: 'include',
      }
    );
    return handleResponse(response);
  },
};

// Join Request APIs
export const joinRequestApi = {
  async getAll(): Promise<JoinRequest[]> {
    const response = await fetch(`${API_BASE}/join-requests`, {
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async getByCommunity(communityId: number): Promise<JoinRequest[]> {
    const response = await fetch(`${API_BASE}/communities/${communityId}/join-requests`, {
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async create(communityId: number, message?: string): Promise<JoinRequest> {
    const response = await fetch(`${API_BASE}/join-requests`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ communityId, message }),
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async approve(requestId: number, role: UserRole = 'user'): Promise<JoinRequest> {
    const response = await fetch(`${API_BASE}/join-requests/${requestId}/approve`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ role }),
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async reject(requestId: number): Promise<JoinRequest> {
    const response = await fetch(`${API_BASE}/join-requests/${requestId}/reject`, {
      method: 'POST',
      credentials: 'include',
    });
    return handleResponse(response);
  },
};

// Service Request APIs
export const serviceRequestApi = {
  async getAll(params?: { community_id?: number; status?: string }): Promise<ServiceRequest[]> {
    const queryParams = new URLSearchParams();
    if (params?.community_id) queryParams.append('community_id', params.community_id.toString());
    if (params?.status) queryParams.append('status', params.status);
    
    const response = await fetch(`${API_BASE}/service-requests?${queryParams.toString()}`, {
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async getById(id: number): Promise<ServiceRequest> {
    const response = await fetch(`${API_BASE}/service-requests/${id}`, {
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async create(data: {
    title: string;
    description: string;
    category?: string;
    budget?: number;
    community_id: number;
  }): Promise<ServiceRequest> {
    const response = await fetch(`${API_BASE}/service-requests`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async update(id: number, data: Partial<ServiceRequest>): Promise<ServiceRequest> {
    const response = await fetch(`${API_BASE}/service-requests/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async acceptOffer(requestId: number, offerId: number): Promise<ServiceRequest> {
    const response = await fetch(`${API_BASE}/service-requests/${requestId}/accept-offer`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ offer_id: offerId }),
      credentials: 'include',
    });
    return handleResponse(response);
  },
};

// Service Offer APIs
export const serviceOfferApi = {
  async getAll(params?: { provider_id?: number; status?: string }): Promise<ServiceOffer[]> {
    const queryParams = new URLSearchParams();
    if (params?.provider_id) queryParams.append('provider_id', params.provider_id.toString());
    if (params?.status) queryParams.append('status', params.status);
    
    const response = await fetch(`${API_BASE}/service-offers?${queryParams.toString()}`, {
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async getById(id: number): Promise<ServiceOffer> {
    const response = await fetch(`${API_BASE}/service-offers/${id}`, {
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async create(data: {
    service_request_id: number;
    description: string;
    proposed_price?: number;
    estimated_duration?: string;
  }): Promise<ServiceOffer> {
    const response = await fetch(`${API_BASE}/service-offers`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async withdraw(id: number): Promise<ServiceOffer> {
    const response = await fetch(`${API_BASE}/service-offers/${id}/withdraw`, {
      method: 'POST',
      credentials: 'include',
    });
    return handleResponse(response);
  },

  async getAcceptedOffers(params?: { provider_id?: number }): Promise<ServiceOffer[]> {
    const queryParams = new URLSearchParams();
    queryParams.append('status', 'accepted');
    if (params?.provider_id) queryParams.append('provider_id', params.provider_id.toString());
    
    const response = await fetch(`${API_BASE}/service-offers?${queryParams.toString()}`, {
      credentials: 'include',
    });
    return handleResponse(response);
  },
};
