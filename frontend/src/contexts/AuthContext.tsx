import { createContext, useContext, useState, useEffect } from 'react';
import type { ReactNode } from 'react';
import type { User, Community, UserCommunity, AuthContextType } from '@/types';
import { authApi, userApi } from '@/api/client';

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [currentCommunity, setCurrentCommunity] = useState<Community | null>(null);
  const [userCommunities, setUserCommunities] = useState<UserCommunity[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Load user on mount
  useEffect(() => {
    loadUser();
  }, []);

  // Load current community from localStorage when user changes
  useEffect(() => {
    if (user && userCommunities.length > 0) {
      const savedCommunityId = localStorage.getItem('currentCommunityId');
      if (savedCommunityId) {
        const community = userCommunities.find(
          (uc) => uc.CommunityID === parseInt(savedCommunityId)
        );
        if (community?.Community) {
          setCurrentCommunity(community.Community);
        } else {
          // Default to first community
          setCurrentCommunity(userCommunities[0]?.Community || null);
        }
      } else {
        // Default to first community
        setCurrentCommunity(userCommunities[0]?.Community || null);
      }
    }
  }, [user, userCommunities]);

  const loadUser = async () => {
    try {
      const currentUser = await authApi.getCurrentUser();
      setUser(currentUser);

      // Load user's communities
      const communities = await userApi.getUserCommunities(currentUser.ID);
      setUserCommunities(communities);
    } catch (error) {
      // Not authenticated
      setUser(null);
      setUserCommunities([]);
      setCurrentCommunity(null);
    } finally {
      setIsLoading(false);
    }
  };

  const login = async (email: string, password: string) => {
    const { user: loggedInUser } = await authApi.login(email, password);
    setUser(loggedInUser);

    // Load user's communities
    const communities = await userApi.getUserCommunities(loggedInUser.ID);
    setUserCommunities(communities);
  };

  const logout = async () => {
    await authApi.logout();
    setUser(null);
    setUserCommunities([]);
    setCurrentCommunity(null);
    localStorage.removeItem('currentCommunityId');
  };

  const switchCommunity = (communityId: number) => {
    const community = userCommunities.find((uc) => uc.CommunityID === communityId);
    if (community?.Community) {
      setCurrentCommunity(community.Community);
      localStorage.setItem('currentCommunityId', communityId.toString());
    }
  };

  const refreshUser = async () => {
    await loadUser();
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        currentCommunity,
        userCommunities,
        isLoading,
        login,
        logout,
        switchCommunity,
        refreshUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
