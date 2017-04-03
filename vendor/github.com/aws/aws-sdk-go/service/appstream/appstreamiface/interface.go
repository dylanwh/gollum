// THIS FILE IS AUTOMATICALLY GENERATED. DO NOT EDIT.

// Package appstreamiface provides an interface to enable mocking the Amazon AppStream service client
// for testing your code.
//
// It is important to note that this interface will have breaking changes
// when the service model is updated and adds new API operations, paginators,
// and waiters.
package appstreamiface

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/appstream"
)

// AppStreamAPI provides an interface to enable mocking the
// appstream.AppStream service client's API operation,
// paginators, and waiters. This make unit testing your code that calls out
// to the SDK's service client's calls easier.
//
// The best way to use this interface is so the SDK's service client's calls
// can be stubbed out for unit testing your code with the SDK without needing
// to inject custom request handlers into the the SDK's request pipeline.
//
//    // myFunc uses an SDK service client to make a request to
//    // Amazon AppStream.
//    func myFunc(svc appstreamiface.AppStreamAPI) bool {
//        // Make svc.AssociateFleet request
//    }
//
//    func main() {
//        sess := session.New()
//        svc := appstream.New(sess)
//
//        myFunc(svc)
//    }
//
// In your _test.go file:
//
//    // Define a mock struct to be used in your unit tests of myFunc.
//    type mockAppStreamClient struct {
//        appstreamiface.AppStreamAPI
//    }
//    func (m *mockAppStreamClient) AssociateFleet(input *appstream.AssociateFleetInput) (*appstream.AssociateFleetOutput, error) {
//        // mock response/functionality
//    }
//
//    func TestMyFunc(t *testing.T) {
//        // Setup Test
//        mockSvc := &mockAppStreamClient{}
//
//        myfunc(mockSvc)
//
//        // Verify myFunc's functionality
//    }
//
// It is important to note that this interface will have breaking changes
// when the service model is updated and adds new API operations, paginators,
// and waiters. Its suggested to use the pattern above for testing, or using
// tooling to generate mocks to satisfy the interfaces.
type AppStreamAPI interface {
	AssociateFleet(*appstream.AssociateFleetInput) (*appstream.AssociateFleetOutput, error)
	AssociateFleetWithContext(aws.Context, *appstream.AssociateFleetInput, ...request.Option) (*appstream.AssociateFleetOutput, error)
	AssociateFleetRequest(*appstream.AssociateFleetInput) (*request.Request, *appstream.AssociateFleetOutput)

	CreateFleet(*appstream.CreateFleetInput) (*appstream.CreateFleetOutput, error)
	CreateFleetWithContext(aws.Context, *appstream.CreateFleetInput, ...request.Option) (*appstream.CreateFleetOutput, error)
	CreateFleetRequest(*appstream.CreateFleetInput) (*request.Request, *appstream.CreateFleetOutput)

	CreateStack(*appstream.CreateStackInput) (*appstream.CreateStackOutput, error)
	CreateStackWithContext(aws.Context, *appstream.CreateStackInput, ...request.Option) (*appstream.CreateStackOutput, error)
	CreateStackRequest(*appstream.CreateStackInput) (*request.Request, *appstream.CreateStackOutput)

	CreateStreamingURL(*appstream.CreateStreamingURLInput) (*appstream.CreateStreamingURLOutput, error)
	CreateStreamingURLWithContext(aws.Context, *appstream.CreateStreamingURLInput, ...request.Option) (*appstream.CreateStreamingURLOutput, error)
	CreateStreamingURLRequest(*appstream.CreateStreamingURLInput) (*request.Request, *appstream.CreateStreamingURLOutput)

	DeleteFleet(*appstream.DeleteFleetInput) (*appstream.DeleteFleetOutput, error)
	DeleteFleetWithContext(aws.Context, *appstream.DeleteFleetInput, ...request.Option) (*appstream.DeleteFleetOutput, error)
	DeleteFleetRequest(*appstream.DeleteFleetInput) (*request.Request, *appstream.DeleteFleetOutput)

	DeleteStack(*appstream.DeleteStackInput) (*appstream.DeleteStackOutput, error)
	DeleteStackWithContext(aws.Context, *appstream.DeleteStackInput, ...request.Option) (*appstream.DeleteStackOutput, error)
	DeleteStackRequest(*appstream.DeleteStackInput) (*request.Request, *appstream.DeleteStackOutput)

	DescribeFleets(*appstream.DescribeFleetsInput) (*appstream.DescribeFleetsOutput, error)
	DescribeFleetsWithContext(aws.Context, *appstream.DescribeFleetsInput, ...request.Option) (*appstream.DescribeFleetsOutput, error)
	DescribeFleetsRequest(*appstream.DescribeFleetsInput) (*request.Request, *appstream.DescribeFleetsOutput)

	DescribeImages(*appstream.DescribeImagesInput) (*appstream.DescribeImagesOutput, error)
	DescribeImagesWithContext(aws.Context, *appstream.DescribeImagesInput, ...request.Option) (*appstream.DescribeImagesOutput, error)
	DescribeImagesRequest(*appstream.DescribeImagesInput) (*request.Request, *appstream.DescribeImagesOutput)

	DescribeSessions(*appstream.DescribeSessionsInput) (*appstream.DescribeSessionsOutput, error)
	DescribeSessionsWithContext(aws.Context, *appstream.DescribeSessionsInput, ...request.Option) (*appstream.DescribeSessionsOutput, error)
	DescribeSessionsRequest(*appstream.DescribeSessionsInput) (*request.Request, *appstream.DescribeSessionsOutput)

	DescribeStacks(*appstream.DescribeStacksInput) (*appstream.DescribeStacksOutput, error)
	DescribeStacksWithContext(aws.Context, *appstream.DescribeStacksInput, ...request.Option) (*appstream.DescribeStacksOutput, error)
	DescribeStacksRequest(*appstream.DescribeStacksInput) (*request.Request, *appstream.DescribeStacksOutput)

	DisassociateFleet(*appstream.DisassociateFleetInput) (*appstream.DisassociateFleetOutput, error)
	DisassociateFleetWithContext(aws.Context, *appstream.DisassociateFleetInput, ...request.Option) (*appstream.DisassociateFleetOutput, error)
	DisassociateFleetRequest(*appstream.DisassociateFleetInput) (*request.Request, *appstream.DisassociateFleetOutput)

	ExpireSession(*appstream.ExpireSessionInput) (*appstream.ExpireSessionOutput, error)
	ExpireSessionWithContext(aws.Context, *appstream.ExpireSessionInput, ...request.Option) (*appstream.ExpireSessionOutput, error)
	ExpireSessionRequest(*appstream.ExpireSessionInput) (*request.Request, *appstream.ExpireSessionOutput)

	ListAssociatedFleets(*appstream.ListAssociatedFleetsInput) (*appstream.ListAssociatedFleetsOutput, error)
	ListAssociatedFleetsWithContext(aws.Context, *appstream.ListAssociatedFleetsInput, ...request.Option) (*appstream.ListAssociatedFleetsOutput, error)
	ListAssociatedFleetsRequest(*appstream.ListAssociatedFleetsInput) (*request.Request, *appstream.ListAssociatedFleetsOutput)

	ListAssociatedStacks(*appstream.ListAssociatedStacksInput) (*appstream.ListAssociatedStacksOutput, error)
	ListAssociatedStacksWithContext(aws.Context, *appstream.ListAssociatedStacksInput, ...request.Option) (*appstream.ListAssociatedStacksOutput, error)
	ListAssociatedStacksRequest(*appstream.ListAssociatedStacksInput) (*request.Request, *appstream.ListAssociatedStacksOutput)

	StartFleet(*appstream.StartFleetInput) (*appstream.StartFleetOutput, error)
	StartFleetWithContext(aws.Context, *appstream.StartFleetInput, ...request.Option) (*appstream.StartFleetOutput, error)
	StartFleetRequest(*appstream.StartFleetInput) (*request.Request, *appstream.StartFleetOutput)

	StopFleet(*appstream.StopFleetInput) (*appstream.StopFleetOutput, error)
	StopFleetWithContext(aws.Context, *appstream.StopFleetInput, ...request.Option) (*appstream.StopFleetOutput, error)
	StopFleetRequest(*appstream.StopFleetInput) (*request.Request, *appstream.StopFleetOutput)

	UpdateFleet(*appstream.UpdateFleetInput) (*appstream.UpdateFleetOutput, error)
	UpdateFleetWithContext(aws.Context, *appstream.UpdateFleetInput, ...request.Option) (*appstream.UpdateFleetOutput, error)
	UpdateFleetRequest(*appstream.UpdateFleetInput) (*request.Request, *appstream.UpdateFleetOutput)

	UpdateStack(*appstream.UpdateStackInput) (*appstream.UpdateStackOutput, error)
	UpdateStackWithContext(aws.Context, *appstream.UpdateStackInput, ...request.Option) (*appstream.UpdateStackOutput, error)
	UpdateStackRequest(*appstream.UpdateStackInput) (*request.Request, *appstream.UpdateStackOutput)

	WaitUntilFleetStarted(*appstream.DescribeFleetsInput) error
	WaitUntilFleetStartedWithContext(aws.Context, *appstream.DescribeFleetsInput, ...request.WaiterOption) error

	WaitUntilFleetStopped(*appstream.DescribeFleetsInput) error
	WaitUntilFleetStoppedWithContext(aws.Context, *appstream.DescribeFleetsInput, ...request.WaiterOption) error
}

var _ AppStreamAPI = (*appstream.AppStream)(nil)
