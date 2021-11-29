// Code generated by go-swagger; DO NOT EDIT.

// Copyright 2017-2021 Authors of Cilium
// SPDX-License-Identifier: Apache-2.0

package recorder

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewGetRecorderParams creates a new GetRecorderParams object
// with the default values initialized.
func NewGetRecorderParams() *GetRecorderParams {

	return &GetRecorderParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewGetRecorderParamsWithTimeout creates a new GetRecorderParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewGetRecorderParamsWithTimeout(timeout time.Duration) *GetRecorderParams {

	return &GetRecorderParams{

		timeout: timeout,
	}
}

// NewGetRecorderParamsWithContext creates a new GetRecorderParams object
// with the default values initialized, and the ability to set a context for a request
func NewGetRecorderParamsWithContext(ctx context.Context) *GetRecorderParams {

	return &GetRecorderParams{

		Context: ctx,
	}
}

// NewGetRecorderParamsWithHTTPClient creates a new GetRecorderParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewGetRecorderParamsWithHTTPClient(client *http.Client) *GetRecorderParams {

	return &GetRecorderParams{
		HTTPClient: client,
	}
}

/*GetRecorderParams contains all the parameters to send to the API endpoint
for the get recorder operation typically these are written to a http.Request
*/
type GetRecorderParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the get recorder params
func (o *GetRecorderParams) WithTimeout(timeout time.Duration) *GetRecorderParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get recorder params
func (o *GetRecorderParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get recorder params
func (o *GetRecorderParams) WithContext(ctx context.Context) *GetRecorderParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get recorder params
func (o *GetRecorderParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get recorder params
func (o *GetRecorderParams) WithHTTPClient(client *http.Client) *GetRecorderParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get recorder params
func (o *GetRecorderParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *GetRecorderParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
