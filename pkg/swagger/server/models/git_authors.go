// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// GitAuthors git authors
//
// swagger:model GitAuthors
type GitAuthors struct {

	// additions
	Additions int64 `json:"additions,omitempty"`

	// deletions
	Deletions int64 `json:"deletions,omitempty"`

	// name
	Name string `json:"name,omitempty"`
}

// Validate validates this git authors
func (m *GitAuthors) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this git authors based on context it is used
func (m *GitAuthors) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *GitAuthors) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *GitAuthors) UnmarshalBinary(b []byte) error {
	var res GitAuthors
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
