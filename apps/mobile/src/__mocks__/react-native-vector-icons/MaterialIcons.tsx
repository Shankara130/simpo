/**
 * Mock for react-native-vector-icons/MaterialIcons
 * Used in tests to avoid requiring the native module
 */

import React from 'react';

const MockIcon = ({ name, size, color }: any) => {
  // Return a simple View or Text as mock
  return React.createElement('Text', {
    style: { width: size, height: size, color },
  }, `[Icon: ${name}]`);
};

export default MockIcon;
