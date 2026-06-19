# validation

[![GoDoc](https://godoc.org/github.com/monetr/validation?status.png)](http://godoc.org/github.com/monetr/validation)
[![Go Report](https://goreportcard.com/badge/github.com/monetr/validation)](https://goreportcard.com/report/github.com/monetr/validation)

## Description

This package is a fork of the original https://github.com/go-ozzo/ozzo-validation

validation is a Go package that provides configurable and extensible data validation capabilities. It has the following
features:

* use normal programming constructs rather than error-prone struct tags to specify how data should be validated.
* can validate data of different types, e.g., structs, strings, byte slices, slices, maps, arrays.
* can validate custom data types as long as they implement the `Validatable` interface.
* can validate data types that implement the `sql.Valuer` interface (e.g. `sql.NullString`).
* customizable and well-formatted validation errors.
* error code and message translation support.
* provide a rich set of validation rules right out of box.
* extremely easy to create and use custom validation rules.

## Requirements

Go 1.24 or above.

## Getting Started

This package mainly includes a set of validation rules and several validation methods. You use
validation rules to describe how a value should be considered valid, and you call `validation.Validate()`,
`validation.ValidateStruct()`, or `validation.Map()` to validate the value. Many of the rules are generic,
so the value you compare against is type-checked at compile time.

### Installation

Run the following command to install the package:

```
go get github.com/monetr/validation
```

### Validating a Simple Value

For a simple value, such as a string or an integer, you may use `validation.Validate()` to validate it. For example, 

```go
package main

import (
	"fmt"

	"github.com/monetr/validation"
	"github.com/monetr/validation/is"
)

func main() {
	data := "example"
	err := validation.Validate(data,
		validation.Required,       // not empty
		validation.Length(5, 100), // length between 5 and 100
		is.URL,                    // is a valid URL
	)
	fmt.Println(err)
	// Output:
	// must be a valid URL
}
```

The method `validation.Validate()` will run through the rules in the order that they are listed. If a rule fails
the validation, the method will return the corresponding error and skip the rest of the rules. The method will
return nil if the value passes all validation rules.


### Validating a Struct

For a struct value, you usually want to check if its fields are valid. For example, in a RESTful application, you
may unmarshal the request payload into a struct and then validate the struct fields. If one or multiple fields
are invalid, you may want to get an error describing which fields are invalid. You can use `validation.ValidateStruct()`
to achieve this purpose. A single struct can have rules for multiple fields, and a field can be associated with multiple 
rules. For example,

```go
type Address struct {
	Street string
	City   string
	State  string
	Zip    string
}

func (a Address) Validate() error {
	return validation.ValidateStruct(&a,
		// Street cannot be empty, and the length must between 5 and 50
		validation.Field(&a.Street, validation.Required, validation.Length(5, 50)),
		// City cannot be empty, and the length must between 5 and 50
		validation.Field(&a.City, validation.Required, validation.Length(5, 50)),
		// State cannot be empty, and must be a string consisting of two letters in upper case
		validation.Field(&a.State, validation.Required, validation.Match(regexp.MustCompile("^[A-Z]{2}$"))),
		// State cannot be empty, and must be a string consisting of five digits
		validation.Field(&a.Zip, validation.Required, validation.Match(regexp.MustCompile("^[0-9]{5}$"))),
	)
}

a := Address{
    Street: "123",
    City:   "Unknown",
    State:  "Virginia",
    Zip:    "12345",
}

err := a.Validate()
fmt.Println(err)
// Output:
// Street: the length must be between 5 and 50; State: must be in a valid format.
```

Note that when calling `validation.ValidateStruct` to validate a struct, you should pass to the method a pointer 
to the struct instead of the struct itself. Similarly, when calling `validation.Field` to specify the rules
for a struct field, you should use a pointer to the struct field. 

When the struct validation is performed, the fields are validated in the order they are specified in `ValidateStruct`. 
And when each field is validated, its rules are also evaluated in the order they are associated with the field.
If a rule fails, an error is recorded for that field, and the validation will continue with the next field.


### Validating a Map

Sometimes you might need to work with dynamic data stored in maps rather than a typed model. You can use `validation.Map()`
in this situation. A single map can have rules for multiple keys, and a key can be associated with multiple 
rules. For example,

```go
c := map[string]interface{}{
	"Name":  "Qiang Xue",
	"Email": "q",
	"Address": map[string]interface{}{
		"Street": "123",
		"City":   "Unknown",
		"State":  "Virginia",
		"Zip":    "12345",
	},
}

err := validation.Validate(c,
	validation.Map(
		// Name cannot be empty, and the length must be between 5 and 20.
		validation.Key("Name", validation.Required, validation.Length(5, 20)),
		// Email cannot be empty and should be in a valid email format.
		validation.Key("Email", validation.Required, is.Email),
		// Validate Address using its own validation rules
		validation.Key("Address", validation.Map(
			// Street cannot be empty, and the length must between 5 and 50
			validation.Key("Street", validation.Required, validation.Length(5, 50)),
			// City cannot be empty, and the length must between 5 and 50
			validation.Key("City", validation.Required, validation.Length(5, 50)),
			// State cannot be empty, and must be a string consisting of two letters in upper case
			validation.Key("State", validation.Required, validation.Match(regexp.MustCompile("^[A-Z]{2}$"))),
			// State cannot be empty, and must be a string consisting of five digits
			validation.Key("Zip", validation.Required, validation.Match(regexp.MustCompile("^[0-9]{5}$"))),
		)),
	),
)
fmt.Println(err)
// Output:
// Address: (State: must be in a valid format; Street: the length must be between 5 and 50.); Email: must be a valid email address.
```

When the map validation is performed, the keys are validated in the order they are specified in `Map`. 
And when each key is validated, its rules are also evaluated in the order they are associated with the key.
If a rule fails, an error is recorded for that key, and the validation will continue with the next key.


### Validating Unions (one-of)

Sometimes a value is valid if it matches *one of* several shapes — a discriminated union. A login
payload, for example, might be `{email, password}` **or** `{username, password}`, but never both. The
`MatchOneOf` family validates a value against a set of alternative schemas: the first schema that
fully validates wins, and its index is returned so you can act on the matched shape (parse it, merge
it, branch on it). If none match, a `OneOfError` is returned describing every alternative that was
tried.

#### Maps

A map schema is just a `Map(...)` rule, and a variant forbids a field simply by **not listing it** —
`Map` already reports unlisted keys as "key not expected" (unless you call `AllowExtraKeys`), so no
special "forbidden" rule is needed.

```go
withEmail := validation.Map(
	validation.Key("email", validation.Required, is.EmailFormat),
	validation.Key("password", validation.Required),
)
withUsername := validation.Map(
	validation.Key("username", validation.Required),
	validation.Key("password", validation.Required),
)

i, err := validation.MatchOneOf(input, withEmail, withUsername)
// i == 0  -> matched withEmail
// i == 1  -> matched withUsername
// i == -1 -> err is a OneOfError describing both attempts
```

Use `MatchOneOfWithContext` to thread a `context.Context` through to the schema rules.

#### Structs

In a struct every field always exists (as its zero value), so a variant marks the fields it forbids
with the `Never` rule. `MatchOneOfStruct` takes a pointer to the struct and one `[]*FieldRules` per
variant:

```go
emailLogin := []*validation.FieldRules{
	validation.Field(&l.Email, validation.Required, is.EmailFormat),
	validation.Field(&l.Password, validation.Required),
	validation.Field(&l.Username, validation.Never), // forbidden in this variant
}
usernameLogin := []*validation.FieldRules{
	validation.Field(&l.Email, validation.Never),
	validation.Field(&l.Username, validation.Required),
	validation.Field(&l.Password, validation.Required),
}

i, err := validation.MatchOneOfStruct(ctx, &l, emailLogin, usernameLogin)
```

`Never` gives pointer fields nil semantics (a present pointer fails even if it points at an empty
value) and non-pointer fields zero-value semantics (an empty string or `0` passes).

#### The `OneOf` rule and strict matching

When you only need pass/fail and not the winning index, `OneOf(schemas ...Rule)` is the rule form and
composes anywhere a `Rule` is accepted:

```go
err := validation.Validate(input, validation.OneOf(withEmail, withUsername))
```

By default the first matching schema wins (`anyOf` semantics). Call `.Strict()` to require that
*exactly* one schema match; if more than one does, it fails with `ErrAmbiguousMatch` — a useful guard
against schemas that are not mutually exclusive.

#### Per-field unions with `AllOf`

`OneOf` treats each alternative as a *single* `Rule`, which is exactly what you want when each branch
is a whole map or struct schema. But sometimes a single field needs to match one *set* of rules OR a
different set — for example a key that must either be null, OR be a present string of a certain length
and shape. `OneOf` alone can't express the multi-rule branch because it only takes one rule per
alternative. `AllOf(rules ...Rule)` bundles several rules into one (the AND counterpart to `OneOf`'s
OR) so it can stand in as a branch:

```go
keyPattern := regexp.MustCompile(`^[a-z0-9_]+$`)

err := validation.Validate(key, validation.OneOf(
    validation.Nil, // either the key is absent,
    validation.AllOf( // OR it is present and satisfies all three rules.
        validation.Required,
        validation.Length(10, 64),
        validation.Match(keyPattern),
    ),
))
```

The single-rule branch (`Nil`) needs no `AllOf` — a lone `Rule` is already a valid alternative.
`AllOf` only earns its keep on branches made of more than one rule. Like `Validate`, it runs its rules
in order and stops at the first failure, so when an `AllOf` branch fails the `OneOfError` records the
specific sub-rule that did not pass rather than a merged blob. It threads `context.Context` through to
its rules, so a context-aware rule (or a nested `OneOf`) inside the set still sees it.

#### Nesting unions

Because `OneOf` and `AllOf` are both ordinary `Rule`s, they compose to any depth — a union can be a branch of
another union, and a single field inside one variant can carry a union of its own. A realistic case is a
discriminated union of object shapes (told apart by a `type` field) where one of those shapes has a field that is
itself a small union:

```go
// A notification destination is one of three shapes, distinguished by "type".
emailDest := validation.Map(
	validation.Key("type", validation.Required, validation.In("email")),
	validation.Key("address", validation.Required, is.EmailFormat),
)
webhookDest := validation.Map(
	validation.Key("type", validation.Required, validation.In("webhook")),
	validation.Key("url", validation.Required, is.URL),
	// A union nested on a single field: the secret may be null, OR a string of
	// 16-64 chars. The key is optional, so absent / null / value all pass.
	validation.Key("secret", validation.OneOf(
		validation.Nil,
		validation.AllOf(is.String, validation.Length(16, 64)),
	)).Required(false),
)
smsDest := validation.Map(
	validation.Key("type", validation.Required, validation.In("sms")),
	validation.Key("phone", validation.Required, is.E164),
)

// The outer union picks the shape. .Strict() guards against an input that
// somehow satisfies more than one shape at once.
destination := validation.OneOf(emailDest, webhookDest, smsDest).Strict()

err := validation.Validate(input, destination)
```

`AllOf` is also handy *wrapping* a `OneOf` rather than inside one, to gate a union behind a coarse type check.
Asserting the structural type first means a value of the wrong type gets one clean "must be an object" error
instead of a `oneOf` array of every shape it failed to be:

```go
err := validation.Validate(input, validation.AllOf(
	is.Map, // first prove it is an object at all,
	validation.OneOf(emailDest, webhookDest, smsDest), // then that it is one of the shapes.
))
```

The errors nest the same way the rules do: a failed inner `OneOf` produces a `OneOfError` that sits inside the
outer `OneOfError`'s `oneOf` array (or inside a parent `validation.Errors` when the union is one field of a larger
object), so a nested structure serializes to nested JSON rather than a flattened string. See the next section for
the exact shape.

#### The error shape

A failed union produces a `OneOfError`, which marshals to JSON as a single `oneOf` field whose value
is an array of the per-variant error maps, one entry per schema, in order. Each entry contains the
fields that failed for that variant:

```go
b, _ := json.Marshal(err)
fmt.Println(string(b))
// {"oneOf":[{"username":"must not be provided"},{"email":"must not be provided"}]}
```

Because `OneOfError` is returned unwrapped, it nests correctly inside a parent `validation.Errors`
(for example when a union is one field of a larger object), serializing structurally rather than
collapsing to a string.


### Validation Errors

The `validation.ValidateStruct` method returns validation errors found in struct fields in terms of `validation.Errors` 
which is a map of fields and their corresponding errors. Nil is returned if validation passes.

By default, `validation.Errors` uses the struct tags named `json` to determine what names should be used to 
represent the invalid fields. The type also implements the `json.Marshaler` interface so that it can be marshaled 
into a proper JSON object. For example,

```go
type Address struct {
	Street string `json:"street"`
	City   string `json:"city"`
	State  string `json:"state"`
	Zip    string `json:"zip"`
}

// ...perform validation here...

err := a.Validate()
b, _ := json.Marshal(err)
fmt.Println(string(b))
// Output:
// {"street":"the length must be between 5 and 50","state":"must be in a valid format"}
```

You may modify `validation.ErrorTag` to use a different struct tag name.

If you do not like the magic that `ValidateStruct` determines error keys based on struct field names or corresponding
tag values, you may use the following alternative approach:

```go
c := Customer{
	Name:  "Qiang Xue",
	Email: "q",
	Address: Address{
		State:  "Virginia",
	},
}

err := validation.Errors{
	"name": validation.Validate(c.Name, validation.Required, validation.Length(5, 20)),
	"email": validation.Validate(c.Name, validation.Required, is.Email),
	"zip": validation.Validate(c.Address.Zip, validation.Required, validation.Match(regexp.MustCompile("^[0-9]{5}$"))),
}.Filter()
fmt.Println(err)
// Output:
// email: must be a valid email address; zip: cannot be blank.
```

In the above example, we build a `validation.Errors` by a list of names and the corresponding validation results. 
At the end we call `Errors.Filter()` to remove from `Errors` all nils which correspond to those successful validation 
results. The method will return nil if `Errors` is empty.

The above approach is very flexible as it allows you to freely build up your validation error structure. You can use
it to validate both struct and non-struct values. Compared to using `ValidateStruct` to validate a struct, 
it has the drawback that you have to redundantly specify the error keys while `ValidateStruct` can automatically 
find them out.


### Internal Errors

Internal errors are different from validation errors in that internal errors are caused by malfunctioning code (e.g.
a validator making a remote call to validate some data when the remote service is down) rather
than the data being validated. When an internal error happens during data validation, you may allow the user to resubmit
the same data to perform validation again, hoping the program resumes functioning. On the other hand, if data validation
fails due to data error, the user should generally not resubmit the same data again.

To differentiate internal errors from validation errors, when an internal error occurs in a validator, wrap it
into `validation.InternalError` by calling `validation.NewInternalError()`. The user of the validator can then check
if a returned error is an internal error or not. For example,

```go
if err := a.Validate(); err != nil {
	if e, ok := err.(validation.InternalError); ok {
		// an internal error happened
		fmt.Println(e.InternalError())
	}
}
```


## Validatable Types

A type is validatable if it implements the `validation.Validatable` interface. 

When `validation.Validate` is used to validate a validatable value, if it does not find any error with the 
given validation rules, it will further call the value's `Validate()` method. 

Similarly, when `validation.ValidateStruct` is validating a struct field whose type is validatable, it will call 
the field's `Validate` method after it passes the listed rules.

> Note: When implementing `validation.Validatable`, do not call `validation.Validate()` to validate the value in its
> original type because this will cause infinite loops. For example, if you define a new type `MyString` as `string`
> and implement `validation.Validatable` for `MyString`, within the `Validate()` function you should cast the value 
> to `string` first before calling `validation.Validate()` to validate it.

In the following example, the `Address` field of `Customer` is validatable because `Address` implements 
`validation.Validatable`. Therefore, when validating a `Customer` struct with `validation.ValidateStruct`,
validation will "dive" into the `Address` field.

```go
type Customer struct {
	Name    string
	Gender  string
	Email   string
	Address Address
}

func (c Customer) Validate() error {
	return validation.ValidateStruct(&c,
		// Name cannot be empty, and the length must be between 5 and 20.
		validation.Field(&c.Name, validation.Required, validation.Length(5, 20)),
		// Gender is optional, and should be either "Female" or "Male".
		validation.Field(&c.Gender, validation.In("Female", "Male")),
		// Email cannot be empty and should be in a valid email format.
		validation.Field(&c.Email, validation.Required, is.Email),
		// Validate Address using its own validation rules
		validation.Field(&c.Address),
	)
}

c := Customer{
	Name:  "Qiang Xue",
	Email: "q",
	Address: Address{
		Street: "123 Main Street",
		City:   "Unknown",
		State:  "Virginia",
		Zip:    "12345",
	},
}

err := c.Validate()
fmt.Println(err)
// Output:
// Address: (State: must be in a valid format.); Email: must be a valid email address.
```

Sometimes, you may want to skip the invocation of a type's `Validate` method. To do so, simply associate
a `validation.Skip` rule with the value being validated.

### Maps/Slices/Arrays of Validatables

When validating an iterable (map, slice, or array), whose element type implements the `validation.Validatable` interface,
the `validation.Validate` method will call the `Validate` method of every non-nil element.
The validation errors of the elements will be returned as `validation.Errors` which maps the keys of the
invalid elements to their corresponding validation errors. For example,

```go
addresses := []Address{
	Address{State: "MD", Zip: "12345"},
	Address{Street: "123 Main St", City: "Vienna", State: "VA", Zip: "12345"},
	Address{City: "Unknown", State: "NC", Zip: "123"},
}
err := validation.Validate(addresses)
fmt.Println(err)
// Output:
// 0: (City: cannot be blank; Street: cannot be blank.); 2: (Street: cannot be blank; Zip: must be in a valid format.).
```

When using `validation.ValidateStruct` to validate a struct, the above validation procedure also applies to those struct 
fields which are map/slices/arrays of validatables. 

#### Each

The `Each` validation rule allows you to apply a set of rules to each element of an array, slice, or map.

```go
type Customer struct {
    Name      string
    Emails    []string
}

func (c Customer) Validate() error {
    return validation.ValidateStruct(&c,
        // Name cannot be empty, and the length must be between 5 and 20.
		validation.Field(&c.Name, validation.Required, validation.Length(5, 20)),
		// Emails are optional, but if given must be valid.
		validation.Field(&c.Emails, validation.Each(is.Email)),
    )
}

c := Customer{
    Name:   "Qiang Xue",
    Emails: []Email{
        "valid@example.com",
        "invalid",
    },
}

err := c.Validate()
fmt.Println(err)
// Output:
// Emails: (1: must be a valid email address.).
```

### Pointers

When a value being validated is a pointer, most validation rules will validate the actual value pointed to by the pointer.
If the pointer is nil, these rules will skip the validation.

An exception is the `validation.Required` and `validation.NotNil` rules. When a pointer is nil, they
will report a validation error.


### Types Implementing `sql.Valuer`

If a data type implements the `sql.Valuer` interface (e.g. `sql.NullString`), the built-in validation rules will handle
it properly. In particular, when a rule is validating such data, it will call the `Value()` method and validate
the returned value instead.


### Required vs. Not Nil

When validating input values, there are two different scenarios about checking if input values are provided or not.

In the first scenario, an input value is considered missing if it is not entered or it is entered as a zero value
(e.g. an empty string, a zero integer). You can use the `validation.Required` rule in this case.

In the second scenario, an input value is considered missing only if it is not entered. A pointer field is usually
used in this case so that you can detect if a value is entered or not by checking if the pointer is nil or not.
You can use the `validation.NotNil` rule to ensure a value is entered (even if it is a zero value).

#### Presence and nullability are two different questions

"Required" tends to get used as a single knob, but for JSON input there are really two independent questions, and
it helps to keep them apart:

- **Presence** — is the field in the payload at all? (`{"x": ...}` vs `{}`)
- **Nullability** — if it is present, may its value be `null`? (`{"x": null}` vs `{"x": <value>}`)

These are orthogonal, so crossing them gives four states, not three:

|                     | null not allowed                                | null allowed              |
| ------------------- | ----------------------------------------------- | ------------------------- |
| **must be present** | required, non-nullable (the classic "required") | **required but nullable** |
| **may be absent**   | optional, but not null if present               | optional and nullable     |

The rules line up with the two axes like this:

- Presence is handled by the `Map` machinery: `validation.Key("x").Required(true)` reports a missing key as
  "required key is missing", and `.Required(false)` lets the key be absent. (Required is the default, so a `Key`
  you do not call `.Required(false)` on must be present.)
- Nullability is a value rule: `validation.NotNil` forbids a `null` value ("must not be nil") and
  `validation.Nil` requires one ("must be nil").
- `validation.Required` is a third, separate thing — it rejects the zero value (empty string, `0`, ...), not just
  `null`. Reach for `NotNil` when you only care about null, and `Required` when an empty value is also "missing".

So all four states are expressible against a map:

```go
schema := validation.Map(
	// required, non-nullable: must be present AND not null.
	validation.Key("a", validation.NotNil).Required(true),
	// required but nullable: must be present, null is fine.
	validation.Key("b").Required(true),
	// optional, but not null if present.
	validation.Key("c", validation.NotNil).Required(false),
	// optional and nullable: present-null, present-value, or absent all pass.
	validation.Key("d").Required(false),
)
```

#### PATCH endpoints (absent vs null)

The reason this distinction matters in practice is partial updates. In a `PATCH` the three input states usually
carry three different meanings:

- field **absent** -> leave it alone,
- field **null** -> clear it,
- field **present with a value** -> set it to that value.

A plain Go struct cannot tell absent from null: `encoding/json` decodes both `{}` and `{"name": null}` into the
same zero value (or the same `nil` pointer), so the distinction is gone before any rule runs. Decoding the body
into a `map[string]any` keeps it — the key is simply not in the map when it was absent. For example, a profile
patch where `name` may be updated but never cleared, while `nickname` may be updated or cleared:

```go
patch := validation.Map(
	// "name" is optional (absent -> leave alone), but if you do send it, it
	// must be a real non-null value, you cannot null out the name.
	validation.Key("name", validation.NotNil, validation.Length(1, 250)).Required(false),
	// "nickname" is optional too, and it IS nullable: send null to clear it, or
	// a string to set it. Absent still means leave alone.
	validation.Key("nickname", validation.Length(1, 250)).Required(false),
)
```

`Length` (like the other value rules) treats a `null` as valid and skips it, so on `nickname` it only constrains a
present, non-null string while still letting `null` through to mean "clear". `NotNil` on `name` is what rejects an
explicit `null` there. Your handler then branches on which keys are actually present in the map:

```go
if v, ok := body["name"]; ok {
	// Present, and NotNil already guaranteed it is not nil -> set it.
	account.Name = v.(string)
}
if v, ok := body["nickname"]; ok {
	// Present: nil means clear, a value means set.
	if v == nil {
		account.Nickname = nil
	} else {
		s := v.(string)
		account.Nickname = &s
	}
}
// Keys that are not in the map were absent, so they are left untouched.
```

If you only ever need "absent vs present" and a meaningful `null` is not a case you care about, a struct of
pointer fields with `NotNil` is simpler — the distinction you give up (null vs absent) is one you would not be
acting on anyway.


### Embedded Structs

The `validation.ValidateStruct` method will properly validate a struct that contains embedded structs. In particular,
the fields of an embedded struct are treated as if they belong directly to the containing struct. For example,

```go
type Employee struct {
	Name string
}

type Manager struct {
	Employee
	Level int
}

m := Manager{}
err := validation.ValidateStruct(&m,
	validation.Field(&m.Name, validation.Required),
	validation.Field(&m.Level, validation.Required),
)
fmt.Println(err)
// Output:
// Level: cannot be blank; Name: cannot be blank.
```

In the above code, we use `&m.Name` to specify the validation of the `Name` field of the embedded struct `Employee`.
And the validation error uses `Name` as the key for the error associated with the `Name` field as if `Name` a field
directly belonging to `Manager`.

If `Employee` implements the `validation.Validatable` interface, we can also use the following code to validate
`Manager`, which generates the same validation result:

```go
func (e Employee) Validate() error {
	return validation.ValidateStruct(&e,
		validation.Field(&e.Name, validation.Required),
	)
}

err := validation.ValidateStruct(&m,
	validation.Field(&m.Employee),
	validation.Field(&m.Level, validation.Required),
)
fmt.Println(err)
// Output:
// Level: cannot be blank; Name: cannot be blank.
```


### Conditional Validation

Sometimes, we may want to validate a value only when certain condition is met. For example, we want to ensure the 
`unit` struct field is not empty only when the `quantity` field is not empty; or we may want to ensure either `email`
or `phone` is provided. The so-called conditional validation can be achieved with the help of `validation.When`.
The following code implements the aforementioned examples:

```go
result := validation.ValidateStruct(&a,
    validation.Field(&a.Unit, validation.When(a.Quantity != "", validation.Required).Else(validation.Nil)),
    validation.Field(&a.Phone, validation.When(a.Email == "", validation.Required.Error('Either phone or Email is required.')),
    validation.Field(&a.Email, validation.When(a.Phone == "", validation.Required.Error('Either phone or Email is required.')),
)
```

Note that `validation.When` and `validation.When.Else` can take a list of validation rules. These rules will be executed only when the condition is true (When) or false (Else).

The above code can also be simplified using the shortcut `validation.Required.When`:

```go
result := validation.ValidateStruct(&a,
    validation.Field(&a.Unit, validation.Required.When(a.Quantity != ""), validation.Nil.When(a.Quantity == "")),
    validation.Field(&a.Phone, validation.Required.When(a.Email == "").Error('Either phone or Email is required.')),
    validation.Field(&a.Email, validation.Required.When(a.Phone == "").Error('Either phone or Email is required.')),
)
```

### Comparing Two Fields

`Eq` and `NotEq` compare a value against a constant. To compare one struct field against another
(for example a password and its confirmation), use `EqField`/`NotEqField` and pass a pointer to the
sibling field, just like you pass field pointers to `Field`:

```go
type Registration struct {
	Password        string
	ConfirmPassword string
}

func (r Registration) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Password, validation.Required, validation.Length(8, 100)),
		validation.Field(&r.ConfirmPassword, validation.EqField(&r.Password)),
	)
}

r := Registration{Password: "hunter2!!", ConfirmPassword: "typo"}
fmt.Println(r.Validate())
// Output:
// ConfirmPassword: must be equal to the other value.
```


### Customizing Error Messages

All built-in validation rules allow you to customize their error messages. To do so, simply call the `Error()` method
of the rules. For example,

```go
data := "2123"
err := validation.Validate(data,
	validation.Required.Error("is required"),
	validation.Match(regexp.MustCompile("^[0-9]{5}$")).Error("must be a string with five digits"),
)
fmt.Println(err)
// Output:
// must be a string with five digits
```

You can also customize the pre-defined error(s) of a built-in rule such that the customization applies to *every*
instance of the rule. For example, the `Required` rule uses the pre-defined error `ErrRequired`. You can customize it
during the application initialization:
```go
validation.ErrRequired = validation.ErrRequired.SetMessage("the value is required") 
```

### Error Code and Message Translation

The errors returned by the validation rules implement the `Error` interface which contains the `Code()` method 
to provide the error code information. While the message of a validation error is often customized, the code is immutable.
You can use error code to programmatically check a validation error or look for the translation of the corresponding message.

If you are developing your own validation rules, you can use `validation.NewError()` to create a validation error which
implements the aforementioned `Error` interface.

## Creating Custom Rules

Creating a custom rule is as simple as implementing the `validation.Rule` interface. The interface contains a single
method as shown below, which should validate the value and return the validation error, if any:

```go
// Validate validates a value and returns an error if validation fails.
Validate(value any) error
```

If you already have a function with the same signature as shown above, you can call `validation.By()` to turn
it into a validation rule. For example,

```go
func checkAbc(value interface{}) error {
	s, _ := value.(string)
	if s != "abc" {
		return errors.New("must be abc")
	}
	return nil
}

err := validation.Validate("xyz", validation.By(checkAbc))
fmt.Println(err)
// Output: must be abc
```

If your validation function takes additional parameters, you can use the following closure trick:

```go
func stringEquals(str string) validation.RuleFunc {
	return func(value interface{}) error {
		s, _ := value.(string)
        if s != str {
            return errors.New("unexpected string")
        }
        return nil
    }
}

err := validation.Validate("xyz", validation.By(stringEquals("abc")))
fmt.Println(err)
// Output: unexpected string
```


### Rule Groups

When a combination of several rules are used in multiple places, you may use the following trick to create a 
rule group so that your code is more maintainable.

```go
var NameRule = []validation.Rule{
	validation.Required,
	validation.Length(5, 20),
}

type User struct {
	FirstName string
	LastName  string
}

func (u User) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(&u.FirstName, NameRule...),
		validation.Field(&u.LastName, NameRule...),
	)
}
```

In the above example, we create a rule group `NameRule` which consists of two validation rules. We then use this rule
group to validate both `FirstName` and `LastName`.


## Context-aware Validation

While most validation rules are self-contained, some rules may depend dynamically on a context. A rule may implement the
`validation.RuleWithContext` interface to support the so-called context-aware validation.
 
To validate an arbitrary value with a context, call `validation.ValidateWithContext()`. The `context.Conext` parameter 
will be passed along to those rules that implement `validation.RuleWithContext`.

To validate the fields of a struct with a context, call `validation.ValidateStructWithContext()`. 

You can define a context-aware rule from scratch by implementing both `validation.Rule` and `validation.RuleWithContext`. 
You can also use `validation.WithContext()` to turn a function into a context-aware rule. For example,


```go
rule := validation.WithContext(func(ctx context.Context, value interface{}) error {
	if ctx.Value("secret") == value.(string) {
	    return nil
	}
	return errors.New("value incorrect")
})
value := "xyz"
ctx := context.WithValue(context.Background(), "secret", "example")
err := validation.ValidateWithContext(ctx, value, rule)
fmt.Println(err)
// Output: value incorrect
```

When performing context-aware validation, if a rule does not implement `validation.RuleWithContext`, its
`validation.Rule` will be used instead.


## Built-in Validation Rules

The following rules are provided in the `validation` package:

Many of these rules are generic. The generic type parameter is usually inferred from the argument,
so you can write `validation.In("a", "b")` or `validation.Min(10)` without spelling it out.

**Equality and membership**

* `Eq[T any](expected T)`: checks if a value is equal to `expected` (using `reflect.DeepEqual`).
* `NotEq[T any](forbidden T)`: checks if a value is NOT equal to `forbidden`.
* `EqField[T any](other *T)`: checks if a value equals the field pointed to by `other`. Intended for
  comparing two struct fields, e.g. a password and its confirmation. Pass a pointer to the sibling
  field the same way you pass it to `Field`.
* `NotEqField[T any](other *T)`: checks if a value differs from the field pointed to by `other`,
  e.g. a new password that must not match the current one.
* `In[T any](...T)`: checks if a value can be found in the given list of values.
* `NotIn[T any](...T)`: checks if a value is NOT among the given list of values.

**Numeric and ordered comparisons**

* `Min[T Threshold](min T)` and `Max[T Threshold](max T)`: checks if a value is within the specified
  bound (inclusive). Call `.Exclusive()` to make the bound strict. Supports int, uint, float and
  `time.Time`.
* `Gt[T Threshold](min T)`, `Gte[T Threshold](min T)`, `Lt[T Threshold](max T)`, `Lte[T Threshold](max T)`:
  readable shorthands for strict/inclusive greater-than and less-than comparisons. `Gt` is equivalent
  to `Min(min).Exclusive()`, `Gte` to `Min(min)`, and likewise for `Lt`/`Lte` and `Max`.
* `Between[T Threshold](min, max T)`: checks if a value is within the inclusive range `[min, max]`.
  Call `.Exclusive()` to exclude both boundaries.
* `MultipleOf[T Integer](base T)`: checks if a value is a multiple of `base`.

**Strings**

* `Length(min, max int)`: checks if the length of a value is within the specified range.
  This rule should only be used for validating strings, slices, maps, and arrays.
* `RuneLength(min, max int)`: checks if the length of a string is within the specified range.
  This rule is similar as `Length` except that when the value being validated is a string, it checks
  its rune length instead of byte length.
* `Match(*regexp.Regexp)`: checks if a value matches the specified regular expression.
  This rule should only be used for strings and byte slices.
* `HasPrefix(prefix string)`: checks if a string starts with `prefix`.
* `HasSuffix(suffix string)`: checks if a string ends with `suffix`.
* `Contains(substring string)`: checks if a string contains `substring`.
* `Date(layout string)`: checks if a string value is a date whose format is specified by the layout.
  By calling `Min()` and/or `Max()`, you can check additionally if the date is within the specified range.

**Presence**

* `Required`: checks if a value is not empty (neither nil nor zero).
* `NotNil`: checks if a pointer value is not nil. Non-pointer values are considered valid.
* `NilOrNotEmpty`: checks if a value is a nil pointer or a non-empty value. This differs from `Required` in that it treats a nil pointer as valid.
* `Nil`: checks if a value is a nil pointer.
* `Empty`: checks if a value is empty. nil pointers are considered valid.
* `Never`: checks that a value is absent — a nil pointer/interface, or the zero value for any other
  type. Use it on the struct fields that a discriminated-union variant forbids (the counterpart to
  `Required`). Unlike `Empty`, a *present* pointer fails even when it references an empty value.

**Composition and control flow**

* `Each(rules ...Rule)`: checks the elements within an iterable (map/slice/array) with other rules.
* `When(condition, rules ...Rule)`: validates with the specified rules only when the condition is true.
* `Else(rules ...Rule)`: must be used with `When(condition, rules ...Rule)`, validates with the specified rules only when the condition is false.
* `Skip`: this is a special rule used to indicate that all rules following it should be skipped (including the nested ones).

**Unions (one-of)**

* `OneOf(schemas ...Rule)`: passes when the value matches at least one of the schemas (the first
  match wins). Call `.Strict()` to require exactly one match. On failure the error is a `OneOfError`
  that marshals to `{"oneOf": [...]}`.
* `AllOf(rules ...Rule)`: passes only when the value matches every one of the rules (the AND
  counterpart to `OneOf`). Its main use is bundling several rules into one so a whole SET of rules can
  be a single branch of a `OneOf`. Stops at the first failing rule.
* `MatchOneOf(value, schemas ...Rule) (int, error)` / `MatchOneOfWithContext`: like `OneOf` but
  returns the index of the matching schema so you can act on the matched shape. Schemas are typically
  `Map(...)` rules; a variant forbids a key by omitting it.
* `MatchOneOfStruct[T](ctx, *T, schemas ...[]*FieldRules) (int, error)`: the struct counterpart,
  taking one `[]*FieldRules` per variant. Use `Never` to forbid the fields a variant disallows.

The `is` sub-package provides a list of commonly used string validation rules that can be used to check if the format
of a value satisfies certain requirements. Note that these rules only handle strings and byte slices and if a string
 or byte slice is empty, it is considered valid. You may use a `Required` rule to ensure a value is not empty.
Below is the whole list of the rules provided by the `is` package:

* `Email`: validates if a string is an email or not. It also checks if the MX record exists for the email domain.
* `EmailFormat`: validates if a string is an email or not. It does NOT check the existence of the MX record.
* `URL`: validates if a string is a valid URL
* `RequestURL`: validates if a string is a valid request URL
* `RequestURI`: validates if a string is a valid request URI
* `Alpha`: validates if a string contains English letters only (a-zA-Z)
* `Digit`: validates if a string contains digits only (0-9)
* `Alphanumeric`: validates if a string contains English letters and digits only (a-zA-Z0-9)
* `UTFLetter`: validates if a string contains unicode letters only
* `UTFDigit`: validates if a string contains unicode decimal digits only
* `UTFLetterNumeric`: validates if a string contains unicode letters and numbers only
* `UTFNumeric`: validates if a string contains unicode number characters (category N) only
* `LowerCase`: validates if a string contains lower case unicode letters only
* `UpperCase`: validates if a string contains upper case unicode letters only
* `Hexadecimal`: validates if a string is a valid hexadecimal number
* `HexColor`: validates if a string is a valid hexadecimal color code
* `RGBColor`: validates if a string is a valid RGB color in the form of rgb(R, G, B)
* `Int`: validates if a string is a valid integer number
* `Float`: validates if a string is a floating point number
* `UUIDv3`: validates if a string is a valid version 3 UUID
* `UUIDv4`: validates if a string is a valid version 4 UUID
* `UUIDv5`: validates if a string is a valid version 5 UUID
* `UUID`: validates if a string is a valid UUID
* `CreditCard`: validates if a string is a valid credit card number
* `ISBN10`: validates if a string is an ISBN version 10
* `ISBN13`: validates if a string is an ISBN version 13
* `ISBN`: validates if a string is an ISBN (either version 10 or 13)
* `JSON`: validates if a string is in valid JSON format
* `ASCII`: validates if a string contains ASCII characters only
* `PrintableASCII`: validates if a string contains printable ASCII characters only
* `PrintableUnicode`: validates if a string contains printable characters only (via `unicode.IsPrint`), allowing international text and emoji while rejecting tabs, newlines, and invisible characters
* `Multibyte`: validates if a string contains multibyte characters
* `FullWidth`: validates if a string contains full-width characters
* `HalfWidth`: validates if a string contains half-width characters
* `VariableWidth`: validates if a string contains both full-width and half-width characters
* `Base64`: validates if a string is encoded in Base64
* `Base32`: validates if a string is encoded in Base32
* `DataURI`: validates if a string is a valid base64-encoded data URI
* `E164`: validates if a string is a valid E164 phone number (+19251232233)
* `CountryCode2`: validates if a string is a valid ISO3166 Alpha 2 country code
* `CountryCode3`: validates if a string is a valid ISO3166 Alpha 3 country code
* `DialString`: validates if a string is a valid dial string that can be passed to Dial()
* `MAC`: validates if a string is a MAC address
* `IP`: validates if a string is a valid IP address (either version 4 or 6)
* `IPv4`: validates if a string is a valid version 4 IP address
* `IPv6`: validates if a string is a valid version 6 IP address
* `Subdomain`: validates if a string is valid subdomain
* `Domain`: validates if a string is valid domain
* `DNSName`: validates if a string is valid DNS name
* `Host`: validates if a string is a valid IP (both v4 and v6) or a valid DNS name
* `Port`: validates if a string is a valid port number
* `MongoID`: validates if a string is a valid Mongo ID
* `Latitude`: validates if a string is a valid latitude
* `Longitude`: validates if a string is a valid longitude
* `SSN`: validates if a string is a social security number (SSN)
* `Semver`: validates if a string is a valid semantic version

## Credits

The `is` sub-package wraps the excellent validators provided by the [govalidator](https://github.com/asaskevich/govalidator) package.
