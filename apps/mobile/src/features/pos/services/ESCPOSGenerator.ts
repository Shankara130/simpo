/**
 * ESC/POS Command Generator
 * Generates ESC/POS protocol commands for thermal printers
 * Supports 58mm and 80mm paper widths with text formatting, barcodes, and images
 */

import { PaperWidth } from '../types/receipt.types';

/**
 * ESC/POS Command Generator Class
 * Implements standard ESC/POS commands for thermal printers
 */
export class ESCPOSGenerator {
  private paperWidth: PaperWidth;
  private currentEncoding: BufferEncoding;

  // ESC/POS Command Constants
  private static readonly ESC = 0x1B;
  private static readonly GS = 0x1D;
  private static readonly AT = 0x40; // @ - Initialize
  private static readonly A = 0x61; // a - Alignment
  private static readonly E = 0x45; // E - Bold
  private static readonly V = 0x56; // V - Cut
  private static readonly LF = 0x0A; // Line Feed
  private static readonly ZERO = 0x00;
  private static readonly ONE = 0x01;
  private static readonly TWO = 0x02;
  private static readonly THREE = 0x03;

  constructor(paperWidth: PaperWidth = 58) {
    this.paperWidth = paperWidth;
    this.currentEncoding = 'utf8';
  }

  /**
   * Initialize printer
   * ESC @ - Initialize printer
   */
  public initialize(): Uint8Array {
    return new Uint8Array([ESCPOSGenerator.ESC, ESCPOSGenerator.AT]);
  }

  /**
   * Generate text command
   */
  public text(text: string): Uint8Array {
    // Transliterate UTF-8 characters to ASCII for printer compatibility
    const asciiText = this.transliterateToASCII(text);
    const encoder = new TextEncoder();
    const textBytes = encoder.encode(asciiText);
    return textBytes;
  }

  /**
   * Transliterate UTF-8 characters to ASCII equivalents
   * Converts accented and special characters to plain ASCII for thermal printer compatibility
   * Handles Indonesian characters and common diacritics
   */
  private transliterateToASCII(text: string): string {
    // Character mapping for common accented characters
    const charMap: { [key: string]: string } = {
      // Latin Extended-A (common in Indonesian names)
      'Ā': 'A', 'ā': 'a', 'Ē': 'E', 'ē': 'e', 'Ī': 'I', 'ī': 'i',
      'Ō': 'O', 'ō': 'o', 'Ū': 'U', 'ū': 'u',
      'À': 'A', 'Á': 'A', 'Â': 'A', 'Ã': 'A', 'Ä': 'A', 'Å': 'A', 'Æ': 'AE',
      'à': 'a', 'á': 'a', 'â': 'a', 'ã': 'a', 'ä': 'a', 'å': 'a', 'æ': 'ae',
      'È': 'E', 'É': 'E', 'Ê': 'E', 'Ë': 'E',
      'è': 'e', 'é': 'e', 'ê': 'e', 'ë': 'e',
      'Ì': 'I', 'Í': 'I', 'Î': 'I', 'Ï': 'I',
      'ì': 'i', 'í': 'i', 'î': 'i', 'ï': 'i',
      'Ò': 'O', 'Ó': 'O', 'Ô': 'O', 'Õ': 'O', 'Ö': 'O', 'Ø': 'O',
      'ò': 'o', 'ó': 'o', 'ô': 'o', 'õ': 'o', 'ö': 'o', 'ø': 'o',
      'Ù': 'U', 'Ú': 'U', 'Û': 'U', 'Ü': 'U',
      'ù': 'u', 'ú': 'u', 'û': 'u', 'ü': 'u',
      'Ý': 'Y', 'ý': 'y', 'ÿ': 'y',
      'Ç': 'C', 'ç': 'c',
      'Ñ': 'N', 'ñ': 'n',
      'Š': 'S', 'š': 's',
      'Ž': 'Z', 'ž': 'z',
      'Đ': 'D', 'đ': 'd',
      // Common symbols
      '€': 'EUR', '£': 'GBP', '¥': 'JPY', '©': '(c)', '®': '(r)',
      '°': 'deg', '±': '+/-', 'µ': 'u', '¶': 'P',
      '¹': '1', '²': '2', '³': '3',
      '¼': '1/4', '½': '1/2', '¾': '3/4',
    };

    return text.split('').map(char => charMap[char] || char).join('');
  }

  /**
   * Generate bold text
   * ESC E n - Bold on (1) / off (0)
   */
  public bold(text: string): Uint8Array {
    // Transliterate for printer compatibility
    const asciiText = this.transliterateToASCII(text);
    const encoder = new TextEncoder();
    const textBytes = encoder.encode(asciiText);

    const boldOn = new Uint8Array([ESCPOSGenerator.ESC, ESCPOSGenerator.E, ESCPOSGenerator.ONE]);
    const boldOff = new Uint8Array([ESCPOSGenerator.ESC, ESCPOSGenerator.E, ESCPOSGenerator.ZERO]);

    return this.appendCommands(this.appendCommands(boldOn, textBytes), boldOff);
  }

