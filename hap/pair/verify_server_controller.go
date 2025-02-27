package pair

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/justinkiang/hc/crypto"
	"github.com/justinkiang/hc/crypto/chacha20poly1305"
	"github.com/justinkiang/hc/db"
	"github.com/justinkiang/hc/hap"
	"github.com/justinkiang/hc/log"
	"github.com/justinkiang/hc/util"
)

// VerifyServerController verifies the stored client public key and negotiates a shared secret
// which is used encrypt the upcoming session.
//
// Verification fails when the client is not known, the public key for the client was not found,
// or the packet's seal could not be verified.
type VerifyServerController struct {
	database db.Database
	context  hap.Context
	session  *VerifySession
	step     VerifyStepType
}

// NewVerifyServerController returns a new verify server controller.
func NewVerifyServerController(database db.Database, context hap.Context) *VerifyServerController {
	controller := VerifyServerController{
		database: database,
		context:  context,
		session:  NewVerifySession(),
		step:     VerifyStepWaiting,
	}

	return &controller
}

// SharedKey returns the shared key which was negotiated with the client.
func (verify *VerifyServerController) SharedKey() [32]byte {
	return verify.session.SharedKey
}

// Handle processes a container to verify if a client is paired correctly.
func (verify *VerifyServerController) Handle(in util.Container) (util.Container, error) {
	var out util.Container
	var err error

	method := PairMethodType(in.GetByte(TagPairingMethod))

	// It is valid that method is not sent
	// If method is sent then it must be 0x00
	if method != PairingMethodDefault {
		return nil, errInvalidPairMethod(method)
	}

	seq := VerifyStepType(in.GetByte(TagSequence))

	switch seq {
	case VerifyStepStartRequest:
		if verify.step != VerifyStepWaiting {
			verify.reset()
			return nil, errInvalidInternalVerifyStep(verify.step)
		}
		out, err = verify.handlePairVerifyStart(in)
	case VerifyStepFinishRequest:
		defer verify.reset()
		if verify.step != VerifyStepStartResponse {
			return nil, errInvalidInternalVerifyStep(verify.step)
		}

		out, err = verify.handlePairVerifyFinish(in)
	default:
		return nil, errInvalidVerifyStep(seq)
	}

	return out, err
}

// Client -> Server
// - Public key `A`
//
// Server
// - Create public `B` and secret key `S` based on `A`

// Server -> Client
// - B: server public key
// - signature: from server session public key, server name, client session public key
func (verify *VerifyServerController) handlePairVerifyStart(in util.Container) (util.Container, error) {
	verify.step = VerifyStepStartResponse

	clientPublicKey := in.GetBytes(TagPublicKey)
	log.Debug.Println("->     A:", hex.EncodeToString(clientPublicKey))
	if len(clientPublicKey) != 32 {
		return nil, errInvalidClientKeyLength
	}

	var otherPublicKey [32]byte
	copy(otherPublicKey[:], clientPublicKey)

	verify.session.GenerateSharedKeyWithOtherPublicKey(otherPublicKey)
	verify.session.SetupEncryptionKey([]byte("Pair-Verify-Encrypt-Salt"), []byte("Pair-Verify-Encrypt-Info"))

	device := verify.context.GetSecuredDevice()
	var material []byte
	material = append(material, verify.session.PublicKey[:]...)
	material = append(material, device.Name()...)
	material = append(material, clientPublicKey...)
	signature, err := crypto.ED25519Signature(device.PrivateKey(), material)
	if err != nil {
		log.Info.Println(err)
		return nil, err
	}

	// Encrypt
	encryptedOut := util.NewTLV8Container()
	encryptedOut.SetString(TagUsername, device.Name())
	encryptedOut.SetBytes(TagSignature, signature)

	encryptedBytes, mac, _ := chacha20poly1305.EncryptAndSeal(verify.session.EncryptionKey[:], []byte("PV-Msg02"), encryptedOut.BytesBuffer().Bytes(), nil)

	out := util.NewTLV8Container()
	out.SetByte(TagSequence, verify.step.Byte())
	out.SetBytes(TagPublicKey, verify.session.PublicKey[:])
	out.SetBytes(TagEncryptedData, append(encryptedBytes, mac[:]...))

	log.Debug.Println("       K:", hex.EncodeToString(verify.session.EncryptionKey[:]))
	log.Debug.Println("       B:", hex.EncodeToString(verify.session.PublicKey[:]))
	log.Debug.Println("       S:", hex.EncodeToString(verify.session.PrivateKey[:]))
	log.Debug.Println("  Shared:", hex.EncodeToString(verify.session.SharedKey[:]))

	log.Debug.Println("<-     B:", hex.EncodeToString(out.GetBytes(TagPublicKey)))

	return out, nil
}

// Client -> Server
// - Encrypted tlv8: username and signature
//
// Server enrypty tlv8 and validates signature

// Server -> Client
// - only sequence number
// - error code (optional)
func (verify *VerifyServerController) handlePairVerifyFinish(in util.Container) (util.Container, error) {
	verify.step = VerifyStepFinishResponse

	data := in.GetBytes(TagEncryptedData)
	message := data[:(len(data) - 16)]
	var mac [16]byte
	copy(mac[:], data[len(message):]) // 16 byte (MAC)
	log.Debug.Println("->     Message:", hex.EncodeToString(message))
	log.Debug.Println("->     MAC:", hex.EncodeToString(mac[:]))

	decryptedBytes, err := chacha20poly1305.DecryptAndVerify(verify.session.EncryptionKey[:], []byte("PV-Msg03"), message, mac, nil)

	out := util.NewTLV8Container()
	out.SetByte(TagSequence, verify.step.Byte())

	if err != nil {
		verify.reset()
		log.Info.Panic(err)
		out.SetByte(TagErrCode, ErrCodeAuthenticationFailed.Byte()) // return error 2
	} else {
		in, err := util.NewTLV8ContainerFromReader(bytes.NewBuffer(decryptedBytes))
		if err != nil {
			return nil, err
		}

		username := in.GetString(TagUsername)
		signature := in.GetBytes(TagSignature)
		log.Debug.Println("    client:", username)
		log.Debug.Println(" signature:", hex.EncodeToString(signature))

		entity, err := verify.database.EntityWithName(username)
		if err != nil {
			return nil, fmt.Errorf("Client %s is unknown", username)
		}

		if len(entity.PublicKey) == 0 {
			return nil, fmt.Errorf("No LTPK available for client %s", username)
		}

		var material []byte
		material = append(material, verify.session.OtherPublicKey[:]...)
		material = append(material, []byte(username)...)
		material = append(material, verify.session.PublicKey[:]...)

		if crypto.ValidateED25519Signature(entity.PublicKey, material, signature) == false {
			log.Debug.Println("signature is invalid")
			verify.reset()
			out.SetByte(TagErrCode, ErrCodeUnknownPeer.Byte()) // return error 4
		} else {
			log.Debug.Println("signature is valid")
		}
	}

	return out, nil
}

func (verify *VerifyServerController) reset() {
	verify.step = VerifyStepWaiting
}
