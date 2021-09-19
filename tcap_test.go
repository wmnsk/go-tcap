package tcap

import (
	"encoding/hex"
	"log"
	"math/rand"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	hexforUSSDRequest_long_length_impl := "6281f248040000000a6b3f283d060700118605010101a032603080020780a109060704000001001302be1f281d060704000001010101a012a01080069121436587f981069121436587f96c81a8a181a502010002013b30819c04010f04818c5474d8bd06e5df7590f92d07d5e769f71944479741c7373bec3e83aad32911342fcbed65b90b94a683d27310b96c2fb3dff0321904afcbcbec3ce8ed069ddfecb0fb0c0abbc9a0341d549f97e7a0e9700865819ab36a90059a0ea95016283875b24043e194053a4e9b3750d84d0625a955d09c1e7693c372f2dc0542bee16550fe5d0795ddea771e4447bbf1800891111111111111f1"
	UssdRequestBytes_long_length_impl, err := hex.DecodeString(hexforUSSDRequest_long_length_impl)
	if err != nil {
		panic(err)
	}

	hexforUSSDRequest_short_length_impl := "62664804000000026b3f283d060700118605010101a032603080020780a109060704000001001302be1f281d060704000001010101a012a01080069121436587f981069121436587f96c1da11b02010002013b301304010f0404f4f29c0e800891111111111111f1"
	UssdRequestBytes_short_length_impl, err := hex.DecodeString(hexforUSSDRequest_short_length_impl)
	if err != nil {
		panic(err)
	}

	hex_for_ussd_gsm_7_160_chars := "5474D8BD06E5DF7590F92D07D5E769F71944479741C7373BEC3E83AAD32911342FCBED65B90B94A683D27310B96C2FB3DFF0321904AFCBCBEC3CE8ED069DDFECB0FB0C0ABBC9A0341D549F97E7A0E9700865819AB36A90059A0EA95016283875B24043E194053A4E9B3750D84D0625A955D09C1E7693C372F2DC0542BEE16550FE5D0795DDEA771E4447BBF1"
	bytes_ussd_gsm_7_160_chars, err := hex.DecodeString(hex_for_ussd_gsm_7_160_chars)
	if err != nil {
		panic(err)
	}

	customGenerateBytes := GenerateNewEndReturnResultWithDialogue(60)
	log.Printf("customGenerateBytes: %x", customGenerateBytes)
	customGenerateBytes2 := GenerateNewEndReturnResultWithDialogue(90)
	log.Printf("customGenerateBytes2: %x", customGenerateBytes2)
	customGenerateBytes3 := GenerateNewEndReturnResultWithDialogueWithMessage(bytes_ussd_gsm_7_160_chars)
	log.Printf("customGenerateBytes3: %x", customGenerateBytes3)

	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: string("Testing Long Length USSD Packet"), args: struct{ b []byte }{b: UssdRequestBytes_long_length_impl} , want: "networkUnstructuredSsContext" , wantErr: false },
		{name: string("Testing Short Length USSD Packet"), args: struct{ b []byte }{b: UssdRequestBytes_short_length_impl} , want: "networkUnstructuredSsContext" , wantErr: false },
		{name: string("Testing Short Length Generated USSD Packet"), args: struct{ b []byte }{b: customGenerateBytes} , want: "networkUnstructuredSsContext" , wantErr: false },
		{name: string("Testing Long Length Generated USSD Packet"), args: struct{ b []byte }{b: customGenerateBytes2} , want: "networkUnstructuredSsContext" , wantErr: false },
		{name: string("Testing Long Length Generated USSD Packet GSM string"), args: struct{ b []byte }{b: customGenerateBytes3} , want: "networkUnstructuredSsContext" , wantErr: false },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Dialogue.Context(), tt.want) {
				t.Errorf("Parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func GenerateNewEndReturnResultWithDialogue(paramlength int) []byte{
	tc2 := NewIE(0x04, []byte(RandStringBytes(paramlength)))
	tc := NewIE(0x04, []byte{0x0f})
	ik, _ := tc.MarshalBinary()
	ik2, _ := tc2.MarshalBinary()
	sample := NewEndReturnResultWithDialogue(uint32(rand.Intn(len(letterBytes))), AARE, NetworkUnstructuredSsContext, 2, 0, 59, true, append(ik, ik2...))
	da, _ := sample.MarshalBinary()
	return da
}
func GenerateNewEndReturnResultWithDialogueWithMessage(message []byte) []byte{
	tc2 := NewIE(0x04, message)
	tc := NewIE(0x04, []byte{0x0f})
	ik, _ := tc.MarshalBinary()
	ik2, _ := tc2.MarshalBinary()
	sample := NewEndReturnResultWithDialogue(uint32(rand.Intn(len(letterBytes))), AARE, NetworkUnstructuredSsContext, 2, 0, 59, true, append(ik, ik2...))
	da, _ := sample.MarshalBinary()
	return da
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}