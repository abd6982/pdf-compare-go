// Sasha Trubetskoy
// sasha@kartographia.com

package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"index/suffixarray"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	// "bytes"
	// "github.com/ledongthuc/pdf"
	"github.com/gen2brain/go-fitz"
)

// UTILS - pure utility functions
//------------------------------------------------------------------------------
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

func parseFilenameString(s string) []string {
	// string 'data/file1.pdf,"data folder/file2.pdf"'
	//				-> ['data/file1.pdf', 'data folder/file2.pdf']
	// https://stackoverflow.com/questions/47489745/splitting-a-string-at-space-except-inside-quotation-marks
	r := csv.NewReader(strings.NewReader(s))
	r.Comma = ',' // separator
	fields, err := r.Read()
	if err != nil {
		panic(err)
	}
	var cleanedFields []string
	for _, s := range fields {
		s2 := strings.Trim(s, "\"")
		cleanedFields = append(cleanedFields, s2)
	}
	return fields
}

// COMMON SUBSTRING - functions and structs for lcs and text comparison
//------------------------------------------------------------------------------
func getSuffixArray(text string) []int {
	// Get suffix array of string.
	// Suppose we have a string s = "banana$"
	// suffixArray(s) = [6, 5, 3, 1, 0, 4, 2]
	// In the suffix array, 6 means s[6:] which is "$",
	// 5 means s[5:] which is "a$", etc. in alphabetical order of suffixes.
	// The suffixes, in alphabetical order, are:
	// 	- $
	// 	- a$
	// 	- ana$
	// 	- anana$
	// 	- banana$
	// 	- na$
	// 	- nana$

	var byteContent = []byte(text)
	var index *suffixarray.Index = suffixarray.New(byteContent)

	// Hack to extract out suffix array.
	// The go standard library has a highly optimized suffix array
	// function but the actual suffix array is a hidden value
	// so we have to access it using weird creative methods.
	indexCopy := reflect.Indirect(reflect.ValueOf(index))
	var saVal reflect.Value = indexCopy.FieldByName("sa")
	var saStr string = fmt.Sprintf("%v", saVal)
	var numsOnly string = strings.Trim(saStr, "{}[] ")
	var numStrsArr []string = strings.Fields(numsOnly)
	var suffArr = []int{}

	for _, i := range numStrsArr {
		j, err := strconv.Atoi(i)
		if err != nil {
			panic(err)
		}
		suffArr = append(suffArr, j)
	}

	return suffArr
}

func kasai(txt string, suffixArr []int) []int {
	// Suppose we have a string s = "banana$"
	// suffixArray(s) = [6, 5, 3, 1, 0, 4, 2]
	//
	// LCP(s, suffixArray(s)) = [0, 0, 1, 3, 0, 0, 2]
	// LCP iterates over the suffix array and looks at the current suffix,
	// compared to the previous suffix.
	// In the LCP array, the first value is 0 because there is no previous
	// suffix. Then the next value is 0 because suffixes "$" and "a$" do not share
	// a prefix. Then the value is 1 because "a$" and "ana$" share a 1-letter
	// prefix "a". The next value is 3 because "ana$" and "anana$" share a 3-letter
	// prefix "ana". And so on.

	var n int = len(suffixArr)

	// To store LCP array
	var lcp = make([]int, n)

	// An auxiliary array to store inverse of suffix array
	// elements. For example if suffixArr[0] is 5, the
	// invSuff[5] would store 0.  This is used to get next
	// suffix string from suffix array.
	var invSuff = make([]int, n)

	// Fill values in invSuff[]
	for i := 0; i < n; i++ {
		invSuff[suffixArr[i]] = i
	}

	// Initialize length of previous LCP
	var k int = 0

	// Process all suffixes one by one starting from
	// first suffix in txt[]
	for i := 0; i < n; i++ {
		// If the current suffix is at i=0, then we don’t
		// have prev substring to consider. So lcp is not
		// defined for this substring, we put zero.
		if invSuff[i] == 0 {
			k = 0
			continue
		}

		// j contains index of the prev substring to
		// be considered  to compare with the present
		// substring, i.e., prev string in suffix array
		var j int = suffixArr[invSuff[i]-1]

		// Directly start matching from k'th index as
		// at-least k-1 characters will match
		for i+k < n && j+k < n && txt[i+k] == txt[j+k] {
			k++
		}

		lcp[invSuff[i]] = k // lcp for the present suffix.

		// Deleting the starting character from the string.
		if k > 0 {
			k--
		}
	}

	// return the constructed lcp array
	return lcp
}

