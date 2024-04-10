// Code generated by go-swagger; DO NOT EDIT.

// Copyright Authors of Cilium
// SPDX-License-Identifier: Apache-2.0

package daemon

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// GetMapNameEventsReader is a Reader for the GetMapNameEvents structure.
type GetMapNameEventsReader struct {
	formats strfmt.Registry
	writer  io.Writer
}

// ReadResponse reads a server response into the received o.
func (o *GetMapNameEventsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetMapNameEventsOK(o.writer)
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 404:
		result := NewGetMapNameEventsNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("[GET /map/{name}/events] GetMapNameEvents", response, response.Code())
	}
}

// NewGetMapNameEventsOK creates a GetMapNameEventsOK with default headers values
func NewGetMapNameEventsOK(writer io.Writer) *GetMapNameEventsOK {
	return &GetMapNameEventsOK{

		Payload: writer,
	}
}

/*
GetMapNameEventsOK describes a response with status code 200, with default header values.

Success
*/
type GetMapNameEventsOK struct {
	Payload io.Writer
}

// IsSuccess returns true when this get map name events o k response has a 2xx status code
func (o *GetMapNameEventsOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this get map name events o k response has a 3xx status code
func (o *GetMapNameEventsOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get map name events o k response has a 4xx status code
func (o *GetMapNameEventsOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this get map name events o k response has a 5xx status code
func (o *GetMapNameEventsOK) IsServerError() bool {
	return false
}

// IsCode returns true when this get map name events o k response a status code equal to that given
func (o *GetMapNameEventsOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the get map name events o k response
func (o *GetMapNameEventsOK) Code() int {
	return 200
}

func (o *GetMapNameEventsOK) Error() string {
	return fmt.Sprintf("[GET /map/{name}/events][%d] getMapNameEventsOK  %+v", 200, o.Payload)
}

func (o *GetMapNameEventsOK) String() string {
	return fmt.Sprintf("[GET /map/{name}/events][%d] getMapNameEventsOK  %+v", 200, o.Payload)
}

func (o *GetMapNameEventsOK) GetPayload() io.Writer {
	return o.Payload
}

func (o *GetMapNameEventsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetMapNameEventsNotFound creates a GetMapNameEventsNotFound with default headers values
func NewGetMapNameEventsNotFound() *GetMapNameEventsNotFound {
	return &GetMapNameEventsNotFound{}
}

/*
GetMapNameEventsNotFound describes a response with status code 404, with default header values.

Map not found
*/
type GetMapNameEventsNotFound struct {
}

// IsSuccess returns true when this get map name events not found response has a 2xx status code
func (o *GetMapNameEventsNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get map name events not found response has a 3xx status code
func (o *GetMapNameEventsNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get map name events not found response has a 4xx status code
func (o *GetMapNameEventsNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this get map name events not found response has a 5xx status code
func (o *GetMapNameEventsNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this get map name events not found response a status code equal to that given
func (o *GetMapNameEventsNotFound) IsCode(code int) bool {
	return code == 404
}

// Code gets the status code for the get map name events not found response
func (o *GetMapNameEventsNotFound) Code() int {
	return 404
}

func (o *GetMapNameEventsNotFound) Error() string {
	return fmt.Sprintf("[GET /map/{name}/events][%d] getMapNameEventsNotFound ", 404)
}

func (o *GetMapNameEventsNotFound) String() string {
	return fmt.Sprintf("[GET /map/{name}/events][%d] getMapNameEventsNotFound ", 404)
}

func (o *GetMapNameEventsNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}
