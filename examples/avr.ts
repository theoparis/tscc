/// <reference path="llvm.d.ts" />
class Arduino {
  private static readonly DDRB = LLVM.indexAddress<number>(0x24); // Data Direction Register B
  private static readonly PORTB = LLVM.indexAddress<number>(0x25); // Port B Data Register

  public static setup(): void {
    // Set PB5 (Digital pin 13) as output
    Arduino.DDRB[0] |= 1 << 5;
  }

  public static loop(): void {
    // Toggle PB5
    Arduino.PORTB[0] ^= 1 << 5;
  }
}

// Entry point
Arduino.setup();
while (true) {
  Arduino.loop();
}
