package domain

import (
	"errors"
	"strings"
)

type Domain string

type Tree struct {
	mark uint8
	sub  domainMap
}

type domainMap map[Domain]*Tree

func (d Domain) nextLevel() Domain {
	if pointIndex := strings.LastIndex(string(d), "."); pointIndex == -1 {
		return ""
	} else {
		return d[:pointIndex]
	}
}

func (d Domain) topLevel() Domain {
	if pointIndex := strings.LastIndex(string(d), "."); pointIndex == -1 {
		return d
	} else {
		return d[pointIndex+1:]
	}
}

func DefaultDomainTree() *Tree {
	return NewDomainTree()
}

func NewDomainTree() (dt *Tree) {
	dt = new(Tree)
	dt.sub = make(domainMap)
	return
}

func (dt *Tree) Has(d Domain) bool {
	if len(dt.sub) == 0 {
		return true
	}

	if sub, ok := dt.sub[d.topLevel()]; ok {
		return sub.Has(d.nextLevel())
	}
	return false
}

func (dt *Tree) insert(sections []Domain) {

	if len(sections) == 0 {
		return
	}

	var lastIndex, lastSec = len(sections)-1, sections[len(sections)-1]

	if sec, ok := dt.sub[lastSec]; ok {
		sec.insert(sections[:lastIndex])
	} else {
		dt.sub[lastSec] = NewDomainTree()
		dt.sub[lastSec].insert(sections[:lastIndex])
	}
}

func (dt *Tree) Insert(d string) error {
	sections := strings.Split(d, ".")
	if len(sections) == 0 {
		return errors.New("Split Domain error\n")
	}

	domainSec := make([]Domain, len(sections))

	for i := range sections {
		domainSec[i] = Domain(sections[i])
	}

	dt.insert(domainSec)
	return nil
}
