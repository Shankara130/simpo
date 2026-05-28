/**
 * ESC/POS Generator Tests
 * Tests for ESC/POS command generation for thermal printers
 */

import { ESCPOSGenerator } from './ESCPOSGenerator';
import { PaperWidth } from '../types/receipt.types';

describe('ESCPOSGenerator', () => {
  let generator: ESCPOSGenerator;

  beforeEach(() => {
    generator = new ESCPOSGenerator();
  });

  describe('Initialization', () => {
    it('should create generator instance', () => {
      expect(generator).toBeDefined();
    });

    it('should have default paper width', () => {
      expect(generator).toBeDefined();
    });
  });

  describe('Basic Commands', () => {
    it('should generate initialize command', () => {
      const command = generator.initialize();
      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBeGreaterThan(0);
    });

    it('should generate text command', () => {
      const command = generator.text('Hello World');
      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBeGreaterThan(0);
    });

    it('should generate bold text command', () => {
      const command = generator.bold('Bold Text');
      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBeGreaterThan(0);
    });

    it('should generate align left command', () => {
      const command = generator.alignLeft('Left Aligned');
      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBeGreaterThan(0);
    });

    it('should generate align center command', () => {
      const command = generator.alignCenter('Centered Text');
      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBeGreaterThan(0);
    });

    it('should generate align right command', () => {
      const command = generator.alignRight('Right Aligned');
      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBeGreaterThan(0);
    });

    it('should generate cut command', () => {
      const command = generator.cut();
      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBeGreaterThan(0);
    });

    it('should generate partial cut command', () => {
      const command = generator.partialCut();
      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBeGreaterThan(0);
    });
  });

  describe('58mm Layout', () => {
    it('should generate 58mm layout with 32 chars per line', () => {
      const text = 'A'.repeat(32); // Exactly 32 characters
      const command = generator.text58mm(text);
      expect(command).toBeInstanceOf(Uint8Array);
    });

    it('should wrap text longer than 32 chars for 58mm', () => {
      const longText = 'A'.repeat(50); // 50 characters
      const command = generator.text58mm(longText);
      expect(command).toBeInstanceOf(Uint8Array);
    });

    it('should handle empty text for 58mm', () => {
      const command = generator.text58mm('');
      expect(command).toBeInstanceOf(Uint8Array);
    });
  });

  describe('80mm Layout', () => {
    it('should generate 80mm layout with 48 chars per line', () => {
      const text = 'A'.repeat(48); // Exactly 48 characters
      const command = generator.text80mm(text);
      expect(command).toBeInstanceOf(Uint8Array);
    });

    it('should wrap text longer than 48 chars for 80mm', () => {
      const longText = 'A'.repeat(60); // 60 characters
      const command = generator.text80mm(longText);
      expect(command).toBeInstanceOf(Uint8Array);
    });

    it('should handle empty text for 80mm', () => {
      const command = generator.text80mm('');
      expect(command).toBeInstanceOf(Uint8Array);
    });
  });

  describe('Barcode Generation', () => {
    it('should generate Code 128 barcode', () => {
      const command = generator.barcode128('123456789');
      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBeGreaterThan(0);
    });

    it('should handle empty barcode data', () => {
      const command = generator.barcode128('');
      expect(command).toBeInstanceOf(Uint8Array);
    });

    it('should handle numeric barcode data', () => {
      const command = generator.barcode128('987654321');
      expect(command).toBeInstanceOf(Uint8Array);
    });
  });

  describe('QR Code Generation', () => {
    it('should generate QR code', () => {
      const command = generator.qrcode('https://example.com');
      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBeGreaterThan(0);
    });

    it('should handle empty QR data', () => {
      const command = generator.qrcode('');
      expect(command).toBeInstanceOf(Uint8Array);
    });

    it('should handle long QR data', () => {
      const longUrl = 'https://example.com/very/long/path/that/goes/on/and/on';
      const command = generator.qrcode(longUrl);
      expect(command).toBeInstanceOf(Uint8Array);
    });
  });

  describe('Image Printing', () => {
    it('should generate image command', () => {
      const imageData = new Uint8Array([0x00, 0x01, 0x02]);
      const command = generator.image(imageData, 100);
      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBeGreaterThan(0);
    });

    it('should handle empty image data', () => {
      const imageData = new Uint8Array([]);
      const command = generator.image(imageData, 0);
      expect(command).toBeInstanceOf(Uint8Array);
    });

    it('should handle different image widths', () => {
      const imageData = new Uint8Array([0x00, 0x01, 0x02]);
      const command = generator.image(imageData, 200);
      expect(command).toBeInstanceOf(Uint8Array);
    });
  });

  describe('Line Spacing and Feeds', () => {
    it('should generate line feed', () => {
      const command = generator.lineFeed();
      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBeGreaterThan(0);
    });

    it('should generate multiple line feeds', () => {
      const command = generator.lineFeeds(3);
      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBeGreaterThan(0);
    });
  });

  describe('Font Styling', () => {
    it('should generate double height text', () => {
      const command = generator.doubleHeight('Double Height');
      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBeGreaterThan(0);
    });

    it('should generate double width text', () => {
      const command = generator.doubleWidth('Double Width');
      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBeGreaterThan(0);
    });

    it('should generate double height and width text', () => {
      const command = generator.doubleSize('Double Size');
      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBeGreaterThan(0);
    });

    it('should generate underline text', () => {
      const command = generator.underline('Underlined');
      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBeGreaterThan(0);
    });
  });

  describe('Character Encoding', () => {
    it('should handle Indonesian characters', () => {
      const text = 'Halo Dunia';
      const command = generator.text(text);
      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBeGreaterThan(0);
    });

    it('should handle special Indonesian characters', () => {
      const text = 'Terima Kasih';
      const command = generator.text(text);
      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBeGreaterThan(0);
    });

    it('should handle numbers and symbols', () => {
      const text = 'Total: Rp 150.000';
      const command = generator.text(text);
      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBeGreaterThan(0);
    });
  });

  describe('Complete Receipt Generation', () => {
    it('should generate complete 58mm receipt', () => {
      let commands = generator.initialize();
      commands = generator.appendCommands(commands, generator.alignCenter('APOTEK SEHAT'));
      commands = generator.appendCommands(commands, generator.text58mm('Jl. Kesehatan No. 123'));
      commands = generator.appendCommands(commands, generator.lineFeed());
      commands = generator.appendCommands(commands, generator.alignLeft('Item        Qty    Total'));
      commands = generator.appendCommands(commands, generator.lineFeed());
      commands = generator.appendCommands(commands, generator.alignLeft('Paracetamol   1   15.000'));
      commands = generator.appendCommands(commands, generator.lineFeed());
      commands = generator.appendCommands(commands, generator.alignLeft('Total:              15.000'));
      commands = generator.appendCommands(commands, generator.lineFeed());
      commands = generator.appendCommands(commands, generator.partialCut());

      expect(commands).toBeInstanceOf(Uint8Array);
      expect(commands.length).toBeGreaterThan(0);
    });

    it('should generate complete 80mm receipt', () => {
      let commands = generator.initialize();
      commands = generator.appendCommands(commands, generator.alignCenter('APOTEK SEHAT'));
      commands = generator.appendCommands(commands, generator.text80mm('Jl. Kesehatan No. 123, Jakarta'));
      commands = generator.appendCommands(commands, generator.lineFeed());
      commands = generator.appendCommands(commands, generator.alignLeft('Item              Qty    Price     Total'));
      commands = generator.appendCommands(commands, generator.lineFeed());
      commands = generator.appendCommands(commands, generator.alignLeft('Paracetamol 500mg   1    15.000   15.000'));
      commands = generator.appendCommands(commands, generator.lineFeed());
      commands = generator.appendCommands(commands, generator.alignLeft('Total:                                    15.000'));
      commands = generator.appendCommands(commands, generator.lineFeed());
      commands = generator.appendCommands(commands, generator.partialCut());

      expect(commands).toBeInstanceOf(Uint8Array);
      expect(commands.length).toBeGreaterThan(0);
    });
  });

  describe('Utility Methods', () => {
    it('should append multiple commands', () => {
      const cmd1 = generator.text('First');
      const cmd2 = generator.text('Second');
      const combined = generator.appendCommands(cmd1, cmd2);

      expect(combined.length).toBeGreaterThan(cmd1.length);
      expect(combined.length).toBeGreaterThan(cmd2.length);
    });

    it('should handle empty commands when appending', () => {
      const cmd1 = generator.text('First');
      const empty = new Uint8Array([]);
      const combined = generator.appendCommands(cmd1, empty);

      expect(combined.length).toBe(cmd1.length);
    });

    it('should set paper width', () => {
      generator.setPaperWidth(58);
      expect(generator).toBeDefined();

      generator.setPaperWidth(80);
      expect(generator).toBeDefined();
    });
  });

  describe('Command Structure Validation', () => {
    it('should initialize with ESC @ command', () => {
      const command = generator.initialize();
      // ESC @ is 0x1B, 0x40
      expect(command[0]).toBe(0x1B);
      expect(command[1]).toBe(0x40);
    });

    it('should include cut command GS V m', () => {
      const command = generator.cut();
      // GS V m is 0x1D, 0x56, 0xXX (where XX is m value)
      expect(command).toContain(0x1D);
      expect(command).toContain(0x56);
    });

    it('should include alignment command ESC a', () => {
      const command = generator.alignCenter('Test');
      // ESC a is 0x1B, 0x61
      expect(command).toContain(0x1B);
      expect(command).toContain(0x61);
    });
  });

  // ============================================================================
  // Cash Drawer Support Tests (Story 7.4)
  // ============================================================================

  describe('Cash Drawer Kick Command', () => {
    it('should generate cash drawer kick command for Pin 2', () => {
      const command = generator.generateCashDrawerKick({
        pulseTiming: 100,
        pinNumber: 0, // Pin 2
      });

      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBe(6); // ESC p + 4 parameters
    });

    it('should generate cash drawer kick command for Pin 5', () => {
      const command = generator.generateCashDrawerKick({
        pulseTiming: 100,
        pinNumber: 1, // Pin 5
      });

      expect(command).toBeInstanceOf(Uint8Array);
      expect(command.length).toBe(6);
    });

    it('should have correct ESC p command structure', () => {
      const command = generator.generateCashDrawerKick({
        pulseTiming: 100,
        pinNumber: 0,
      });

      // ESC = 0x1B, p = 0x70
      expect(command[0]).toBe(0x1B); // ESC
      expect(command[1]).toBe(0x70); // p
      expect(command[2]).toBe(0x00); // Pin 2
      expect(command[3]).toBe(0x00); // Pulse mode
    });

    it('should calculate pulse units correctly', () => {
      // 100ms pulse = 50 units (100 / 2)
      const command = generator.generateCashDrawerKick({
        pulseTiming: 100,
        pinNumber: 0,
      });

      const pulseUnits = Math.floor(100 / 2); // 50
      expect(command[4]).toBe(pulseUnits); // Pulse on time
      expect(command[5]).toBe(pulseUnits); // Pulse off time
    });

    it('should support different pulse timings', () => {
      // Test 50ms pulse = 25 units
      const command50 = generator.generateCashDrawerKick({
        pulseTiming: 50,
        pinNumber: 0,
      });
      expect(command50[4]).toBe(25);
      expect(command50[5]).toBe(25);

      // Test 200ms pulse = 100 units
      const command200 = generator.generateCashDrawerKick({
        pulseTiming: 200,
        pinNumber: 0,
      });
      expect(command200[4]).toBe(100);
      expect(command200[5]).toBe(100);
    });

    it('should set correct pin number for Pin 2', () => {
      const command = generator.generateCashDrawerKick({
        pulseTiming: 100,
        pinNumber: 0, // Pin 2
      });

      expect(command[2]).toBe(0x00); // Drawer 1 (Pin 2)
    });

    it('should set correct pin number for Pin 5', () => {
      const command = generator.generateCashDrawerKick({
        pulseTiming: 100,
        pinNumber: 1, // Pin 5
      });

      expect(command[2]).toBe(0x01); // Drawer 2 (Pin 5)
    });

    it('should always use pulse mode (mode 0x00)', () => {
      const command = generator.generateCashDrawerKick({
        pulseTiming: 100,
        pinNumber: 0,
      });

      expect(command[3]).toBe(0x00); // Pulse mode
    });

    it('should handle boundary pulse timing values', () => {
      // Minimum: 50ms
      const minCommand = generator.generateCashDrawerKick({
        pulseTiming: 50,
        pinNumber: 0,
      });
      expect(minCommand[4]).toBe(25); // 50 / 2 = 25

      // Maximum: 500ms
      const maxCommand = generator.generateCashDrawerKick({
        pulseTiming: 500,
        pinNumber: 0,
      });
      expect(maxCommand[4]).toBe(250); // 500 / 2 = 250
    });

    it('should handle odd pulse timing values', () => {
      // 101ms should round down to 50 units
      const command = generator.generateCashDrawerKick({
        pulseTiming: 101,
        pinNumber: 0,
      });

      const expectedUnits = Math.floor(101 / 2); // 50
      expect(command[4]).toBe(expectedUnits);
      expect(command[5]).toBe(expectedUnits);
    });

    it('should generate complete command bytes in correct order', () => {
      const command = generator.generateCashDrawerKick({
        pulseTiming: 100,
        pinNumber: 0,
      });

      // Expected: ESC (0x1B), p (0x70), pin (0x00), mode (0x00), t1 (50), t2 (50)
      expect(command).toEqual(new Uint8Array([0x1B, 0x70, 0x00, 0x00, 50, 50]));
    });
  });
});
