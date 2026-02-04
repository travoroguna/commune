import { createFileRoute, Outlet, Navigate, Link } from '@tanstack/react-router';
import { useAuth } from '@/contexts/AuthContext';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { ChevronDown, Building2, LogOut, User as UserIcon, Settings } from 'lucide-react';

export const Route = createFileRoute('/_authenticated')({
  component: AuthenticatedLayout,
});

function AuthenticatedLayout() {
  const { user, currentCommunity, userCommunities, logout, switchCommunity } = useAuth();

  // Redirect to login if not authenticated
  if (!user) {
    return <Navigate to="/login" />;
  }

  const canManageUsers = user.Role === 'super_admin' || user.Role === 'admin';

  return (
    <div className="min-h-screen bg-slate-50">
      {/* Header */}
      <header className="bg-white border-b sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center space-x-8">
              <Link to="/" className="text-xl font-bold">
                Commune
              </Link>

              <nav className="flex space-x-4">
                <Link
                  to="/dashboard"
                  className="text-sm font-medium text-slate-600 hover:text-slate-900"
                >
                  Dashboard
                </Link>
                <Link
                  to="/services"
                  className="text-sm font-medium text-slate-600 hover:text-slate-900"
                >
                  Services
                </Link>
                {canManageUsers && (
                  <Link
                    to="/users"
                    className="text-sm font-medium text-slate-600 hover:text-slate-900"
                  >
                    Users
                  </Link>
                )}
                {user.Role === 'super_admin' && (
                  <Link
                    to="/communities"
                    className="text-sm font-medium text-slate-600 hover:text-slate-900"
                  >
                    Communities
                  </Link>
                )}
              </nav>
            </div>

            <div className="flex items-center space-x-4">
              {/* Community Switcher (Super Admin only) */}
              {user.Role === 'super_admin' && userCommunities.length > 0 && (
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="outline" className="gap-2">
                      <Building2 className="h-4 w-4" />
                      {currentCommunity?.Name || 'Select Community'}
                      <ChevronDown className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-56">
                    <DropdownMenuLabel>Switch Community</DropdownMenuLabel>
                    <DropdownMenuSeparator />
                    {userCommunities.map((uc) => (
                      <DropdownMenuItem
                        key={uc.CommunityID}
                        onClick={() => switchCommunity(uc.CommunityID)}
                        className={
                          currentCommunity?.ID === uc.CommunityID
                            ? 'bg-slate-100'
                            : ''
                        }
                      >
                        {uc.Community?.Name}
                      </DropdownMenuItem>
                    ))}
                  </DropdownMenuContent>
                </DropdownMenu>
              )}

              {/* User Menu */}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" className="gap-2">
                    <UserIcon className="h-4 w-4" />
                    {user.Name}
                    <ChevronDown className="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-56">
                  <DropdownMenuLabel>
                    <div className="flex flex-col">
                      <span>{user.Name}</span>
                      <span className="text-xs text-slate-500 font-normal">
                        {user.Email}
                      </span>
                      <span className="text-xs text-slate-500 font-normal mt-1">
                        Role: {user.Role.replace('_', ' ')}
                      </span>
                    </div>
                  </DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem asChild>
                    <Link to="/profile" className="flex items-center gap-2">
                      <Settings className="h-4 w-4" />
                      Profile Settings
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={logout} className="flex items-center gap-2">
                    <LogOut className="h-4 w-4" />
                    Logout
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <Outlet />
      </main>
    </div>
  );
}

