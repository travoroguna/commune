import { createFileRoute } from '@tanstack/react-router';
import { useState, useEffect } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Search, MapPin, DollarSign, Clock, User } from 'lucide-react';
import type { ServiceRequest } from '@/types';

export const Route = createFileRoute('/_authenticated/services')({
  component: ServicesPage,
});

const statusColors = {
  open: 'bg-green-100 text-green-800',
  in_progress: 'bg-blue-100 text-blue-800',
  completed: 'bg-gray-100 text-gray-800',
  cancelled: 'bg-red-100 text-red-800',
};

const statusLabels = {
  open: 'Open',
  in_progress: 'In Progress',
  completed: 'Completed',
  cancelled: 'Cancelled',
};

function ServicesPage() {
  const { currentCommunity } = useAuth();
  const [services, setServices] = useState<ServiceRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('all');
  const [statusFilter, setStatusFilter] = useState('all');

  useEffect(() => {
    fetchServices();
  }, [currentCommunity, categoryFilter, statusFilter]);

  const fetchServices = async () => {
    try {
      setLoading(true);
      const params = new URLSearchParams();
      
      if (currentCommunity) {
        params.append('community_id', currentCommunity.ID.toString());
      }
      if (categoryFilter !== 'all') {
        params.append('category', categoryFilter);
      }
      if (statusFilter !== 'all') {
        params.append('status', statusFilter);
      }

      const response = await fetch(`/api/services?${params.toString()}`);
      if (response.ok) {
        const data = await response.json();
        setServices(data || []);
      }
    } catch (error) {
      console.error('Failed to fetch services:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = async () => {
    if (!searchTerm.trim()) {
      fetchServices();
      return;
    }

    try {
      setLoading(true);
      const params = new URLSearchParams();
      params.append('search', searchTerm);
      
      if (currentCommunity) {
        params.append('community_id', currentCommunity.ID.toString());
      }
      if (categoryFilter !== 'all') {
        params.append('category', categoryFilter);
      }
      if (statusFilter !== 'all') {
        params.append('status', statusFilter);
      }

      const response = await fetch(`/api/services?${params.toString()}`);
      if (response.ok) {
        const data = await response.json();
        setServices(data || []);
      }
    } catch (error) {
      console.error('Failed to search services:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const formatCurrency = (amount?: number) => {
    if (!amount) return 'N/A';
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(amount);
  };

  // Extract unique categories from services
  const categories = Array.from(
    new Set(services.map((s) => s.Category).filter(Boolean))
  );

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-slate-900">Available Services</h1>
        <p className="mt-2 text-slate-600">
          {currentCommunity
            ? `Browse and search services in ${currentCommunity.Name}`
            : 'Browse and search all available services'}
        </p>
      </div>

      {/* Search and Filter Controls */}
      <Card>
        <CardHeader>
          <CardTitle>Search Services</CardTitle>
          <CardDescription>
            Find services by title, description, category, or status
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {/* Search Bar */}
            <div className="flex gap-2">
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
                <Input
                  type="text"
                  placeholder="Search services..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  onKeyPress={(e) => {
                    if (e.key === 'Enter') {
                      handleSearch();
                    }
                  }}
                  className="pl-9"
                />
              </div>
              <Button onClick={handleSearch}>Search</Button>
              {searchTerm && (
                <Button
                  variant="outline"
                  onClick={() => {
                    setSearchTerm('');
                    fetchServices();
                  }}
                >
                  Clear
                </Button>
              )}
            </div>

            {/* Filters */}
            <div className="grid gap-4 md:grid-cols-2">
              <div>
                <label className="text-sm font-medium text-slate-700 mb-1.5 block">
                  Category
                </label>
                <Select value={categoryFilter} onValueChange={setCategoryFilter}>
                  <SelectTrigger>
                    <SelectValue placeholder="All Categories" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Categories</SelectItem>
                    {categories.map((category) => (
                      <SelectItem key={category} value={category!}>
                        {category}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div>
                <label className="text-sm font-medium text-slate-700 mb-1.5 block">
                  Status
                </label>
                <Select value={statusFilter} onValueChange={setStatusFilter}>
                  <SelectTrigger>
                    <SelectValue placeholder="All Statuses" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Statuses</SelectItem>
                    <SelectItem value="open">Open</SelectItem>
                    <SelectItem value="in_progress">In Progress</SelectItem>
                    <SelectItem value="completed">Completed</SelectItem>
                    <SelectItem value="cancelled">Cancelled</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Results */}
      {loading ? (
        <div className="text-center py-12">
          <p className="text-slate-600">Loading services...</p>
        </div>
      ) : services.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <p className="text-slate-600">No services found.</p>
            <p className="text-sm text-slate-500 mt-2">
              Try adjusting your search criteria or filters.
            </p>
          </CardContent>
        </Card>
      ) : (
        <>
          <div className="flex justify-between items-center">
            <p className="text-sm text-slate-600">
              Found {services.length} service{services.length !== 1 ? 's' : ''}
            </p>
          </div>

          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {services.map((service) => (
              <Card
                key={service.ID}
                className="hover:shadow-lg transition-shadow cursor-pointer"
              >
                <CardHeader>
                  <div className="flex justify-between items-start gap-2">
                    <CardTitle className="text-lg line-clamp-2">
                      {service.Title}
                    </CardTitle>
                    <span
                      className={`px-2 py-1 text-xs font-medium rounded-full whitespace-nowrap ${
                        statusColors[service.Status]
                      }`}
                    >
                      {statusLabels[service.Status]}
                    </span>
                  </div>
                  {service.Category && (
                    <CardDescription className="text-xs font-medium text-blue-600">
                      {service.Category}
                    </CardDescription>
                  )}
                </CardHeader>
                <CardContent className="space-y-3">
                  <p className="text-sm text-slate-600 line-clamp-3">
                    {service.Description}
                  </p>

                  <div className="space-y-2 text-sm text-slate-600">
                    {service.Community && (
                      <div className="flex items-center gap-2">
                        <MapPin className="h-4 w-4 text-slate-400" />
                        <span className="truncate">{service.Community.Name}</span>
                      </div>
                    )}

                    {service.Requester && (
                      <div className="flex items-center gap-2">
                        <User className="h-4 w-4 text-slate-400" />
                        <span className="truncate">{service.Requester.Name}</span>
                      </div>
                    )}

                    {service.Budget && (
                      <div className="flex items-center gap-2">
                        <DollarSign className="h-4 w-4 text-slate-400" />
                        <span>Budget: {formatCurrency(service.Budget)}</span>
                      </div>
                    )}

                    <div className="flex items-center gap-2">
                      <Clock className="h-4 w-4 text-slate-400" />
                      <span>Posted {formatDate(service.CreatedAt)}</span>
                    </div>

                    {service.ServiceOffers && service.ServiceOffers.length > 0 && (
                      <div className="pt-2 border-t">
                        <span className="text-xs font-medium text-slate-700">
                          {service.ServiceOffers.length} offer
                          {service.ServiceOffers.length !== 1 ? 's' : ''} received
                        </span>
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
