import { createFileRoute, Navigate } from '@tanstack/react-router';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuth } from '@/contexts/AuthContext';
import { joinRequestApi } from '@/api/client';
import type { UserRole } from '@/types';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Check, X } from 'lucide-react';
import { useState } from 'react';

export const Route = createFileRoute('/_authenticated/join-requests')({
  component: JoinRequestsPage,
});

function JoinRequestsPage() {
  const { user } = useAuth();
  const queryClient = useQueryClient();

  const canManageRequests = user?.Role === 'super_admin' || user?.Role === 'admin' || user?.Role === 'moderator';

  const { data: requests, isLoading } = useQuery({
    queryKey: ['join-requests'],
    queryFn: joinRequestApi.getAll,
    enabled: canManageRequests,
  });

  const approveMutation = useMutation({
    mutationFn: ({ requestId, role }: { requestId: number; role: UserRole }) =>
      joinRequestApi.approve(requestId, role),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['join-requests'] });
    },
  });

  const rejectMutation = useMutation({
    mutationFn: (requestId: number) => joinRequestApi.reject(requestId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['join-requests'] });
    },
  });

  if (!canManageRequests) {
    return <Navigate to="/" />;
  }

  const pendingRequests = requests?.filter((r) => r.Status === 'pending') || [];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Join Requests</h1>
        <p className="text-slate-600 mt-1">Review and manage community join requests</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Pending Requests</CardTitle>
          <CardDescription>
            {pendingRequests.length} request(s) awaiting review
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8">Loading requests...</div>
          ) : pendingRequests.length > 0 ? (
            <div className="space-y-4">
              {pendingRequests.map((request) => (
                <JoinRequestCard
                  key={request.ID}
                  request={request}
                  onApprove={(role) => approveMutation.mutate({ requestId: request.ID, role })}
                  onReject={() => rejectMutation.mutate(request.ID)}
                  isProcessing={approveMutation.isPending || rejectMutation.isPending}
                />
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-slate-600">
              No pending join requests
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

interface JoinRequestCardProps {
  request: any;
  onApprove: (role: UserRole) => void;
  onReject: () => void;
  isProcessing: boolean;
}

function JoinRequestCard({ request, onApprove, onReject, isProcessing }: JoinRequestCardProps) {
  const [selectedRole, setSelectedRole] = useState<UserRole>('user');

  return (
    <div className="border rounded-lg p-4">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <div className="font-semibold">{request.User?.Name || 'Unknown User'}</div>
          <div className="text-sm text-slate-600">{request.User?.Email}</div>
          <div className="text-sm text-slate-600 mt-1">
            Community: {request.Community?.Name || 'Unknown'}
          </div>
          {request.Message && (
            <div className="mt-2 text-sm bg-slate-50 p-2 rounded">
              <span className="font-medium">Message:</span> {request.Message}
            </div>
          )}
          <div className="text-xs text-slate-500 mt-2">
            Requested: {new Date(request.CreatedAt).toLocaleDateString()}
          </div>
        </div>

        <div className="flex gap-2 items-start ml-4">
          <div className="w-40">
            <Select
              value={selectedRole}
              onValueChange={(value) => setSelectedRole(value as UserRole)}
              disabled={isProcessing}
            >
              <SelectTrigger className="h-9">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="user">User</SelectItem>
                <SelectItem value="service_provider">Service Provider</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <Button
            size="sm"
            onClick={() => onApprove(selectedRole)}
            disabled={isProcessing}
            className="gap-1"
          >
            <Check className="h-4 w-4" />
            Approve
          </Button>
          <Button
            size="sm"
            variant="destructive"
            onClick={onReject}
            disabled={isProcessing}
            className="gap-1"
          >
            <X className="h-4 w-4" />
            Reject
          </Button>
        </div>
      </div>
    </div>
  );
}

