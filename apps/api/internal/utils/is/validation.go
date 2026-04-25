package is

import (
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/lania-smp/backend/internal/domain"
)

var IsRFC3339Date = validation.Date(time.RFC3339)

var IsAvatarSize = validation.In(
	domain.AvatarSizeSmall,
	domain.AvatarSizeMedium,
	domain.AvatarSizeLarge,
)

var AccessSource = validation.In(
	domain.AccessSourceFree,
	domain.AccessSourceFreekassa,
)

var MinecraftUsername = validation.Match(regexp.MustCompile(`^[a-zA-Z0-9_]{3,16}$`))

var PaymentMethod = validation.In(
	domain.PaymentMethodFreekassa,
	domain.PaymentMethodDonationAlerts,
	domain.PaymentMethodEasyDonate,
)