func isSubsetOfAnyExisting(new string, existing []string) bool {
	// This is made easier bc it's sorted by length
	// so a subsequent element can never be a superset of an existing.
	for _, strExisting := range existing {
		if strings.Contains(strExisting, new) {
			return true
		}
	}
	return false
}

func findCommonSubstrings(text1 string, text2 string, minLen int) []string {
	// Find all common non-overlapping substrings between two strings.
	// minLen is the minimum acceptable length of resulting common substrings.
	//
	// findCommonSubstrings("abcde", "bcbcd", 2)
	//	-> ["bcd"]
	//	Note: "bc" and "cd" are also common substrings, but they are substrings
	//		of "bcd" and therefore do not count.
	//
	// combined: "abcdebcbcd"
	// suffix array: 	[0 5 7 1 6 8 2 9 3 4]
	// suffixes:
	// 	- abcdebcbcd
	// 	- bcbcd
	// 	- bcd
	// 	- bcdebcbcd
	// 	- cbcd
	// 	- cd
	// 	- cdebcbcd
	// 	- d
	// 	- debcbcd
	// 	- ebcbcd
	// LCP array: 		[0 0 2 3 0 1 2 0 1 0]
	//
	// Iterating through LCP we check to see if the LCP value is greater than
	// minLen (meaning the overlap is long enough), and if the overlap occurs
	// in both texts.
	// We get some candidates:
	// 	 - bc
	// 	 - bcd
	// 	 - cd
	//
	// We sort the candidates by length and remove those that are substrings of
	// any previous candidate. Thus we are left with "bcd".

	// Get suffix array and Longest Common Prefix (LCP) array
	// for combined text
	var textCombined string = text1 + "||" + text2
	var sa []int = getSuffixArray(textCombined)
	var lcp []int = kasai(textCombined, sa)

	// Collect candidates
	var candidates []string

	for i := 1; i < len(sa); i++ {
		var isLongEnough bool = lcp[i] > minLen
		if isLongEnough {
			var j1 int = sa[i-1]
			var j2 int = sa[i]
			var h int = lcp[i]
			var jMin int = min(j1, j2)
			var jMax int = max(j1, j2)
			var isInBoth bool = jMin < len(text1) && jMax > len(text1)
			var doesNotCross bool = !(jMin < len(text1) && jMin+h > len(text1))
			if isInBoth && doesNotCross {
				var substring string = (textCombined)[j1 : j1+h]
				candidates = append(candidates, substring)
			}
		}
	}

	// Remove candidates that are a substring of other candidates.

	// Sort in place by length, descending
	sort.Slice(candidates, func(i, j int) bool {
		return len(candidates[i]) > len(candidates[j])
	})
	// Go through and take out substrings
	var nonOverlapping []string
	for i := 0; i < len(candidates); i++ {
		new := candidates[i]
		existing := candidates[:i]
		if !(isSubsetOfAnyExisting(new, existing)) {
			nonOverlapping = append(nonOverlapping, new)
		}
	}

	return nonOverlapping
}

// READ PDFS - functions for reading and getting data out of pdfs
//------------------------------------------------------------------------------

type PdfData struct {
	PathToFile      string
	Filename        string
	PageTexts       []string
	PageDigits      []string
	PageImageHashes []string
	FullText        string
	FullDigits      string
	FullImageHashes string
}

func getShortFilename(f string) string {
	var sep string = "/"
	split := strings.Split(f, sep)
	return split[len(split)-1]
}

func getPageTexts(fpath string) []string {
	var result []string

	// * USING LEDONGTHUC * (TODO: test if faster)
	// f, r, err := pdf.Open(fpath)
	// 	defer f.Close()
	// if err != nil {
	// 	panic(err)
	// }

	// var buf bytes.Buffer
	// var nPages int = r.NumPage()

	// // Pages are from 1 to nPages
	// for i := 1; i <= nPages; i++ {
	// 	page := r.Page(i)
	// 	pageText, err := page.GetPlainText()
	// }

	// USING MUPDF
	doc, err := fitz.New(fpath)
	if err != nil {
		panic(err)
	}

	defer doc.Close()

	for n := 0; n < doc.NumPage(); n++ {
		text, err := doc.Text(n)
		if err != nil {
			panic(err)
		}
		result = append(result, text)
	}

	return result
}

func getPageDigits(pageTexts []string) []string {
	var result []string

	reg, err := regexp.Compile("[^0-9]+") // non digits
	if err != nil {
		panic(err)
	}

	for _, t := range pageTexts {
		digits := reg.ReplaceAllString(t, "") // Replace non digits with space
		result = append(result, digits)
	}

	return result
}

