package kubernetes

import (
	"context"
	"fmt"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ExtensionClient struct {
	client apiextensionsclient.Interface
}

func NewExtensionClient(client apiextensionsclient.Interface) *ExtensionClient {
	return &ExtensionClient{
		client: client,
	}
}

func (ec *ExtensionClient) DeleteCRD(ctx context.Context, crdName string) error {
	if ec == nil || ec.client == nil {
		return fmt.Errorf("extension client is not initialized")
	}

	err := ec.client.
		ApiextensionsV1().
		CustomResourceDefinitions().
		Delete(ctx, crdName, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}
