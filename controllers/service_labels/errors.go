package service_labels

import "fmt"

type NonTenantObject struct {
	objectName string
}

func NewNonTenantObject(objectName string) error {
	return &NonTenantObject{objectName: objectName}
}

func (n NonTenantObject) Error() string {
	return fmt.Sprintf("Skipping labels sync for %s as it doesn't belong to tenant", n.objectName)
}

type NoServicesMetadata struct {
	objectName string
}

func NewNoServicesMetadata(objectName string) error {
	return &NoServicesMetadata{objectName: objectName}
}

func (n NoServicesMetadata) Error() string {
	return fmt.Sprintf("Skipping labels sync for %s because no AdditionalLabels or AdditionalAnnotations presents in Tenant spec", n.objectName)
}
