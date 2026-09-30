package flatten

import (
	"encoding/json/v2"
	"fmt"
	"slices"

	"github.com/MarkRosemaker/errpath"
	"github.com/MarkRosemaker/openapi"
	"github.com/ettle/strcase"
)

type mode int

const (
	moveIfNecessary mode = iota
	alwaysMove
	neverMove
)

// inlineSchema moves s into the components if it deserves a name of its own, leaving a reference in its place, and flattens what it contains.
func inlineSchema(d *openapi.Document, s *openapi.Schema, name string, mode mode) error {
	if s.Ref != nil {
		return nil // already processed
	}

	if mode == alwaysMove {
		// process the schema itself
		return schema(d, moveSchemaToComponents(d, name, s), name)
	}

	switch s.Type {
	case openapi.TypeInteger, openapi.TypeNumber, openapi.TypeBoolean, openapi.TypeNull: // no need to move to components
	case openapi.TypeString:
		if s.Enum != nil && mode != neverMove {
			s = moveSchemaToComponents(d, name, s)
		} // else just string, no need to move to components
	case openapi.TypeArray:
		if len(s.PrefixItems) > 0 {
			// a tuple has a defined shape, positional but no less real than
			// an object's properties: give it a name of its own the same way.
			if mode != neverMove {
				s = moveSchemaToComponents(d, name, s)
			}
		} else if items := s.Items; items != nil { // no items when only ever seen empty
			switch items.Type {
			case openapi.TypeInteger: // do nothing, just []int
			case openapi.TypeNumber: // do nothing, just []float32 or []float64
			case openapi.TypeString:
				if items.Enum != nil && mode != neverMove {
					s = moveSchemaToComponents(d, name, s)
				} // else just []string, no need to move to components
			case openapi.TypeObject:
				if len(items.Properties) > 0 && mode != neverMove {
					s = moveSchemaToComponents(d, name, s)
				}
			case openapi.TypeBoolean: // do nothing, just []bool
			case openapi.TypeNull: // do nothing, just []null
			case openapi.TypeArray: // TODO: later
			case "": // no explicit type — items uses anyOf / oneOf / allOf (e.g. nullable union)
			default:
				return fmt.Errorf("unimplemented item type %q", items.Type)
			}
		}
	case openapi.TypeObject: // move to components
		if len(s.Properties) > 0 && mode != neverMove {
			s = moveSchemaToComponents(d, name, s)
		}
	case "": // no explicit type — oneOf / anyOf / allOf composition or bare properties
		v := s
		hasComposition := len(v.OneOf) > 0 || len(v.AnyOf) > 0 || len(v.AllOf) > 0
		if (hasComposition || len(v.Properties) > 0) && mode != neverMove {
			s = moveSchemaToComponents(d, name, s)
		}
	default:
		return fmt.Errorf("unimplemented schema ref type %q", s.Type)
	}

	// process the schema itself
	return schema(d, s, name)
}

func schema(d *openapi.Document, s *openapi.Schema, name string) error {
	switch s.Type {
	case openapi.TypeString,
		openapi.TypeInteger,
		openapi.TypeNumber,
		openapi.TypeBoolean,
		openapi.TypeNull: // no need to do anything
		return nil
	case openapi.TypeArray, openapi.TypeObject: // do below
	case "": // is valid if schema contains oneOf, anyOf, allOf, or properties
	default:
		return fmt.Errorf("unimplemented schema type %q", s.Type)
	}

	distributeUnion(s)

	if err := inlineSchemaList(d, s.AllOf, name+"AllOf", neverMove); err != nil {
		return &errpath.ErrField{Field: "allOf", Err: err}
	}

	// a branch with a shape of its own is named like any other; allOf's stay inline, as they merge into s
	if err := inlineBranches(d, s.OneOf, name+"OneOf"); err != nil {
		return &errpath.ErrField{Field: "oneOf", Err: err}
	}

	if err := inlineBranches(d, s.AnyOf, name+"AnyOf"); err != nil {
		return &errpath.ErrField{Field: "anyOf", Err: err}
	}

	// each position is a real, reusable shape, the same as an object property
	// just addressed by index instead of by name.
	if err := inlineSchemaList(d, s.PrefixItems, name+"Item", moveIfNecessary); err != nil {
		return &errpath.ErrField{Field: "prefixItems", Err: err}
	}

	if s.Items != nil {
		if err := inlineSchema(d, s.Items, name+"Item", moveIfNecessary); err != nil {
			return &errpath.ErrField{Field: "items", Err: err}
		}
	}

	if err := inlineSchemas(d, s.Properties, name); err != nil {
		return &errpath.ErrField{Field: "properties", Err: err}
	}

	if ap := s.AdditionalProperties; ap != nil && ap.Schema != nil {
		if err := inlineSchema(d, ap.Schema, name+"Value", moveIfNecessary); err != nil {
			return &errpath.ErrField{Field: "additionalProperties", Err: err}
		}
	}

	if s.PropertyNames != nil {
		if err := inlineSchema(d, s.PropertyNames, name+"Key", moveIfNecessary); err != nil {
			return &errpath.ErrField{Field: "propertyNames", Err: err}
		}
	}

	if s.Not != nil {
		if err := inlineSchema(d, s.Not, name+"Not", moveIfNecessary); err != nil {
			return &errpath.ErrField{Field: "not", Err: err}
		}
	}

	return nil
}

