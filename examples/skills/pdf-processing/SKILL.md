---
name: pdf-processing
description: Extract PDF text and tables, fill forms, merge and split PDFs. Use when working with PDF files or when the user mentions PDFs, forms, or document extraction.
license: MIT
metadata:
  author: agent-sdk
  version: "1.0"
---

# PDF Processing Guide

## Overview

This skill provides PDF processing capabilities. When activated, use the instructions
below to handle PDF-related tasks.

## Quick Start

### Extract Text from PDF

Use `pdftotext` (from poppler-utils) for fast text extraction:

```bash
pdftotext input.pdf output.txt
pdftotext -layout input.pdf output.txt  # Preserve layout
pdftotext -f 1 -l 5 input.pdf output.txt  # Pages 1-5
```

### Merge PDFs

```bash
# Using qpdf
qpdf --empty --pages file1.pdf file2.pdf -- merged.pdf
```

### Split PDFs

```bash
# Extract pages 1-5
qpdf input.pdf --pages . 1-5 -- output.pdf
```

## Python Libraries

### pypdf - Basic Operations

```python
from pypdf import PdfReader, PdfWriter

# Read PDF
reader = PdfReader("document.pdf")
print(f"Pages: {len(reader.pages)}")

# Extract text
for page in reader.pages:
    print(page.extract_text())

# Merge PDFs
writer = PdfWriter()
for pdf_file in ["doc1.pdf", "doc2.pdf"]:
    reader = PdfReader(pdf_file)
    for page in reader.pages:
        writer.add_page(page)
with open("merged.pdf", "wb") as f:
    writer.write(f)
```

### pdfplumber - Table Extraction

```python
import pdfplumber

with pdfplumber.open("document.pdf") as pdf:
    for page in pdf.pages:
        tables = page.extract_tables()
        for table in tables:
            for row in table:
                print(row)
```

## Common Tasks

- **Extract text**: Use `pdftotext` (fastest) or `pypdf` (Python)
- **Extract tables**: Use `pdfplumber`
- **Merge PDFs**: Use `pypdf` or `qpdf`
- **Split PDFs**: Use `qpdf`
- **OCR scanned PDFs**: Use `pytesseract` + `pdf2image`
- **Fill forms**: See `references/FORMS.md`

## References

- See `references/FORMS.md` for PDF form filling instructions
- See `references/TROUBLESHOOTING.md` for common issues
