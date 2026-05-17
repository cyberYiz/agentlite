# PDF Form Filling

This reference covers filling PDF forms programmatically.

## Using pypdf

```python
from pypdf import PdfReader, PdfWriter

reader = PdfReader("form.pdf")
writer = PdfWriter()

# Clone the form
writer.append(reader)

# Fill form fields
writer.update_page_form_field_values(writer.pages[0], {
    "name": "John Doe",
    "email": "john@example.com",
    "date": "2025-01-15"
})

# Make read-only (optional)
writer.set_need_appearances_writer(False)

with open("filled_form.pdf", "wb") as f:
    writer.write(f)
```

## Notes

- Use `reader.get_fields()` to list all form field names
- Form field names are case-sensitive
- Check the field type before filling (text, checkbox, etc.)
