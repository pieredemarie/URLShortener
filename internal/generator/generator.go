package generator

const alpabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const base = uint64(len(alpabet))

func Encode(id uint64) string {
	if id == 0 {
		return string(alpabet[0])
	}

	var encoded []byte
	for id > 0 {
		remainder := id % base
		id /= base
		encoded = append([]byte{alpabet[remainder]}, encoded...)
	}

	return string(encoded)
}
