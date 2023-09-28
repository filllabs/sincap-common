package qapi

import (
	"errors"
	"strings"
)

// Direction defines the type of the Sort
type Direction int

const (
	// ASC 1,2,3,4,5,...
	ASC Direction = iota + 1
	// DSC ...,5,4,3,2,1
	DSC
)

func (dr Direction) String() string {
	names := [...]string{
		"Unknown",
		"asc",
		"desc"}
	return names[dr]
}

// Sort holds the necessary info for a sort param.
type Sort struct {
	Direction Direction
	Name      string
}

// Parse parses and fills the sort
func (sort *Sort) Parse(param string) error {
	param = strings.TrimRight(param, " ")
	if len(param) < 2 {
		return errors.New("Sort param can't be shorter than 2")
	}
	switch param[0] {
	case '-':
		sort.Direction = DSC
	case '+', ' ':
		sort.Direction = ASC
	default:
		return errors.New("Sort param can only start with - or +")
	}
	sort.Name = param[1:]
	return nil
}

// String returns concatted
func (sort *Sort) String() string {
	return sort.Name + " " + sort.Direction.String()
}
