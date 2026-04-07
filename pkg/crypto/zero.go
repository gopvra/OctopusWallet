package crypto

// ZeroBytes overwrites a byte slice with zeros to prevent key material
// from lingering in memory after use.
func ZeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
