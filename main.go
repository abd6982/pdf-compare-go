// Sasha Trubetskoy
// sasha@kartographia.com

package main

import (
	"fmt"
	"flag"
	"sort"
	"reflect"
	"regexp"
	"strings"
	"strconv"
	"encoding/csv"
	"encoding/json"
	"index/suffixarray"
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


// LCS - functions and structs for lcs and text comparison
//------------------------------------------------------------------------------
func getSuffixArray(text string) ([]int) {
	// Get suffix array of string
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


/* To construct and return LCP */
func kasai(txt string, suffixArr []int) ([]int) {
	var n int = len(suffixArr)
 
	// To store LCP array
	var lcp = make([]int, n)
 
	// An auxiliary array to store inverse of suffix array
	// elements. For example if suffixArr[0] is 5, the
	// invSuff[5] would store 0.  This is used to get next
	// suffix string from suffix array.
	var invSuff = make([]int, n)
 
	// Fill values in invSuff[]
	for i:=0; i < n; i++ {
		invSuff[suffixArr[i]] = i
	}
 
	// Initialize length of previous LCP
	var k int = 0
 
	// Process all suffixes one by one starting from
	// first suffix in txt[]
	for i:=0; i<n; i++ {
		/* If the current suffix is at n-1, then we don’t
		   have next substring to consider. So lcp is not
		   defined for this substring, we put zero. */
		if invSuff[i] == n-1 {
			k = 0
			continue;
		}
 
		/* j contains index of the next substring to
		   be considered  to compare with the present
		   substring, i.e., next string in suffix array */
		var j int = suffixArr[invSuff[i]+1];
 
		// Directly start matching from k'th index as
		// at-least k-1 characters will match
		for i+k<n && j+k<n && txt[i+k]==txt[j+k] {
			k++
		}
 
		lcp[invSuff[i]] = k // lcp for the present suffix.
 
		// Deleting the starting character from the string.
		if k>0 {
			k--
		}
	}
 
	// return the constructed lcp array
	return lcp
}


func findCommonSubstrings(text1 string, text2 string, minLen int) []string {
	// minLen is the minimum acceptable length of resulting common substrings.
	//
	// findCommonSubstrings("banana", "anagram", 2)
	//	-> ["ana", "na"]
	
	// Get suffix array and Longest Common Prefix (LCP) array
	// for combined text
	var sa []int = getSuffixArray(text1 + text2)
	var lcp []int = kasai(text1 + text2, sa)

	// Collecting the substrings here
	var crossDocSubstrs []Substr

	for i := 1; i < len(text); i++ {
		if lcp[i] > minLen {
			var j1 int = sa[i - 1]
			var j2 int = sa[i]
			var h int = lcp[i]
			var jMin int = min(j1, j2)
			var jMax int = max(j1, j2)
			var isCrossDoc bool = jMin < lenA && jMax > lenA
			if isCrossDoc {
				var substring string = text[j1:j1 + h]
				newSubstrs := Substr{
					Text: substring,
					Indexes: []int{jMin, jMax},
				}
				crossDocSubstrs = append(crossDocSubstrs, newSubstrs)
			}
		}	
	}



}
func longestCommonSubstring(lenA int, text string, minLen int) ([]Substr) {
	// Inputs
	// 	- lenA: length of the first text
	// 	- text: combined text (first and second concatenated)
	// 	- minLen: minimum number of characters in substring
	//
	// Get the longest common substrings and their positions.
	// >>> longest_common_substring('banana')
	// {'ana': [1, 3]}
	// >>> text = "not so Agamemnon, who spoke fiercely to "
	// >>> sorted(longest_common_substring(text).items())
	// [(' s', [3, 21]), ('no', [0, 13]), ('o ', [5, 20, 38])]
	// This function can be easy modified for any criteria, e.g. for searching ten
	// longest non overlapping repeated substrings.

	// Get non overlapping
	isSubsetOfAnyExisting := func(new Substr, existing []Substr) (bool) {
		// This is made easier bc it's sorted by length
		// so a subsequent element can never be a superset of an existing.
		var n1_start int = new.Indexes[0]
		var n1_end int = n1_start + len(new.Text)

		for _, substrExisting := range existing {
			var e1_start int = substrExisting.Indexes[0] // "e" for "existing"
			var e1_end int = e1_start + len(substrExisting.Text)
			// Look only at the zeroth occurrence of both strings
			var newIsSubset bool = e1_start <= n1_start && n1_end <= e1_end
			if newIsSubset {
				return true
			}
		}
		return false
	}

	// Sort by length, descending
	sort.Slice(crossDocSubstrs, func(i, j int) bool {
		return len(crossDocSubstrs[i].Text) > len(crossDocSubstrs[j].Text)
	})
	var nonOverlapping []Substr
	for _, new := range crossDocSubstrs {
		if !(isSubsetOfAnyExisting(new, nonOverlapping)) {
			nonOverlapping = append(nonOverlapping, new)
		}
	}

	return nonOverlapping
}


// READ PDFS - functions for reading and getting data out of pdfs
//------------------------------------------------------------------------------


type PdfData struct {
	PathToFile 		string
	Filename 		string
	PageTexts 		[]string
	PageDigits 		[]string
	PageImageHashes []string
	FullText 		string
	FullDigits		string
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
		PathToFile: fullFilename,
		Filename: filename,
		PageTexts: pageTexts,
		PageDigits: pageDigits,
		PageImageHashes: pageImageHashes,
		FullText: fullText,
		FullDigits: fullDigits,
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
	Filename string 	`json:"filename"`
	Page 	 string 	`json:"page"`
}


type PdfResult struct {
	Kind 			string 			`json:"type"`
	StringPreview 	string 			`json:"string_preview"`
	NumCharacters 	int             `json:"num_characters"`
	Pages			[]ResultPage  	`json:"pages"`
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
			return string(i + 1)
		}
	}

	return "Page not found"
}	



