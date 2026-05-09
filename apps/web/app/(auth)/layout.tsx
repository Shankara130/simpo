export default function AuthenticatedLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div className="min-h-screen flex flex-col">
      {/* Header will be added here */}
      <header className="border-b bg-white">
        <div className="container mx-auto px-4 py-4">
          <h1 className="text-xl font-bold">simpo Admin Dashboard</h1>
        </div>
      </header>

      <div className="flex flex-1">
        {/* Sidebar will be added here */}
        <aside className="w-64 border-r bg-gray-50 p-4">
          <nav className="space-y-2">
            <a href="/dashboard" className="block py-2 px-4 rounded hover:bg-gray-200">
              Dashboard
            </a>
            <a href="/products" className="block py-2 px-4 rounded hover:bg-gray-200">
              Products
            </a>
            <a href="/reports" className="block py-2 px-4 rounded hover:bg-gray-200">
              Reports
            </a>
            <a href="/users" className="block py-2 px-4 rounded hover:bg-gray-200">
              Users
            </a>
            <a href="/settings" className="block py-2 px-4 rounded hover:bg-gray-200">
              Settings
            </a>
          </nav>
        </aside>

        {/* Main content */}
        <main className="flex-1 p-6">
          {children}
        </main>
      </div>

      {/* Footer will be added here */}
      <footer className="border-t bg-white py-4">
        <div className="container mx-auto px-4 text-center text-sm text-gray-600">
          © 2026 simpo. Pharmacy Management System.
        </div>
      </footer>
    </div>
  );
}
