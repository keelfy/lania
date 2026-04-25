package domain

type AvatarSize string

const (
	AvatarSizeSmall  AvatarSize = "sm"
	AvatarSizeMedium AvatarSize = "md"
	AvatarSizeLarge  AvatarSize = "lg"
)

var AvatarSizes = map[AvatarSize]int{AvatarSizeSmall: 32, AvatarSizeMedium: 64, AvatarSizeLarge: 128}
