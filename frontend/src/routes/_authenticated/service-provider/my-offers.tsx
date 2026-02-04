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
  CheckCircle,
  XCircle,
  FileText,
  Briefcase,
} from 'lucide-react';

export const Route = createFileRoute('/_authenticated/service-provider/my-offers')({
  component: MyOffersPage,
});

function MyOffersPage() {
  const { user } = useAuth();
  const [offers, setOffers] = useState<ServiceOffer[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (user) {
      fetchOffers();
    }
  }, [user]);

  const fetchOffers = async () => {
    if (!user) return;

    try {
      setLoading(true);
      const allOffers = await serviceOfferApi.getAll({ provider_id: user.ID });
      setOffers(allOffers || []);
    } catch (error) {
      console.error('Failed to fetch offers:', error);
      setOffers([]);
    } finally {
      setLoading(false);
    }
  };

  const handleWithdraw = async (offerId: number) => {
    if (!confirm('Are you sure you want to withdraw this offer?')) return;

    try {
      await serviceOfferApi.withdraw(offerId);
      await fetchOffers();
    } catch (error) {
      console.error('Failed to withdraw offer:', error);
      alert('Failed to withdraw offer. Please try again.');
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'pending':
        return <Clock className="h-5 w-5 text-yellow-600" />;
      case 'accepted':
        return <CheckCircle className="h-5 w-5 text-green-600" />;
      case 'rejected':
        return <XCircle className="h-5 w-5 text-red-600" />;
      case 'withdrawn':
        return <AlertCircle className="h-5 w-5 text-gray-600" />;
      default:
        return <Clock className="h-5 w-5 text-gray-600" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending':
        return 'bg-yellow-100 text-yellow-800';
      case 'accepted':
        return 'bg-green-100 text-green-800';
      case 'rejected':
        return 'bg-red-100 text-red-800';
      case 'withdrawn':
        return 'bg-gray-100 text-gray-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const groupOffersByStatus = () => {
    const pending = offers.filter((o) => o.Status === 'pending');
    const accepted = offers.filter((o) => o.Status === 'accepted');
    const rejected = offers.filter((o) => o.Status === 'rejected');
    const withdrawn = offers.filter((o) => o.Status === 'withdrawn');

    return { pending, accepted, rejected, withdrawn };
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-[400px]">
        <div className="text-slate-600">Loading your offers...</div>
      </div>
    );
  }

  const { pending, accepted, rejected, withdrawn } = groupOffersByStatus();

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-slate-900">My Offers</h1>
          <p className="mt-2 text-slate-600">
            Manage all your service offers in one place
          </p>
        </div>
        <div className="flex gap-3">
          <Link to="/service-provider/dashboard">
            <Button variant="outline">
              <Briefcase className="h-4 w-4 mr-2" />
              Browse Requests
            </Button>
          </Link>
          <Link to="/service-provider/contacts">
            <Button>View Contacts</Button>
          </Link>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Pending</CardDescription>
            <CardTitle className="text-3xl text-yellow-600">{pending.length}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Accepted</CardDescription>
            <CardTitle className="text-3xl text-green-600">{accepted.length}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Rejected</CardDescription>
            <CardTitle className="text-3xl text-red-600">{rejected.length}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Withdrawn</CardDescription>
            <CardTitle className="text-3xl text-gray-600">{withdrawn.length}</CardTitle>
          </CardHeader>
        </Card>
      </div>

      {offers.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <FileText className="h-12 w-12 text-slate-400 mb-4" />
            <p className="text-slate-600 text-center mb-4">You haven't created any offers yet</p>
            <Link to="/service-provider/dashboard">
              <Button>Browse Service Requests</Button>
            </Link>
          </CardContent>
        </Card>
      ) : (
        <>
          {/* Pending Offers */}
          {pending.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Clock className="h-5 w-5 text-yellow-600" />
                  Pending Offers ({pending.length})
                </CardTitle>
                <CardDescription>Awaiting response from requesters</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {pending.map((offer) => (
                  <OfferCard
                    key={offer.ID}
                    offer={offer}
                    onWithdraw={handleWithdraw}
                    getStatusIcon={getStatusIcon}
                    getStatusColor={getStatusColor}
                  />
                ))}
              </CardContent>
            </Card>
          )}

          {/* Accepted Offers */}
          {accepted.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <CheckCircle className="h-5 w-5 text-green-600" />
                  Accepted Offers ({accepted.length})
                </CardTitle>
                <CardDescription>Your accepted service engagements</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {accepted.map((offer) => (
                  <OfferCard
                    key={offer.ID}
                    offer={offer}
                    onWithdraw={handleWithdraw}
                    getStatusIcon={getStatusIcon}
                    getStatusColor={getStatusColor}
                  />
                ))}
              </CardContent>
            </Card>
          )}

          {/* Rejected Offers */}
          {rejected.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <XCircle className="h-5 w-5 text-red-600" />
                  Rejected Offers ({rejected.length})
                </CardTitle>
                <CardDescription>Offers that were not accepted</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {rejected.map((offer) => (
                  <OfferCard
                    key={offer.ID}
                    offer={offer}
                    onWithdraw={handleWithdraw}
                    getStatusIcon={getStatusIcon}
                    getStatusColor={getStatusColor}
                  />
                ))}
              </CardContent>
            </Card>
          )}

          {/* Withdrawn Offers */}
          {withdrawn.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <AlertCircle className="h-5 w-5 text-gray-600" />
                  Withdrawn Offers ({withdrawn.length})
                </CardTitle>
                <CardDescription>Offers you have withdrawn</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {withdrawn.map((offer) => (
                  <OfferCard
                    key={offer.ID}
                    offer={offer}
                    onWithdraw={handleWithdraw}
                    getStatusIcon={getStatusIcon}
                    getStatusColor={getStatusColor}
                  />
                ))}
              </CardContent>
            </Card>
          )}
        </>
      )}
    </div>
  );
}

