package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	NotebookInstanceInServiceTimeout = 10 * time.Minute
	NotebookInstanceStoppedTimeout   = 10 * time.Minute
	NotebookInstanceDeletedTimeout   = 10 * time.Minute
	ImageCreatedTimeout              = 10 * time.Minute
	ImageDeletedTimeout              = 10 * time.Minute
)

// NotebookInstanceInService waits for a NotebookInstance to return InService
func NotebookInstanceInService(conn *sagemaker.SageMaker, notebookName string) (*sagemaker.DescribeNotebookInstanceOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			SagemakerNotebookInstanceStatusNotFound,
			sagemaker.NotebookInstanceStatusUpdating,
			sagemaker.NotebookInstanceStatusPending,
			sagemaker.NotebookInstanceStatusStopped,
		},
		Target:  []string{sagemaker.NotebookInstanceStatusInService},
		Refresh: NotebookInstanceStatus(conn, notebookName),
		Timeout: NotebookInstanceInServiceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeNotebookInstanceOutput); ok {
		return output, err
	}

	return nil, err
}

// NotebookInstanceStopped waits for a NotebookInstance to return Stopped
func NotebookInstanceStopped(conn *sagemaker.SageMaker, notebookName string) (*sagemaker.DescribeNotebookInstanceOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.NotebookInstanceStatusUpdating,
			sagemaker.NotebookInstanceStatusStopping,
		},
		Target:  []string{sagemaker.NotebookInstanceStatusStopped},
		Refresh: NotebookInstanceStatus(conn, notebookName),
		Timeout: NotebookInstanceStoppedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeNotebookInstanceOutput); ok {
		return output, err
	}

	return nil, err
}

// NotebookInstanceDeleted waits for a NotebookInstance to return Deleted
func NotebookInstanceDeleted(conn *sagemaker.SageMaker, notebookName string) (*sagemaker.DescribeNotebookInstanceOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.NotebookInstanceStatusDeleting,
		},
		Target:  []string{},
		Refresh: NotebookInstanceStatus(conn, notebookName),
		Timeout: NotebookInstanceDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeNotebookInstanceOutput); ok {
		return output, err
	}

	return nil, err
}

// ImageCreated waits for a Image to return Created
func ImageCreated(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeImageOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.ImageStatusCreating,
			sagemaker.ImageStatusUpdating,
		},
		Target:  []string{sagemaker.ImageStatusCreated},
		Refresh: ImageStatus(conn, name),
		Timeout: ImageCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeImageOutput); ok {
		return output, err
	}

	return nil, err
}

// ImageDeleted waits for a Image to return Deleted
func ImageDeleted(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeImageOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{sagemaker.ImageStatusDeleting},
		Target:  []string{},
		Refresh: ImageStatus(conn, name),
		Timeout: ImageDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeImageOutput); ok {
		return output, err
	}

	return nil, err
}
