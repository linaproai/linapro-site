// This file verifies that public, publisher, and reviewer read handlers pass
// the correct authorization scope into the marketplace service boundary.

package market

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/plugin/capability/bizctxcap"
	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	marketplacesvc "linapro-plugin-marketplace/backend/internal/service/marketplace"
)

func TestManagementReadEndpointsDeclareRolePermissions(t *testing.T) {
	tests := []struct {
		name       string
		permission string
		request    any
	}{
		{name: "owned list", permission: "market:plugin:publish", request: marketv1.MyPluginListReq{}},
		{name: "owned detail", permission: "market:plugin:publish", request: marketv1.MyPluginDetailReq{}},
		{name: "managed list", permission: "market:plugin:review", request: marketv1.ManagedPluginListReq{}},
		{name: "managed detail", permission: "market:plugin:review", request: marketv1.ManagedPluginDetailReq{}},
		{name: "review queue", permission: "market:plugin:review", request: marketv1.ReviewQueueListReq{}},
		{name: "review decision", permission: "market:plugin:review", request: marketv1.ReleaseReviewReq{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			requestType := reflect.TypeOf(tc.request)
			metaField, ok := requestType.FieldByName("Meta")
			if !ok {
				t.Fatalf("%s request has no g.Meta field", requestType.Name())
			}
			if got := metaField.Tag.Get("permission"); got != tc.permission {
				t.Fatalf("expected permission %q, got %q", tc.permission, got)
			}
		})
	}
}

// scopeCaptureMarketplaceService records the read scope received by one handler.
type scopeCaptureMarketplaceService struct {
	marketplacesvc.Service
	operation string
	subject   marketplacesvc.VisibilitySubject
}

var _ marketplacesvc.Service = (*scopeCaptureMarketplaceService)(nil)

// writeCaptureMarketplaceService records ownership inputs received by write handlers.
type writeCaptureMarketplaceService struct {
	marketplacesvc.Service
	pluginDraft   marketplacesvc.SavePluginDraftInput
	sourceUpload  marketplacesvc.UploadSourcePackageInput
	dynamicUpload marketplacesvc.UploadDynamicPackageInput
	submitReview  marketplacesvc.SubmitReleaseReviewInput
}

var _ marketplacesvc.Service = (*writeCaptureMarketplaceService)(nil)

// SavePluginDraft captures plugin identity write ownership.
func (s *writeCaptureMarketplaceService) SavePluginDraft(
	_ context.Context,
	in marketplacesvc.SavePluginDraftInput,
) (*marketplacesvc.PluginRecord, error) {
	s.pluginDraft = in
	return &marketplacesvc.PluginRecord{}, nil
}

// UploadSourcePackage captures source-package upload ownership.
func (s *writeCaptureMarketplaceService) UploadSourcePackage(
	_ context.Context,
	in marketplacesvc.UploadSourcePackageInput,
) (*marketplacesvc.SourcePackageUploadResult, error) {
	s.sourceUpload = in
	return &marketplacesvc.SourcePackageUploadResult{Release: &marketplacesvc.ReleaseRecord{}}, nil
}

// UploadDynamicPackage captures dynamic-package upload ownership.
func (s *writeCaptureMarketplaceService) UploadDynamicPackage(
	_ context.Context,
	in marketplacesvc.UploadDynamicPackageInput,
) (*marketplacesvc.DynamicPackageUploadResult, error) {
	s.dynamicUpload = in
	return &marketplacesvc.DynamicPackageUploadResult{Release: &marketplacesvc.ReleaseRecord{}}, nil
}

// SubmitReleaseReview captures review-submission ownership.
func (s *writeCaptureMarketplaceService) SubmitReleaseReview(
	_ context.Context,
	in marketplacesvc.SubmitReleaseReviewInput,
) (*marketplacesvc.ReleaseRecord, error) {
	s.submitReview = in
	return &marketplacesvc.ReleaseRecord{}, nil
}

// GetPluginDetail captures the visibility subject passed by a detail handler.
func (s *scopeCaptureMarketplaceService) GetPluginDetail(
	_ context.Context,
	in marketplacesvc.GetPluginDetailInput,
) (*marketplacesvc.PluginDetailOutput, error) {
	s.operation = "detail"
	s.subject = in.Visibility
	return &marketplacesvc.PluginDetailOutput{}, nil
}

// ListReleases captures the visibility subject passed by a release-list handler.
func (s *scopeCaptureMarketplaceService) ListReleases(
	_ context.Context,
	in marketplacesvc.ListReleasesInput,
) (*marketplacesvc.ReleaseListOutput, error) {
	s.operation = "releases"
	s.subject = in.Visibility
	return &marketplacesvc.ReleaseListOutput{}, nil
}

