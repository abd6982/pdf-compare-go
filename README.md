# pdf-compare-go

## Usage
Note: filenames are separated by commas
```zsh
go run main.go -f data/test_pdfs/small_test/copied_data.pdf,data/test_pdfs/small_test/orig_data.pdf
```

## Current features

- Duplicate text
- Duplicate digits 

Currently this tool can detect text overlaps between two PDFs, e.g. when text has been copy-pasted from one file into another.

It can also detect duplicated digit sequences, which is useful in finding tables that may have been copy-pasted between two files.