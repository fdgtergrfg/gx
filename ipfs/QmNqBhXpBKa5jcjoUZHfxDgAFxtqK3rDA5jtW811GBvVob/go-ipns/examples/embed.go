package examples

import (
	"time"

	pb "gx/ipfs/QmNqBhXpBKa5jcjoUZHfxDgAFxtqK3rDA5jtW811GBvVob/go-ipns/pb"

	ipns "gx/ipfs/QmNqBhXpBKa5jcjoUZHfxDgAFxtqK3rDA5jtW811GBvVob/go-ipns"
	crypto "gx/ipfs/QmPvyPwuCgJ7pDmrKDxRtsScJgBaM5h4EpRL2qQJsmXf4n/go-libp2p-crypto"
)

// CreateEntryWithEmbed shows how you can create an IPNS entry
// and embed it with a public key. For ed25519 keys this is not needed
// so attempting to embed with an ed25519 key, will not actually embed the key
func CreateEntryWithEmbed(ipfsPath string, publicKey crypto.PubKey, privateKey crypto.PrivKey) (*pb.IpnsEntry, error) {
	ipfsPathByte := []byte(ipfsPath)
	eol := time.Now().Add(time.Hour * 48)
	entry, err := ipns.Create(privateKey, ipfsPathByte, 1, eol)
	if err != nil {
		return nil, err
	}
	err = ipns.EmbedPublicKey(publicKey, entry)
	if err != nil {
		return nil, err
	}
	return entry, nil
}
