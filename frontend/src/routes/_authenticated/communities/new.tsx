import { createFileRoute, Navigate, useNavigate } from '@tanstack/react-router';
import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuth } from '@/contexts/AuthContext';
import { communityApi } from '@/api/client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';

export const Route = createFileRoute('/_authenticated/communities/new')({
  component: CreateCommunityPage,
});

function CreateCommunityPage() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const canManageCommunities = user?.Role === 'super_admin';

  const [formData, setFormData] = useState({
    name: '',
    description: '',
    subdomain: '',
    address: '',
    city: '',
    state: '',
    country: '',
    zipCode: '',
  });
  const [error, setError] = useState('');

  const createCommunityMutation = useMutation({
    mutationFn: communityApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['communities'] });
      navigate({ to: '/_authenticated/communities' });
    },
    onError: (err: Error) => {
      setError(err.message);
    },
  });

  if (!canManageCommunities) {
    return <Navigate to="/" />;
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!formData.name) {
      setError('Community name is required');
      return;
    }

    // Create the payload with only non-empty fields
    const payload: Record<string, string> = {
      Name: formData.name,
    };
    
    if (formData.description) payload.Description = formData.description;
    if (formData.subdomain) payload.Subdomain = formData.subdomain;
    if (formData.address) payload.Address = formData.address;
    if (formData.city) payload.City = formData.city;
    if (formData.state) payload.State = formData.state;
    if (formData.country) payload.Country = formData.country;
    if (formData.zipCode) payload.ZipCode = formData.zipCode;

    createCommunityMutation.mutate(payload as any);
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Create Community</h1>
        <p className="text-slate-600 mt-1">Add a new community to the system</p>
      </div>

      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle>Community Details</CardTitle>
          <CardDescription>Enter the information for the new community</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name">Community Name *</Label>
              <Input
                id="name"
                type="text"
                placeholder="Sunset Apartments"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                disabled={createCommunityMutation.isPending}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Input
                id="description"
                type="text"
                placeholder="A vibrant community..."
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                disabled={createCommunityMutation.isPending}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="subdomain">Subdomain</Label>
              <Input
                id="subdomain"
                type="text"
                placeholder="sunset (for sunset.commune.com)"
                value={formData.subdomain}
                onChange={(e) => setFormData({ ...formData, subdomain: e.target.value })}
                disabled={createCommunityMutation.isPending}
              />
              <p className="text-xs text-slate-500">
                Optional: This will be used for subdomain routing (e.g., sunset.commune.com)
              </p>
            </div>

            <div className="border-t pt-4">
              <h3 className="font-semibold mb-3">Location (Optional)</h3>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2 col-span-2">
                  <Label htmlFor="address">Address</Label>
                  <Input
                    id="address"
                    type="text"
                    value={formData.address}
                    onChange={(e) => setFormData({ ...formData, address: e.target.value })}
                    disabled={createCommunityMutation.isPending}
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="city">City</Label>
                  <Input
                    id="city"
                    type="text"
                    value={formData.city}
                    onChange={(e) => setFormData({ ...formData, city: e.target.value })}
                    disabled={createCommunityMutation.isPending}
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="state">State/Province</Label>
                  <Input
                    id="state"
                    type="text"
                    value={formData.state}
                    onChange={(e) => setFormData({ ...formData, state: e.target.value })}
                    disabled={createCommunityMutation.isPending}
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="country">Country</Label>
                  <Input
                    id="country"
                    type="text"
                    value={formData.country}
                    onChange={(e) => setFormData({ ...formData, country: e.target.value })}
                    disabled={createCommunityMutation.isPending}
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="zipCode">Zip/Postal Code</Label>
                  <Input
                    id="zipCode"
                    type="text"
                    value={formData.zipCode}
                    onChange={(e) => setFormData({ ...formData, zipCode: e.target.value })}
                    disabled={createCommunityMutation.isPending}
                  />
                </div>
              </div>
            </div>

            {error && (
              <div className="bg-red-50 text-red-600 px-4 py-2 rounded-md text-sm">
                {error}
              </div>
            )}

            <div className="flex gap-3 pt-4">
              <Button type="submit" disabled={createCommunityMutation.isPending}>
                {createCommunityMutation.isPending ? 'Creating...' : 'Create Community'}
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => navigate({ to: '/_authenticated/communities' })}
                disabled={createCommunityMutation.isPending}
              >
                Cancel
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

