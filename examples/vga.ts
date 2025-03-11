class VGA {
  private static readonly VGA_ADDRESS = 0xb8000;
  private static readonly VGA_WIDTH = 80;
  private static readonly VGA_HEIGHT = 25;
  private static readonly VGA_SIZE = VGA.VGA_WIDTH * VGA.VGA_HEIGHT * 2;

  private static readonly VGA_BUFFER = LLVM.addressToArrayBuffer<number>(
    VGA.VGA_ADDRESS
  );

  private static cursorX = 0;
  private static cursorY = 0;

  private static setCursor(x: number, y: number): void {
    VGA.cursorX = x;
    VGA.cursorY = y;
  }

  public static getCursor(): [number, number] {
    return [VGA.cursorX, VGA.cursorY];
  }

  public static clearScreen(): void {
    for (let i = 0; i < VGA.VGA_SIZE; i++) {
      VGA.VGA_BUFFER[i] = 0;
    }

    VGA.setCursor(0, 0);
  }

  private static writeChar(char: string, x: number, y: number): void {
    const buffer = VGA.VGA_BUFFER;
    const width = VGA.VGA_WIDTH;
    const index = y * width * 2 + x * 2;

    buffer[index] = char.charCodeAt(0);
    buffer[index + 1] = 0x07;
  }

  public static writeString(str: string): void {
    const [x, y] = VGA.getCursor();

    for (let i = 0; i < str.length; i++) {
      const char = str[i];

      if (char === "\n") {
        VGA.setCursor(0, y + 1);
      } else {
        VGA.writeChar(char, x + i, y);
      }
    }
  }
}

export const kernelMain = (): void => {
  VGA.clearScreen();
  VGA.writeString("Hello, World!\n");

  while (true) {}
}
