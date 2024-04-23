// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package dtls

import (
	"context"
	"testing"
	"time"

	"github.com/adrian38/dtls/v2/internal/ciphersuite"
	"github.com/adrian38/dtls/v2/pkg/protocol/alert"
	"github.com/adrian38/dtls/v2/pkg/protocol/handshake"
	"github.com/pion/transport/v3/test"
)

type flight4TestMockFlightConn struct{}

func (f *flight4TestMockFlightConn) notify(context.Context, alert.Level, alert.Description) error {
	return nil
}
func (f *flight4TestMockFlightConn) writePackets(context.Context, []*packet) error { return nil }
func (f *flight4TestMockFlightConn) recvHandshake() <-chan chan struct{}           { return nil }
func (f *flight4TestMockFlightConn) setLocalEpoch(uint16)                          {}
func (f *flight4TestMockFlightConn) handleQueuedPackets(context.Context) error     { return nil }
func (f *flight4TestMockFlightConn) sessionKey() []byte                            { return nil }

type flight4TestMockCipherSuite struct {
	ciphersuite.TLSEcdheEcdsaWithAes128GcmSha256

	t *testing.T
}

func (f *flight4TestMockCipherSuite) IsInitialized() bool {
	f.t.Fatal("IsInitialized called with Certificate but not CertificateVerify")
	return true
}

