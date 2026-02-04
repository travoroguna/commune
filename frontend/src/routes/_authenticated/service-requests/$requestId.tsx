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
  Clock,
  CheckCircle,
  XCircle,
  AlertCircle,
  DollarSign,
  User,
  Calendar,
  Mail,
} from 'lucide-react';

export const Route = createFileRoute('/_authenticated/service-requests/$requestId')({
  component: ServiceRequestDetailPage,
});

function ServiceRequestDetailPage() {
  const { requestId } = Route.useParams();
  const { user } = useAuth();
  const [serviceRequest, setServiceRequest] = useState<ServiceRequest | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [showOfferForm, setShowOfferForm] = useState(false);
  const [offerForm, setOfferForm] = useState({
    description: '',
    proposedPrice: '',
    estimatedDuration: '',
  });

  useEffect(() => {
    fetchServiceRequest();
  }, [requestId]);

  const fetchServiceRequest = async () => {
    try {
      setLoading(true);
      const request = await serviceRequestApi.getById(parseInt(requestId));
      setServiceRequest(request);
    } catch (error) {
      console.error('Failed to fetch service request:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateOffer = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!serviceRequest) return;

    try {
      setSubmitting(true);
      await serviceOfferApi.create({
        service_request_id: serviceRequest.ID,
        description: offerForm.description,
        proposed_price: offerForm.proposedPrice ? parseFloat(offerForm.proposedPrice) : undefined,
        estimated_duration: offerForm.estimatedDuration || undefined,
      });

      setOfferForm({ description: '', proposedPrice: '', estimatedDuration: '' });
      setShowOfferForm(false);
      await fetchServiceRequest();
    } catch (error) {
      console.error('Failed to create offer:', error);
      alert('Failed to create offer. Please try again.');
    } finally {
      setSubmitting(false);
    }
  };

  const handleAcceptOffer = async (offerId: number) => {
    if (!serviceRequest) return;
    if (!confirm('Are you sure you want to accept this offer?')) return;

    try {
      await serviceRequestApi.acceptOffer(serviceRequest.ID, offerId);
      await fetchServiceRequest();
    } catch (error) {
      console.error('Failed to accept offer:', error);
      alert('Failed to accept offer. Please try again.');
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'open':
        return <Clock className="h-5 w-5 text-blue-600" />;
      case 'in_progress':
        return <AlertCircle className="h-5 w-5 text-yellow-600" />;
      case 'completed':
        return <CheckCircle className="h-5 w-5 text-green-600" />;
      case 'cancelled':
        return <XCircle className="h-5 w-5 text-red-600" />;
      default:
        return <Clock className="h-5 w-5 text-gray-600" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'open':
        return 'bg-blue-100 text-blue-800';
      case 'in_progress':
        return 'bg-yellow-100 text-yellow-800';
      case 'completed':
        return 'bg-green-100 text-green-800';
      case 'cancelled':
        return 'bg-red-100 text-red-800';
      case 'pending':
        return 'bg-gray-100 text-gray-800';
      case 'accepted':
        return 'bg-green-100 text-green-800';
      case 'rejected':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-[400px]">
        <div className="text-slate-600">Loading service request...</div>
      </div>
    );
  }

  if (!serviceRequest) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[400px]">
        <AlertCircle className="h-12 w-12 text-slate-400 mb-4" />
        <p className="text-slate-600 mb-4">Service request not found</p>
        <Link to="/service-requests">
          <Button>Back to Service Requests</Button>
        </Link>
      </div>
    );
  }

  const isRequester = user?.ID === serviceRequest.RequesterID;
  const canCreateOffer = !isRequester && serviceRequest.Status === 'open';
  const hasAcceptedOffer = serviceRequest.AcceptedOfferID !== undefined && serviceRequest.AcceptedOfferID !== null;

  return (
    <div className="space-y-6 max-w-5xl mx-auto">
      <div className="flex items-center justify-between">
        <Link to="/service-requests">
          <Button variant="outline">← Back to Requests</Button>
        </Link>
      </div>

      {/* Service Request Details */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-3 mb-2">
            {getStatusIcon(serviceRequest.Status)}
            <span
              className={`text-sm px-3 py-1 rounded-full font-medium ${getStatusColor(
                serviceRequest.Status
              )}`}
            >
              {serviceRequest.Status.replace('_', ' ').toUpperCase()}
            </span>
          </div>
          <CardTitle className="text-2xl">{serviceRequest.Title}</CardTitle>
          <CardDescription className="text-base mt-2">
            {serviceRequest.Description}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            {serviceRequest.Category && (
              <div className="flex items-center gap-2">
                <span className="font-medium text-slate-700">Category:</span>
                <span className="text-blue-600">{serviceRequest.Category}</span>
              </div>
            )}
            {serviceRequest.Budget && (
              <div className="flex items-center gap-2">
                <DollarSign className="h-4 w-4 text-green-600" />
                <span className="font-medium text-slate-700">Budget:</span>
                <span className="text-green-600">${serviceRequest.Budget}</span>
              </div>
            )}
            <div className="flex items-center gap-2">
              <User className="h-4 w-4 text-slate-600" />
              <span className="font-medium text-slate-700">Requester:</span>
              <span>{serviceRequest.Requester?.Name || 'Unknown'}</span>
            </div>
            <div className="flex items-center gap-2">
              <Calendar className="h-4 w-4 text-slate-600" />
              <span className="font-medium text-slate-700">Created:</span>
              <span>{new Date(serviceRequest.CreatedAt).toLocaleDateString()}</span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Accepted Offer Contact Info */}
      {hasAcceptedOffer && serviceRequest.AcceptedOffer && (
        <Card className="border-green-500 border-2">
          <CardHeader>
            <CardTitle className="text-green-700">Accepted Offer</CardTitle>
            <CardDescription>This offer has been accepted for this service request</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="bg-green-50 p-4 rounded-lg">
              <p className="text-sm text-slate-700 mb-2">{serviceRequest.AcceptedOffer.Description}</p>
              <div className="grid grid-cols-2 gap-3 text-sm">
                {serviceRequest.AcceptedOffer.ProposedPrice && (
                  <div className="flex items-center gap-2">
                    <DollarSign className="h-4 w-4 text-green-600" />
                    <span className="font-medium">Price:</span>
                    <span className="text-green-700">${serviceRequest.AcceptedOffer.ProposedPrice}</span>
                  </div>
                )}
                {serviceRequest.AcceptedOffer.EstimatedDuration && (
                  <div className="flex items-center gap-2">
                    <Clock className="h-4 w-4 text-slate-600" />
                    <span className="font-medium">Duration:</span>
                    <span>{serviceRequest.AcceptedOffer.EstimatedDuration}</span>
                  </div>
                )}
              </div>
            </div>
            <div className="border-t pt-4">
              <h4 className="font-medium text-slate-900 mb-3">Service Provider Contact</h4>
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <User className="h-4 w-4 text-slate-600" />
                  <span className="font-medium">Name:</span>
                  <span>{serviceRequest.AcceptedOffer.Provider?.Name || 'Unknown'}</span>
                </div>
                <div className="flex items-center gap-2">
                  <Mail className="h-4 w-4 text-slate-600" />
                  <span className="font-medium">Email:</span>
                  <a
                    href={`mailto:${serviceRequest.AcceptedOffer.Provider?.Email}`}
                    className="text-blue-600 hover:underline"
                  >
                    {serviceRequest.AcceptedOffer.Provider?.Email}
                  </a>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Create Offer Form */}
      {canCreateOffer && (
        <Card>
          <CardHeader>
            <CardTitle>Create an Offer</CardTitle>
            <CardDescription>Submit your proposal for this service request</CardDescription>
          </CardHeader>
          <CardContent>
            {!showOfferForm ? (
              <Button onClick={() => setShowOfferForm(true)}>Create Offer</Button>
            ) : (
              <form onSubmit={handleCreateOffer} className="space-y-4">
                <div>
                  <Label htmlFor="description">Description *</Label>
                  <textarea
                    id="description"
                    value={offerForm.description}
                    onChange={(e) => setOfferForm({ ...offerForm, description: e.target.value })}
                    placeholder="Describe your proposed service..."
                    required
                    className="w-full min-h-[100px] px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="proposedPrice">Proposed Price ($)</Label>
                    <Input
                      id="proposedPrice"
                      type="number"
                      step="0.01"
                      min="0"
                      value={offerForm.proposedPrice}
                      onChange={(e) =>
                        setOfferForm({ ...offerForm, proposedPrice: e.target.value })
                      }
                      placeholder="e.g., 150.00"
                    />
                  </div>
                  <div>
                    <Label htmlFor="estimatedDuration">Estimated Duration</Label>
                    <Input
                      id="estimatedDuration"
                      value={offerForm.estimatedDuration}
                      onChange={(e) =>
                        setOfferForm({ ...offerForm, estimatedDuration: e.target.value })
                      }
                      placeholder="e.g., 2 hours, 3 days"
                    />
                  </div>
                </div>
                <div className="flex gap-3">
                  <Button type="submit" disabled={submitting}>
                    {submitting ? 'Submitting...' : 'Submit Offer'}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => {
                      setShowOfferForm(false);
                      setOfferForm({ description: '', proposedPrice: '', estimatedDuration: '' });
                    }}
                  >
                    Cancel
                  </Button>
                </div>
              </form>
            )}
          </CardContent>
        </Card>
      )}

      {/* Service Offers List */}
      <Card>
        <CardHeader>
          <CardTitle>Service Offers ({serviceRequest.ServiceOffers?.length || 0})</CardTitle>
          <CardDescription>Proposals from service providers</CardDescription>
        </CardHeader>
        <CardContent>
          {!serviceRequest.ServiceOffers || serviceRequest.ServiceOffers.length === 0 ? (
            <div className="text-center py-8 text-slate-600">
              <AlertCircle className="h-8 w-8 text-slate-400 mx-auto mb-2" />
              <p>No offers yet</p>
            </div>
          ) : (
            <div className="space-y-4">
              {serviceRequest.ServiceOffers.map((offer) => (
                <div
                  key={offer.ID}
                  className={`border rounded-lg p-4 ${
                    offer.ID === serviceRequest.AcceptedOfferID ? 'border-green-500 bg-green-50' : ''
                  }`}
                >
                  <div className="flex justify-between items-start mb-3">
                    <div className="flex items-center gap-2">
                      <User className="h-4 w-4 text-slate-600" />
                      <span className="font-medium">{offer.Provider?.Name || 'Unknown'}</span>
                    </div>
                    <span
                      className={`text-xs px-2 py-1 rounded-full font-medium ${getStatusColor(
                        offer.Status
                      )}`}
                    >
                      {offer.Status.toUpperCase()}
                    </span>
                  </div>
                  <p className="text-slate-700 mb-3">{offer.Description}</p>
                  <div className="flex gap-4 text-sm text-slate-600">
                    {offer.ProposedPrice && (
                      <div className="flex items-center gap-1">
                        <DollarSign className="h-4 w-4 text-green-600" />
                        <span className="font-medium">${offer.ProposedPrice}</span>
                      </div>
                    )}
                    {offer.EstimatedDuration && (
                      <div className="flex items-center gap-1">
                        <Clock className="h-4 w-4 text-slate-600" />
                        <span>{offer.EstimatedDuration}</span>
                      </div>
                    )}
                    <div className="flex items-center gap-1">
                      <Calendar className="h-4 w-4 text-slate-600" />
                      <span>{new Date(offer.CreatedAt).toLocaleDateString()}</span>
                    </div>
                  </div>
                  {isRequester &&
                    offer.Status === 'pending' &&
                    serviceRequest.Status === 'open' &&
                    !hasAcceptedOffer && (
                      <div className="mt-3 pt-3 border-t">
                        <Button size="sm" onClick={() => handleAcceptOffer(offer.ID)}>
                          Accept Offer
                        </Button>
                      </div>
                    )}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
