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
import { Button } from '@/components/ui/button';
import { Link } from '@tanstack/react-router';
import { UserPlus } from 'lucide-react';

export const Route = createFileRoute('/_authenticated/users/')({
  component: UsersListPage,
});

function UsersListPage() {
  const { user } = useAuth();

  // Check permissions
  const canManageUsers = user?.Role === 'super_admin' || user?.Role === 'admin';

  const { data: users, isLoading } = useQuery({
    queryKey: ['users'],
    queryFn: userApi.getAll,
    enabled: canManageUsers,
  });

  if (!canManageUsers) {
    return <Navigate to="/" />;
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Users</h1>
          <p className="text-slate-600 mt-1">Manage user accounts</p>
        </div>
        <Link to="/users/new">
          <Button>
            <UserPlus className="mr-2 h-4 w-4" />
            Create User
          </Button>
        </Link>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>All Users</CardTitle>
          <CardDescription>View and manage all user accounts</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8">Loading users...</div>
          ) : users && users.length > 0 ? (
            <div className="space-y-2">
              <div className="grid grid-cols-4 font-semibold text-sm text-slate-600 pb-2 border-b">
                <div>Name</div>
                <div>Email</div>
                <div>Role</div>
                <div>Status</div>
              </div>
              {users.map((u) => (
                <Link
                  key={u.ID}
                  to="/users/$userId"
                  params={{ userId: u.ID.toString() }}
                  className="grid grid-cols-4 py-3 hover:bg-slate-50 rounded-lg px-2 -mx-2 transition-colors"
                >
                  <div className="font-medium">{u.Name}</div>
                  <div className="text-slate-600">{u.Email}</div>
                  <div className="text-slate-600">
                    {u.Role.replace('_', ' ')}
                  </div>
                  <div>
                    <span
                      className={`inline-flex px-2 py-1 text-xs rounded-full ${
                        u.IsActive
                          ? 'bg-green-100 text-green-800'
                          : 'bg-red-100 text-red-800'
                      }`}
                    >
                      {u.IsActive ? 'Active' : 'Inactive'}
                    </span>
                  </div>
                </Link>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-slate-600">
              No users found. Create your first user to get started.
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