  /**
   * Generate left-aligned text
   * ESC a n - Alignment (0=left, 1=center, 2=right)
   */
  public alignLeft(text: string): Uint8Array {
    const alignCmd = new Uint8Array([ESCPOSGenerator.ESC, ESCPOSGenerator.A, ESCPOSGenerator.ZERO]);
    const textBytes = this.text(text);
    return this.appendCommands(alignCmd, textBytes);
  }

  /**
   * Generate center-aligned text
   */
  public alignCenter(text: string): Uint8Array {
    const alignCmd = new Uint8Array([ESCPOSGenerator.ESC, ESCPOSGenerator.A, ESCPOSGenerator.ONE]);
    const textBytes = this.text(text);
    return this.appendCommands(alignCmd, textBytes);
  }

  /**
   * Generate right-aligned text
   */
  public alignRight(text: string): Uint8Array {
    const alignCmd = new Uint8Array([ESCPOSGenerator.ESC, ESCPOSGenerator.A, ESCPOSGenerator.TWO]);
    const textBytes = this.text(text);
    return this.appendCommands(alignCmd, textBytes);
  }

  /**
   * Generate full cut command
   * GS V m - Cut paper (65=full, 66=partial)
   */
  public cut(): Uint8Array {
    return new Uint8Array([ESCPOSGenerator.GS, ESCPOSGenerator.V, 0x41, ESCPOSGenerator.ZERO]);
  }

  /**
   * Generate partial cut command
   * GS V m - Partial cut with 3 dot feed
   */
  public partialCut(): Uint8Array {
    return new Uint8Array([ESCPOSGenerator.GS, ESCPOSGenerator.V, 0x42, ESCPOSGenerator.ZERO]);
  }

  /**
   * Generate text for 58mm paper (32 chars per line)
   */
  public text58mm(text: string): Uint8Array {
    const maxChars = 32;
    return this.wrapText(text, maxChars);
  }

  /**
   * Generate text for 80mm paper (48 chars per line)
   */
  public text80mm(text: string): Uint8Array {
    const maxChars = 48;
    return this.wrapText(text, maxChars);
  }

  /**
   * Generate Code 128 barcode
   * GS k m - Barcode printing
   */
  public barcode128(data: string): Uint8Array {
    if (!data || data.length === 0) {
      return new Uint8Array([]);
    }

    const encoder = new TextEncoder();
    const dataBytes = encoder.encode(data);

    // Code 128 barcode command: GS k m n d1...dn
    // m = barcode type (73 for Code 128)
    // n = data length
    const barcodeCmd = new Uint8Array([
      ESCPOSGenerator.GS,
      0x6B, // k
      73, // Code 128
      dataBytes.length,
      ...dataBytes
    ]);

    return barcodeCmd;
  }

  /**
   * Generate QR code
   * GS k m - QR code printing
   */
  public qrcode(data: string): Uint8Array {
    if (!data || data.length === 0) {
      return new Uint8Array([]);
    }

    const encoder = new TextEncoder();
    const dataBytes = encoder.encode(data);

    // QR code command structure
    // This is a simplified implementation
    const qrCmd = new Uint8Array([
      ESCPOSGenerator.GS,
      0x6B, // k
      0x51, // QR code
      dataBytes.length,
      ...dataBytes
    ]);

    return qrCmd;
  }

  /**
   * Generate image printing command
   * GS v 0 - Print raster image
   */
  public image(imageData: Uint8Array, width: number): Uint8Array {
    if (!imageData || imageData.length === 0 || width <= 0) {
      return new Uint8Array([]);
    }

    // Image printing command structure
    // This is a simplified implementation
    const height = Math.ceil(imageData.length / (width / 8));

    const imageCmd = new Uint8Array([
      ESCPOSGenerator.GS,
      0x76, // v
      0x30, // 0
      (width >> 8) & 0xFF,
      width & 0xFF,
      (height >> 8) & 0xFF,
      height & 0xFF,
      ...imageData
    ]);

    return imageCmd;
  }

  /**
   * Generate line feed
   */
  public lineFeed(): Uint8Array {
    return new Uint8Array([ESCPOSGenerator.LF]);
  }

  /**
   * Generate multiple line feeds
   */
  public lineFeeds(count: number): Uint8Array {
    const feeds = new Uint8Array(count);
    for (let i = 0; i < count; i++) {
      feeds[i] = ESCPOSGenerator.LF;
    }
    return feeds;
  }

  /**
   * Generate double height text
   * GS ! n - Select size multiplier
   */
  public doubleHeight(text: string): Uint8Array {
    const sizeCmd = new Uint8Array([ESCPOSGenerator.GS, 0x21, 0x01]); // Double height
    const resetCmd = new Uint8Array([ESCPOSGenerator.GS, 0x21, 0x00]); // Reset
    const textBytes = this.text(text);

    return this.appendCommands(this.appendCommands(sizeCmd, textBytes), resetCmd);
  }

