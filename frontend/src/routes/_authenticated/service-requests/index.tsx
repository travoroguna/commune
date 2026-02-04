import { createFileRoute, Link } from '@tanstack/react-router';
import { useAuth } from '@/contexts/AuthContext';
import { useEffect, useState } from 'react';
import { serviceRequestApi } from '@/api/client';
import type { ServiceRequest, ServiceRequestStatus } from '@/types';
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
import { Plus, Clock, CheckCircle, XCircle, AlertCircle } from 'lucide-react';

export const Route = createFileRoute('/_authenticated/service-requests/')({
  component: ServiceRequestsPage,
});

function ServiceRequestsPage() {
  const { currentCommunity } = useAuth();
  const [serviceRequests, setServiceRequests] = useState<ServiceRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState<ServiceRequestStatus | 'all'>('all');

  useEffect(() => {
    fetchServiceRequests();
  }, [currentCommunity, statusFilter]);

  const fetchServiceRequests = async () => {
    try {
      setLoading(true);
      const params: { community_id?: number; status?: string } = {};
      if (currentCommunity) {
        params.community_id = currentCommunity.ID;
      }
      if (statusFilter !== 'all') {
        params.status = statusFilter;
      }

      const requests = await serviceRequestApi.getAll(params);
      setServiceRequests(requests || []);
    } catch (error) {
      console.error('Failed to fetch service requests:', error);
      setServiceRequests([]);
    } finally {
      setLoading(false);
    }
  };

  const getStatusIcon = (status: ServiceRequestStatus) => {
    switch (status) {
      case 'open':
        return <Clock className="h-4 w-4 text-blue-600" />;
      case 'in_progress':
        return <AlertCircle className="h-4 w-4 text-yellow-600" />;
      case 'completed':
        return <CheckCircle className="h-4 w-4 text-green-600" />;
      case 'cancelled':
        return <XCircle className="h-4 w-4 text-red-600" />;
    }
  };

  const getStatusColor = (status: ServiceRequestStatus) => {
    switch (status) {
      case 'open':
        return 'bg-blue-100 text-blue-800';
      case 'in_progress':
        return 'bg-yellow-100 text-yellow-800';
      case 'completed':
        return 'bg-green-100 text-green-800';
      case 'cancelled':
        return 'bg-red-100 text-red-800';
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-[400px]">
        <div className="text-slate-600">Loading service requests...</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-slate-900">Service Requests</h1>
          <p className="mt-2 text-slate-600">
            Browse and manage service requests in your community
          </p>
        </div>
        <Link to="/service-requests/new">
          <Button>
            <Plus className="h-4 w-4 mr-2" />
            New Request
          </Button>
        </Link>
      </div>

      <div className="flex gap-4 items-center">
        <div className="w-48">
          <Select
            value={statusFilter}
            onValueChange={(value) => setStatusFilter(value as ServiceRequestStatus | 'all')}
          >
            <SelectTrigger>
              <SelectValue placeholder="Filter by status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Requests</SelectItem>
              <SelectItem value="open">Open</SelectItem>
              <SelectItem value="in_progress">In Progress</SelectItem>
              <SelectItem value="completed">Completed</SelectItem>
              <SelectItem value="cancelled">Cancelled</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {serviceRequests.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <AlertCircle className="h-12 w-12 text-slate-400 mb-4" />
            <p className="text-slate-600 text-center">
              No service requests found.
              {statusFilter !== 'all' && ' Try changing the filter.'}
            </p>
            {statusFilter === 'all' && (
              <Link to="/service-requests/new" className="mt-4">
                <Button>Create First Request</Button>
              </Link>
            )}
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {serviceRequests.map((request) => (
            <Link
              key={request.ID}
              to="/service-requests/$requestId"
              params={{ requestId: request.ID.toString() }}
            >
              <Card className="hover:shadow-lg transition-shadow cursor-pointer h-full">
                <CardHeader>
                  <div className="flex justify-between items-start mb-2">
                    <div className="flex items-center gap-2">
                      {getStatusIcon(request.Status)}
                      <span
                        className={`text-xs px-2 py-1 rounded-full font-medium ${getStatusColor(
                          request.Status
                        )}`}
                      >
                        {request.Status.replace('_', ' ').toUpperCase()}
                      </span>
                    </div>
                  </div>
                  <CardTitle className="line-clamp-2">{request.Title}</CardTitle>
                  <CardDescription className="line-clamp-2">
                    {request.Description}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2 text-sm text-slate-600">
                    {request.Category && (
                      <div className="flex items-center gap-2">
                        <span className="font-medium">Category:</span>
                        <span className="text-blue-600">{request.Category}</span>
                      </div>
                    )}
                    {request.Budget && (
                      <div className="flex items-center gap-2">
                        <span className="font-medium">Budget:</span>
                        <span className="text-green-600">${request.Budget}</span>
                      </div>
                    )}
                    <div className="flex items-center gap-2">
                      <span className="font-medium">By:</span>
                      <span>{request.Requester?.Name || 'Unknown'}</span>
                    </div>
                    {request.ServiceOffers && request.ServiceOffers.length > 0 && (
                      <div className="flex items-center gap-2 pt-2 border-t">
                        <span className="font-medium text-purple-600">
                          {request.ServiceOffers.length} offer(s)
                        </span>
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
