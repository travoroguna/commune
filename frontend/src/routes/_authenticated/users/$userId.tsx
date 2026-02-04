import { createFileRoute, Navigate } from '@tanstack/react-router';
import { useQuery } from '@tanstack/react-query';
import { useAuth } from '@/contexts/AuthContext';
import { userApi } from '@/api/client';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';

export const Route = createFileRoute('/_authenticated/users/$userId')({
  component: UserDetailPage,
});

function UserDetailPage() {
  const { userId } = Route.useParams();
  const { user } = useAuth();

  const canManageUsers = user?.Role === 'super_admin' || user?.Role === 'admin';

  const { data: userData, isLoading } = useQuery({
    queryKey: ['user', userId],
    queryFn: () => userApi.getAll().then(users => users.find(u => u.ID.toString() === userId)),
    enabled: canManageUsers,
  });

  if (!canManageUsers) {
    return <Navigate to="/" />;
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">User Details</h1>
        <p className="text-slate-600 mt-1">View user information</p>
      </div>

      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle>User Information</CardTitle>
          <CardDescription>Details for user #{userId}</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8">Loading user...</div>
          ) : userData ? (
            <div className="space-y-4">
              <div>
                <div className="text-sm font-medium text-slate-600">Name</div>
                <div className="text-lg">{userData.Name}</div>
              </div>
              <div>
                <div className="text-sm font-medium text-slate-600">Email</div>
                <div className="text-lg">{userData.Email}</div>
              </div>
              <div>
                <div className="text-sm font-medium text-slate-600">Role</div>
                <div className="text-lg">{userData.Role.replace('_', ' ')}</div>
              </div>
              <div>
                <div className="text-sm font-medium text-slate-600">Status</div>
                <div>
                  <span
                    className={`inline-flex px-2 py-1 text-xs rounded-full ${
                      userData.IsActive
                        ? 'bg-green-100 text-green-800'
                        : 'bg-red-100 text-red-800'
                    }`}
                  >
                    {userData.IsActive ? 'Active' : 'Inactive'}
                  </span>
                </div>
              </div>
            </div>
          ) : (
            <div className="text-center py-8 text-slate-600">User not found</div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
