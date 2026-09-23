package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

/** SYSTEM */

func GetPort() string {
	return os.Getenv("PORT")
}

func IsDebug() bool {
	return os.Getenv("DEBUG") == "true"
}

/** API */

func GetApiKey() string {
	return os.Getenv("API_KEY")
}

func GetContextTimeoutMs() time.Duration {
	value, err := strconv.Atoi(os.Getenv("CONTEXT_TIMEOUT_MS"))
	if err != nil {
		log.Printf("Error parsing CONTEXT_TIMEOUT_MS: %v", err)
		return 1000 * 60 * time.Millisecond
	}
	return time.Duration(value) * time.Millisecond
}

/** Ory */

func GetOryUrl() string {
	return os.Getenv("ORY_URL")
}

// GetOryAdminUrl returns the Kratos admin API address. It must never be reachable from outside the private network.
func GetOryAdminUrl() string {
	return os.Getenv("ORY_ADMIN_URL")
}

/** JWT */

func GetJWTSecret() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

/** DATABASE */

func GetDatabaseHost() string {
	return os.Getenv("DATABASE_HOST")
}

func GetDatabasePort() string {
	return os.Getenv("DATABASE_PORT")
}

func GetDatabaseUser() string {
	return os.Getenv("DATABASE_USER")
}

func GetDatabasePassword() string {
	return os.Getenv("DATABASE_PASSWORD")
}

func GetDatabaseName() string {
	return os.Getenv("DATABASE_NAME")
}

/** SHELL */

// GetShellAddress returns host:port of the shell gRPC service of the primary season.
// It only seeds the season setting once, the address is kept in the seasons table afterwards.
func GetShellAddress() string {
	return os.Getenv("SHELL_ADDRESS")
}

func GetShellToken() string {
	return os.Getenv("SHELL_TOKEN")
}

/** ROLE SYNC */

const defaultRoleSyncWindowMinutes = 10

// GetRoleSyncWindow returns how far back role sync looks for changed roles.
// A change is pushed to the shells on every run within this window, so a shell that was down for a shorter time catches up.
func GetRoleSyncWindow() time.Duration {
	value := os.Getenv("ROLE_SYNC_WINDOW_MINUTES")
	if value == "" {
		return defaultRoleSyncWindowMinutes * time.Minute
	}
	minutes, err := strconv.Atoi(value)
	if err != nil || minutes <= 0 {
		log.Printf("Error parsing ROLE_SYNC_WINDOW_MINUTES: %q is not a positive number", value)
		return defaultRoleSyncWindowMinutes * time.Minute
	}
	return time.Duration(minutes) * time.Minute
}

/** REDIS */

func GetRedisURL() string {
	return os.Getenv("REDIS_URL")
}

/** ELASTICSEARCH */

func GetElasticsearchUrls() []string {
	return strings.Split(os.Getenv("ELASTICSEARCH_URLS"), ";")
}

func GetElasticsearchUsername() string {
	return os.Getenv("ELASTICSEARCH_USERNAME")
}

func GetElasticsearchPassword() string {
	return os.Getenv("ELASTICSEARCH_PASSWORD")
}

/** imgproxy */

func GetImgProxyUrl() string {
	return os.Getenv("IMGPROXY_URL")
}

func GetImgProxyKey() string {
	return os.Getenv("IMGPROXY_KEY")
}

func GetImgProxySalt() string {
	return os.Getenv("IMGPROXY_SALT")
}

/** S3 */

func GetS3Bucket() string {
	return os.Getenv("S3_BUCKET")
}

func GetS3Region() string {
	return os.Getenv("S3_REGION")
}

// GetS3Endpoint returns the object storage endpoint. Empty means the AWS default endpoint.
func GetS3Endpoint() string {
	return os.Getenv("S3_ENDPOINT")
}

// IsS3PathStyleForced reports whether the client must address the bucket as part of the URL path.
// Most non-AWS S3-compatible providers need this.
func IsS3PathStyleForced() bool {
	return os.Getenv("S3_FORCE_PATH_STYLE") == "true"
}

/** CORS */

func GetCorsAllowedOrigins() []string {
	return strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ";")
}

/** Twitch */

func GetTwitchClientID() string {
	return os.Getenv("TWITCH_CLIENT_ID")
}

func GetTwitchClientSecret() string {
	return os.Getenv("TWITCH_CLIENT_SECRET")
}

func GetIGDBImageURLFormat() string {
	return os.Getenv("IGDB_IMAGE_URL_FORMAT")
}

/** Business Constraints */

func GetMaxProfilesPerUser() int {
	value, err := strconv.Atoi(os.Getenv("MAX_PROFILES_PER_USER"))
	if err != nil {
		log.Printf("Error parsing MAX_PROFILES_PER_USER: %v", err)
		return 1
	}
	return value
}

func GetDefaultNameColorID() uuid.UUID {
	return uuid.MustParse(os.Getenv("DEFAULT_NAME_COLOR_ID"))
}

/** Freekassa */

func GetFreekassaBasePaymentURL() string {
	return os.Getenv("FREEKASSA_BASE_PAYMENT_URL")
}

func GetFreekassaMerchantID() int64 {
	value, err := strconv.ParseInt(os.Getenv("FREEKASSA_MERCHANT_ID"), 10, 64)
	if err != nil {
		log.Printf("Error parsing FREEKASSA_MERCHANT_ID: %v", err)
		return 0
	}
	return value
}

func GetFreekassaMerchantPassword1() string {
	return os.Getenv("FREEKASSA_MERCHANT_PASSWORD_1")
}

func GetFreekassaMerchantPassword2() string {
	return os.Getenv("FREEKASSA_MERCHANT_PASSWORD_2")
}

/** Donation Alerts */

func GetDonationAlertsClientID() string {
	return os.Getenv("DONATION_ALERTS_CLIENT_ID")
}

func GetDonationAlertsClientSecret() string {
	return os.Getenv("DONATION_ALERTS_CLIENT_SECRET")
}

func GetDonationAlertsScope() string {
	return os.Getenv("DONATION_ALERTS_SCOPE")
}

func GetDonationAlertSocketConnectionToken() string {
	return os.Getenv("DONATION_ALERTS_SOCKET_CONNECTION_TOKEN")
}

func GetDonationAlertsUserID() int64 {
	value, err := strconv.ParseInt(os.Getenv("DONATION_ALERTS_USER_ID"), 10, 64)
	if err != nil {
		log.Printf("Error parsing DONATION_ALERTS_USER_ID: %v", err)
		return 0
	}
	return value
}

/** Easy Donate */

func GetEasyDonateCreatePaymentEndpoint() string {
	return os.Getenv("ED_CREATE_PAYMENT_ENDPOINT")
}

func GetEasyDonateProxyServerID() int64 {
	value, err := strconv.ParseInt(os.Getenv("ED_PROXY_SERVER_ID"), 10, 64)
	if err != nil {
		log.Printf("Error parsing ED_PROXY_SERVER_ID: %v", err)
		return 0
	}
	return value
}

func GetEasyDonateKey() string {
	return os.Getenv("ED_KEY")
}

func GetEasyDonateSuccessURL() string {
	return os.Getenv("ED_SUCCESS_URL")
}

func IsEasyDonateSignatureVerificationSkipped() bool {
	return os.Getenv("ED_SKIP_SIGNATURE_VERIFICATION") == "true"
}
