import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'

function App() {
  const [count, setCount] = useState(0)
  const [health, setHealth] = useState<string>('')

  useEffect(() => {
    // Test the API connection
    fetch('/api/health')
      .then(res => res.json())
      .then(data => setHealth(data.status))
      .catch(() => setHealth('error'))
  }, [])

  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-8">
      <div className="max-w-2xl w-full space-y-8">
        <div className="text-center space-y-4">
          <h1 className="text-4xl font-bold tracking-tight">
            Commune
          </h1>
          <p className="text-muted-foreground">
            Go Backend + React Frontend + Vite + shadcn/ui
          </p>
        </div>

        <div className="border rounded-lg p-6 space-y-4">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium">API Status:</span>
            <span className={`text-sm ${health === 'ok' ? 'text-green-600' : 'text-red-600'}`}>
              {health || 'checking...'}
            </span>
          </div>

          <div className="flex items-center justify-between">
            <span className="text-sm font-medium">Counter:</span>
            <span className="text-sm font-mono">{count}</span>
          </div>

          <div className="flex gap-2">
            <Button onClick={() => setCount(count + 1)}>
              Increment
            </Button>
            <Button variant="outline" onClick={() => setCount(0)}>
              Reset
            </Button>
          </div>
        </div>

        <div className="text-center text-sm text-muted-foreground">
          <p>
            This is a starting point for building full-stack applications.
          </p>
          <p className="mt-2">
            Backend: Go + GORM + gormmigrate | Frontend: React + Vite + Tailwind + shadcn/ui
          </p>
        </div>
      </div>
    </div>
  )
}

export default App
