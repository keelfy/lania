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

func GetDatabaseFlectoneName() string {
	return os.Getenv("DATABASE_FLECTONE_NAME")
}

func GetDatabasePlanName() string {
	return os.Getenv("DATABASE_PLAN_NAME")
}

func GetLuckpermsUserPermissionsTableName() string {
	tableName := os.Getenv("LUCKPERMS_USER_PERMISSIONS_TABLE_NAME")
	if tableName == "" {
		return "luckperms_user_permissions"
	}
	return tableName
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

func GetActiveSeasonID() uuid.UUID {
	return uuid.MustParse(os.Getenv("ACTIVE_SEASON_ID"))
}

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

func IsPreRegistrationEnabled() bool {
	return os.Getenv("PREREGISTRATION") == "true"
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

func GetEasyDonateShopKeyHeader() string {
	return os.Getenv("ED_SHOP_KEY_HEADER")
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
