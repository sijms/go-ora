package tns

import "unicode"

const stopRune rune = 0

// lookAhead returns next rune, without consuming it
func (p *parser) lookAhead() rune {
	if p.isEndReached() {
		return stopRune
	}
	return p.tns[p.pos]
}

// isEndReached returns true if current parser position is out of bounds of the provided TNS slice.
func (p *parser) isEndReached() bool {
	return p.pos >= len(p.tns)
}

// consumeRune consumes one rune
func (p *parser) consumeRune() rune {
	if p.isEndReached() {
		return stopRune
	}
	r := p.tns[p.pos]
	p.pos += 1
	return r
}

// consumeWhitespace consumes all whitespace (if any) and stops before the next non-space rune
//
// reports number of runes consumed
func (p *parser) consumeWhitespace() int {
	var cnt int
	for {
		r := p.lookAhead()
		if r == stopRune {
			break
		}
		if !unicode.IsSpace(r) {
			break
		}
		_ = p.consumeRune()
		cnt++
	}
	return cnt
}
