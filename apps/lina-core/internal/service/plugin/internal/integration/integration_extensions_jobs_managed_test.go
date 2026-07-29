// This file white-box tests managedJobCollector.Add JobSpec policy handling
// in integration_extensions_jobs_managed.go.

package integration

import (
	"context"
	jobv1 "lina-core/api/job/v1"
	"lina-core/pkg/plugin/pluginhost"
	"testing"
)

// TestManagedJobCollectorAddPreservesScopeAndConcurrency ensures master_only +
// singleton declarations survive collection for multi-node scheduling.
func TestManagedJobCollectorAddPreservesScopeAndConcurrency(t *testing.T) {
	t.Parallel()

	collector := &managedJobCollector{
		pluginID: "linapro-plugin-marketplace",
		items:    make([]ManagedJob, 0, 1),
	}
	err := collector.Add(context.Background(), pluginhost.JobSpec{
		Pattern:     "@every 10s",
		Name:        "linapro-plugin-marketplace-process-pipeline",
		DisplayName: "Marketplace process pipeline",
		Description: "Discover, verify, and auto-submit marketplace plugins",
		Scope:       pluginhost.JobScopeMasterOnly,
		Concurrency: pluginhost.JobConcurrencySingleton,
		Handler:     func(context.Context) error { return nil },
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if len(collector.items) != 1 {
		t.Fatalf("expected 1 job, got %d", len(collector.items))
	}
	item := collector.items[0]
	if item.Pattern != "@every 10s" {
		t.Fatalf("pattern=%q", item.Pattern)
	}
	if item.Scope != jobv1.ScopeMasterOnly {
		t.Fatalf("scope=%q want %q", item.Scope, jobv1.ScopeMasterOnly)
	}
	if item.Concurrency != jobv1.ConcurrencySingleton {
		t.Fatalf("concurrency=%q want %q", item.Concurrency, jobv1.ConcurrencySingleton)
	}
	if item.MaxConcurrency != 1 {
		t.Fatalf("maxConcurrency=%d", item.MaxConcurrency)
	}
}

// TestManagedJobCollectorAddLeavesOptionalPolicyEmpty keeps host defaults when
// Scope and Concurrency are omitted from JobSpec.
func TestManagedJobCollectorAddLeavesOptionalPolicyEmpty(t *testing.T) {
	t.Parallel()

	collector := &managedJobCollector{
		pluginID: "demo-plugin",
		items:    make([]ManagedJob, 0, 1),
	}
	if err := collector.Add(context.Background(), pluginhost.JobSpec{
		Pattern:     "0 * * * * *",
		Name:        "demo-job",
		DisplayName: "Demo Job",
		Description: "Demo description",
		Handler:     func(context.Context) error { return nil },
	}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if len(collector.items) != 1 {
		t.Fatalf("expected 1 job, got %d", len(collector.items))
	}
	if collector.items[0].Scope != "" {
		t.Fatalf("scope should stay empty for host default, got %q", collector.items[0].Scope)
	}
	if collector.items[0].Concurrency != "" {
		t.Fatalf("concurrency should stay empty for host default, got %q", collector.items[0].Concurrency)
	}
}

// TestManagedJobCollectorAddRejectsMissingRequiredFields covers Pattern/Name/Handler.
func TestManagedJobCollectorAddRejectsMissingRequiredFields(t *testing.T) {
	t.Parallel()

	collector := &managedJobCollector{pluginID: "demo-plugin", items: make([]ManagedJob, 0)}
	handler := func(context.Context) error { return nil }

	if err := collector.Add(context.Background(), pluginhost.JobSpec{
		Name:    "demo",
		Handler: handler,
	}); err == nil {
		t.Fatal("expected empty Pattern to fail")
	}
	if err := collector.Add(context.Background(), pluginhost.JobSpec{
		Pattern: "@every 1s",
		Handler: handler,
	}); err == nil {
		t.Fatal("expected empty Name to fail")
	}
	if err := collector.Add(context.Background(), pluginhost.JobSpec{
		Pattern: "@every 1s",
		Name:    "demo",
	}); err == nil {
		t.Fatal("expected nil Handler to fail")
	}
}

// TestManagedJobCollectorAddRejectsInvalidEnums covers typed scope/concurrency.
func TestManagedJobCollectorAddRejectsInvalidEnums(t *testing.T) {
	t.Parallel()

	collector := &managedJobCollector{pluginID: "demo-plugin", items: make([]ManagedJob, 0)}
	handler := func(context.Context) error { return nil }

	if err := collector.Add(context.Background(), pluginhost.JobSpec{
		Pattern: "@every 1s",
		Name:    "demo",
		Scope:   pluginhost.JobScope("not-a-scope"),
		Handler: handler,
	}); err == nil {
		t.Fatal("expected invalid Scope to fail")
	}
	if err := collector.Add(context.Background(), pluginhost.JobSpec{
		Pattern:     "@every 1s",
		Name:        "demo",
		Concurrency: pluginhost.JobConcurrency("not-a-mode"),
		Handler:     handler,
	}); err == nil {
		t.Fatal("expected invalid Concurrency to fail")
	}
}
