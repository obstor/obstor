package quick

import "reflect"

// Field represents a single struct field.
type Field struct {
	name  string
	value interface{}
	kind  reflect.Kind
}

// Name returns the name of the field.
func (f *Field) Name() string {
	return f.name
}

// Value returns the underlying value of the field.
func (f *Field) Value() interface{} {
	return f.value
}

// Kind returns the reflect.Kind of the field.
func (f *Field) Kind() reflect.Kind {
	return f.kind
}

// Struct wraps a struct value for introspection.
type Struct struct {
	value reflect.Value
	typ   reflect.Type
}

// newStruct creates a Struct from a value, dereferencing pointers as needed.
func newStruct(v interface{}) *Struct {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	return &Struct{
		value: rv,
		typ:   rv.Type(),
	}
}

// Name returns the struct type name.
func (s *Struct) Name() string {
	return s.typ.Name()
}

// Field returns a *Field for the given field name.
// It panics if the field does not exist.
func (s *Struct) Field(name string) *Field {
	f, ok := s.FieldOk(name)
	if !ok {
		panic("quick/structs: field not found: " + name)
	}
	return f
}

// FieldOk returns a *Field for the given field name and a bool indicating
// whether the field was found.
func (s *Struct) FieldOk(name string) (*Field, bool) {
	sf, ok := s.typ.FieldByName(name)
	if !ok {
		return nil, false
	}
	fv := s.value.FieldByIndex(sf.Index)
	return &Field{
		name:  sf.Name,
		value: fv.Interface(),
		kind:  sf.Type.Kind(),
	}, true
}

// fields returns all exported fields of the struct as a slice of *Field.
func (s *Struct) fields() []*Field {
	var result []*Field
	for i := 0; i < s.typ.NumField(); i++ {
		sf := s.typ.Field(i)
		if !sf.IsExported() {
			continue
		}
		fv := s.value.Field(i)
		result = append(result, &Field{
			name:  sf.Name,
			value: fv.Interface(),
			kind:  sf.Type.Kind(),
		})
	}
	return result
}

// isStruct reports whether v is a struct or a pointer to a struct.
func isStruct(v interface{}) bool {
	if v == nil {
		return false
	}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	return rv.Kind() == reflect.Struct
}

// structFields returns all exported fields of the given struct (or pointer
// to struct) as a slice of *Field.
func structFields(v interface{}) []*Field {
	return newStruct(v).fields()
}
