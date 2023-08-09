package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCharacters(t *testing.T) {

	testcases := []struct {
		ch    string
		pos   int
		match bool
		i     int
		err   error
	}{
		{
			ch:    "0",
			pos:   0,
			match: true,
			i:     1,
		},
		{
			ch:    "0",
			pos:   1,
			match: false,
			i:     0,
		},
		{
			ch:    "1",
			pos:   1,
			match: true,
			i:     1,
		},
		{
			ch:    "1",
			pos:   0,
			match: false,
			i:     0,
		},
	}

	for _, tt := range testcases {
		t.Run("", func(t *testing.T) {
			assertMatch(t, tt.match, characters(tt.ch), "0123456789", tt.pos, tt.i)
		})
	}
}

func TestCharacterRange(t *testing.T) {

	testcases := []struct {
		pos   int
		match bool
		i     int
	}{
		{
			pos:   0,
			match: false,
			i:     0,
		},
		{
			pos:   1,
			match: true,
			i:     1,
		},
		{
			pos:   8,
			match: true,
			i:     1,
		},
		{
			pos:   9,
			match: false,
			i:     0,
		},
	}

	for _, tt := range testcases {
		t.Run("", func(t *testing.T) {
			assertMatch(t, tt.match, characterRange('1', '8'), "0123456789", tt.pos, tt.i)
		})
	}
}

func TestDigit(t *testing.T) {

	testcases := "/0123456789:"
	digit0Expected := []bool{false, true, false, false, false, false, false, false, false, false, false, false}
	digit09Expected := []bool{false, true, true, true, true, true, true, true, true, true, true, false}
	digit19Expected := []bool{false, false, true, true, true, true, true, true, true, true, true, false}

	for i := 0; i < len(testcases); i++ {
		t.Run(testcases[i:i+1], func(t *testing.T) {
			assertMatch(t, digit0Expected[i], digit0, testcases[i:i+1], 0, 1)
		})

		t.Run(testcases[i:i+1], func(t *testing.T) {
			assertMatch(t, digit09Expected[i], digit09, testcases[i:i+1], 0, 1)
		})

		t.Run(testcases[i:i+1], func(t *testing.T) {
			assertMatch(t, digit19Expected[i], digit19, testcases[i:i+1], 0, 1)
		})
	}
}

func TestMay(t *testing.T) {
	assertMatch(t, true, question(digit0), "0", 0, 1)
	assertMatch(t, true, question(digit0), "1", 0, 0)
}

func TestMany0(t *testing.T) {
	assertMatch(t, true, many0(digit0), "1", 0, 0)
	assertMatch(t, true, many0(digit0), "0", 0, 1)
	assertMatch(t, true, many0(digit0), "00", 0, 2)
}

func TestMany1(t *testing.T) {
	assertMatch(t, false, many1(digit0), "1", 0, 0)
	assertMatch(t, true, many1(digit0), "0", 0, 1)
	assertMatch(t, true, many1(digit0), "00", 0, 2)
}

func TestAnd(t *testing.T) {
	assertMatch(t, true, and(digit19, digit09), "12", 0, 2)
}

func TestOr(t *testing.T) {
	assertMatch(t, true, or(characters("a"), characters("b")), "b", 0, 1)
	assertMatch(t, false, or(characters("a"), characters("b")), "c", 0, 0)
}

func assertMatch(t *testing.T, match bool, parseFunc ParseFunc, input string, pos int, expected int) {
	t.Helper()
	v, i, err := parseFunc(input, pos)

	assert.Equal(t, JsonValue{}, v)
	if match {
		assert.Equal(t, expected, i)
		assert.Nil(t, err)
	} else {
		assert.Equal(t, 0, i)
		assert.Equal(t, err, NotMatched)
	}
}

