package endpoints

import (
	"context"
	"fmt"
	"os"

	"github.com/krateoplatformops/plumbing/endpoints"
	"k8s.io/client-go/rest"

	finopsdatatypes "github.com/krateoplatformops/finops-data-types/api/v1"
)

func FromSecret(ctx context.Context, rc *rest.Config, ref *finopsdatatypes.ObjectRef) (endpoints.Endpoint, error) {
	if ref == nil {
		tokenData, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
		if err != nil {
			return endpoints.Endpoint{}, fmt.Errorf("there has been an error reading the cert-file: %w", err)
		}
		certData, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
		if err != nil {
			return endpoints.Endpoint{}, fmt.Errorf("there has been an error reading the cert-file: %w", err)
		}
		return endpoints.Endpoint{
			ServerURL:                "https://kubernetes.default.svc",
			Token:                    string(tokenData),
			CertificateAuthorityData: string(certData),
			Insecure:                 true,
		}, nil

	} else {
		return endpoints.FromSecret(ctx, rc, ref.Name, ref.Namespace)
	}
}