func getPdfData(fullFilename string) PdfData {
	var filename string = getShortFilename(fullFilename)
	var pageTexts []string = getPageTexts(fullFilename)
	var pageDigits []string = getPageDigits(pageTexts)
	var pageImageHashes []string // = getPageImageHashes(pageTexts) TODO
	var fullText string = strings.Join(pageTexts, "|")
	var fullDigits string = strings.Join(pageDigits, "|")
	var fullImageHashes string // = TODO

	var result PdfData = PdfData{
		PathToFile:      fullFilename,
		Filename:        filename,
		PageTexts:       pageTexts,
		PageDigits:      pageDigits,
		PageImageHashes: pageImageHashes,
		FullText:        fullText,
		FullDigits:      fullDigits,
		FullImageHashes: fullImageHashes,
	}
	return result
}

func getStringPreview(s string) string {
	var result string = strings.Replace(s, "\n", " ", -1)
	if len(s) >= 97 {
		result = result[:97] + "..."
	}
	return result
}

// RESULTS - functions and structs for gathering up the results
//------------------------------------------------------------------------------
type ResultPage struct {
	Filename string `json:"filename"`
	Page     string `json:"page"`
}

type PdfResult struct {
	Kind          string       `json:"type"`
	StringPreview string       `json:"string_preview"`
	NumCharacters int          `json:"num_characters"`
	Pages         []ResultPage `json:"pages"`
}

func findPage(pages []string, susSubstr string) string {
	// Find the page that the bad substring was on.

	// Take only the part before the page separator, if the page separator is
	// in the string.
	if strings.Contains(susSubstr, "|") {
		susSubstr = strings.Split(susSubstr, "|")[0]
	}

	// Go through pages and check where the sus substr occurs
	for i, pageText := range pages {
		if strings.Contains(pageText, susSubstr) {
			pageNum := strconv.Itoa(i + 1)
			return pageNum
		}
	}

	return "Page not found"
}

func compareFiles(pdf1 PdfData, pdf2 PdfData, minLen int) []PdfResult {
	var results []PdfResult

	// Compare text
	commonSubstrings := findCommonSubstrings(pdf1.FullText, pdf2.FullText, minLen)

	for _, s := range commonSubstrings {
		var strPreview string = getStringPreview(s)

		resultPage1 := ResultPage{
			Filename: pdf1.Filename,
			Page:     findPage(pdf1.PageTexts, s),
		}
		resultPage2 := ResultPage{
			Filename: pdf2.Filename,
			Page:     findPage(pdf2.PageTexts, s),
		}

		var result PdfResult = PdfResult{
			Kind:          "Common text string",
			StringPreview: strPreview,
			NumCharacters: len(s),
			Pages:         []ResultPage{resultPage1, resultPage2},
		}

		results = append(results, result)
	}

	// Compare digits
	// minLen = 15
	commonDigits := findCommonSubstrings(pdf1.FullDigits, pdf2.FullDigits, 15)

	for _, s := range commonDigits {
		var strPreview string = getStringPreview(s)

		resultPage1 := ResultPage{
			Filename: pdf1.Filename,
			Page:     findPage(pdf1.PageDigits, s),
		}
		resultPage2 := ResultPage{
			Filename: pdf2.Filename,
			Page:     findPage(pdf2.PageDigits, s),
		}

		var result PdfResult = PdfResult{
			Kind:          "Common digit string",
			StringPreview: strPreview,
			NumCharacters: len(s),
			Pages:         []ResultPage{resultPage1, resultPage2},
		}

		results = append(results, result)
	}

	// Compare images
	// TODO
	return results
}

// MAIN
//------------------------------------------------------------------------------
func main() {
	// Command line flags
	filenamePtr := flag.String("f", "default", "Filenames to look at")
	wordPtr := flag.Int("minlen", 280, "Minimum length of text match")
	flag.Parse()
	var minLen int = *wordPtr
	var filenames []string = parseFilenameString(*filenamePtr)

	// Make sure we have files to actually read
	if len(filenames) < 2 {
		panic("Need at least 2 files to compare!")
	}

	// Read the PDFs
	var allPDFs []PdfData
	for _, filename := range filenames {
		var pdfData PdfData = getPdfData(filename)
		allPDFs = append(allPDFs, pdfData)
	}

	// Compare the files
	var results []PdfResult
	// Iterate over only top triangle of pairs matrix
	for i := 0; i < len(allPDFs); i++ {
		for j := i + 1; j < len(allPDFs); j++ {
			var pdf1 PdfData = allPDFs[i]
			var pdf2 PdfData = allPDFs[j]
			var newResults []PdfResult = compareFiles(pdf1, pdf2, minLen)
			results = append(results, newResults...)
		}
	}

	resultsJson, _ := json.Marshal(results)
	fmt.Println(string(resultsJson))
	return
}
