import { createFileRoute, redirect } from '@tanstack/react-router';

export const Route = createFileRoute('/')({
  beforeLoad: async () => {
    // Check if user is authenticated via fetch
    try {
      const response = await fetch('/api/auth/me', { credentials: 'include' });
      if (response.ok) {
        // User is authenticated, redirect to dashboard
        throw redirect({ to: '/_authenticated/dashboard' });
      }
    } catch (err) {
      // If error is a redirect, rethrow it
      if (err && typeof err === 'object' && 'to' in err) {
        throw err;
      }
    }

    // Check if it's first boot
    try {
      const response = await fetch('/api/auth/first-boot');
      if (response.ok) {
        const data = await response.json();
        if (data.needsSetup) {
          throw redirect({ to: '/setup' });
        }
      }
    } catch (err) {
      // If error is a redirect, rethrow it
      if (err && typeof err === 'object' && 'to' in err) {
        throw err;
      }
    }

    // Not authenticated and not first boot, redirect to login
    throw redirect({ to: '/login' });
  },
});

