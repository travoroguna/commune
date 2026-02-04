import { createFileRoute, Navigate, Link } from '@tanstack/react-router';
import { useQuery } from '@tanstack/react-query';
import { useAuth } from '@/contexts/AuthContext';
import { communityApi } from '@/api/client';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Building2 } from 'lucide-react';

export const Route = createFileRoute('/_authenticated/communities/')({
  component: CommunitiesListPage,
});

function CommunitiesListPage() {
  const { user } = useAuth();

  const canManageCommunities = user?.Role === 'super_admin';

  const { data: communities, isLoading } = useQuery({
    queryKey: ['communities'],
    queryFn: communityApi.getAll,
    enabled: canManageCommunities,
  });

  if (!canManageCommunities) {
    return <Navigate to="/" />;
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Communities</h1>
          <p className="text-slate-600 mt-1">Manage all communities</p>
        </div>
        <Link to="/_authenticated/communities/new">
          <Button>
            <Building2 className="mr-2 h-4 w-4" />
            Create Community
          </Button>
        </Link>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>All Communities</CardTitle>
          <CardDescription>View and manage all communities</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8">Loading communities...</div>
          ) : communities && communities.length > 0 ? (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {communities.map((community) => (
                <Link
                  key={community.ID}
                  to="/_authenticated/communities/$communityId"
                  params={{ communityId: community.ID.toString() }}
                >
                  <Card className="hover:shadow-md transition-shadow cursor-pointer">
                    <CardHeader>
                      <CardTitle className="flex items-center gap-2">
                        <Building2 className="h-5 w-5" />
                        {community.Name}
                      </CardTitle>
                      <CardDescription className="line-clamp-2">
                        {community.Description || 'No description'}
                      </CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-1 text-sm text-slate-600">
                        {community.Slug && (
                          <div>Slug: <span className="font-mono">{community.Slug}</span></div>
                        )}
                        {community.Subdomain && (
                          <div>Subdomain: <span className="font-mono">{community.Subdomain}</span></div>
                        )}
                        <div>
                          <span
                            className={`inline-flex px-2 py-1 text-xs rounded-full ${
                              community.IsActive
                                ? 'bg-green-100 text-green-800'
                                : 'bg-red-100 text-red-800'
                            }`}
                          >
                            {community.IsActive ? 'Active' : 'Inactive'}
                          </span>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </Link>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-slate-600">
              No communities found. Create your first community to get started.
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

