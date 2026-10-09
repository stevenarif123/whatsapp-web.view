# Generates assets/notif-telegram.png, the Telegram toast icon: a blue circle with a white paper plane.
# Pure Python (no Pillow): 4x supersampled polygon fill, written as an RGBA PNG.
import struct, zlib, os

N = 128
SS = 4
BLUE = (36, 161, 222)
WHITE = (255, 255, 255)
FOLD = (200, 218, 234)

def inside(poly, x, y):
    c = False
    j = len(poly) - 1
    for i in range(len(poly)):
        xi, yi = poly[i]; xj, yj = poly[j]
        if (yi > y) != (yj > y) and x < (xj - xi) * (y - yi) / (yj - yi) + xi:
            c = not c
        j = i
    return c

plane = [(0.20, 0.50), (0.78, 0.27), (0.60, 0.76), (0.47, 0.62)]
fold = [(0.47, 0.62), (0.60, 0.76), (0.44, 0.73)]

rows = []
for py in range(N):
    row = bytearray()
    for px in range(N):
        acc = [0, 0, 0, 0]
        for sy in range(SS):
            for sx in range(SS):
                x = (px + (sx + 0.5) / SS) / N
                y = (py + (sy + 0.5) / SS) / N
                if (x - 0.5) ** 2 + (y - 0.5) ** 2 > 0.47 ** 2:
                    continue
                col = BLUE
                if inside(plane, x, y): col = WHITE
                if inside(fold, x, y): col = FOLD
                acc[0] += col[0]; acc[1] += col[1]; acc[2] += col[2]; acc[3] += 255
        n = SS * SS
        a = acc[3] // n
        if acc[3]:
            row += bytes([acc[0] * 255 // acc[3], acc[1] * 255 // acc[3], acc[2] * 255 // acc[3], a])
        else:
            row += bytes(4)
    rows.append(b'\x00' + bytes(row))

def chunk(t, c):
    return struct.pack('>I', len(c)) + t + c + struct.pack('>I', zlib.crc32(t + c))

png = (b'\x89PNG\r\n\x1a\n' + chunk(b'IHDR', struct.pack('>IIBBBBB', N, N, 8, 6, 0, 0, 0))
       + chunk(b'IDAT', zlib.compress(b''.join(rows), 9)) + chunk(b'IEND', b''))
out = os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', 'assets', 'notif-telegram.png')
open(out, 'wb').write(png)
print('wrote', os.path.normpath(out), len(png), 'bytes')
