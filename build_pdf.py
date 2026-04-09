#!/usr/bin/env python3
"""Convert VULNERABILITY_REPORT.md -> VULNERABILITY_REPORT.pdf (formal document style)"""

import markdown
import re
import subprocess
import sys
import os

MD_FILE   = "VULNERABILITY_REPORT.md"
HTML_FILE = "VULNERABILITY_REPORT.html"
PDF_FILE  = "VULNERABILITY_REPORT.pdf"

with open(MD_FILE, encoding="utf-8") as f:
    md_text = f.read()

extensions = ["tables", "fenced_code", "codehilite", "toc", "nl2br", "sane_lists"]
md = markdown.Markdown(extensions=extensions,
                       extension_configs={"codehilite": {"css_class": "highlight", "linenums": False}})
body = md.convert(md_text)

# Convert GitHub-style alerts to plain formal callout boxes
ALERT_LABELS = {
    "NOTE":      "NOTE",
    "TIP":       "TIP",
    "IMPORTANT": "IMPORTANT",
    "WARNING":   "WARNING",
    "CAUTION":   "CAUTION",
}

def replace_alert(m):
    kind = m.group(1).upper()
    content = m.group(2).strip()
    label = ALERT_LABELS.get(kind, kind)
    return (
        '<div class="callout">'
        f'<span class="callout-label">{label}:</span> {content}'
        '</div>'
    )

body = re.sub(
    r'<blockquote>\s*<p>\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]\s*(.*?)</p>\s*</blockquote>',
    replace_alert, body, flags=re.DOTALL | re.IGNORECASE
)

# Insert page break before each issue heading (h1 that is not the cover heading)
body = re.sub(
    r'(<h1[^>]*>(?!Security &amp;))',
    r'<div class="page-break"></div>\1',
    body
)

