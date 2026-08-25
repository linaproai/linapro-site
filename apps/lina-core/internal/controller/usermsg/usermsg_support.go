// This file rewrites the current-user identity and inbox category labels for
// the user-message HTTP surface. Data access and unauthenticated errors live
// in notify.

package usermsg

import (
	"context"
)

// Stable i18n key convention used to localize inbox category labels and tag
// colors. The host does not enumerate specific category codes here; senders
// register translations at `notify.category.{code}.label` and
// `notify.category.{code}.color`.
const (
	// usermsgCategoryI18nNamespace is the parent i18n namespace shared by all category labels and colors.
	usermsgCategoryI18nNamespace = "notify.category."
	// usermsgCategoryLabelI18nSuffix is the i18n key suffix that resolves the category display label.
	usermsgCategoryLabelI18nSuffix = ".label"
	// usermsgCategoryColorI18nSuffix is the i18n key suffix that resolves the category tag color.
	usermsgCategoryColorI18nSuffix = ".color"
	// usermsgCategoryFallbackCode is the canonical fallback category used when a message has no declared category code.
	usermsgCategoryFallbackCode = "other"
	// usermsgCategoryDefaultColor is the safety color rendered when no category color resource is configured.
	usermsgCategoryDefaultColor = "default"
	// usermsgCategoryDefaultLabel is the safety label rendered when no category label resource is configured.
	usermsgCategoryDefaultLabel = "Notification"
)

// currentUserID returns the authenticated user identifier from request context.
// Missing identity returns 0 so notify inbox methods can reject the call with
// CodeNotifyNotAuthenticated.
func (c *ControllerV1) currentUserID(ctx context.Context) int64 {
	if c == nil || c.bizCtxSvc == nil {
		return 0
	}
	bizCtx := c.bizCtxSvc.Get(ctx)
	if bizCtx == nil {
		return 0
	}
	return int64(bizCtx.UserId)
}

// resolveCategoryCode normalizes an inbox message category code, falling back
// to the canonical "other" bucket when the sender did not declare one.
func resolveCategoryCode(categoryCode string) string {
	if categoryCode == "" {
		return usermsgCategoryFallbackCode
	}
	return categoryCode
}

// localizeCategoryLabel resolves the localized category display label for the
// given category code. Translation is looked up at
// `notify.category.{code}.label`. If the requested code has no translation
// resource, it falls back to the canonical "other" bucket and finally to a
// safety literal so the inbox never renders an empty category cell.
func (c *ControllerV1) localizeCategoryLabel(ctx context.Context, categoryCode string) string {
	if c == nil || c.i18nSvc == nil {
		return usermsgCategoryDefaultLabel
	}
	code := resolveCategoryCode(categoryCode)
	key := usermsgCategoryI18nNamespace + code + usermsgCategoryLabelI18nSuffix
	if label := c.i18nSvc.Translate(ctx, key, ""); label != "" {
		return label
	}
	if code != usermsgCategoryFallbackCode {
		fallbackKey := usermsgCategoryI18nNamespace + usermsgCategoryFallbackCode + usermsgCategoryLabelI18nSuffix
		if label := c.i18nSvc.Translate(ctx, fallbackKey, ""); label != "" {
			return label
		}
	}
	return usermsgCategoryDefaultLabel
}

// localizeCategoryColor resolves the localized category tag color for the
// given category code. Color is treated as a localizable display attribute;
// the resolution path mirrors localizeCategoryLabel and falls back to a
// generic neutral color.
func (c *ControllerV1) localizeCategoryColor(ctx context.Context, categoryCode string) string {
	if c == nil || c.i18nSvc == nil {
		return usermsgCategoryDefaultColor
	}
	code := resolveCategoryCode(categoryCode)
	key := usermsgCategoryI18nNamespace + code + usermsgCategoryColorI18nSuffix
	if color := c.i18nSvc.Translate(ctx, key, ""); color != "" {
		return color
	}
	if code != usermsgCategoryFallbackCode {
		fallbackKey := usermsgCategoryI18nNamespace + usermsgCategoryFallbackCode + usermsgCategoryColorI18nSuffix
		if color := c.i18nSvc.Translate(ctx, fallbackKey, ""); color != "" {
			return color
		}
	}
	return usermsgCategoryDefaultColor
}
