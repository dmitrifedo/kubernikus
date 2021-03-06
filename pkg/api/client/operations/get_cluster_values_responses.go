// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/sapcc/kubernikus/pkg/api/models"
)

// GetClusterValuesReader is a Reader for the GetClusterValues structure.
type GetClusterValuesReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetClusterValuesReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewGetClusterValuesOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		result := NewGetClusterValuesDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewGetClusterValuesOK creates a GetClusterValuesOK with default headers values
func NewGetClusterValuesOK() *GetClusterValuesOK {
	return &GetClusterValuesOK{}
}

/*GetClusterValuesOK handles this case with default header values.

OK
*/
type GetClusterValuesOK struct {
	Payload *models.GetClusterValuesOKBody
}

func (o *GetClusterValuesOK) Error() string {
	return fmt.Sprintf("[GET /api/v1/{account}/clusters/{name}/values][%d] getClusterValuesOK  %+v", 200, o.Payload)
}

func (o *GetClusterValuesOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.GetClusterValuesOKBody)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetClusterValuesDefault creates a GetClusterValuesDefault with default headers values
func NewGetClusterValuesDefault(code int) *GetClusterValuesDefault {
	return &GetClusterValuesDefault{
		_statusCode: code,
	}
}

/*GetClusterValuesDefault handles this case with default header values.

Error
*/
type GetClusterValuesDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the get cluster values default response
func (o *GetClusterValuesDefault) Code() int {
	return o._statusCode
}

func (o *GetClusterValuesDefault) Error() string {
	return fmt.Sprintf("[GET /api/v1/{account}/clusters/{name}/values][%d] GetClusterValues default  %+v", o._statusCode, o.Payload)
}

func (o *GetClusterValuesDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
