package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type ValueType int

const (
	Invalid ValueType = iota
	False
	Null
	True
	Object
	Array
	Number
	String
)

type JsonPair struct {
	Key   JsonValue
	Value JsonValue
}

type JsonValue struct {
	ValueType    ValueType
	RawValue     string
	ObjectMember []JsonPair
	ArrayMember  []JsonValue
}

var NotMatched = fmt.Errorf("not match")

type ParseFunc func(input string, pos int) (JsonValue, int, error)

func and(fs ...ParseFunc) ParseFunc {
	return func(input string, pos int) (JsonValue, int, error) {
		i := 0
		for _, f := range fs {
			_, ret, err := f(input, pos+i)
			if err != nil {
				return JsonValue{}, 0, err
			}
			i += ret
		}
		return JsonValue{}, i, nil
	}
}

func or(fs ...ParseFunc) ParseFunc {
	return func(input string, pos int) (JsonValue, int, error) {
		for _, f := range fs {
			v, ret, err := f(input, pos)
			if err != nil {
				continue
			}
			return v, ret, nil
		}
		return JsonValue{}, 0, NotMatched
	}
}

func many0(f ParseFunc) ParseFunc {
	return func(input string, pos int) (JsonValue, int, error) {
		i := 0
		for {
			_, ret, err := f(input, pos+i)
			if err != nil {
				return JsonValue{}, i, nil
			}
			i += ret
		}
	}
}

func many1(f ParseFunc) ParseFunc {
	return func(input string, pos int) (JsonValue, int, error) {
		_, i, err := f(input, pos)
		if err != nil {
			return JsonValue{}, 0, err
		}

		for {
			_, ret, err := f(input, pos+i)
			if err != nil {
				return JsonValue{}, i, nil
			}
			i += ret
		}
	}
}

func question(f ParseFunc) ParseFunc {
	return func(input string, pos int) (JsonValue, int, error) {
		_, i, err := f(input, pos)
		if err != nil {
			return JsonValue{}, 0, nil
		}
		return JsonValue{}, i, nil
	}
}

func characters(ch string) ParseFunc {
	return func(input string, pos int) (JsonValue, int, error) {
		if strings.HasPrefix(input[pos:], ch) {
			return JsonValue{}, len(ch), nil
		}
		return JsonValue{}, 0, NotMatched
	}
}

func characterRange(min rune, max rune) ParseFunc {
	return func(input string, pos int) (JsonValue, int, error) {
		r, size := utf8.DecodeRune([]byte(input[pos:]))
		if r == utf8.RuneError {
			return JsonValue{}, 0, NotMatched
		}
		if r < min || max < r {
			return JsonValue{}, 0, NotMatched
		}
		return JsonValue{}, size, nil
	}
}

var digit0 = characters("0")
var digit19 = characterRange('1', '9')
var digit09 = characterRange('0', '9')

var ws = many0(
	or(
		characters("\x20"),
		characters("\x09"),
		characters("\x0a"),
		characters("\x0d"),
	),
)

func parseLiteral(input string, pos int) (JsonValue, int, error) {

	var i int
	var err error

	_, i, err = characters("false")(input, pos)
	if err == nil {
		value := JsonValue{
			ValueType: False,
			RawValue:  input[pos : pos+i],
		}
		return value, i, nil
	}

	_, i, err = characters("null")(input, pos)
	if err == nil {
		value := JsonValue{
			ValueType: Null,
			RawValue:  input[pos : pos+i],
		}
		return value, i, nil
	}

	_, i, err = characters("true")(input, pos)
	if err == nil {
		value := JsonValue{
			ValueType: True,
			RawValue:  input[pos : pos+i],
		}
		return value, i, nil
	}

	return JsonValue{}, 0, NotMatched
}

func parseNumber(input string, pos int) (JsonValue, int, error) {

	_, i, err := and(
		question(
			characters("-"),
		),
		or(
			digit0,
			and(
				digit19,
				many0(digit09),
			),
		),
		question(
			and(
				characters("."),
				many1(digit09),
			),
		),
		question(
			and(
				characters("e"),
				or(
					characters("-"),
					characters("+"),
				),
				many1(digit09),
			),
		),
	)(input, pos)

	if err != nil {
		return JsonValue{}, 0, NotMatched
	}

	return JsonValue{
		ValueType: Number,
		RawValue:  input[pos : pos+i],
	}, i, nil
}

