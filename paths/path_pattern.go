package paths

import (
	"fmt"
	"github.com/stretchr/objx"
	stewstrings "github.com/stretchr/stew/strings"
	"strings"
)

/*
  PathPattern represents a path that can contain special matching segments.

  Valid paths include:

    /*** - Matches everything
  	/*** /literal/ *** - match anything before and after literal
    /literal
    /{placeholder}
    /[optional placeholder]
    /* - matches like a placeholder but doesn't care what it is
    /something/*** - Matches the start plus anything after it

*/
type PathPattern struct {

	// RawPath is the raw path.
	RawPath string

	path *Path
}

func NewPathPattern(path string) (*PathPattern, error) {

	p := new(PathPattern)
	p.RawPath = path
	p.path = NewPath(path)

	return p, nil
}

func (p *PathPattern) String() string {
	return stewstrings.MergeStrings("{PathPattern:\"", p.RawPath, "\"}")
}

/*
	GetPathMatch gets the PathMatch object describing the match or otherwise
	between this PathPattern and the specified Path.
*/
func (p *PathPattern) GetPathMatch(path *Path) *PathMatch {

	pathMatch := new(PathMatch)
	pathMatch.Matches = true

	// if this is the root catch all, just return yes
	if p.RawPath == segmentCatchAll {
		return pathMatch
	}

	pathMatch.Parameters = make(objx.Map)

	checkSegments := p.path.Segments()
	pathSegments := path.Segments()

	lastCheckSegment := checkSegments[len(checkSegments)-1]
	lastCheckSegmentType := getSegmentType(lastCheckSegment)

	fmt.Printf("checkseg: %v, pathSeg %v\n", checkSegments, pathSegments)

	// determine if this is a wildcard with a literal at the end
	if len(checkSegments) == 2 &&
		getSegmentType(checkSegments[0]) == segmentTypeCatchall &&
		lastCheckSegmentType == segmentTypeLiteral {
		if strings.ToLower(lastCheckSegment) == strings.ToLower(pathSegments[len(pathSegments)-1]) {
			return pathMatch
		}
	}

	// make sure the segments match in length, or there is a catchall
	// at the end of the check path
	if len(checkSegments) < len(pathSegments) {

		// check segments: /people/{id}
		// path segments:  /people

		// is the last segment a catchall?
		if lastCheckSegmentType != segmentTypeCatchall {
			return PathDoesntMatch
		}

	} else if len(checkSegments) > len(pathSegments) {

		// check segments: /poeple/{something}
		// path segments: /people/something/more

		// this situation is only OK if the last segment is optional, or
		// if it's the catch-all.
		if lastCheckSegmentType != segmentTypeDynamicOptional &&
			lastCheckSegmentType != segmentTypeCatchall {
			return PathDoesntMatch
		}

	}

	// check each segment
	for segmentIndex, checkSegment := range checkSegments {

		switch getSegmentType(checkSegment) {
		case segmentTypeLiteral:

			pathSegment := pathSegments[segmentIndex]
			if strings.ToLower(pathSegment) != strings.ToLower(checkSegment) {
				return PathDoesntMatch
			}

		case segmentTypeDynamic, segmentTypeDynamicOptional:

			if segmentIndex < len(pathSegments) {
				pathMatch.Parameters[cleanSegmentName(checkSegment)] = pathSegments[segmentIndex]
			}
		}

	}

	return pathMatch

}
