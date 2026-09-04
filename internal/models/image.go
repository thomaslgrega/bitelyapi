package models

import "strings"

// ImageLocator turns the stored key for a Recipe Image into the URL a client
// fetches it from (ADR-0006). Rows hold keys rather than URLs, so the bucket's
// public hostname stays configuration and a move costs an environment variable
// instead of a table rewrite.
type ImageLocator struct {
	baseURL string
}

func NewImageLocator(baseURL string) ImageLocator {
	return ImageLocator{baseURL: strings.TrimSuffix(baseURL, "/")}
}

// URLFor answers "" for a Recipe with no image, so an imageless Recipe carries
// no image_url at all rather than a URL pointing at the bare bucket.
func (l ImageLocator) URLFor(key string) string {
	if key == "" {
		return ""
	}

	return l.baseURL + "/" + key
}

// resolveImage composes the URL a response carries from the key a row holds.
func resolveImage(key *string, url *string, locator ImageLocator) {
	*url = locator.URLFor(*key)
}

func (r *Recipe) ResolveImage(locator ImageLocator) {
	resolveImage(&r.ImageKey, &r.ImageUrl, locator)
}

func (s *RecipeSummary) ResolveImage(locator ImageLocator) {
	resolveImage(&s.ImageKey, &s.ImageUrl, locator)
}
