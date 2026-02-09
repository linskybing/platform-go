package image

// CraneCopyArgs builds the argument slice to run `crane copy` from source to target.
func CraneCopyArgs(src, dst string) []string {
	return []string{"crane", "copy", src, dst, "--insecure"}
}
