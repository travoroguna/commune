import { createFileRoute, Navigate } from '@tanstack/react-router';
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

export const Route = createFileRoute('/_authenticated/communities/$communityId')({
  component: CommunityDetailPage,
});

function CommunityDetailPage() {
  const { communityId } = Route.useParams();
  const { user } = useAuth();

  const canManageCommunities = user?.Role === 'super_admin';

  const { data: community, isLoading } = useQuery({
    queryKey: ['community', communityId],
    queryFn: () => communityApi.getAll().then(communities => communities.find(c => c.ID.toString() === communityId)),
    enabled: canManageCommunities,
  });

  if (!canManageCommunities) {
    return <Navigate to="/" />;
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Community Details</h1>
        <p className="text-slate-600 mt-1">View community information</p>
      </div>

      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle>Community Information</CardTitle>
          <CardDescription>Details for community #{communityId}</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8">Loading community...</div>
          ) : community ? (
            <div className="space-y-4">
              <div>
                <div className="text-sm font-medium text-slate-600">Name</div>
                <div className="text-lg">{community.Name}</div>
              </div>
              {community.Description && (
                <div>
                  <div className="text-sm font-medium text-slate-600">Description</div>
                  <div className="text-lg">{community.Description}</div>
                </div>
              )}
              {community.Slug && (
                <div>
                  <div className="text-sm font-medium text-slate-600">Slug</div>
                  <div className="text-lg font-mono">{community.Slug}</div>
                </div>
              )}
              {community.Subdomain && (
                <div>
                  <div className="text-sm font-medium text-slate-600">Subdomain</div>
                  <div className="text-lg font-mono">{community.Subdomain}</div>
                </div>
              )}
              <div>
                <div className="text-sm font-medium text-slate-600">Status</div>
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
              {community.Address && (
                <div>
                  <div className="text-sm font-medium text-slate-600">Location</div>
                  <div className="text-lg">
                    {community.Address}
                    {community.City && `, ${community.City}`}
                    {community.State && `, ${community.State}`}
                    {community.ZipCode && ` ${community.ZipCode}`}
                    {community.Country && `, ${community.Country}`}
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="text-center py-8 text-slate-600">Community not found</div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
