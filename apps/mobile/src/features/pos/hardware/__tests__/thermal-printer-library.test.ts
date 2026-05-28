/**
 * Thermal Printer Library Installation Test
 * Tests that @finan-me/react-native-thermal-printer library is properly installed and configured
 */

import fs from 'fs';
import path from 'path';

describe('Thermal Printer Library Installation', () => {
  describe('Package Installation', () => {
    it('should be installed in node_modules', () => {
      // Verify the library exists in node_modules
      const libPath = path.resolve(__dirname, '../../../../../node_modules/@finan-me/react-native-thermal-printer');
      expect(fs.existsSync(libPath)).toBe(true);
    });

    it('should have package.json with correct metadata', () => {
      // Verify library metadata
      const libPath = path.resolve(__dirname, '../../../../../node_modules/@finan-me/react-native-thermal-printer');
      const packageJsonPath = path.join(libPath, 'package.json');

      if (fs.existsSync(packageJsonPath)) {
        const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf-8'));
        expect(packageJson.name).toBe('@finan-me/react-native-thermal-printer');
        expect(packageJson.version).toBeDefined();
      }
    });
  });

  describe('Android Permissions Configuration', () => {
    it('should have USB permissions in AndroidManifest.xml', () => {
      // Verify Android permissions are configured
      const manifestPath = path.resolve(__dirname, '../../../../../android/app/src/main/AndroidManifest.xml');
      const manifestContent = fs.readFileSync(manifestPath, 'utf-8');

      expect(manifestContent).toContain('android.permission.USB_PERMISSION');
      expect(manifestContent).toContain('android.hardware.usb.host');
    });

    it('should have Bluetooth permissions in AndroidManifest.xml', () => {
      // Verify Bluetooth permissions are configured
      const manifestPath = path.resolve(__dirname, '../../../../../android/app/src/main/AndroidManifest.xml');
      const manifestContent = fs.readFileSync(manifestPath, 'utf-8');

      expect(manifestContent).toContain('android.permission.BLUETOOTH_SCAN');
      expect(manifestContent).toContain('android.permission.BLUETOOTH_CONNECT');
    });

    it('should have Network permissions in AndroidManifest.xml', () => {
      // Verify Network permissions are configured
      const manifestPath = path.resolve(__dirname, '../../../../../android/app/src/main/AndroidManifest.xml');
      const manifestContent = fs.readFileSync(manifestPath, 'utf-8');

      expect(manifestContent).toContain('android.permission.ACCESS_WIFI_STATE');
      expect(manifestContent).toContain('android.permission.ACCESS_NETWORK_STATE');
    });
  });

  describe('App Configuration', () => {
    it('should have library in dependencies', () => {
      // Verify library is in package.json dependencies
      const packageJsonPath = path.resolve(__dirname, '../../../../../package.json');
      const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf-8'));

      expect(packageJson.dependencies).toBeDefined();
      expect(packageJson.dependencies['@finan-me/react-native-thermal-printer']).toBeDefined();
    });

    it('should have Expo configuration with printer permissions', () => {
      // Verify app.json has proper configuration
      const appJsonPath = path.resolve(__dirname, '../../../../../app.json');
      const appJson = JSON.parse(fs.readFileSync(appJsonPath, 'utf-8'));

      expect(appJson.expo).toBeDefined();
      expect(appJson.expo.android).toBeDefined();
      expect(appJson.expo.android.permissions).toBeDefined();

      const permissions = appJson.expo.android.permissions;
      expect(permissions).toContain('BLUETOOTH_SCAN');
      expect(permissions).toContain('BLUETOOTH_CONNECT');
    });
  });

  describe('Jest Configuration', () => {
    it('should have library in transformIgnorePatterns', () => {
      // Verify Jest can transform the library
      const jestConfigPath = path.resolve(__dirname, '../../../../../jest.config.js');
      const jestConfigContent = fs.readFileSync(jestConfigPath, 'utf-8');

      expect(jestConfigContent).toContain('@finan-me');
    });
  });
});
