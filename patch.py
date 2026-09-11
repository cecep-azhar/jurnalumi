import sys

with open('Note/task-qa.md', 'r') as f:
    content = f.read()

content += """### 2026-09-11 12:55 WIB — Idle
Status: skip
PR: -
Ringkasan: Plafon tercapai. 5 PR terbuka (#29, #30, #31, #32, #33). Menunggu review Prof. Cron idle.
"""

with open('Note/task-qa.md', 'w') as f:
    f.write(content)