// GetReleaseDocument captures the visibility subject passed by a document handler.
func (s *scopeCaptureMarketplaceService) GetReleaseDocument(
	_ context.Context,
	in marketplacesvc.GetReleaseDocumentInput,
) (*marketplacesvc.DocumentOutput, error) {
	s.operation = "docs"
	s.subject = in.Visibility
	return &marketplacesvc.DocumentOutput{}, nil
}

// ListReleaseRisks captures the visibility subject passed by a risk-list handler.
func (s *scopeCaptureMarketplaceService) ListReleaseRisks(
	_ context.Context,
	in marketplacesvc.ListReleaseRisksInput,
) (*marketplacesvc.RiskListOutput, error) {
	s.operation = "risks"
	s.subject = in.Visibility
	return &marketplacesvc.RiskListOutput{}, nil
}

// scopeCaptureBizCtx returns one stable request identity for handler tests.
type scopeCaptureBizCtx struct {
	current bizctxcap.CurrentContext
}

// Current returns the configured request identity snapshot.
func (s scopeCaptureBizCtx) Current(context.Context) bizctxcap.CurrentContext {
	return s.current
}

func TestReadHandlersPassExpectedVisibilityScope(t *testing.T) {
	type invokeHandler func(context.Context, *ControllerV1) error

	tests := []struct {
		name       string
		operation  string
		invoke     invokeHandler
		canPublish bool
		canReview  bool
	}{
		{
			name:      "public plugin detail",
			operation: "detail",
			invoke: func(ctx context.Context, controller *ControllerV1) error {
				_, err := controller.PluginDetail(ctx, &marketv1.PluginDetailReq{PluginId: "demo"})
				return err
			},
		},
		{
			name:       "publisher plugin detail",
			operation:  "detail",
			canPublish: true,
			invoke: func(ctx context.Context, controller *ControllerV1) error {
				_, err := controller.MyPluginDetail(ctx, &marketv1.MyPluginDetailReq{PluginId: "demo"})
				return err
			},
		},
		{
			name:      "reviewer plugin detail",
			operation: "detail",
			canReview: true,
			invoke: func(ctx context.Context, controller *ControllerV1) error {
				_, err := controller.ManagedPluginDetail(ctx, &marketv1.ManagedPluginDetailReq{PluginId: "demo"})
				return err
			},
		},
		{
			name:      "public release list",
			operation: "releases",
			invoke: func(ctx context.Context, controller *ControllerV1) error {
				_, err := controller.ReleaseList(ctx, &marketv1.ReleaseListReq{PluginId: "demo"})
				return err
			},
		},
		{
			name:       "publisher release list",
			operation:  "releases",
			canPublish: true,
			invoke: func(ctx context.Context, controller *ControllerV1) error {
				_, err := controller.MyReleaseList(ctx, &marketv1.MyReleaseListReq{PluginId: "demo"})
				return err
			},
		},
		{
			name:      "reviewer release list",
			operation: "releases",
			canReview: true,
			invoke: func(ctx context.Context, controller *ControllerV1) error {
				_, err := controller.ManagedReleaseList(ctx, &marketv1.ManagedReleaseListReq{PluginId: "demo"})
				return err
			},
		},
		{
			name:      "public release docs",
			operation: "docs",
			invoke: func(ctx context.Context, controller *ControllerV1) error {
				_, err := controller.ReleaseDocs(ctx, &marketv1.ReleaseDocsReq{PluginId: "demo", Version: "v1.0.0"})
				return err
			},
		},
		{
			name:       "publisher release docs",
			operation:  "docs",
			canPublish: true,
			invoke: func(ctx context.Context, controller *ControllerV1) error {
				_, err := controller.MyReleaseDocs(ctx, &marketv1.MyReleaseDocsReq{PluginId: "demo", Version: "v1.0.0"})
				return err
			},
		},
		{
			name:      "reviewer release docs",
			operation: "docs",
			canReview: true,
			invoke: func(ctx context.Context, controller *ControllerV1) error {
				_, err := controller.ManagedReleaseDocs(ctx, &marketv1.ManagedReleaseDocsReq{PluginId: "demo", Version: "v1.0.0"})
				return err
			},
		},
		{
			name:      "public release risks",
			operation: "risks",
			invoke: func(ctx context.Context, controller *ControllerV1) error {
				_, err := controller.ReleaseRisks(ctx, &marketv1.ReleaseRisksReq{PluginId: "demo", Version: "v1.0.0"})
				return err
			},
		},
		{
			name:       "publisher release risks",
			operation:  "risks",
			canPublish: true,
			invoke: func(ctx context.Context, controller *ControllerV1) error {
				_, err := controller.MyReleaseRisks(ctx, &marketv1.MyReleaseRisksReq{PluginId: "demo", Version: "v1.0.0"})
				return err
			},
		},
		{
			name:      "reviewer release risks",
			operation: "risks",
			canReview: true,
			invoke: func(ctx context.Context, controller *ControllerV1) error {
				_, err := controller.ManagedReleaseRisks(ctx, &marketv1.ManagedReleaseRisksReq{PluginId: "demo", Version: "v1.0.0"})
				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			capture := &scopeCaptureMarketplaceService{}
			controller := &ControllerV1{
				marketSvc: capture,
				bizCtx: scopeCaptureBizCtx{current: bizctxcap.CurrentContext{
					UserID:   1001,
					TenantID: 2001,
				}},
			}

			if err := tc.invoke(context.Background(), controller); err != nil {
				t.Fatalf("invoke handler: %v", err)
			}
			if capture.operation != tc.operation {
				t.Fatalf("expected %s operation, got %s", tc.operation, capture.operation)
			}
			if capture.subject.UserID != 1001 || capture.subject.TenantID != 2001 {
				t.Fatalf("unexpected identity scope: %#v", capture.subject)
			}
			if capture.subject.CanPublish != tc.canPublish || capture.subject.CanReview != tc.canReview {
				t.Fatalf(
					"expected CanPublish=%t CanReview=%t, got %#v",
					tc.canPublish,
					tc.canReview,
					capture.subject,
				)
			}
		})
	}
}

