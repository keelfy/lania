package domain

// UploadedImage is the result of putting an admin-uploaded image into the project's S3 bucket.
// Location is an s3://bucket/key string, the same form every image field in the catalog stores.
type UploadedImage struct {
	Location string
	Width    int
	Height   int
}
