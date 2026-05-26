'use client';

import { useState, useEffect, FormEvent } from 'react';
import { useAuth } from '@/context/AuthContext';
import apiClient, { getSystemSettings, updateSystemSettings, SystemSettingsResponse } from '@/lib/apiClient';
import { ApiError } from '@/lib/apiClient';

export default function SettingsPage() {
  const { user, loading: authLoading } = useAuth();
  const [settings, setSettings] = useState<SystemSettingsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [formData, setFormData] = useState({
    businessName: '',
    address: '',
    phone: '',
    email: '',
  });
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({});

  // Check if user is Admin (AC8: RBAC enforcement)
  const isAdmin = user?.role === 'Admin' || user?.role === 'SYSTEM_ADMIN';

  // Load settings on mount - only when auth is complete and user is admin
  useEffect(() => {
    const loadSettings = async () => {
      if (!authLoading && isAdmin) {
        try {
          const data = await getSystemSettings();
          setSettings(data);
          setFormData({
            businessName: data.businessName,
            address: data.address,
            phone: data.phone,
            email: data.email,
          });
        } catch (err) {
          setError(err instanceof Error ? err.message : 'Failed to load settings');
        } finally {
          setLoading(false);
        }
      } else if (!authLoading && !isAdmin) {
        setLoading(false);
      }
    };

    loadSettings();
  }, [authLoading, isAdmin]);

  // Validate form
  const validateForm = (): boolean => {
    const errors: Record<string, string> = {};

    if (!formData.businessName.trim()) {
      errors.businessName = 'Business name is required';
    }

    if (!formData.email.trim()) {
      errors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      errors.email = 'Invalid email format';
    }

    // Phone validation: allow international formats with optional + prefix
    if (formData.phone.trim() && !/^[\d\s\-+()]+$/.test(formData.phone)) {
      errors.phone = 'Invalid phone format';
    }

    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  // Handle form submission
  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (!validateForm()) {
      return;
    }

    setSaving(true);

    try {
      const response = await updateSystemSettings({
        businessName: formData.businessName,
        address: formData.address,
        phone: formData.phone,
        email: formData.email,
      });

      setSuccess('Settings updated successfully!');

      // Update local state with response data (includes updatedAt)
      setSettings(response);
      // Form data is already current, no need to update
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.detail);
      } else {
        setError(err instanceof Error ? err.message : 'Failed to update settings');
      }
    } finally {
      setSaving(false);
    }
  };

  // Handle input change
  const handleInputChange = (field: string, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    // Clear validation error for this field
    if (validationErrors[field]) {
      setValidationErrors(prev => {
        const newErrors = { ...prev };
        delete newErrors[field];
        return newErrors;
      });
    }
  };

  // Show loading state
  if (authLoading || loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  // Show access denied for non-admin users (AC8)
  if (!isAdmin) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="bg-red-50 border border-red-200 text-red-700 px-6 py-4 rounded-lg">
            <h2 className="text-xl font-bold mb-2">Access Denied</h2>
            <p>Only System Administrators can access system settings.</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">System Settings</h1>
        <p className="text-gray-600 mt-1">
          Configure pharmacy business information and contact details
        </p>
      </div>

      {/* Success notification */}
      {success && (
        <div className="mb-6 bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded">
          {success}
        </div>
      )}

      {/* Error notification */}
      {error && (
        <div className="mb-6 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
          {error}
        </div>
      )}

      <div className="bg-white rounded-lg border shadow-sm">
        <form onSubmit={handleSubmit} className="p-6 space-y-6">
          {/* Business Name (AC1) */}
          <div>
            <label htmlFor="businessName" className="block text-sm font-medium text-gray-700 mb-2">
              Business Name <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="businessName"
              value={formData.businessName}
              onChange={(e) => handleInputChange('businessName', e.target.value)}
              className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                validationErrors.businessName ? 'border-red-500' : 'border-gray-300'
              }`}
              placeholder="Simpo Pharmacy"
              disabled={saving}
            />
            {validationErrors.businessName && (
              <p className="mt-1 text-sm text-red-600">{validationErrors.businessName}</p>
            )}
          </div>

          {/* Address (AC2) */}
          <div>
            <label htmlFor="address" className="block text-sm font-medium text-gray-700 mb-2">
              Address
            </label>
            <textarea
              id="address"
              value={formData.address}
              onChange={(e) => handleInputChange('address', e.target.value)}
              rows={3}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="123 Main St, Jakarta, Indonesia"
              disabled={saving}
            />
          </div>

          {/* Phone (AC3) */}
          <div>
            <label htmlFor="phone" className="block text-sm font-medium text-gray-700 mb-2">
              Phone Number
            </label>
            <input
              type="tel"
              id="phone"
              value={formData.phone}
              onChange={(e) => handleInputChange('phone', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="+62-21-1234-5678"
              disabled={saving}
            />
          </div>

          {/* Email (AC4) */}
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-2">
              Email Address <span className="text-red-500">*</span>
            </label>
            <input
              type="email"
              id="email"
              value={formData.email}
              onChange={(e) => handleInputChange('email', e.target.value)}
              className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                validationErrors.email ? 'border-red-500' : 'border-gray-300'
              }`}
              placeholder="admin@simpo.pharmacy"
              disabled={saving}
            />
            {validationErrors.email && (
              <p className="mt-1 text-sm text-red-600">{validationErrors.email}</p>
            )}
          </div>

          {/* Last updated info (AC5) */}
          {settings && (
            <div className="bg-gray-50 border border-gray-200 rounded-md p-4">
              <p className="text-sm text-gray-600">
                <span className="font-medium">Last updated:</span>{' '}
                {new Date(settings.updatedAt).toLocaleString()}
              </p>
            </div>
          )}

          {/* Submit button */}
          <div className="flex justify-end pt-4 border-t">
            <button
              type="submit"
              disabled={saving}
              className="bg-blue-600 text-white py-2 px-6 rounded-md hover:bg-blue-700 disabled:bg-blue-300 disabled:cursor-not-allowed transition-colors"
            >
              {saving ? 'Saving...' : 'Save Settings'}
            </button>
          </div>
        </form>

        {/* Info section */}
        <div className="px-6 py-4 bg-gray-50 border-t rounded-b-lg">
          <p className="text-sm text-gray-600">
            <span className="font-medium">Note:</span> Changes to system settings will be reflected
            throughout the application including receipts, reports, and the user interface.
            All changes are logged for audit purposes (AC7).
          </p>
        </div>
      </div>
    </div>
  );
}
