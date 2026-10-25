# Builds the app and notification icons from the official logos:
#   WhatsApp.svg.webp          -> icon.ico (16-256 px, app/window/tray icon) and assets/notif-whatsapp.png
#   assets/telegram-logo.png   -> assets/notif-telegram.png (Telegram toast icon; from telegram.org/img/t_logo_2x.png)
# Needs Pillow:  pip install Pillow  then  python tools/make_icons.py
# After changing icon.ico, rebuild rsrc.syso (see README) so the exe embeds the new icon.
import os
from PIL import Image

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
ICO_SIZES = [16, 20, 24, 32, 40, 48, 64, 96, 128, 256]


def square(img):
    """Pads to a square canvas (the WhatsApp logo is 1280x1283) without cropping."""
    img = img.convert('RGBA')
    side = max(img.size)
    canvas = Image.new('RGBA', (side, side), (0, 0, 0, 0))
    canvas.paste(img, ((side - img.width) // 2, (side - img.height) // 2), img)
    return canvas


def main():
    wa = square(Image.open(os.path.join(ROOT, 'WhatsApp.svg.webp')))
    wa256 = wa.resize((256, 256), Image.LANCZOS)
    wa256.save(os.path.join(ROOT, 'icon.ico'), sizes=[(s, s) for s in ICO_SIZES])
    wa256.save(os.path.join(ROOT, 'assets', 'notif-whatsapp.png'), optimize=True)

    tg = square(Image.open(os.path.join(ROOT, 'assets', 'telegram-logo.png')))
    tg.resize((256, 256), Image.LANCZOS).save(os.path.join(ROOT, 'assets', 'notif-telegram.png'), optimize=True)
    print('icon.ico, assets/notif-whatsapp.png, assets/notif-telegram.png written')


if __name__ == '__main__':
    main()
