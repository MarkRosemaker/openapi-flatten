package openapi

import (
	"net/url"

	"github.com/MarkRosemaker/errpath"
	"github.com/go-api-libs/types"
)

// Contact information for the exposed API.
// ([Specification])
//
// [Specification]: https://spec.openapis.org/oas/v3.1.0#contact-object
type Contact struct {
	// The identifying name of the contact person/organization.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// The URL pointing to the contact information. MUST be in the format of a URL.
	URL *url.URL `json:"url,omitempty" yaml:"url,omitempty"`
	// The email address of the contact person/organization. This MUST be in the form of an email address.
	Email types.Email `json:"email,omitempty" yaml:"email,omitempty"`
	// This object MAY be extended with Specification Extensions.
	Extensions Extensions `json:",inline" yaml:",inline"`
}

// Validate checks the contact for consistency.
func (c *Contact) Validate() error {
	// assume that the scheme is https and add it if it is missing
	fixScheme(c.URL)

	if c.Email != "" {
		if err := c.Email.Validate(); err != nil {
			return &errpath.ErrField{Field: "email", Err: err}
		}
	}

	return validateExtensions(c.Extensions)
}
