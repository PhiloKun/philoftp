#!/usr/bin/env python3
"""生成 PhiloFTP 图标（.ico 供安装程序/快捷方式，.png 供托盘/源码 embed）。
用法：python3 gen_icon.py <输出目录>
"""
import sys
import os
from PIL import Image, ImageDraw

def make_icon(size):
    img = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    d = ImageDraw.Draw(img)
    bg = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    bd = ImageDraw.Draw(bg)
    for y in range(size):
        t = y / size
        r = int(13 + 20 * t)
        g = int(211 + 24 * t)
        b = int(238 + 0 * t)
        bd.line([(0, y), (size, y)], fill=(r, g, b, 255))
    mask = Image.new('L', (size, size), 0)
    md = ImageDraw.Draw(mask)
    radius = int(size * 0.22)
    md.rounded_rectangle([0, 0, size - 1, size - 1], radius=radius, fill=255)
    img = Image.composite(bg, img, mask)
    d = ImageDraw.Draw(img)
    lw = max(2, int(size * 0.08))
    cy = size * 0.55
    d.rounded_rectangle([size * 0.28, cy + size * 0.16, size * 0.72, cy + size * 0.22], radius=lw, fill=(255, 255, 255, 255))
    d.rounded_rectangle([size * 0.30, cy - size * 0.10, size * 0.70, cy + size * 0.16], radius=lw,
                        fill=(6, 18, 26, 255), outline=(255, 255, 255, 255), width=lw)
    d.line([(size * 0.5, cy - size * 0.02), (size * 0.5, cy + size * 0.06)], fill=(34, 211, 238, 255), width=lw)
    d.polygon([(size * 0.5 - lw * 2, cy - size * 0.02), (size * 0.5 + lw * 2, cy - size * 0.02),
               (size * 0.5, cy - size * 0.12)], fill=(34, 211, 238, 255))
    d.ellipse([size * 0.40, cy - size * 0.05, size * 0.44, cy - size * 0.01], fill=(34, 211, 238, 255))
    d.ellipse([size * 0.56, cy - size * 0.05, size * 0.60, cy - size * 0.01], fill=(34, 211, 238, 255))
    return img

def main():
    out = sys.argv[1] if len(sys.argv) > 1 else '.'
    os.makedirs(out, exist_ok=True)
    sizes = [16, 24, 32, 48, 64, 128, 256]
    imgs = [make_icon(s) for s in sizes]
    imgs[-1].save(os.path.join(out, 'philoftp.ico'), format='ICO', sizes=[(s, s) for s in sizes])
    imgs[-1].save(os.path.join(out, 'philoftp-icon.png'))
    # 同时输出 256 png 供托盘 embed（写到 assets/icon.png）
    imgs[-1].save(os.path.join(out, '..', '..', 'assets', 'icon.png'))
    print('icons generated to', out)

if __name__ == '__main__':
    main()
