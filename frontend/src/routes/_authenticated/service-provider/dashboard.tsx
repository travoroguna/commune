import { createFileRoute, Link } from '@tanstack/react-router';
import { useAuth } from '@/contexts/AuthContext';
import { useEffect, useState } from 'react';
import { serviceRequestApi, serviceOfferApi } from '@/api/client';
import type { ServiceRequest } from '@/types';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  DollarSign,
  User,
  Calendar,
  AlertCircle,
  Briefcase,
  Filter,
} from 'lucide-react';

export const Route = createFileRoute('/_authenticated/service-provider/dashboard')({
  component: ServiceProviderDashboard,
});

function ServiceProviderDashboard() {
  const { user, currentCommunity } = useAuth();
  const [serviceRequests, setServiceRequests] = useState<ServiceRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [categoryFilter, setCategoryFilter] = useState<string>('all');
  const [categories, setCategories] = useState<string[]>([]);
  const [selectedRequest, setSelectedRequest] = useState<number | null>(null);
  const [offerForm, setOfferForm] = useState({
    description: '',
    proposedPrice: '',
    estimatedDuration: '',
  });
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    fetchServiceRequests();
  }, [currentCommunity]);

  const fetchServiceRequests = async () => {
    try {
      setLoading(true);
      const params: { community_id?: number; status?: string } = {
        status: 'open',
      };
      if (currentCommunity) {
        params.community_id = currentCommunity.ID;
      }

      const requests = await serviceRequestApi.getAll(params);
      setServiceRequests(requests || []);

      // Extract unique categories
      const uniqueCategories = Array.from(
        new Set(requests.filter((r) => r.Category).map((r) => r.Category as string))
      );
      setCategories(uniqueCategories);
    } catch (error) {
      console.error('Failed to fetch service requests:', error);
      setServiceRequests([]);
    } finally {
      setLoading(false);
    }
  };

  const handleQuickOffer = async (requestId: number, e: React.FormEvent) => {
    e.preventDefault();
    if (!offerForm.description) return;

    try {
      setSubmitting(true);
      await serviceOfferApi.create({
        service_request_id: requestId,
        description: offerForm.description,
        proposed_price: offerForm.proposedPrice ? parseFloat(offerForm.proposedPrice) : undefined,
        estimated_duration: offerForm.estimatedDuration || undefined,
      });

      setOfferForm({ description: '', proposedPrice: '', estimatedDuration: '' });
      setSelectedRequest(null);
      await fetchServiceRequests();
    } catch (error) {
      console.error('Failed to create offer:', error);
      alert('Failed to create offer. Please try again.');
    } finally {
      setSubmitting(false);
    }
  };

  const filteredRequests = serviceRequests.filter((request) => {
    if (categoryFilter === 'all') return true;
    return request.Category === categoryFilter;
  });

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
          <h1 className="text-3xl font-bold text-slate-900">Service Provider Dashboard</h1>
          <p className="mt-2 text-slate-600">
            Browse open service requests and submit your offers
          </p>
        </div>
        <Link to="/service-provider/my-offers">
          <Button variant="outline">
            <Briefcase className="h-4 w-4 mr-2" />
            My Offers
          </Button>
        </Link>
      </div>

      {/* Filter Section */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <Filter className="h-5 w-5" />
            Filters
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-4 items-center">
            <div className="w-64">
              <Label htmlFor="category">Filter by Category</Label>
              <Select value={categoryFilter} onValueChange={setCategoryFilter}>
                <SelectTrigger id="category">
                  <SelectValue placeholder="All Categories" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Categories</SelectItem>
                  {categories.map((category) => (
                    <SelectItem key={category} value={category}>
                      {category}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            {categoryFilter !== 'all' && (
              <Button variant="outline" size="sm" onClick={() => setCategoryFilter('all')}>
                Clear Filter
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Service Requests Grid */}
      {filteredRequests.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <AlertCircle className="h-12 w-12 text-slate-400 mb-4" />
            <p className="text-slate-600 text-center">
              No open service requests found
              {categoryFilter !== 'all' && ' in this category'}.
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {filteredRequests.map((request) => {
            const userHasOffered = request.ServiceOffers?.some(
              (offer) => offer.ProviderID === user?.ID
            );

            return (
              <Card key={request.ID} className="hover:shadow-lg transition-shadow">
                <CardHeader>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-xs px-2 py-1 rounded-full font-medium bg-blue-100 text-blue-800">
                      OPEN
                    </span>
                    {userHasOffered && (
                      <span className="text-xs px-2 py-1 rounded-full font-medium bg-purple-100 text-purple-800">
                        OFFERED
                      </span>
                    )}
                  </div>
                  <CardTitle className="line-clamp-2">{request.Title}</CardTitle>
                  <CardDescription className="line-clamp-3">
                    {request.Description}
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-2 text-sm">
                    {request.Category && (
                      <div className="flex items-center gap-2 text-slate-600">
                        <Briefcase className="h-4 w-4" />
                        <span className="font-medium">Category:</span>
                        <span className="text-blue-600">{request.Category}</span>
                      </div>
                    )}
                    {request.Budget && (
                      <div className="flex items-center gap-2 text-slate-600">
                        <DollarSign className="h-4 w-4 text-green-600" />
                        <span className="font-medium">Budget:</span>
                        <span className="text-green-600">${request.Budget}</span>
                      </div>
                    )}
                    <div className="flex items-center gap-2 text-slate-600">
                      <User className="h-4 w-4" />
                      <span className="font-medium">By:</span>
                      <span>{request.Requester?.Name || 'Unknown'}</span>
                    </div>
                    <div className="flex items-center gap-2 text-slate-600">
                      <Calendar className="h-4 w-4" />
                      <span>{new Date(request.CreatedAt).toLocaleDateString()}</span>
                    </div>
                    {request.ServiceOffers && request.ServiceOffers.length > 0 && (
                      <div className="text-slate-500 text-xs pt-1">
                        {request.ServiceOffers.length} offer(s) submitted
                      </div>
                    )}
                  </div>

                  <div className="pt-3 border-t space-y-3">
                    {selectedRequest === request.ID ? (
                      <form onSubmit={(e) => handleQuickOffer(request.ID, e)} className="space-y-3">
                        <div>
                          <Label htmlFor={`description-${request.ID}`}>Your Offer *</Label>
                          <textarea
                            id={`description-${request.ID}`}
                            value={offerForm.description}
                            onChange={(e) =>
                              setOfferForm({ ...offerForm, description: e.target.value })
                            }
                            placeholder="Describe your service..."
                            required
                            className="w-full min-h-[80px] px-3 py-2 text-sm border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                          />
                        </div>
                        <div className="grid grid-cols-2 gap-2">
                          <div>
                            <Label htmlFor={`price-${request.ID}`} className="text-xs">
                              Price ($)
                            </Label>
                            <Input
                              id={`price-${request.ID}`}
                              type="number"
                              step="0.01"
                              min="0"
                              value={offerForm.proposedPrice}
                              onChange={(e) =>
                                setOfferForm({ ...offerForm, proposedPrice: e.target.value })
                              }
                              placeholder="150.00"
                              className="text-sm"
                            />
                          </div>
                          <div>
                            <Label htmlFor={`duration-${request.ID}`} className="text-xs">
                              Duration
                            </Label>
                            <Input
                              id={`duration-${request.ID}`}
                              value={offerForm.estimatedDuration}
                              onChange={(e) =>
                                setOfferForm({ ...offerForm, estimatedDuration: e.target.value })
                              }
                              placeholder="2 hours"
                              className="text-sm"
                            />
                          </div>
                        </div>
                        <div className="flex gap-2">
                          <Button type="submit" size="sm" disabled={submitting} className="flex-1">
                            {submitting ? 'Submitting...' : 'Submit'}
                          </Button>
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() => {
                              setSelectedRequest(null);
                              setOfferForm({
                                description: '',
                                proposedPrice: '',
                                estimatedDuration: '',
                              });
                            }}
                          >
                            Cancel
                          </Button>
                        </div>
                      </form>
                    ) : (
                      <div className="flex gap-2">
                        <Link
                          to="/service-requests/$requestId"
                          params={{ requestId: request.ID.toString() }}
                          className="flex-1"
                        >
                          <Button variant="outline" size="sm" className="w-full">
                            View Details
                          </Button>
                        </Link>
                        {!userHasOffered && (
                          <Button
                            size="sm"
                            className="flex-1"
                            onClick={() => setSelectedRequest(request.ID)}
                          >
                            Quick Offer
                          </Button>
                        )}
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}