func TestPublishingWriteHandlersPassCurrentOwnerUserID(t *testing.T) {
	capture := &writeCaptureMarketplaceService{}
	controller := &ControllerV1{
		marketSvc: capture,
		bizCtx: scopeCaptureBizCtx{current: bizctxcap.CurrentContext{
			UserID: 1001,
		}},
	}

	if _, err := controller.PluginCreate(context.Background(), &marketv1.PluginCreateReq{
		PublisherKey: "publisher-a",
		PluginId:     "plugin-a",
	}); err != nil {
		t.Fatalf("create plugin draft: %v", err)
	}
	if capture.pluginDraft.OwnerUserID != 1001 {
		t.Fatalf("plugin draft owner user ID = %d, want 1001", capture.pluginDraft.OwnerUserID)
	}

	if _, err := controller.ReleaseSubmitReview(context.Background(), &marketv1.ReleaseSubmitReviewReq{
		PluginId: "plugin-a",
		Version:  "v1.0.0",
	}); err != nil {
		t.Fatalf("submit release review: %v", err)
	}
	if capture.submitReview.OwnerUserID != 1001 {
		t.Fatalf("review submission owner user ID = %d, want 1001", capture.submitReview.OwnerUserID)
	}

	for _, pluginType := range []marketv1.MarketplacePluginType{
		marketv1.MarketplacePluginTypeSource,
		marketv1.MarketplacePluginTypeDynamic,
	} {
		ctx := newMarketplaceUploadContext(t, "publisher-a")
		if _, err := controller.ReleaseUpload(ctx, &marketv1.ReleaseUploadReq{
			PluginId:   "plugin-a",
			Version:    "v1.0.0",
			PluginType: pluginType,
		}); err != nil {
			t.Fatalf("upload %s release: %v", pluginType, err)
		}
	}
	if capture.sourceUpload.OwnerUserID != 1001 || capture.sourceUpload.PublisherKey != "publisher-a" {
		t.Fatalf("unexpected source upload ownership: %#v", capture.sourceUpload)
	}
	if capture.dynamicUpload.OwnerUserID != 1001 || capture.dynamicUpload.PublisherKey != "publisher-a" {
		t.Fatalf("unexpected dynamic upload ownership: %#v", capture.dynamicUpload)
	}
}

// newMarketplaceUploadContext builds one real multipart request for upload handlers.
func newMarketplaceUploadContext(t *testing.T, publisherKey string) context.Context {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("publisherKey", publisherKey); err != nil {
		t.Fatalf("write publisher key: %v", err)
	}
	part, err := writer.CreateFormFile("file", "plugin.zip")
	if err != nil {
		t.Fatalf("create upload file: %v", err)
	}
	if _, err = part.Write([]byte("package")); err != nil {
		t.Fatalf("write upload file: %v", err)
	}
	if err = writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/market/plugins/plugin-a/releases", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return (&ghttp.Request{Request: request, Server: g.Server("marketplace-upload-handler-test")}).Context()
}
