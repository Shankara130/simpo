import Link from 'next/link';

export default function ReportsPage() {
  return (
    <div>
      <h2 className="text-2xl font-bold mb-6">Laporan Keuangan</h2>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Daily Sales Summary Report */}
        {/* Story 5.1, Task 6: Daily Sales Summary Report */}
        <Link
          href="/reports/daily"
          className="bg-white p-6 rounded-lg border shadow-sm hover:shadow-md transition-shadow"
        >
          <div className="flex items-start">
            <div className="flex-1">
              <h3 className="text-lg font-semibold text-gray-900 mb-2">
                Laporan Penjualan Harian
              </h3>
              <p className="text-gray-600 text-sm mb-4">
                Ringkasan penjualan harian dengan breakdown metode pembayaran, produk terlaris, dan tren penjualan per jam.
              </p>
              <div className="flex flex-wrap gap-2">
                <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-blue-100 text-blue-800">
                  Total Penjualan
                </span>
                <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-green-100 text-green-800">
                  Metode Pembayaran
                </span>
                <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-purple-100 text-purple-800">
                  Produk Terlaris
                </span>
                <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-orange-100 text-orange-800">
                  Tren Per Jam
                </span>
              </div>
            </div>
            <svg
              className="w-6 h-6 text-gray-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 5l7 7-7 7"
              />
            </svg>
          </div>
        </Link>

        {/* Profit & Loss Report */}
        {/* Story 5.2, Task 6: Profit/Loss Report */}
        <Link
          href="/reports/profit-loss"
          className="bg-white p-6 rounded-lg border shadow-sm hover:shadow-md transition-shadow"
        >
          <div className="flex items-start">
            <div className="flex-1">
              <h3 className="text-lg font-semibold text-gray-900 mb-2">
                Laporan Laba Rugi
              </h3>
              <p className="text-gray-600 text-sm mb-4">
                Analisis profitabilitas dengan breakdown pendapatan, harga pokok penjualan, dan margin laba kotor.
              </p>
              <div className="flex flex-wrap gap-2">
                <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-blue-100 text-blue-800">
                  Total Pendapatan
                </span>
                <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-green-100 text-green-800">
                  Harga Pokok
                </span>
                <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-purple-100 text-purple-800">
                  Laba Kotor
                </span>
                <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-orange-100 text-orange-800">
                  Margin
                </span>
              </div>
            </div>
            <svg
              className="w-6 h-6 text-gray-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 5l7 7-7 7"
              />
            </svg>
          </div>
        </Link>
      </div>

      {/* Report Features Section */}
      <div className="mt-8 bg-white p-6 rounded-lg border shadow-sm">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Fitur Laporan</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="flex items-start">
            <div className="flex-shrink-0">
              <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div className="ml-3">
              <h4 className="text-sm font-medium text-gray-900">Real-time Data</h4>
              <p className="text-xs text-gray-600 mt-1">Data laporan diambil langsung dari sistem transaksi</p>
            </div>
          </div>

          <div className="flex items-start">
            <div className="flex-shrink-0">
              <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 21v-4m0 0V5a2 2 0 012-2h6.5l1 1H21l-3 6 3 6h-8.5l-1-1H5a2 2 0 00-2 2zm9-13.5V9" />
              </svg>
            </div>
            <div className="ml-3">
              <h4 className="text-sm font-medium text-gray-900">Multi-Cabang</h4>
              <p className="text-xs text-gray-600 mt-1">Lihat laporan per cabang atau agregasi semua cabang</p>
            </div>
          </div>

          <div className="flex items-start">
            <div className="flex-shrink-0">
              <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
            </div>
            <div className="ml-3">
              <h4 className="text-sm font-medium text-gray-900">Export PDF/Excel</h4>
              <p className="text-xs text-gray-600 mt-1">Unduh laporan untuk keperluan akuntansi (Coming Soon)</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
