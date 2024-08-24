package slices

func BufAppend[S ~string | ~[]byte](buf *[]byte, s S) {
	*buf = append(*buf, s...)
}
