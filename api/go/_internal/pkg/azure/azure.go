package azure

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/huangc28/kikichoice-be/api/go/_internal/configs"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

const (
	ProductImageContainerName = "products"
)

func NewSharedKeyCredential(cfg *configs.Config) (*azblob.SharedKeyCredential, error) {
	cred, err := azblob.NewSharedKeyCredential(cfg.Azure.BlobStorageAccountName, cfg.Azure.BlobStorageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create shared key credential: %v", err)
	}
	return cred, nil
}

func NewBlobStorageClient(cfg *configs.Config, cred *azblob.SharedKeyCredential) (*azblob.Client, error) {
	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", cfg.Azure.BlobStorageAccountName)

	log.Printf("serviceURL: %+v", cfg.Azure)
	log.Println("serviceURL", serviceURL)

	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	return client, nil
}

type BlobStorageWrapperClient struct {
	Client             *azblob.Client
	StorageAccountName string
}

func NewBlobStorageWrapperClient(cfg *configs.Config, client *azblob.Client) (*BlobStorageWrapperClient, error) {
	return &BlobStorageWrapperClient{
		Client:             client,
		StorageAccountName: cfg.Azure.BlobStorageAccountName,
	}, nil
}

func (c *BlobStorageWrapperClient) UploadProductImage(ctx context.Context, blobName string, contentReader io.Reader) (string, error) {
	_, err := c.Client.UploadStream(
		ctx,
		ProductImageContainerName,
		blobName,
		contentReader,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to Azure blob storage: %v", err)
	}

	return c.GetPublicURL(ProductImageContainerName, blobName), nil
}

func (c *BlobStorageWrapperClient) GetPublicURL(containerName, blobName string) string {
	return fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s",
		c.StorageAccountName,
		containerName,
		blobName,
	)
}
