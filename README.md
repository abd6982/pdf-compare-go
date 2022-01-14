# PDF comparison tool written in Go

## Setup
Go can be downloaded and installed from here: https://go.dev/doc/install

This package requires several dependencies. To download all of them, simply run:
```zsh
go get
```

This automatically goes through the code and downloads all needed dependencies.

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

Performance numbers for 2019 MacBook Pro, 2.3 GHz 8-Core Intel Core i9, 16 GB RAM, 4 GB Graphics:

| Step 				| Run time (10 page files) | Run time (500 page files) |
| ----------------- | ------------------------ | ------------------------- |
| Read PDF 			| 0.3 sec 	 			   | 1 sec 					   |
| Run analytics 	| <0.1 sec    			   | 0.4 sec 				   |

This script can be described as a two-step process. The first step is reading the PDF file. For this we currently use the Go bindings for MuPDF, a popular package written in C++. Reading a PDF takes about 0.3-1 seconds depending on size.

The second step is the analytics to compare the texts and digits. For smaller files this can take dozens of milliseconds. For a pair of larger files (hundreds of pages) this takes about half a second.

### Profiling

[Click here](https://go.dev/blog/pprof) for general info about profiling in Go.

Run to generate profile:
```zsh
go run main.go -f file1.pdf,file2.pdf -cpuprofile myprofile.prof
```

Interpret profile:
```zsh
go tool pprof myprofile.prof
```

You will enter "interactive mode". Type `top10` to see most time intensive processes.
