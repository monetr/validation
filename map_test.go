package validation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	var m0 map[string]interface{}
	m1 := map[string]interface{}{"A": "abc", "B": "xyz", "c": "abc", "D": (*string)(nil), "F": (*String123)(nil), "H": []string{"abc", "abc"}, "I": map[string]string{"foo": "abc"}}
	m2 := map[string]interface{}{"E": String123("xyz"), "F": (*String123)(nil)}
	m3 := map[string]interface{}{"M3": Model3{}}
	m4 := map[string]interface{}{"M3": Model3{A: "abc"}}
	m5 := map[string]interface{}{"A": "internal", "B": ""}
	m6 := map[int]string{11: "abc", 22: "xyz"}
	// Note: Map/Key are now generic over the key type (Map[K]/Key[K]). The rule
	// is built per-case so that string-keyed and int-keyed maps can coexist in
	// one table. Map[string]() is used where there are no key rules.
	tests := []struct {
		tag   string
		model any
		rule  Rule
		err   string
	}{
		// empty rules
		{"t1.1", m1, Map[string]().AllowExtraKeys(), ""},
		{"t1.2", m1, Map(Key("A"), Key("B")).AllowExtraKeys(), ""},
		// normal rules
		{"t2.1", m1, Map(Key("A", &validateAbc{}), Key("B", &validateXyz{})).AllowExtraKeys(), ""},
		{"t2.2", m1, Map(Key("A", &validateXyz{}), Key("B", &validateAbc{})).AllowExtraKeys(), "A: error xyz; B: error abc."},
		{"t2.3", m1, Map(Key("A", &validateXyz{}), Key("c", &validateXyz{})).AllowExtraKeys(), "A: error xyz; c: error xyz."},
		{"t2.4", m1, Map(Key("D", Length(0, 5))).AllowExtraKeys(), ""},
		{"t2.5", m1, Map(Key("F", Length(0, 5))).AllowExtraKeys(), ""},
		{"t2.6", m1, Map(Key("H", Each(&validateAbc{})), Key("I", Each(&validateAbc{}))).AllowExtraKeys(), ""},
		{"t2.7", m1, Map(Key("H", Each(&validateXyz{})), Key("I", Each(&validateXyz{}))).AllowExtraKeys(), "H: (0: error xyz; 1: error xyz.); I: (foo: error xyz.)."},
		{"t2.8", m1, Map(Key("I", Map(Key("foo", &validateAbc{})))).AllowExtraKeys(), ""},
		{"t2.9", m1, Map(Key("I", Map(Key("foo", &validateXyz{})))).AllowExtraKeys(), "I: (foo: error xyz.)."},
		// non-map value
		{"t3.1", &m1, Map[string]().AllowExtraKeys(), ""},
		{"t3.2", nil, Map[string]().AllowExtraKeys(), ErrNotMap.Error()},
		{"t3.3", m0, Map[string]().AllowExtraKeys(), ""},
		{"t3.4", &m0, Map[string]().AllowExtraKeys(), ""},
		{"t3.5", 123, Map[string]().AllowExtraKeys(), ErrNotMap.Error()},
		// invalid key spec
		{"t4.1", m1, Map(Key(123)).AllowExtraKeys(), "123: key not the correct type."},
		{"t4.2", m1, Map(Key("X")).AllowExtraKeys(), "X: required key is missing."},
		{"t4.3", m1, Map(Key("X").Optional()).AllowExtraKeys(), ""},
		// non-string keys
		{"t5.1", m6, Map(Key(11, &validateAbc{}), Key(22, &validateXyz{})).AllowExtraKeys(), ""},
		{"t5.2", m6, Map(Key(11, &validateXyz{}), Key(22, &validateAbc{})).AllowExtraKeys(), "11: error xyz; 22: error abc."},
		// validatable value
		{"t6.1", m2, Map(Key("E")).AllowExtraKeys(), "E: error 123."},
		{"t6.2", m2, Map(Key("E", Skip)).AllowExtraKeys(), ""},
		{"t6.3", m2, Map(Key("E", Skip.When(true))).AllowExtraKeys(), ""},
		{"t6.4", m2, Map(Key("E", Skip.When(false))).AllowExtraKeys(), "E: error 123."},
		// Required, NotNil
		{"t7.1", m2, Map(Key("F", Required)).AllowExtraKeys(), "F: cannot be blank."},
		{"t7.2", m2, Map(Key("F", NotNil)).AllowExtraKeys(), "F: is required."},
		{"t7.3", m2, Map(Key("F", Skip, Required)).AllowExtraKeys(), ""},
		{"t7.4", m2, Map(Key("F", Skip, NotNil)).AllowExtraKeys(), ""},
		{"t7.5", m2, Map(Key("F", Skip.When(true), Required)).AllowExtraKeys(), ""},
		{"t7.6", m2, Map(Key("F", Skip.When(true), NotNil)).AllowExtraKeys(), ""},
		{"t7.7", m2, Map(Key("F", Skip.When(false), Required)).AllowExtraKeys(), "F: cannot be blank."},
		{"t7.8", m2, Map(Key("F", Skip.When(false), NotNil)).AllowExtraKeys(), "F: is required."},
		// validatable structs
		{"t8.1", m3, Map(Key("M3", Skip)).AllowExtraKeys(), ""},
		{"t8.2", m3, Map(Key("M3")).AllowExtraKeys(), "M3: (A: error abc.)."},
		{"t8.3", m4, Map(Key("M3")).AllowExtraKeys(), ""},
		// internal error
		{"t9.1", m5, Map(Key("A", &validateAbc{}), Key("B", Required), Key("A", &validateInternalError{})).AllowExtraKeys(), "error internal"},
	}
	for _, test := range tests {
		err1 := Validate(test.model, test.rule)
		err2 := ValidateWithContext(context.Background(), test.model, test.rule)
		assertError(t, test.err, err1, test.tag)
		assertError(t, test.err, err2, test.tag)
	}

	a := map[string]interface{}{"Name": "name", "Value": "demo", "Extra": true}
	err := Validate(a, Map(
		Key("Name", Required),
		Key("Value", Required, Length(5, 10)),
	))
	assert.EqualError(t, err, "Extra: key not expected; Value: the length must be between 5 and 10.")
}