interface OfferCardProps {
  offer: ServiceOffer;
  onWithdraw: (offerId: number) => void;
  getStatusIcon: (status: string) => React.ReactElement;
  getStatusColor: (status: string) => string;
}

function OfferCard({ offer, onWithdraw, getStatusIcon, getStatusColor }: OfferCardProps) {
  return (
    <div className="border rounded-lg p-4 hover:shadow-md transition-shadow">
      <div className="flex justify-between items-start mb-3">
        <div>
          <Link
            to="/service-requests/$requestId"
            params={{ requestId: offer.ServiceRequestID.toString() }}
            className="font-medium text-lg text-blue-600 hover:underline"
          >
            {offer.ServiceRequest?.Title || 'Service Request'}
          </Link>
          {offer.ServiceRequest?.Category && (
            <p className="text-sm text-slate-600 mt-1">
              Category: <span className="text-blue-600">{offer.ServiceRequest.Category}</span>
            </p>
          )}
        </div>
        <div className="flex items-center gap-2">
          {getStatusIcon(offer.Status)}
          <span className={`text-xs px-2 py-1 rounded-full font-medium ${getStatusColor(offer.Status)}`}>
            {offer.Status.toUpperCase()}
          </span>
        </div>
      </div>

      <p className="text-slate-700 mb-3">{offer.Description}</p>

      <div className="grid grid-cols-3 gap-4 text-sm text-slate-600 mb-3">
        {offer.ProposedPrice && (
          <div className="flex items-center gap-2">
            <DollarSign className="h-4 w-4 text-green-600" />
            <span className="font-medium">${offer.ProposedPrice}</span>
          </div>
        )}
        {offer.EstimatedDuration && (
          <div className="flex items-center gap-2">
            <Clock className="h-4 w-4 text-slate-600" />
            <span>{offer.EstimatedDuration}</span>
          </div>
        )}
        <div className="flex items-center gap-2">
          <Calendar className="h-4 w-4 text-slate-600" />
          <span>{new Date(offer.CreatedAt).toLocaleDateString()}</span>
        </div>
      </div>

      {offer.ServiceRequest && (
        <div className="bg-slate-50 p-3 rounded-md mb-3">
          <p className="text-sm text-slate-700 line-clamp-2">
            {offer.ServiceRequest.Description}
          </p>
          {offer.ServiceRequest.Budget && (
            <p className="text-sm text-slate-600 mt-2">
              Request Budget:{' '}
              <span className="text-green-600 font-medium">${offer.ServiceRequest.Budget}</span>
            </p>
          )}
        </div>
      )}

      <div className="flex gap-2 pt-3 border-t">
        <Link
          to="/service-requests/$requestId"
          params={{ requestId: offer.ServiceRequestID.toString() }}
        >
          <Button variant="outline" size="sm">
            View Request
          </Button>
        </Link>
        {offer.Status === 'pending' && (
          <Button variant="destructive" size="sm" onClick={() => onWithdraw(offer.ID)}>
            Withdraw
          </Button>
        )}
      </div>
    </div>
  );
}