  /**
   * Generate double width text
   */
  public doubleWidth(text: string): Uint8Array {
    const sizeCmd = new Uint8Array([ESCPOSGenerator.GS, 0x21, 0x10]); // Double width
    const resetCmd = new Uint8Array([ESCPOSGenerator.GS, 0x21, 0x00]); // Reset
    const textBytes = this.text(text);

    return this.appendCommands(this.appendCommands(sizeCmd, textBytes), resetCmd);
  }

  /**
   * Generate double height and width text
   */
  public doubleSize(text: string): Uint8Array {
    const sizeCmd = new Uint8Array([ESCPOSGenerator.GS, 0x21, 0x11]); // Double both
    const resetCmd = new Uint8Array([ESCPOSGenerator.GS, 0x21, 0x00]); // Reset
    const textBytes = this.text(text);

    return this.appendCommands(this.appendCommands(sizeCmd, textBytes), resetCmd);
  }

  /**
   * Generate underlined text
   * ESC - n - Underline (1/2=on, 0=off)
   */
  public underline(text: string): Uint8Array {
    const underlineOn = new Uint8Array([ESCPOSGenerator.ESC, 0x2D, ESCPOSGenerator.ONE]);
    const underlineOff = new Uint8Array([ESCPOSGenerator.ESC, 0x2D, ESCPOSGenerator.ZERO]);
    const textBytes = this.text(text);

    return this.appendCommands(this.appendCommands(underlineOn, textBytes), underlineOff);
  }

  /**
   * Wrap text to fit specified line width
   */
  private wrapText(text: string, maxChars: number): Uint8Array {
    if (!text || text.length === 0) {
      return new Uint8Array([]);
    }

    const lines: string[] = [];
    let remainingText = text;

    while (remainingText.length > 0) {
      if (remainingText.length <= maxChars) {
        lines.push(remainingText);
        break;
      }

      lines.push(remainingText.substring(0, maxChars));
      remainingText = remainingText.substring(maxChars);
    }

    const encoder = new TextEncoder();
    const lineBytes = lines.map(line => encoder.encode(line + '\n'));

    // Combine all lines
    const totalLength = lineBytes.reduce((sum, bytes) => sum + bytes.length, 0);
    const combined = new Uint8Array(totalLength);

    let offset = 0;
    for (const bytes of lineBytes) {
      combined.set(bytes, offset);
      offset += bytes.length;
    }

    return combined;
  }

  /**
   * Append multiple commands into single Uint8Array
   */
  public appendCommands(...commands: Uint8Array[]): Uint8Array {
    const totalLength = commands.reduce((sum, cmd) => sum + cmd.length, 0);
    const combined = new Uint8Array(totalLength);

    let offset = 0;
    for (const command of commands) {
      if (command.length > 0) {
        combined.set(command, offset);
        offset += command.length;
      }
    }

    return combined;
  }

  /**
   * Set paper width
   */
  public setPaperWidth(width: PaperWidth): void {
    this.paperWidth = width;
  }

  /**
   * Get current paper width
   */
  public getPaperWidth(): PaperWidth {
    return this.paperWidth;
  }

  // ============================================================================
  // Cash Drawer Support (Story 7.4)
  // ============================================================================

  /**
   * Generate ESC/POS cash drawer kick command
   * Format: ESC p 0 m t1 t2
   * ESC = 0x1B, p = 0x70, 0 = drawer number, m = 0x00 (pulse), t1 = on time, t2 = off time
   * @param options - Cash drawer options including pulse timing and pin number
   * @returns Uint8Array containing the cash drawer kick command
   */
  public generateCashDrawerKick(options: {
    pulseTiming: number;
    pinNumber: 0 | 1;
  }): Uint8Array {
    const { pulseTiming, pinNumber } = options;

    // Validate pulse timing range (0-500ms supported by UI, max 255 units = 510ms)
    if (pulseTiming < 0 || pulseTiming > 500) {
      throw new Error(`Pulse timing must be between 0 and 500ms, got ${pulseTiming}ms`);
    }

    // Convert pulse timing from milliseconds to ESC/POS units (2ms per unit)
    const pulseUnits = Math.floor(pulseTiming / 2);

    // Validate pulse units fit in ESC/POS single byte (max 255)
    if (pulseUnits > 255) {
      throw new Error(`Pulse timing ${pulseTiming}ms exceeds maximum supported (510ms)`);
    }

    // Validate pin number is 0 or 1
    if (pinNumber !== 0 && pinNumber !== 1) {
      throw new Error(`Pin number must be 0 (pin 2) or 1 (pin 5), got ${pinNumber}`);
    }

    // ESC p command: 0x1B 0x70 [drawer] [mode] [t1] [t2]
    // drawer: 0x00 = pin 2, 0x01 = pin 5
    // mode: 0x00 = pulse mode, 0x01 = steady mode
    return new Uint8Array([
      ESCPOSGenerator.ESC, // 0x1B
      0x70,                 // p
      pinNumber,            // Drawer pin (0 = pin 2, 1 = pin 5)
      0x00,                 // Pulse mode
      pulseUnits,           // Pulse on time (t1)
      pulseUnits,           // Pulse off time (t2)
    ]);
  }
}

// Export for convenience
export default ESCPOSGenerator;
