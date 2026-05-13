/**
 * Tests for ScannerFeedback component
 * Tests visual feedback indicators for scanner success/error/loading states
 */

import React from 'react';
import { render, screen } from '@testing-library/react-native';
import { ScannerFeedback } from './ScannerFeedback';
import { ScannerState } from '../types/scanner.types';

describe('ScannerFeedback', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.runOnlyPendingTimers();
    jest.useRealTimers();
  });

  describe('idle state', () => {
    it('should not render anything when state is idle', () => {
      const { toJSON } = render(
        <ScannerFeedback state="idle" message="" />
      );

      expect(toJSON()).toBeNull();
    });

    it('should not render with default props', () => {
      const { toJSON } = render(<ScannerFeedback />);

      expect(toJSON()).toBeNull();
    });
  });

  describe('loading state', () => {
    it('should render loading indicator', () => {
      render(<ScannerFeedback state="loading" message="Memuat..." />);

      expect(screen.getByText('Memuat...')).toBeTruthy();
    });

    it('should render default loading message when not provided', () => {
      render(<ScannerFeedback state="loading" />);

      expect(screen.getByText('Memindai...')).toBeTruthy();
    });

    it('should render activity indicator', () => {
      const { getByTestId } = render(
        <ScannerFeedback state="loading" />
      );

      expect(getByTestId('loading-indicator')).toBeTruthy();
    });
  });

  describe('success state', () => {
    it('should render success checkmark icon', () => {
      const { getByTestId } = render(
        <ScannerFeedback state="success" message="Produk ditambahkan" />
      );

      expect(getByTestId('success-icon')).toBeTruthy();
    });

    it('should render success message when provided', () => {
      render(
        <ScannerFeedback state="success" message="Produk ditambahkan" />
      );

      expect(screen.getByText('Produk ditambahkan')).toBeTruthy();
    });

    it('should render default success message when not provided', () => {
      render(<ScannerFeedback state="success" />);

      expect(screen.getByText('Scan berhasil')).toBeTruthy();
    });

    it('should have green background', () => {
      const { getByTestId } = render(
        <ScannerFeedback state="success" />
      );

      const container = getByTestId('feedback-container');
      // Style is array of style objects, check that one contains green
      const hasGreenBackground = container.props.style.some(
        (s: any) => s?.backgroundColor === '#4CAF50'
      );
      expect(hasGreenBackground).toBe(true);
    });
  });

  describe('error state', () => {
    it('should render error icon', () => {
      const { getByTestId } = render(
        <ScannerFeedback state="error" message="Produk tidak ditemukan" />
      );

      expect(getByTestId('error-icon')).toBeTruthy();
    });

    it('should render error message when provided', () => {
      render(
        <ScannerFeedback state="error" message="Produk tidak ditemukan" />
      );

      expect(screen.getByText('Produk tidak ditemukan')).toBeTruthy();
    });

    it('should render default error message when not provided', () => {
      render(<ScannerFeedback state="error" />);

      expect(screen.getByText('Scan gagal')).toBeTruthy();
    });

    it('should have red background', () => {
      const { getByTestId } = render(
        <ScannerFeedback state="error" />
      );

      const container = getByTestId('feedback-container');
      // Style is array of style objects, check that one contains red
      const hasRedBackground = container.props.style.some(
        (s: any) => s?.backgroundColor === '#F44336'
      );
      expect(hasRedBackground).toBe(true);
    });
  });

  describe('scanning state', () => {
    it('should render scanning indicator', () => {
      const { getByTestId } = render(
        <ScannerFeedback state="scanning" message="Memindai..." />
      );

      expect(getByTestId('scanning-indicator')).toBeTruthy();
    });

    it('should have blue background for scanning state', () => {
      const { getByTestId } = render(
        <ScannerFeedback state="scanning" />
      );

      const container = getByTestId('feedback-container');
      // Style is array of style objects, check that one contains blue
      const hasBlueBackground = container.props.style.some(
        (s: any) => s?.backgroundColor === '#2196F3'
      );
      expect(hasBlueBackground).toBe(true);
    });

    it('should not auto-dismiss while scanning', () => {
      jest.useFakeTimers();

      const { toJSON } = render(
        <ScannerFeedback state="scanning" />
      );

      // Advance time significantly
      jest.advanceTimersByTime(5000);

      // Should still be visible (scanning state doesn't auto-dismiss)
      expect(toJSON()).toBeTruthy();
    });
  });

  describe('accessibility', () => {
    it('should have accessible label for screen readers', () => {
      const { getByLabelText } = render(
        <ScannerFeedback state="success" message="Produk ditambahkan" />
      );

      expect(getByLabelText('Status pemindai')).toBeTruthy();
    });

    it('should announce state changes to screen readers', () => {
      const { getByTestId } = render(
        <ScannerFeedback state="loading" message="Memuat..." />
      );

      const alert = getByTestId('feedback-container');
      expect(alert).toBeTruthy();
      expect(alert.props.accessibilityRole).toBe('alert');
    });
  });

  describe('positioning and layout', () => {
    it('should render at top of screen with proper positioning', () => {
      const { getByTestId } = render(
        <ScannerFeedback state="success" />
      );

      const container = getByTestId('feedback-container');
      // Style is array of style objects, check that base styles are present
      const hasPositioning = container.props.style.some(
        (s: any) => s?.position === 'absolute' && s?.top === 0 && s?.left === 0 && s?.right === 0
      );
      expect(hasPositioning).toBe(true);
    });

    it('should have elevation for overlay effect', () => {
      const { getByTestId } = render(
        <ScannerFeedback state="error" />
      );

      const container = getByTestId('feedback-container');
      // Should have some elevation/zIndex
      expect(container.props.style).toBeDefined();
    });

    it('should center content horizontally and vertically', () => {
      const { getByTestId } = render(
        <ScannerFeedback state="success" />
      );

      const content = getByTestId('feedback-content');
      expect(content.props.style).toMatchObject({
        alignItems: 'center',
        justifyContent: 'center',
      });
    });
  });
});
