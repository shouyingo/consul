package consul

import "crypto/rand"

func genuuidv4(u *[16]byte) {
	// https://tools.ietf.org/html/rfc4122.html#section-4.4
	rand.Read(u[:])
	// version
	u[6] = (u[6] & 0x0f) | 0x40
	// variant
	u[8] = (u[8] & 0x3f) | 0x80
}