// moveSchemaToComponents puts a copy of s in the components and makes s a reference to it, keeping s's place. It returns the copy.
func moveSchemaToComponents(d *openapi.Document, name string, s *openapi.Schema) *openapi.Schema {
	moved := *s
	name = uniqueName(d.Components.Schemas, name)
	d.Components.Schemas.Set(name, &moved)
	s.Replace(&openapi.Schema{Ref: &openapi.SchemaRef{Identifier: newRef("schemas", name).Identifier, Value: &moved}})

	return &moved
}

// inlineBranches flattens the alternatives of a oneOf or anyOf, naming each by its title, or else by prefix and position.
func inlineBranches(d *openapi.Document, ss openapi.SchemaList, prefix string) error {
	for i, s := range ss {
		name := fmt.Sprintf("%s%d", prefix, i)
		if s.Title != "" {
			name = strcase.ToGoPascal(s.Title)
		}

		if err := inlineSchema(d, s, name, moveIfNecessary); err != nil {
			return &errpath.ErrIndex{Index: i, Err: err}
		}
	}

	return nil
}

func inlineSchemaList(d *openapi.Document, ss openapi.SchemaList, prefix string, mode mode) error {
	for i, s := range ss {
		if err := inlineSchema(d, s, fmt.Sprintf("%s%d", prefix, i), mode); err != nil {
			return &errpath.ErrIndex{Index: i, Err: err}
		}
	}

	return nil
}

// distributeUnion rewrites allOf: [X, {oneOf: [A, B]}] as oneOf: [{allOf: [X, A]}, {allOf: [X, B]}], and the same for anyOf.
//
// Both say "X, and A or B", but only the second gives each alternative a shape of its own to name. It applies only where
// that is all s says: a single union among its allOf entries, inline or a component that is only a union, carrying
// nothing but the union, and no union or properties of s's own.
func distributeUnion(s *openapi.Schema) {
	if len(s.OneOf) > 0 || len(s.AnyOf) > 0 || !onlyDocumentation(s, func(c *openapi.Schema) {
		c.AllOf, c.Type = nil, ""
	}) || s.Type != "" && s.Type != openapi.TypeObject {
		return
	}

	at := -1

	var union *openapi.Schema

	for i, e := range s.AllOf {
		u := e
		if e.Ref != nil {
			// a component that is only a union counts as one, and stays as it is for whatever else refers to it
			if !onlyDocumentation(e, func(c *openapi.Schema) { c.Ref = nil }) {
				continue
			}

			u = e.Ref.Value
		}

		if len(u.OneOf) == 0 && len(u.AnyOf) == 0 {
			continue
		}

		if at >= 0 || len(u.OneOf) > 0 && len(u.AnyOf) > 0 ||
			!onlyDocumentation(u, func(c *openapi.Schema) { c.OneOf, c.AnyOf, c.Discriminator = nil, nil, nil }) {
			return // more than one union, or one that says more than its alternatives
		}

		at, union = i, u
	}

	if at < 0 {
		return
	}

	alternatives := union.OneOf
	if len(alternatives) == 0 {
		alternatives = union.AnyOf
	}

	branches := make(openapi.SchemaList, len(alternatives))
	for i, alt := range alternatives {
		allOf := slices.Clone(s.AllOf)
		allOf[at] = alt
		branches[i] = &openapi.Schema{AllOf: allOf}

		// the title names the branch now, the combination the alternative stands for, unless others share the alternative
		if union == s.AllOf[at] {
			branches[i].Title, alt.Title = alt.Title, ""
		}
	}

	if len(union.OneOf) > 0 {
		s.OneOf = branches
	} else {
		s.AnyOf = branches
	}

	s.AllOf, s.Discriminator = nil, union.Discriminator
}

// onlyDocumentation reports whether s says nothing but documentation once clear has removed what the caller accounts for.
func onlyDocumentation(s *openapi.Schema, clear func(*openapi.Schema)) bool {
	c := *s
	clear(&c)
	c.Title, c.Description, c.Deprecated = "", "", false
	c.Default, c.Example, c.Examples, c.Extensions = nil, nil, nil, nil

	b, err := json.Marshal(&c)

	return err == nil && string(b) == "{}"
}