HTML = """<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Vulnerability Report - mx-chain-proxy-go</title>
<style>

* { box-sizing: border-box; margin: 0; padding: 0; }

body {
  font-family: "Times New Roman", Times, serif;
  font-size: 11pt;
  line-height: 1.65;
  color: #000;
  background: #fff;
}

/* Document header block */
.doc-header {
  border-bottom: 2px solid #000;
  padding-bottom: 14px;
  margin-bottom: 22px;
}
.doc-header .doc-title {
  font-size: 16pt;
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 1px;
  margin-bottom: 10px;
}
.doc-meta-table {
  width: auto;
  border-collapse: collapse;
  font-size: 10.5pt;
}
.doc-meta-table td {
  border: none;
  padding: 2px 20px 2px 0;
  background: none;
  vertical-align: top;
}
.doc-meta-table td:first-child {
  font-weight: bold;
  width: 160px;
  white-space: nowrap;
}

/* Headings */
h1 {
  font-family: "Times New Roman", Times, serif;
  font-size: 13pt;
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-top: 2px solid #000;
  border-bottom: 1px solid #000;
  padding: 6px 0;
  margin: 28px 0 14px 0;
  page-break-after: avoid;
}

h2 {
  font-family: "Times New Roman", Times, serif;
  font-size: 11.5pt;
  font-weight: bold;
  border-bottom: 1px solid #000;
  padding-bottom: 3px;
  margin: 22px 0 10px 0;
  page-break-after: avoid;
}

h3 {
  font-family: "Times New Roman", Times, serif;
  font-size: 11pt;
  font-weight: bold;
  margin: 14px 0 6px 0;
  page-break-after: avoid;
}

h4 {
  font-size: 11pt;
  font-weight: bold;
  font-style: italic;
  margin: 10px 0 5px 0;
}

/* Text */
p { margin: 0 0 8px 0; text-align: justify; }
ul, ol { margin: 0 0 8px 0; padding-left: 22px; }
li { margin-bottom: 3px; }

/* Inline code */
code {
  font-family: "Courier New", Courier, monospace;
  font-size: 9.5pt;
  background: #efefef;
  border: 1px solid #ccc;
  padding: 0 3px;
  border-radius: 1px;
}

/* Code blocks */
pre {
  font-family: "Courier New", Courier, monospace;
  font-size: 9pt;
  background: #f5f5f5;
  border: 1px solid #bbb;
  border-left: 3px solid #444;
  padding: 10px 14px;
  margin: 10px 0;
  line-height: 1.45;
  white-space: pre-wrap;
  word-wrap: break-word;
  page-break-inside: avoid;
}
pre code {
  background: none;
  border: none;
  padding: 0;
  font-size: inherit;
}

/* Strip syntax highlight colors - keep monochrome */
.highlight { background: #f5f5f5; }
.highlight span { color: inherit !important; }
.highlight .k, .highlight .kd, .highlight .kn { font-weight: bold; }
.highlight .c, .highlight .c1, .highlight .cm { color: #555 !important; font-style: italic; }
.highlight pre { background: #f5f5f5; border: none; padding: 0; margin: 0; }

/* Tables */
table {
  border-collapse: collapse;
  width: 100%;
  margin: 10px 0;
  font-size: 10.5pt;
  page-break-inside: avoid;
}
th {
  background: #e0e0e0;
  border: 1px solid #888;
  padding: 6px 10px;
  text-align: left;
  font-weight: bold;
}
td {
  border: 1px solid #888;
  padding: 5px 10px;
  vertical-align: top;
}
tbody tr:nth-child(even) { background: #f9f9f9; }

/* Formal callout box (replaces colourful alerts) */
.callout {
  border: 1px solid #555;
  border-left: 4px solid #000;
  padding: 8px 12px;
  margin: 12px 0;
  font-size: 10.5pt;
  background: #f5f5f5;
  page-break-inside: avoid;
}
.callout-label {
  font-weight: bold;
  font-variant: small-caps;
  letter-spacing: 0.5px;
}

/* Horizontal rule */
hr {
  border: none;
  border-top: 1px solid #888;
  margin: 28px 0;
}

/* Page break */
.page-break {
  page-break-before: always;
  height: 0;
}

/* Fallback blockquote */
blockquote {
  border-left: 3px solid #555;
  margin: 10px 0;
  padding: 5px 12px;
  color: #333;
  font-style: italic;
}

/* Links - plain black for print */
a { color: #000; text-decoration: none; }

</style>
</head>
<body>

<div class="doc-header">
  <div class="doc-title">Vulnerability &amp; Code Review Report</div>
  <table class="doc-meta-table">
    <tr><td>Project:</td><td>mx-chain-proxy-go</td></tr>
    <tr><td>Domain:</td><td>RWA / DRWA / Core</td></tr>
    <tr><td>Date:</td><td>2026-04-08</td></tr>
    <tr><td>Reviewer:</td><td>Antigravity Code Analysis</td></tr>
    <tr><td>Template:</td><td>8-Section PR Analysis Format</td></tr>
    <tr><td>Total Issues:</td><td>10 &nbsp;(2 Critical &middot; 5 Medium &middot; 3 Informational)</td></tr>
    <tr><td>Overall Status:</td><td><strong>REJECT &mdash; 8 issues require fixes before merge</strong></td></tr>
  </table>
</div>

""" + body + """

</body>
</html>
"""

with open(HTML_FILE, "w", encoding="utf-8") as f:
    f.write(HTML)
print("HTML written -> " + HTML_FILE)

# Chrome headless PDF
chrome_candidates = [
    "/usr/bin/google-chrome", "google-chrome",
    "google-chrome-stable", "chromium-browser", "chromium",
]
chrome = None
for c in chrome_candidates:
    if os.path.exists(c):
        chrome = c
        break
    r = subprocess.run(["which", c], capture_output=True)
    if r.returncode == 0:
        chrome = c
        break

if not chrome:
    print("Chrome not found. Open HTML file in browser and print to PDF.")
    sys.exit(1)

abs_html = os.path.abspath(HTML_FILE)
abs_pdf  = os.path.abspath(PDF_FILE)

cmd = [
    chrome,
    "--headless=new",
    "--disable-gpu",
    "--no-sandbox",
    "--disable-dev-shm-usage",
    "--run-all-compositor-stages-before-draw",
    "--virtual-time-budget=5000",
    "--print-to-pdf=" + abs_pdf,
    "--print-to-pdf-no-header",
    "file://" + abs_html,
]

print("Generating PDF via Chrome headless...")
result = subprocess.run(cmd, capture_output=True, text=True, timeout=90)

if result.returncode == 0 and os.path.exists(abs_pdf) and os.path.getsize(abs_pdf) > 20_000:
    size_kb = os.path.getsize(abs_pdf) // 1024
    print("PDF written -> " + PDF_FILE + "  (" + str(size_kb) + " KB)")
else:
    print("Chrome PDF generation failed")
    if result.stderr:
        print(result.stderr[-800:])
    sys.exit(1)
