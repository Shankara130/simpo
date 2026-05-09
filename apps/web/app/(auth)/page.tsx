export default function DashboardPage() {
  return (
    <div>
      <h2 className="text-2xl font-bold mb-6">Dashboard</h2>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div className="bg-white p-6 rounded-lg border shadow-sm">
          <h3 className="text-sm font-medium text-gray-600">{"Today's Sales"}</h3>
          <p className="text-2xl font-bold mt-2">Rp 0</p>
        </div>
        <div className="bg-white p-6 rounded-lg border shadow-sm">
          <h3 className="text-sm font-medium text-gray-600">Transactions</h3>
          <p className="text-2xl font-bold mt-2">0</p>
        </div>
        <div className="bg-white p-6 rounded-lg border shadow-sm">
          <h3 className="text-sm font-medium text-gray-600">Products</h3>
          <p className="text-2xl font-bold mt-2">0</p>
        </div>
        <div className="bg-white p-6 rounded-lg border shadow-sm">
          <h3 className="text-sm font-medium text-gray-600">Low Stock Items</h3>
          <p className="text-2xl font-bold mt-2">0</p>
        </div>
      </div>

      <div className="bg-white p-6 rounded-lg border shadow-sm">
        <h3 className="text-lg font-semibold mb-4">Welcome to simpo Admin Dashboard</h3>
        <p className="text-gray-600">
          This dashboard provides business oversight for pharmacy owners and system admins.
          Connect to the backend API to view real-time data.
        </p>
      </div>
    </div>
  );
}
