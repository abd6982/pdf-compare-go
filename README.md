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

## Performance

| Step 				| Run time (10 page files) | Run time (500 page files) |
| ----------------- | ------------------------ | ------------------------- |
| Read PDF 			| 0.3 sec 	 			   | 1 sec 					   |
| Run analytics 	| 0.1 sec     			   | 0.4 sec 				   |

This script can be described as a two-step process. The first step is reading the PDF file. For this we currently use the Go bindings for MuPDF, a popular package written in C++. Reading a PDF takes about 0.3-0.5 seconds depending on size.

### Profiling

(Read more)[https://go.dev/blog/pprof]

Run to generate profile:
```zsh
go run main.go -f file1.pdf,file2.pdf -cpuprofile myprofile.prof
```

Interpret profile:
```zsh
go tool pprof myprofile.prof
```

You will enter "interactive mode". Type `top10` to see most time intensive processes.
