package internal

import (
	"strconv"
	"strings"
	"text/template"

	"github.com/serenize/snaker"
)

// NewTemplateFuncs returns a set of template funcs bound to the supplied args.
func (a *ArgType) NewTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"colcount":       a.colcount,
		"colnames":       a.colnames,
		"colnamesquery":  a.colnamesquery,
		"colprefixnames": a.colprefixnames,
		"colvals":        a.colvals,
		"fieldnames":     a.fieldnames,
		"goparamlist":    a.goparamlist,
		"inc":            a.inc,
		"reniltype":      a.reniltype,
		"retype":         a.retype,
		"shortname":      a.shortname,
		"nulltypeext":    a.nulltypeext,
		"schema":         a.schemafn,
	}
}

// retype checks the type against the known types, adding the custom type
// package (if any).
func (a *ArgType) retype(typ string) string {
	if strings.Contains(typ, ".") {
		return typ
	}

	prefix := ""
	for strings.HasPrefix(typ, "[]") {
		typ = typ[2:]
		prefix = prefix + "[]"
	}

	if _, ok := a.KnownTypeMap[typ]; !ok {
		pkg := a.CustomTypePackage
		if pkg != "" {
			pkg = pkg + "."
		}

		return prefix + pkg + typ
	}

	return prefix + typ
}

// reniltype checks the nil type against the known types (similar to retype),
// adding the custom type package (if applicable).
func (a *ArgType) reniltype(typ string) string {
	if strings.Contains(typ, ".") {
		return typ
	}

	if strings.HasSuffix(typ, "{}") {
		if _, ok := a.KnownTypeMap[typ[:len(typ)-2]]; ok {
			return typ
		}

		pkg := a.CustomTypePackage
		if pkg != "" {
			pkg = pkg + "."
		}

		return pkg + typ
	}

	return typ
}

// shortname checks the passed type against the ShortNameTypeMap and returns
// the value for it. If the type is not found, then the value is calculated and
// stored in the ShortNameTypeMap for use in the future.
func (a *ArgType) shortname(typ string) string {
	var v string
	var ok bool

	// check short name map
	if v, ok = a.ShortNameTypeMap[typ]; !ok {
		// calc the short name
		u := []string{}
		for _, s := range strings.Split(strings.ToLower(snaker.CamelToSnake(typ)), "_") {
			if len(s) > 0 && s != "id" {
				u = append(u, s[:1])
			}
		}
		v = strings.Join(u, "")

		// store back in short name map
		a.ShortNameTypeMap[typ] = v
	}

	return v
}

// inc increements i by 1.
func (a *ArgType) inc(i int) int {
	return i + 1
}

// colnames creates a list of the column names found in fields, excluding any
// Field.Name in ignoreNames.
//
// Used to present a comma separated list of column names, ie in a sql select,
// or update. (ie, field_1, field_2, field_3, ...)
func (a *ArgType) colnames(fields []*Field, ignoreNames ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		if i != 0 {
			str = str + ", "
		}
		str = str + f.Col.ColumnName
		i++
	}

	return str
}

// colnamesquery creates a list of the column names found in as a query and
// joined by sep.
//
// Used to create a sql query list of column names in a where clause (ie,
// field_1 = $1 AND field_2 = $2 AND ... ) or in an update clause (ie, field =
// $1, field = $2, ...)
func (a *ArgType) colnamesquery(fields []*Field, sep string, ignoreNames ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		if i != 0 {
			str = str + sep
		}
		str = str + f.Col.ColumnName + " = " + a.Loader.NthParam(i)
		i++
	}

	return str
}

// colprefixnames creates a list of the column names found in fields with the
// supplied prefix, excluding any FieldName in ignoreNames.
//
// Used to present a comma separated list of column names, with a prefix, ie in
// a sql select, or update. (ie, t.field_1, t.field_2, t.field_3, ...)
func (a *ArgType) colprefixnames(fields []*Field, prefix string, ignoreNames ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		if i != 0 {
			str = str + ", "
		}
		str = str + prefix + "." + f.Col.ColumnName
		i++
	}

	return str
}

// colvals creates a list of value place holders for the fields found in
// fields, excluding any FieldName in ignoreNames.
//
// Used to present a comma separated list of column names, ie in a sql select,
// or update. (ie, $1, $2, $3 ...)
func (a *ArgType) colvals(fields []*Field, ignoreNames ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		if i != 0 {
			str = str + ", "
		}
		str = str + a.Loader.NthParam(i)
		i++
	}

	return str
}

// fieldnames creates a list of field names from the field names of the
// provided fields, adding the prefix provided, and excluding any field name
// in ignoreNames.
//
// Used to present a comma separated list of field names, ie in a Go statement
// (ie, t.Field1, t.Field2, t.Field3 ...)
func (a *ArgType) fieldnames(fields []*Field, prefix string, ignoreNames ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		if i != 0 {
			str = str + ", "
		}
		str = str + prefix + "." + f.Name
		i++
	}

	return str
}

// count returns the 1-based count of fields, excluding any field name in
// ignoreNames.
//
// Used to get the count of fields, and useful for specifying the last sql
// parameter.
func (a *ArgType) colcount(fields []*Field, ignoreNames ...string) int {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	i := 1
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		i++
	}
	return i
}

// goparamlist converts a list of fields into their named Go parameters,
// skipping any field names in ignoreNames.
func (a *ArgType) goparamlist(fields []*Field, addType bool, ignoreNames ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		s := "v" + strconv.Itoa(i)
		if len(f.Name) > 0 {
			n := strings.Split(snaker.CamelToSnake(f.Name), "_")
			s = strings.ToLower(n[0]) + f.Name[len(n[0]):]
		}

		str = str + ", " + s
		if addType {
			str = str + " " + a.retype(f.Type)
		}

		i++
	}

	return str
}

// nulltypeext determines if the type begins with Null (ie, NullString,
// NullInt64, etc) and returns the remainder with the prefix.
func (a *ArgType) nulltypeext(typ string) string {
	if strings.HasPrefix(typ, "sql.Null") {
		return "." + typ[8:]
	}

	return ""
}

// schemafn takes a series of names and joins them with the schema name.
func (a *ArgType) schemafn(s string, names ...string) string {
	n := strings.Join(names, ".")

	if s == "" && n == "" {
		return ""
	}

	if s != "" && n != "" {
		s = s + "."
	}

	return s + n
}