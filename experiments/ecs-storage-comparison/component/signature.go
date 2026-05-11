package component

type Signature uint64

func (s *Signature) Set(cid ID) {
	*s |= 1 << (cid - 1)
}

func (s Signature) Has(cid ID) bool {
	return s&(1<<(cid-1)) != 0
}

func (s Signature) Contains(other Signature) bool {
	return s&other == other
}

func SignatureOf(cids []ID) Signature {
	var sig Signature
	for _, cid := range cids {
		sig.Set(cid)
	}
	return sig
}
