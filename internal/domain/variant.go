package domain

// Variant represents the structural variant of an architecture.
type Variant string

const (
	VariantClassic Variant = "classic"
	VariantModular Variant = "modular"
)

// VariantInfo holds human-readable metadata for a variant.
type VariantInfo struct {
	Value       Variant
	DisplayName string
	Description string
}

// variantInfos is the ordered list of all supported variants.
var variantInfos = []VariantInfo{
	{
		Value:       VariantClassic,
		DisplayName: "Classic",
		Description: "By-the-book — canonical layer naming, strict separation, true to the original pattern",
	},
	{
		Value:       VariantModular,
		DisplayName: "Modular",
		Description: "Business-module-first — each module encapsulates its own layers internally",
	},
}

// IsValid reports whether v is a recognized variant.
func (v Variant) IsValid() bool {
	for _, info := range variantInfos {
		if info.Value == v {
			return true
		}
	}
	return false
}

// Info returns the VariantInfo for v, and whether it was found.
func (v Variant) Info() (VariantInfo, bool) {
	for _, info := range variantInfos {
		if info.Value == v {
			return info, true
		}
	}
	return VariantInfo{}, false
}

// AllVariants returns a copy of the ordered slice of all variants.
func AllVariants() []VariantInfo {
	result := make([]VariantInfo, len(variantInfos))
	copy(result, variantInfos)
	return result
}