func TestMapWithContext(t *testing.T) {
	m1 := map[string]interface{}{"A": "abc", "B": "xyz", "c": "abc", "g": "xyz"}
	m2 := map[string]interface{}{"A": "internal", "B": ""}
	tests := []struct {
		tag   string
		model any
		rule  Rule
		err   string
	}{
		// normal rules
		{"t1.1", m1, Map(Key("A", &validateContextAbc{}), Key("B", &validateContextXyz{})).AllowExtraKeys(), ""},
		{"t1.2", m1, Map(Key("A", &validateContextXyz{}), Key("B", &validateContextAbc{})).AllowExtraKeys(), "A: error xyz; B: error abc."},
		{"t1.3", m1, Map(Key("A", &validateContextXyz{}), Key("c", &validateContextXyz{})).AllowExtraKeys(), "A: error xyz; c: error xyz."},
		{"t1.4", m1, Map(Key("g", &validateContextAbc{})).AllowExtraKeys(), "g: error abc."},
		// skip rule
		{"t2.1", m1, Map(Key("g", Skip, &validateContextAbc{})).AllowExtraKeys(), ""},
		{"t2.2", m1, Map(Key("g", &validateContextAbc{}, Skip)).AllowExtraKeys(), "g: error abc."},
		// internal error
		{"t3.1", m2, Map(Key("A", &validateContextAbc{}), Key("B", Required), Key("A", &validateInternalError{})).AllowExtraKeys(), "error internal"},
	}
	for _, test := range tests {
		err := ValidateWithContext(context.Background(), test.model, test.rule)
		assertError(t, test.err, err, test.tag)
	}

	a := map[string]interface{}{"Name": "name", "Value": "demo", "Extra": true}
	err := ValidateWithContext(context.Background(), a, Map(
		Key("Name", Required),
		Key("Value", Required, Length(5, 10)),
	))
	if assert.NotNil(t, err) {
		assert.Equal(t, "Extra: key not expected; Value: the length must be between 5 and 10.", err.Error())
	}
}
