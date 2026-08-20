#!/usr/bin/env python3
"""Rebuild one document in one locale.

Code fences are copied verbatim; every other non-blank line must appear in the
translation map, so a missing line fails loudly instead of silently shipping
English inside an otherwise translated page.
"""
import json, pathlib, sys

slug, locale, map_file = sys.argv[1], sys.argv[2], sys.argv[3]
root = pathlib.Path(__file__).resolve().parent.parent / 'content'
source = (root / f'{slug}.md').read_text(encoding='utf-8').split('\n')
data = json.loads(pathlib.Path(map_file).read_text(encoding='utf-8'))
lines, out, fence, front, missing = data['lines'], [], False, 0, []

for line in source:
    stripped = line.strip()
    if stripped.startswith('```'):
        fence = not fence
        out.append(line); continue
    if fence or stripped == '':
        out.append(line); continue
    if stripped == '---' and front < 2:
        front += 1
        out.append(line); continue
    if front == 1:
        out.append(f'title: {data["title"]}' if stripped.startswith('title:') else line)
        continue
    if stripped in lines:
        out.append(line.replace(stripped, lines[stripped]))
    else:
        missing.append(stripped)
        out.append(line)

if missing:
    print(f'{slug}/{locale}: {len(missing)} untranslated lines', file=sys.stderr)
    for item in missing[:6]:
        print(f'  · {item[:110]}', file=sys.stderr)
    sys.exit(1)

target = root / locale / f'{slug}.md'
target.parent.mkdir(parents=True, exist_ok=True)
target.write_text('\n'.join(out), encoding='utf-8')
print(f'{locale}/{slug}.md')
