#!/usr/bin/env python3
"""
strip_subpkg_prefix.py
暴力删除所有 blockchain. / p2p. / wallet. 前缀
"""
import re
from pathlib import Path

ROOT   = Path(__file__).resolve().parent
PREFIX = ('blockchain.', 'p2p.', 'wallet.')

for go in ROOT.rglob('*.go'):
    txt = orig = go.read_text(encoding='utf-8')

    for pre in PREFIX:
        # 1. 去掉前缀
        txt = re.sub(rf'\b{re.escape(pre)}', '', txt)
        # 2. 如果行只剩空白或只剩一个标识符，也删掉（可选）
        # txt = re.sub(rf'^\s*{re.escape(pre.rstrip("."))}\s*$\n', '', txt, flags=re.M)

    if txt != orig:
        go.write_text(txt, encoding='utf-8')
        print('fixed ->', go.relative_to(ROOT))