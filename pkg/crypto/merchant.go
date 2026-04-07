package crypto

// MerchantIDToIndex converts a UUID merchant ID to a uint32 index for HD derivation.
func MerchantIDToIndex(id string) uint32 {
	var sum uint32
	for _, b := range []byte(id) {
		sum = sum*31 + uint32(b)
	}
	return sum & 0x7FFFFFFF
}
