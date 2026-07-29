// Package backend tests verify the marketplace source-plugin static
// registration and embedded asset binding remain discoverable by the host.

package backend

import (
	"context"
	"io/fs"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/pluginhost"
	marketplace "linapro-plugin-marketplace"
	"testing"
)

// TestSourcePluginRegistration verifies the compile-time source-plugin
// declaration uses the same plugin ID as the manifest and exposes embedded
// plugin resources to the host scanner.
func TestSourcePluginRegistration(t *testing.T) {
	t.Parallel()

	definition, ok := pluginhost.GetSourcePlugin(marketplace.PluginID)
	if !ok {
		t.Fatalf("expected source plugin %s to be registered", marketplace.PluginID)
	}
	if definition.ID() != marketplace.PluginID {
		t.Fatalf("expected registered plugin ID %q, got %q", marketplace.PluginID, definition.ID())
	}
	if _, err := fs.Stat(definition.GetEmbeddedFiles(), "plugin.yaml"); err != nil {
		t.Fatalf("expected embedded plugin.yaml to be readable: %v", err)
	}
	if len(definition.GetRouteRegistrars()) == 0 {
		t.Fatal("expected marketplace HTTP route registrar to be registered")
	}
}

func TestRegisterRejectsNilDeclarations(t *testing.T) {
	t.Parallel()

	if err := Register(nil); err == nil {
		t.Fatal("expected nil declarations to be rejected")
	}
}

// stubJobsRegistrar captures JobSpec registrations for unit tests.
type stubJobsRegistrar struct {
	specs       []pluginhost.JobSpec
	primaryNode bool
}

func (s *stubJobsRegistrar) Add(_ context.Context, spec pluginhost.JobSpec) error {
	s.specs = append(s.specs, spec)
	return nil
}

func (s *stubJobsRegistrar) IsPrimaryNode() bool { return s.primaryNode }

func (s *stubJobsRegistrar) Services() capability.Services { return nil }

// TestRegisterMarketplaceJobsProcessPipelinePolicy verifies the process
// pipeline runs every 10s as a master-only singleton job.
func TestRegisterMarketplaceJobsProcessPipelinePolicy(t *testing.T) {
	t.Parallel()

	stub := &stubJobsRegistrar{primaryNode: true}
	if err := registerMarketplaceJobs(context.Background(), stub); err != nil {
		t.Fatalf("registerMarketplaceJobs: %v", err)
	}
	if len(stub.specs) < 1 {
		t.Fatal("expected at least process-pipeline job registration")
	}

	var found *pluginhost.JobSpec
	for i := range stub.specs {
		if stub.specs[i].Name == "linapro-plugin-marketplace-process-pipeline" {
			found = &stub.specs[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("process pipeline job not registered, specs=%+v", stub.specs)
	}
	if found.Pattern != "@every 10s" {
		t.Fatalf("pattern=%q want @every 10s", found.Pattern)
	}
	if found.Scope != pluginhost.JobScopeMasterOnly {
		t.Fatalf("scope=%q want %q", found.Scope, pluginhost.JobScopeMasterOnly)
	}
	if found.Concurrency != pluginhost.JobConcurrencySingleton {
		t.Fatalf("concurrency=%q want %q", found.Concurrency, pluginhost.JobConcurrencySingleton)
	}
	if found.Handler == nil {
		t.Fatal("handler is required")
	}
}
