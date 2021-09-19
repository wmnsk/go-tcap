package tcap

import (
	"bytes"
)

func handleMarshalLen(elementLength uint8, len int) int {
	if(elementLength > 127){
		return len + 3
	} else {
		return len + 2
	}
}

/*
ReadLength to read the length based on the ASN1 Implementation. the first byte of the length indicates if it is long or short
 */
func readLength(b []byte) (uint8, int) {
	var length int
	r := bytes.NewReader(b[1:])
	lengthByte, _ := r.ReadByte()
	if((lengthByte & 128) == 0){
		return uint8(int(lengthByte)) , 2
	} else {
		lengthByte = (lengthByte & 127)
		if(lengthByte == 0){
			return 0 , 2
		} else {
			for i := 0; i < int(lengthByte); i++ {
				tmp, _ := r.ReadByte()
				length = int(byte(length) << 8 | 255 & tmp)
			}
			//length++
			return uint8(length) , 3
		}
	}
}

/*
WriteLength to read the length based on the ASN1 Implementation. the first byte of the length indicates if it is long or short
*/
func writeLength(b []byte, length uint8) int {
	var offset int = 2
	if(length <= 127 ){
		b[1] = byte(length)
		return offset
	} else {
		buf := make([]byte, 4)
		length = length - 1
		var count int
		if (int64(length) & int64(-16777216)) > 0 {
			buf[0] = byte(length >> 24 & 255)
			buf[1] = byte(length >> 16 & 255)
			buf[2] = byte(length >> 8 & 255)
			buf[3] = byte(length & 255)
			count = 4
		} else if (int64(length) & 16711680) > 0 {
			buf[0] = byte(length >> 16 & 255)
			buf[1] = byte(length >> 8 & 255)
			buf[2] = byte(length & 255)
			count = 3

		} else if (int64(length) & 65280) > 0 {
			buf[0] = byte(length >> 8 & 255)
			buf[1] = byte(length & 255)
			count = 2
		} else {
			buf[0] = byte(length & 255)
			count = 1
		}
		b[offset-1] = byte(128 | count)
		for i := 0; i < count; i++ {
			b[offset+i] = buf[i]
		}
		offset = offset + count
		return offset
	}
}
