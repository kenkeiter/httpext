package httpext

import (
	"errors"
	"fmt"
	"strconv"
	// "strings"
)

// TODO(kk): When there are 0 records total, response should be Range: */0 and
//           server should return HTTP 416 Request Not Satisfiable.

var (
	// ErrRangeIsSuffix indicates that a range is only a suffix, which is a type
	// of range that indicates a number of records that should be read from
	// the end of a collection.
	ErrRangeIsSuffix = errors.New("first index in range is negative, " +
		"indicating a suffix -- no last index may be supplied")

	// ErrRangeInvalid indicates that basic logic around ranges has not been
	// fulfilled.
	ErrRangeInvalid = errors.New("first index in range must be <= last index")

	// ErrRangeUnsatisfiableZeroLength indicates that the specified range
	// would include zero elements.
	ErrRangeUnsatisfiableZeroLength = errors.New("range can satisfy a " +
		"zero-length set")

	// ErrRangeOutsideConstraints indicates that a specified range includes only
	// elements outside of the range it has been constrained to.
	ErrRangeOutsideConstraints = errors.New("range begins outside of the " +
		"total number of elements")
)

const (
	// RangeUnconstrained is returned whenever a range has not been constrained
	// in a way that the requested value can be calculated.
	RangeUnconstrained = -1
)

func NewContentRange(units string, first, last int) (*ContentRange, error) {
	c := &ContentRange{units: units}
	if err := c.SetFirst(first); err != nil {
		return nil, err
	}
	if err := c.SetLast(last); err != nil {
		return nil, err
	}
	return c, nil
}

// ContentRange represents information provided by a Range header, as specified
// in IETF RFC 7233 (http://tools.ietf.org/html/rfc7233).
type ContentRange struct {
	units string

	first  int
	last   int
	fBound bool
	lBound bool

	total  int
	tBound bool
}

func (c *ContentRange) SetFirst(first int) error {
	if c.lBound && c.first >= c.last {
		return ErrRangeInvalid
	}
	c.first = first
	c.fBound = true
	return nil
}

func (c *ContentRange) SetLast(last int) error {
	if c.fBound && last < c.first {
		return ErrRangeInvalid
	}
	c.last = last
	c.lBound = true
	return nil
}

func (c *ContentRange) First() int {
	if !c.fBound {
		return RangeUnconstrained
	}
	return c.first
}

func (c *ContentRange) Last() int {
	if !c.lBound {
		return RangeUnconstrained
	}
	return c.last
}

func (c *ContentRange) IsSuffix() bool {
	return c.fBound == false
}

func (c *ContentRange) IsFixed() bool {
	return c.fBound && c.lBound
}

func (c *ContentRange) IsUnbounded() bool {
	return !c.fBound || !c.lBound
}

func (c *ContentRange) IsFullRange() bool {
	return (c.fBound && c.first == 0) && !c.lBound
}

func (c *ContentRange) Contains(offset int) bool {
	if offset < 0 {
		if !c.fBound {
			return false
		}
		return c.last >= -offset
	}
	if !c.fBound || !c.lBound {
		return false
	}
	return (c.first <= offset) && (offset <= c.last)
}

func (c *ContentRange) Constrain(size int) error {
	if size == 0 {
		if !c.fBound {
			c.last = 0
			c.lBound = true
			return nil
		}
		return ErrRangeUnsatisfiableZeroLength
	}

	if !c.fBound {
		c.first = size + c.last
		c.fBound = true
		if c.first < 0 {
			c.first = 0
		}
		c.last = size - 1
		c.lBound = true
		return nil
	}

	if c.first > (size - 1) {
		return ErrRangeOutsideConstraints
	}

	if !c.lBound {
		c.last = size - 1
		c.lBound = true
	}

	return nil
}

func (c *ContentRange) Offset() int {
	return c.first
}

func (c *ContentRange) Limit() int {
	if c.IsFixed() {
		return c.last - c.first
	}
	if c.lBound && c.last < 0 {
		return -c.last
	}
	return RangeUnconstrained
}

func (c *ContentRange) SetTotal(total int) error {
	if err := c.Constrain(total); err != nil {
		return err
	}
	c.tBound = true
	c.total = total
	return nil
}

func (c *ContentRange) Units() string {
	return c.units
}

// Format returns a representation of the ContentRange as the body of an HTTP
// Content-Range header.
func (c *ContentRange) Format() (string, error) {
	// Determine how to render the range.
	max := "*"
	if c.tBound {
		max = strconv.FormatInt(int64(c.total), 10)
	}

	// If both upper/lower bounds are missing, render "*/total" pg 12 of RFC 7233.
	if !c.fBound && !c.lBound {
		return fmt.Sprintf("%s */%s", c.units, max), nil
	}

	if (!c.fBound && c.lBound) || (c.fBound && !c.lBound) {
		return "", fmt.Errorf("One or more unbound: %b %b", c.fBound, c.lBound)
	}

	return fmt.Sprintf("%s %d-%d/%s", c.units, c.first, c.last, max), nil
}

// ParseRange parses an HTTP Range header into a *ContentRange. ParseRange only
// supports single ranges, not multiple. It does not support parameters.
//
//   resources=-99   // <- last 100 resources from end of set (suffix range)
//   resources=0-99  // <- 100 resources, from indices [0-99]
//   resources=99-   // <- resources from indices [99-n], where n = len(collection)
//
func ParseRange(r string) (*ContentRange, error) {
	var rng = &ContentRange{}
	var units, s string
	var first, last int
	var err error
	var ok bool

	units, s = expectUnitSpecifier(r)
	rng.units = units

	first, s, err = expectRangeValue(s)
	if err != nil {
		return nil, err
	}
	if first < 0 {
		err = rng.SetLast(int(first))
		return rng, err
	}
	err = rng.SetFirst(int(first))
	if err != nil {
		return nil, err
	}

	if len(s) == 0 {
		return rng, nil
	}

	s, ok = expectSeparator(s, '-')
	if ok && len(s) > 0 {
		last, s, err = expectRangeValue(s)
		if err != nil {
			return nil, err
		}
		err = rng.SetLast(last)
		if err != nil {
			return nil, err
		}
	}

	if len(s) > 0 {
		return nil, ErrRangeInvalid
	}

	return rng, nil
}

func expectUnitSpecifier(s string) (units, rest string) {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '=':
			return s[:i], s[i+1:]
		}
	}
	return "", ""
}

func expectRangeValue(s string) (value int, rest string, err error) {
	// read chars until we encounter a separator or EOL (other than in 1st position)
	for i := 0; i < len(s); i++ {
		isPastFirstPos := i > 0
		isDigit := (s[i] >= '0' && s[i] <= '9')
		isLastChar := i+1 == len(s)

		switch {
		case isPastFirstPos && !isDigit:
			v, err := strconv.ParseInt(s[:i], 10, 64)
			return int(v), s[i:], err
		case isLastChar:
			v, err := strconv.ParseInt(s[:i+1], 10, 64)
			return int(v), s[i+1:], err
		}
	}

	return 0, "", ErrRangeInvalid
}

func expectSeparator(s string, sep uint8) (rest string, found bool) {
	if s[0] == sep {
		return s[1:], true
	}
	return s, false
}
