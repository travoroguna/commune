import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { useAuth } from '@/contexts/AuthContext';
import { useState } from 'react';
import { serviceRequestApi } from '@/api/client';
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

export const Route = createFileRoute('/_authenticated/service-requests/new')({
  component: NewServiceRequestPage,
});

function NewServiceRequestPage() {
  const { currentCommunity } = useAuth();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    title: '',
    description: '',
    category: '',
    budget: '',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!currentCommunity) {
      alert('Please select a community first');
      return;
    }

    try {
      setLoading(true);
      await serviceRequestApi.create({
        title: formData.title,
        description: formData.description,
        category: formData.category || undefined,
        budget: formData.budget ? parseFloat(formData.budget) : undefined,
        community_id: currentCommunity.ID,
      });
      
      navigate({ to: '/service-requests' });
    } catch (error) {
      console.error('Failed to create service request:', error);
      alert('Failed to create service request. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-slate-900">New Service Request</h1>
        <p className="mt-2 text-slate-600">
          Create a new service request for your community
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Request Details</CardTitle>
          <CardDescription>
            Provide details about the service you need
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <Label htmlFor="title">Title *</Label>
              <Input
                id="title"
                value={formData.title}
                onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                placeholder="e.g., Need electrician for outlet repair"
                required
              />
            </div>

            <div>
              <Label htmlFor="description">Description *</Label>
              <textarea
                id="description"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                placeholder="Describe the service you need in detail..."
                required
                className="w-full min-h-[120px] px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div>
              <Label htmlFor="category">Category</Label>
              <Input
                id="category"
                value={formData.category}
                onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                placeholder="e.g., Electrical, Plumbing, Cleaning"
              />
            </div>

            <div>
              <Label htmlFor="budget">Budget ($)</Label>
              <Input
                id="budget"
                type="number"
                step="0.01"
                min="0"
                value={formData.budget}
                onChange={(e) => setFormData({ ...formData, budget: e.target.value })}
                placeholder="e.g., 200.00"
              />
            </div>

            <div className="flex gap-3 pt-4">
              <Button type="submit" disabled={loading}>
                {loading ? 'Creating...' : 'Create Request'}
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => navigate({ to: '/service-requests' })}
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
