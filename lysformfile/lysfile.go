package lysformfile

// ImageMimeTypes is a predefined list of common image MIME types.
// These MIME types have decoder registrations in the Go standard library, allowing dimension validation in ExtractFromRequest.
var ImageMimeTypes = []string{
	"image/gif",
	"image/jpeg",
	"image/png",
}
