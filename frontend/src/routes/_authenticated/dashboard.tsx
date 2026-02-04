import { createFileRoute, Link } from '@tanstack/react-router';
import { useAuth } from '@/contexts/AuthContext';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Users, Building2, UserPlus, Mail, Briefcase } from 'lucide-react';

export const Route = createFileRoute('/_authenticated/dashboard')({
  component: DashboardPage,
});

function DashboardPage() {
  const { user, currentCommunity } = useAuth();

  const canManageUsers = user?.Role === 'super_admin' || user?.Role === 'admin';

  return (
    <div className="space-y-8">
      {/* Welcome Section */}
      <div>
        <h1 className="text-3xl font-bold text-slate-900">
          Welcome back, {user?.Name}!
        </h1>
        <p className="mt-2 text-slate-600">
          {currentCommunity
            ? `You're currently in ${currentCommunity.Name}`
            : 'Manage your communities and users'}
        </p>
      </div>

      {/* Quick Stats */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Role</CardTitle>
            <Users className="h-4 w-4 text-slate-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {user?.Role.replace('_', ' ').toUpperCase()}
            </div>
            <p className="text-xs text-muted-foreground">Your access level</p>
          </CardContent>
        </Card>

        {currentCommunity && (
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Community</CardTitle>
              <Building2 className="h-4 w-4 text-slate-600" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{currentCommunity.Name}</div>
              <p className="text-xs text-muted-foreground">Current context</p>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Quick Actions */}
      <div>
        <h2 className="text-xl font-semibold mb-4">Quick Actions</h2>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {canManageUsers && (
            <Card className="hover:shadow-md transition-shadow cursor-pointer">
              <Link to="/users/new">
                <CardHeader>
                  <div className="flex items-center gap-3">
                    <UserPlus className="h-8 w-8 text-blue-600" />
                    <div>
                      <CardTitle>Create User</CardTitle>
                      <CardDescription>Add a new user to the system</CardDescription>
                    </div>
                  </div>
                </CardHeader>
              </Link>
            </Card>
          )}

          {user?.Role === 'super_admin' && (
            <Card className="hover:shadow-md transition-shadow cursor-pointer">
              <Link to="/communities/new">
                <CardHeader>
                  <div className="flex items-center gap-3">
                    <Building2 className="h-8 w-8 text-green-600" />
                    <div>
                      <CardTitle>Create Community</CardTitle>
                      <CardDescription>Add a new community</CardDescription>
                    </div>
                  </div>
                </CardHeader>
              </Link>
            </Card>
          )}

          {canManageUsers && (
            <Card className="hover:shadow-md transition-shadow cursor-pointer">
              <Link to="/join-requests">
                <CardHeader>
                  <div className="flex items-center gap-3">
                    <Mail className="h-8 w-8 text-purple-600" />
                    <div>
                      <CardTitle>Join Requests</CardTitle>
                      <CardDescription>Review pending requests</CardDescription>
                    </div>
                  </div>
                </CardHeader>
              </Link>
            </Card>
          )}

          <Card className="hover:shadow-md transition-shadow cursor-pointer">
            <Link to="/services">
              <CardHeader>
                <div className="flex items-center gap-3">
                  <Briefcase className="h-8 w-8 text-indigo-600" />
                  <div>
                    <CardTitle>Browse Services</CardTitle>
                    <CardDescription>View available services</CardDescription>
                  </div>
                </div>
              </CardHeader>
            </Link>
          </Card>

          <Card className="hover:shadow-md transition-shadow cursor-pointer">
            <Link to="/profile">
              <CardHeader>
                <div className="flex items-center gap-3">
                  <Users className="h-8 w-8 text-orange-600" />
                  <div>
                    <CardTitle>My Profile</CardTitle>
                    <CardDescription>Manage your account settings</CardDescription>
                  </div>
                </div>
              </CardHeader>
            </Link>
          </Card>
        </div>
      </div>
    </div>
  );
}