func parseString(input string, pos int) (JsonValue, int, error) {

	_, i, err := and(
		characters("\""),
		many0(
			or(
				or(
					characterRange(rune(0x20), rune(0x21)),
					characterRange(rune(0x23), rune(0x5B)),
					characterRange(rune(0x5d), rune(0x10FFFF)),
				),
				and(
					characters("\\"),
					or(
						characters("\""),
						characters("\\"),
						characters("/"),
						characters("b"),
						characters("f"),
						characters("n"),
						characters("r"),
						characters("t"),
						and(
							characters("u"),
							digit09,
							digit09,
							digit09,
							digit09,
						),
					),
				),
			),
		),
		characters("\""),
	)(input, pos)

	if err != nil {
		return JsonValue{}, 0, NotMatched
	}

	return JsonValue{
		ValueType: String,
		RawValue:  input[pos : pos+i],
	}, i, nil
}

func parseArray(input string, pos int) (JsonValue, int, error) {
	var value JsonValue
	var i, ret int
	var err error
	member := []JsonValue{}

	i = 0

	_, ret, err = and(ws, characters("["), ws)(input, pos+i)
	if err != nil {
		return JsonValue{}, 0, NotMatched
	}
	i += ret

	value, ret, err = parse(input, pos+i)
	if err != nil {
		_, ret, err = and(ws, characters("]"), ws)(input, pos+i)
		if err != nil {
			return JsonValue{}, 0, NotMatched
		}
		i += ret

		return JsonValue{
			ValueType:   Array,
			RawValue:    input[pos : pos+i],
			ArrayMember: member,
		}, i, nil
	}
	i += ret
	member = append(member, value)

	_, ret, err = and(ws, characters("]"), ws)(input, pos+i)
	if err == nil {
		i += ret

		return JsonValue{
			ValueType:   Array,
			RawValue:    input[pos : pos+i],
			ArrayMember: member,
		}, i, nil
	}

	for {
		_, ret, err = and(ws, characters(","), ws)(input, pos+i)
		if err != nil {
			break
		}
		i += ret

		value, ret, err = parse(input, pos+i)
		if err != nil {
			return JsonValue{}, 0, NotMatched
		}
		i += ret
		member = append(member, value)
	}

	_, ret, err = and(ws, characters("]"), ws)(input, pos+i)
	if err != nil {
		return JsonValue{}, 0, NotMatched
	}
	i += ret

	return JsonValue{
		ValueType:   Array,
		RawValue:    input[pos : pos+i],
		ArrayMember: member,
	}, i, nil
}

func parseObject(input string, pos int) (JsonValue, int, error) {
	var key, value JsonValue
	var i, ret int
	var err error
	member := []JsonPair{}

	i = 0

	_, ret, err = and(ws, characters("{"), ws)(input, pos+i)
	if err != nil {
		return JsonValue{}, 0, NotMatched
	}
	i += ret

	key, ret, err = parseString(input, pos+i)
	if err != nil {
		_, ret, err = and(ws, characters("}"), ws)(input, pos+i)
		if err != nil {
			return JsonValue{}, 0, NotMatched
		}
		i += ret

		return JsonValue{
			ValueType:    Object,
			RawValue:     input[pos : pos+i],
			ObjectMember: []JsonPair{},
		}, i, nil
	}
	i += ret

	_, ret, err = and(ws, characters(":"), ws)(input, pos+i)
	if err != nil {
		return JsonValue{}, 0, NotMatched
	}
	i += ret

	value, ret, err = parse(input, pos+i)
	if err != nil {
		return JsonValue{}, 0, NotMatched
	}
	i += ret
	member = append(member, JsonPair{Key: key, Value: value})

	for {
		_, ret, err = and(ws, characters(","), ws)(input, pos+i)
		if err != nil {
			break
		}
		i += ret

		key, ret, err = parseString(input, pos+i)
		if err != nil {
			return JsonValue{}, 0, NotMatched
		}
		i += ret

		_, ret, err = and(ws, characters(":"), ws)(input, pos+i)
		if err != nil {
			return JsonValue{}, 0, NotMatched
		}
		i += ret

		value, ret, err = parse(input, pos+i)
		if err != nil {
			return JsonValue{}, 0, NotMatched
		}
		i += ret
		member = append(member, JsonPair{Key: key, Value: value})
	}

	_, ret, err = and(ws, characters("}"), ws)(input, pos+i)
	if err != nil {
		return JsonValue{}, 0, NotMatched
	}
	i += ret

	return JsonValue{
		ValueType:    Object,
		RawValue:     input[pos : pos+i],
		ObjectMember: member,
	}, i, nil
}

func parse(input string, pos int) (JsonValue, int, error) {
	return or(
		parseLiteral,
		parseNumber,
		parseString,
		parseArray,
		parseObject,
	)(input, pos)
}
