// Sasha Trubetskoy
// sasha@kartographia.com

package main

import (
	"fmt"
	"flag"
	"sort"
	"bytes"
	"reflect"
	"strings"
	"strconv"
	"index/suffixarray"
	"github.com/ledongthuc/pdf"
    "encoding/csv"
)


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


func readPdf(path string) (string, error) {
 f, r, err := pdf.Open(path)
 // remember close file
	defer f.Close()
 if err != nil {
	 return "", err
 }
 var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)
 return buf.String(), nil
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


type Substr struct {
	Text 	string
	Indexes 	[]int
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
	
	// Get suffix array and Longest Common Prefix (LCP) array
	var sa []int = getSuffixArray(text)
	var lcp []int = kasai(text, sa)

	// Collecting the substrings here
	var crossDocSubstrs []Substr

	for i := 0; i < len(text); i++ {
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


func parseFilenameString(s string) []string {
	// https://stackoverflow.com/questions/47489745/splitting-a-string-at-space-except-inside-quotation-marks
    r := csv.NewReader(strings.NewReader(s))
    r.Comma = ',' // separator
    fields, err := r.Read()
    if err != nil {
        panic(err)
    }
    return fields
}


func main() {
	// Command line flags
	filenamePtr := flag.String("f", "default", "Filenames to look at")
	wordPtr 	:= flag.Int("minlen", 280, "Minimum length of text match")
	flag.Parse()

	// Params
	var minLen int = *wordPtr
	var filenames []string = parseFilenameString(*filenamePtr)

	// Make sure we have files to actually read
	if len(filenames) > 2 {
		panic("More than 2 files not implemented!")
	} else if len(filenames) < 2 {
		panic("Need at least 2 files to compare!")
	}

	// Read the PDF text
	content1, err := readPdf(filenames[0]) // Read local pdf file
	if err != nil {
		panic(err)
	}
	content2, err := readPdf(filenames[1]) // Read local pdf file
	if err != nil {
		panic(err)
	}	

	var lenA int = len(content1)
	var combinedText string = content1 + "||" + content2
	var lcs = longestCommonSubstring(lenA, combinedText, minLen)

	fmt.Println(lcs)
	// myStr := content[23:len(content)]
	// fmt.Println(myStr)
	return
}