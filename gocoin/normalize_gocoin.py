#!/usr/bin/env python3
"""
normalize_gocoin.py
1. 统一 package 名为 gocoin（main.go 除外）
2. 删除所有子包 import 行
3. 删除空 import 块
4. 删除空目录
"""
import re
from pathlib import Path
import shutil

ROOT = Path(__file__).resolve().parent
SUB_PKGS = ('blockchain', 'p2p', 'wallet')

for go in ROOT.rglob('*.go'):
    txt = orig = go.read_text(encoding='utf-8')

    # 1. package 行
    if go.name == 'main.go':
        txt = re.sub(r'^package \w+$', 'package main', txt, flags=re.M)
    else:
        txt = re.sub(r'^package \w+$', 'package gocoin', txt, flags=re.M)

    # 2. 删除子包 import 行
    for sub in SUB_PKGS:
        txt = re.sub(r'^\s*"(my-blockchain/gocoin/)?' + sub + r'"$\n', '', txt, flags=re.M)

    # 3. 清理空 import 块
    txt = re.sub(r'import\s+\(\s*\)\s*\n', '', txt)

    if txt != orig:
        go.write_text(txt, encoding='utf-8')
        print(f'fixed  ->  {go.relative_to(ROOT)}')

# 4. 删空目录
for folder in sorted(ROOT.rglob('*'), key=lambda p: len(p.parts), reverse=True):
    if folder.is_dir() and not any(folder.iterdir()):
        shutil.rmtree(folder)
        print(f'rmdir  ->  {folder.relative_to(ROOT)}')

print('all done!')