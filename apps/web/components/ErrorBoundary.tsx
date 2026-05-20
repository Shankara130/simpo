/**
 * ErrorBoundary Component
 * React Error Boundary for catching JavaScript errors in component tree
 *
 * This prevents the entire app from crashing when a component throws an error.
 * Displays a fallback UI instead of the broken component.
 */

'use client';

import React, { Component, ErrorInfo, ReactNode } from 'react';

interface ErrorBoundaryProps {
	children: ReactNode;
	fallback?: ReactNode;
	onError?: (error: Error, errorInfo: ErrorInfo) => void;
}

interface ErrorBoundaryState {
	hasError: boolean;
	error: Error | null;
}

/**
 * ErrorBoundary Class Component
 *
 * Catches errors in any component below it in the tree.
 * Logs errors and displays a fallback UI.
 */
export class ErrorBoundary extends Component<
	ErrorBoundaryProps,
	ErrorBoundaryState
> {
	constructor(props: ErrorBoundaryProps) {
		super(props);
		this.state = {
			hasError: false,
			error: null,
		};
	}

	static getDerivedStateFromError(error: Error): ErrorBoundaryState {
		// Update state so the next render will show the fallback UI
		return {
			hasError: true,
			error,
		};
	}

	componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
		// Log the error to an error reporting service (optional)
		console.error('ErrorBoundary caught an error:', error, errorInfo);

		// Call custom error handler if provided
		if (this.props.onError) {
			this.props.onError(error, errorInfo);
		}
	}

	render(): ReactNode {
		if (this.state.hasError) {
			// Custom fallback UI if provided
			if (this.props.fallback) {
				return this.props.fallback;
			}

			// Default fallback UI
			return (
				<div className="bg-red-50 border border-red-500 rounded-md shadow-md p-4 m-4">
					<h2 className="text-red-900 font-semibold text-lg mb-2">
						Something went wrong
					</h2>
					<p className="text-red-700 text-sm mb-2">
						An error occurred while rendering this component.
					</p>
					{this.state.error && (
						<details className="mt-2">
							<summary className="cursor-pointer text-red-800 text-xs font-medium">
								Error details
							</summary>
							<pre className="mt-2 p-2 bg-red-100 rounded text-xs overflow-auto">
								{this.state.error.toString()}
							</pre>
						</details>
					)}
					<button
						onClick={() => window.location.reload()}
						className="mt-3 px-3 py-1.5 text-xs font-medium rounded bg-red-600 text-white hover:bg-red-700 transition-colors"
					>
						Reload Page
					</button>
				</div>
			);
		}

		return this.props.children;
	}
}

/**
 * withErrorBoundary HOC
 *
 * Higher-order component to wrap any component with ErrorBoundary.
 *
 * Usage:
 * ```tsx
 * const MyComponentWithErrorBoundary = withErrorBoundary(MyComponent);
 * ```
 */
export function withErrorBoundary<P extends object>(
	WrappedComponent: React.ComponentType<P>,
	fallback?: ReactNode,
): React.ComponentType<P> {
	return function WithErrorBoundaryWrapper(props: P) {
		return (
			<ErrorBoundary fallback={fallback}>
				<WrappedComponent {...props} />
			</ErrorBoundary>
		);
	};
}
