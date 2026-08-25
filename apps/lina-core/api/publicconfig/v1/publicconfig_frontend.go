// This file defines the public frontend-config API DTOs exposed to unauthenticated clients.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ThemeMode identifies the default frontend theme preference.
type ThemeMode string

const (
	// ThemeModeLight selects the light theme.
	ThemeModeLight ThemeMode = "light"
	// ThemeModeDark selects the dark theme.
	ThemeModeDark ThemeMode = "dark"
	// ThemeModeAuto follows the client system theme.
	ThemeModeAuto ThemeMode = "auto"
)

// CronLogRetentionMode identifies the system-level job-log retention policy.
type CronLogRetentionMode string

const (
	// CronLogRetentionModeDays removes logs older than the configured day count.
	CronLogRetentionModeDays CronLogRetentionMode = "days"
	// CronLogRetentionModeCount keeps only the configured number of latest logs.
	CronLogRetentionModeCount CronLogRetentionMode = "count"
	// CronLogRetentionModeNone disables automatic cleanup.
	CronLogRetentionModeNone CronLogRetentionMode = "none"
)

// FrontendReq defines the request for fetching public frontend config.
type FrontendReq struct {
	g.Meta `path:"/config/public/frontend" method:"get" tags:"Public Configuration" summary:"Get public frontend configuration" dc:"Return the unauthenticated brand, login copy, feature switches, theme, and workspace path that any management workbench can read. The payload is not a Vben layout contract."`
}

// FrontendRes defines the public frontend config response.
type FrontendRes struct {
	App       FrontendAppRes       `json:"app" dc:"Apply brand display configuration" eg:"{}"`
	Auth      FrontendAuthRes      `json:"auth" dc:"Login page display configuration" eg:"{}"`
	User      FrontendUserRes      `json:"user" dc:"User display configuration" eg:"{}"`
	UI        FrontendUIRes        `json:"ui" dc:"UI style configuration" eg:"{}"`
	Cron      FrontendCronRes      `json:"cron" dc:"Scheduled job frontend capability configuration" eg:"{}"`
	Workspace FrontendWorkspaceRes `json:"workspace" dc:"Admin workspace startup routing configuration" eg:"{}"`
}

// FrontendAppRes stores brand-related public settings.
type FrontendAppRes struct {
	Name     string `json:"name" dc:"Application name, used for browser title and workbench logo copywriting" eg:"LinaPro.AI"`
	Logo     string `json:"logo" dc:"Default logo image address" eg:"/logo.webp"`
	LogoDark string `json:"logoDark" dc:"Dark theme logo image address" eg:"/logo.webp"`
}

// FrontendAuthRes stores login-page public copy and entry-switch settings.
type FrontendAuthRes struct {
	PageTitle             string `json:"pageTitle" dc:"Login page main title copywriting" eg:"An AI-native full-stack framework engineered for sustainable delivery"`
	PageDesc              string `json:"pageDesc" dc:"Login page description copy" eg:"Facing business evolution, it provides out-of-the-box management portals and flexible pluggable expansion mechanisms."`
	LoginSubtitle         string `json:"loginSubtitle" dc:"Login form subtitle copywriting" eg:"Please enter your account information to enter the LinaPro hosting workspace"`
	ForgetPasswordEnabled bool   `json:"forgetPasswordEnabled" dc:"Whether the login page exposes the forget-password entry: true=show false=hide" eg:"true"`
	RegisterEnabled       bool   `json:"registerEnabled" dc:"Whether the login page exposes the create-account entry: true=show false=hide" eg:"true"`
	PrivacyPolicy         string `json:"privacyPolicy" dc:"Privacy policy body shown in the registration consent dialog" eg:"Privacy Policy\n\n..."`
	TermsOfService        string `json:"termsOfService" dc:"Terms of service body shown in the registration consent dialog" eg:"Terms of Service\n\n..."`
}

// FrontendUserRes stores user-facing public fallback settings.
type FrontendUserRes struct {
	DefaultAvatar string `json:"defaultAvatar" dc:"The default avatar address used when the user does not set an avatar" eg:"/avatar.webp"`
}

// FrontendUIRes stores public-safe theme and layout preferences.
type FrontendUIRes struct {
	ThemeMode        ThemeMode `json:"themeMode" dc:"Theme mode: light=light dark=dark auto=follow the system" eg:"light"`
	WatermarkEnabled bool      `json:"watermarkEnabled" dc:"Whether to enable watermark: true=enable false=disable" eg:"false"`
	WatermarkContent string    `json:"watermarkContent" dc:"Watermark copy content" eg:"LinaPro"`
}

// FrontendCronRes stores public-safe scheduled-job capability flags.
type FrontendCronRes struct {
	LogRetention FrontendCronLogRetentionRes `json:"logRetention" dc:"System-level scheduled job log retention policy" eg:"{}"`
	Shell        FrontendCronShellRes        `json:"shell" dc:"Shell task frontend capability switch" eg:"{}"`
	Timezone     FrontendCronTimezoneRes     `json:"timezone" dc:"Default time zone configuration for scheduled jobs" eg:"{}"`
}

// FrontendCronLogRetentionRes stores the frontend-visible default log-retention policy.
type FrontendCronLogRetentionRes struct {
	Mode  CronLogRetentionMode `json:"mode" dc:"System-level log retention mode: days=retain by day count=retain by number of entries none=do not automatically clean up" eg:"days"`
	Value int64                `json:"value" dc:"System-level log retention threshold; greater than 0 when mode=days or count, 0 when mode=none" eg:"30"`
}

// FrontendCronShellRes stores the shell-job gate exposed to the frontend.
type FrontendCronShellRes struct {
	Enabled           bool   `json:"enabled" dc:"Whether to allow creation and execution of Shell tasks: true=allowed false=not allowed" eg:"false"`
	Supported         bool   `json:"supported" dc:"Whether the current platform supports Shell tasks: true=supported false=not supported" eg:"true"`
	DisabledReason    string `json:"disabledReason" dc:"English fallback explaining why the Shell task is unavailable" eg:"The current platform does not support shell mode"`
	DisabledReasonKey string `json:"disabledReasonKey" dc:"Runtime i18n key for localizing disabledReason on the frontend" eg:"config.cron.shell.unsupportedReason"`
}

// FrontendCronTimezoneRes stores the default timezone exposed to the frontend.
type FrontendCronTimezoneRes struct {
	Current string `json:"current" dc:"The current host system time zone identifier, used as the default value for the new task time zone" eg:"Asia/Shanghai"`
}

// FrontendWorkspaceRes stores public-safe admin workspace routing settings.
type FrontendWorkspaceRes struct {
	BasePath string `json:"basePath" dc:"Admin workspace entry path loaded from startup configuration; may be / for a dedicated admin domain" eg:"/admin"`
}