// Assert that if a Client sends a certificate they
// must also send a CertificateVerify message.
// The flight4handler must not interact with the CipherSuite
// if the CertificateVerify is missing
func TestFlight4_Process_CertificateVerify(t *testing.T) {
	// Limit runtime in case of deadlocks
	lim := test.TimeOut(5 * time.Second)
	defer lim.Stop()

	// Check for leaking routines
	report := test.CheckRoutines(t)
	defer report()

	mockConn := &flight4TestMockFlightConn{}
	state := &State{
		cipherSuite: &flight4TestMockCipherSuite{t: t},
	}
	cache := newHandshakeCache()
	cfg := &handshakeConfig{}

	rawCertificate := []byte{
		0x0b, 0x00, 0x01, 0x9b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x01, 0x9b, 0x00, 0x01, 0x98, 0x00, 0x01, 0x95, 0x30, 0x82,
		0x01, 0x91, 0x30, 0x82, 0x01, 0x38, 0xa0, 0x03, 0x02, 0x01,
		0x02, 0x02, 0x11, 0x01, 0x65, 0x03, 0x3f, 0x4d, 0x0b, 0x9a,
		0x62, 0x91, 0xdb, 0x4d, 0x28, 0x2c, 0x1f, 0xd6, 0x73, 0x32,
		0x30, 0x0a, 0x06, 0x08, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x04,
		0x03, 0x02, 0x30, 0x00, 0x30, 0x1e, 0x17, 0x0d, 0x32, 0x32,
		0x30, 0x35, 0x31, 0x35, 0x31, 0x38, 0x34, 0x33, 0x35, 0x35,
		0x5a, 0x17, 0x0d, 0x32, 0x32, 0x30, 0x36, 0x31, 0x35, 0x31,
		0x38, 0x34, 0x33, 0x35, 0x35, 0x5a, 0x30, 0x00, 0x30, 0x59,
		0x30, 0x13, 0x06, 0x07, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x02,
		0x01, 0x06, 0x08, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x03, 0x01,
		0x07, 0x03, 0x42, 0x00, 0x04, 0xc3, 0xb7, 0x13, 0x1a, 0x0a,
		0xfc, 0xd0, 0x82, 0xf8, 0x94, 0x5e, 0xc0, 0x77, 0x07, 0x81,
		0x28, 0xc9, 0xcb, 0x08, 0x84, 0x50, 0x6b, 0xf0, 0x22, 0xe8,
		0x79, 0xb9, 0x15, 0x33, 0xc4, 0x56, 0xa1, 0xd3, 0x1b, 0x24,
		0xe3, 0x61, 0xbd, 0x4d, 0x65, 0x80, 0x6b, 0x5d, 0x96, 0x48,
		0xa2, 0x44, 0x9e, 0xce, 0xe8, 0x65, 0xd6, 0x3c, 0xe0, 0x9b,
		0x6b, 0xa1, 0x36, 0x34, 0xb2, 0x39, 0xe2, 0x03, 0x00, 0xa3,
		0x81, 0x92, 0x30, 0x81, 0x8f, 0x30, 0x0e, 0x06, 0x03, 0x55,
		0x1d, 0x0f, 0x01, 0x01, 0xff, 0x04, 0x04, 0x03, 0x02, 0x02,
		0xa4, 0x30, 0x1d, 0x06, 0x03, 0x55, 0x1d, 0x25, 0x04, 0x16,
		0x30, 0x14, 0x06, 0x08, 0x2b, 0x06, 0x01, 0x05, 0x05, 0x07,
		0x03, 0x02, 0x06, 0x08, 0x2b, 0x06, 0x01, 0x05, 0x05, 0x07,
		0x03, 0x01, 0x30, 0x0f, 0x06, 0x03, 0x55, 0x1d, 0x13, 0x01,
		0x01, 0xff, 0x04, 0x05, 0x30, 0x03, 0x01, 0x01, 0xff, 0x30,
		0x1d, 0x06, 0x03, 0x55, 0x1d, 0x0e, 0x04, 0x16, 0x04, 0x14,
		0xb1, 0x1a, 0xe3, 0xeb, 0x6f, 0x7c, 0xc3, 0x8f, 0xba, 0x6f,
		0x1c, 0xe8, 0xf0, 0x23, 0x08, 0x50, 0x8d, 0x3c, 0xea, 0x31,
		0x30, 0x2e, 0x06, 0x03, 0x55, 0x1d, 0x11, 0x01, 0x01, 0xff,
		0x04, 0x24, 0x30, 0x22, 0x82, 0x20, 0x30, 0x30, 0x30, 0x30,
		0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
		0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
		0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x0a,
		0x06, 0x08, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x04, 0x03, 0x02,
		0x03, 0x47, 0x00, 0x30, 0x44, 0x02, 0x20, 0x06, 0x31, 0x43,
		0xac, 0x03, 0x45, 0x79, 0x3c, 0xd7, 0x5f, 0x6e, 0x6a, 0xf8,
		0x0e, 0xfd, 0x35, 0x49, 0xee, 0x1b, 0xbc, 0x47, 0xce, 0xe3,
		0x39, 0xec, 0xe4, 0x62, 0xe1, 0x30, 0x1a, 0xa1, 0x89, 0x02,
		0x20, 0x35, 0xcd, 0x7a, 0x15, 0x68, 0x09, 0x50, 0x49, 0x9e,
		0x3e, 0x05, 0xd7, 0xc2, 0x69, 0x3f, 0x9c, 0x0c, 0x98, 0x92,
		0x65, 0xec, 0xae, 0x44, 0xfe, 0xe5, 0x68, 0xb8, 0x09, 0x78,
		0x7f, 0x6b, 0x77,
	}

	rawClientKeyExchange := []byte{
		0x10, 0x00, 0x00, 0x21, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x21, 0x20, 0x96, 0xed, 0x0c, 0xee, 0xf3, 0x11, 0xb1,
		0x9d, 0x8b, 0x1c, 0x02, 0x7f, 0x06, 0x7c, 0x57, 0x7a, 0x14,
		0xa6, 0x41, 0xde, 0x63, 0x57, 0x9e, 0xcd, 0x34, 0x54, 0xba,
		0x37, 0x4d, 0x34, 0x15, 0x18,
	}

	cache.push(rawCertificate, 0, 0, handshake.TypeCertificate, true)
	cache.push(rawClientKeyExchange, 0, 1, handshake.TypeClientKeyExchange, true)

	if _, _, err := flight4Parse(context.TODO(), mockConn, state, cache, cfg); err != nil {
		t.Fatal(err)
	}
}