func TestParseLiteral(t *testing.T) {
	var v JsonValue
	var i int
	var err error

	v, i, err = parseLiteral("true", 0)
	assert.Equal(t, JsonValue{ValueType: True, RawValue: "true"}, v)
	assert.Equal(t, 4, i)
	assert.Nil(t, err)

	v, i, err = parseLiteral("false", 0)
	assert.Equal(t, JsonValue{ValueType: False, RawValue: "false"}, v)
	assert.Equal(t, 5, i)
	assert.Nil(t, err)

	v, i, err = parseLiteral("null", 0)
	assert.Equal(t, JsonValue{ValueType: Null, RawValue: "null"}, v)
	assert.Equal(t, 4, i)
	assert.Nil(t, err)

	v, i, err = parseLiteral("hoge", 0)
	assert.Equal(t, JsonValue{}, v)
	assert.Equal(t, 0, i)
	assert.Equal(t, NotMatched, err)
}

func TestParseNumber(t *testing.T) {

	testcases := []struct {
		input string
		i     int
	}{
		{"123", 3},
		{"-0.123", 6},
		{"1e-1", 4},
		{"1e+1", 4},
	}

	for _, tt := range testcases {
		v, i, err := parseNumber(tt.input, 0)
		assert.Equal(t, JsonValue{ValueType: Number, RawValue: tt.input}, v)
		assert.Equal(t, tt.i, i)
		assert.Nil(t, err)
	}
}

func TestParseString(t *testing.T) {

	testcases := []struct {
		input string
		i     int
	}{
		{`"0123456789"`, 12},
		{`"\\\"\/\b\f\n\r\t"`, 18},
		{`"\u0123"`, 8},
	}

	for _, tt := range testcases {
		t.Run("", func(t *testing.T) {
			v, i, err := parseString(tt.input, 0)
			assert.Equal(t, JsonValue{ValueType: String, RawValue: tt.input}, v)
			assert.Equal(t, tt.i, i)
			assert.Nil(t, err)
		})
	}
}

func TestParseArray(t *testing.T) {

	testcases := []struct {
		input string
		i     int
		value []JsonValue
	}{
		{`[]`, 2, []JsonValue{}},
		{`[1]`, 3, []JsonValue{
			{ValueType: Number, RawValue: "1"},
		}},
		{`[1,"2"]`, 7, []JsonValue{
			{ValueType: Number, RawValue: "1"},
			{ValueType: String, RawValue: `"2"`},
		}},
	}

	for _, tt := range testcases {
		t.Run("", func(t *testing.T) {
			v, i, err := parseArray(tt.input, 0)
			assert.Equal(t, JsonValue{ValueType: Array, RawValue: tt.input, ArrayMember: tt.value}, v)
			assert.Equal(t, tt.i, i)
			assert.Nil(t, err)
		})
	}
}

func TestParseObject(t *testing.T) {

	testcases := []struct {
		input string
		i     int
		value []JsonPair
	}{
		{`{}`, 2, []JsonPair{}},
		{`{"a":"b"}`, 9,
			[]JsonPair{
				{
					Key:   JsonValue{ValueType: String, RawValue: `"a"`},
					Value: JsonValue{ValueType: String, RawValue: `"b"`},
				},
			},
		},
		{`{"a":"b","c":"d"}`, 17,
			[]JsonPair{
				{
					Key:   JsonValue{ValueType: String, RawValue: `"a"`},
					Value: JsonValue{ValueType: String, RawValue: `"b"`},
				},
				{
					Key:   JsonValue{ValueType: String, RawValue: `"c"`},
					Value: JsonValue{ValueType: String, RawValue: `"d"`},
				},
			},
		},
	}

	for _, tt := range testcases {
		t.Run("", func(t *testing.T) {
			v, i, err := parseObject(tt.input, 0)
			assert.Equal(t, JsonValue{ValueType: Object, RawValue: tt.input, ObjectMember: tt.value}, v)
			assert.Equal(t, tt.i, i)
			assert.Nil(t, err)
		})
	}
}