func compareFiles(pdf1 PdfData, pdf2 PdfData, minLen int) []PdfResult {
	var results []PdfResult

	// Compare text
	commonSubstrings := findCommonSubstrings(pdf1.FullText, pdf2.FullText, minLen)
	var lcs []Substr = longestCommonSubstring(len(text1), combinedText, minLen)

	for _, lcsResult := range lcs {
		var strPreview string = getStringPreview(lcsResult.Text)
		
		resultPage1 := ResultPage{
			Filename: pdf1.Filename,
			Page: findPage(pdf1.PageTexts, lcsResult.Text),
		}
		resultPage2 := ResultPage{
			Filename: pdf2.Filename,
			Page: findPage(pdf2.PageTexts, lcsResult.Text),
		}

		var result PdfResult = PdfResult{
			Kind: "Common text string",
			StringPreview: strPreview,
			NumCharacters: len(lcsResult.Text),
			Pages: []ResultPage{resultPage1, resultPage2},
		}

		results = append(results, result)
	}

	// Compare digits
	var digits1 string = pdf1.FullDigits
	var digits2 string = pdf2.FullDigits
	var combineddigits string = digits1 + "||" + digits2
	lcs = longestCommonSubstring(len(digits1), combineddigits, 15)

	for _, lcsResult := range lcs {
		var strPreview string = getStringPreview(lcsResult.Text)
		
		resultPage1 := ResultPage{
			Filename: pdf1.Filename,
			Page: findPage(pdf1.PageDigits, lcsResult.Text),
		}
		resultPage2 := ResultPage{
			Filename: pdf2.Filename,
			Page: findPage(pdf2.PageDigits, lcsResult.Text),
		}

		var result PdfResult = PdfResult{
			Kind: "Common digit string",
			StringPreview: strPreview,
			NumCharacters: len(lcsResult.Text),
			Pages: []ResultPage{resultPage1, resultPage2},
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
	wordPtr 	:= flag.Int("minlen", 280, "Minimum length of text match")
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
		for j := i+1; j < len(allPDFs); j++ {
			var pdf1 PdfData = allPDFs[i]
			var pdf2 PdfData = allPDFs[j]
			var newResults []PdfResult = compareFiles(pdf1, pdf2, minLen)
			results = append(results, newResults...)
		}
	}

	resultsJson, _ := json.Marshal(results)
	fmt.Println(string(resultsJson))

	lcsTest := longestCommonSubstring(6, "banana banana", 0)
	fmt.Println(lcsTest)
	return
}