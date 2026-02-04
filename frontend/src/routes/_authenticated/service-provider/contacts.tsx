import { createFileRoute, Link } from '@tanstack/react-router';
import { useAuth } from '@/contexts/AuthContext';
import { useEffect, useState } from 'react';
import { serviceOfferApi } from '@/api/client';
import type { ServiceOffer } from '@/types';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
  Clock,
  DollarSign,
  Calendar,
  AlertCircle,
  Mail,
  User,
  Briefcase,
  FileText,
} from 'lucide-react';

export const Route = createFileRoute('/_authenticated/service-provider/contacts')({
  component: ContactsPage,
});

function ContactsPage() {
  const { user } = useAuth();
  const [acceptedOffers, setAcceptedOffers] = useState<ServiceOffer[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (user) {
      fetchAcceptedOffers();
    }
  }, [user]);

  const fetchAcceptedOffers = async () => {
    if (!user) return;

    try {
      setLoading(true);
      const offers = await serviceOfferApi.getAcceptedOffers({ provider_id: user.ID });
      setAcceptedOffers(offers || []);
    } catch (error) {
      console.error('Failed to fetch accepted offers:', error);
      setAcceptedOffers([]);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-[400px]">
        <div className="text-slate-600">Loading your contacts...</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-slate-900">Service Contacts</h1>
          <p className="mt-2 text-slate-600">
            View contact information for your accepted service offers
          </p>
        </div>
        <div className="flex gap-3">
          <Link to="/service-provider/dashboard">
            <Button variant="outline">
              <Briefcase className="h-4 w-4 mr-2" />
              Browse Requests
            </Button>
          </Link>
          <Link to="/service-provider/my-offers">
            <Button variant="outline">
              <FileText className="h-4 w-4 mr-2" />
              My Offers
            </Button>
          </Link>
        </div>
      </div>

      {acceptedOffers.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <AlertCircle className="h-12 w-12 text-slate-400 mb-4" />
            <p className="text-slate-600 text-center mb-2">
              You don't have any accepted offers yet
            </p>
            <p className="text-slate-500 text-center text-sm mb-4">
              Contact information will appear here once your offers are accepted
            </p>
            <Link to="/service-provider/dashboard">
              <Button>Browse Service Requests</Button>
            </Link>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-2">
          {acceptedOffers.map((offer) => (
            <Card key={offer.ID} className="border-green-500 border-2">
              <CardHeader>
                <div className="flex items-center gap-2 mb-2">
                  <span className="text-xs px-3 py-1 rounded-full font-medium bg-green-100 text-green-800">
                    ACCEPTED
                  </span>
                </div>
                <CardTitle className="text-xl">
                  {offer.ServiceRequest?.Title || 'Service Request'}
                </CardTitle>
                <CardDescription>
                  {offer.ServiceRequest?.Description || 'No description available'}
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {/* Service Request Info */}
                <div className="bg-blue-50 p-4 rounded-lg space-y-2">
                  <h4 className="font-medium text-slate-900 mb-2">Service Details</h4>
                  <div className="grid grid-cols-2 gap-3 text-sm">
                    {offer.ServiceRequest?.Category && (
                      <div className="flex items-center gap-2 text-slate-600">
                        <Briefcase className="h-4 w-4" />
                        <span>{offer.ServiceRequest.Category}</span>
                      </div>
                    )}
                    {offer.ServiceRequest?.Budget && (
                      <div className="flex items-center gap-2 text-slate-600">
                        <DollarSign className="h-4 w-4 text-green-600" />
                        <span className="text-green-700 font-medium">
                          ${offer.ServiceRequest.Budget}
                        </span>
                      </div>
                    )}
                  </div>
                </div>

                {/* Your Offer Info */}
                <div className="bg-slate-50 p-4 rounded-lg space-y-2">
                  <h4 className="font-medium text-slate-900 mb-2">Your Offer</h4>
                  <p className="text-sm text-slate-700 mb-2">{offer.Description}</p>
                  <div className="grid grid-cols-2 gap-3 text-sm">
                    {offer.ProposedPrice && (
                      <div className="flex items-center gap-2 text-slate-600">
                        <DollarSign className="h-4 w-4 text-green-600" />
                        <span className="font-medium">${offer.ProposedPrice}</span>
                      </div>
                    )}
                    {offer.EstimatedDuration && (
                      <div className="flex items-center gap-2 text-slate-600">
                        <Clock className="h-4 w-4" />
                        <span>{offer.EstimatedDuration}</span>
                      </div>
                    )}
                    <div className="flex items-center gap-2 text-slate-600">
                      <Calendar className="h-4 w-4" />
                      <span>Accepted {new Date(offer.UpdatedAt).toLocaleDateString()}</span>
                    </div>
                  </div>
                </div>

                {/* Requester Contact Info */}
                <div className="border-t pt-4">
                  <h4 className="font-medium text-slate-900 mb-3 flex items-center gap-2">
                    <User className="h-5 w-5 text-blue-600" />
                    Requester Contact Information
                  </h4>
                  <div className="space-y-3">
                    <div className="flex items-start gap-3">
                      <User className="h-5 w-5 text-slate-600 mt-0.5" />
                      <div>
                        <p className="text-sm font-medium text-slate-700">Name</p>
                        <p className="text-base text-slate-900">
                          {offer.ServiceRequest?.Requester?.Name || 'Unknown'}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-start gap-3">
                      <Mail className="h-5 w-5 text-slate-600 mt-0.5" />
                      <div>
                        <p className="text-sm font-medium text-slate-700">Email</p>
                        <a
                          href={`mailto:${offer.ServiceRequest?.Requester?.Email}`}
                          className="text-base text-blue-600 hover:underline"
                        >
                          {offer.ServiceRequest?.Requester?.Email || 'Not available'}
                        </a>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Actions */}
                <div className="flex gap-2 pt-3 border-t">
                  <Link
                    to="/service-requests/$requestId"
                    params={{ requestId: offer.ServiceRequestID.toString() }}
                    className="flex-1"
                  >
                    <Button variant="outline" size="sm" className="w-full">
                      View Details
                    </Button>
                  </Link>
                  <Button
                    size="sm"
                    className="flex-1"
                    onClick={() => {
                      const email = offer.ServiceRequest?.Requester?.Email;
                      const subject = encodeURIComponent(
                        `Regarding: ${offer.ServiceRequest?.Title || 'Service Request'}`
                      );
                      window.location.href = `mailto:${email}?subject=${subject}`;
                    }}
                  >
                    <Mail className="h-4 w-4 mr-2" />
                    Send Email
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Summary Section */}
      {acceptedOffers.length > 0 && (
        <Card className="bg-blue-50 border-blue-200">
          <CardHeader>
            <CardTitle className="text-lg">Active Engagements Summary</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-center">
              <div>
                <p className="text-3xl font-bold text-blue-600">{acceptedOffers.length}</p>
                <p className="text-sm text-slate-600">Active Services</p>
              </div>
              <div>
                <p className="text-3xl font-bold text-green-600">
                  $
                  {acceptedOffers
                    .reduce((sum, offer) => sum + (offer.ProposedPrice || 0), 0)
                    .toFixed(2)}
                </p>
                <p className="text-sm text-slate-600">Total Value</p>
              </div>
              <div>
                <p className="text-3xl font-bold text-purple-600">
                  {new Set(acceptedOffers.map((o) => o.ServiceRequest?.Category)).size}
                </p>
                <p className="text-sm text-slate-600">Categories</p>
              </div>
              <div>
                <p className="text-3xl font-bold text-orange-600">
                  {new Set(acceptedOffers.map((o) => o.ServiceRequest?.RequesterID)).size}
                </p>
                <p className="text-sm text-slate-600">Unique Clients</p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
