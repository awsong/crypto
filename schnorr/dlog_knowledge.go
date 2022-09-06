/*
 * Copyright 2017 XLAB d.o.o.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package schnorr

import (
	"fmt"
	"math/big"

	"github.com/awsong/crypto/common"
)

// Prover is a generalized Schnorr - while usually Schnorr proof is executed with one base,
// Prover for a given y enables proof of knowledge of secrets x_1,...,x_k
// such that y = g_1^x_1 * ... * g_k^x_k where g_i are given generators (bases) of Schnorr group.
// For a "normal" Schnorr just use bases and secrets arrays with only one element.
type Prover struct {
	Group      *Group
	secrets    []*big.Int
	bases      []*big.Int
	randomVals []*big.Int
	y          *big.Int
}

func NewProver(group *Group, secrets,
	bases []*big.Int, y *big.Int) (*Prover, error) {
	if len(secrets) != len(bases) {
		return nil, fmt.Errorf("number of secrets and representation bases shoud be the same")
	}

	return &Prover{
		Group:   group,
		secrets: secrets,
		bases:   bases,
		y:       y,
	}, nil
}

func (p *Prover) GetProofRandomData() *big.Int {
	// t = g_1^r_1 * ... * g_k^r_k where g_i are bases and r_i are random values
	t := big.NewInt(1)
	var randomVals = make([]*big.Int, len(p.bases))
	for i, _ := range randomVals {
		r := common.GetRandomInt(p.Group.Q)
		randomVals[i] = r
		f := p.Group.Exp(p.bases[i], r)
		t = p.Group.Mul(t, f)
	}
	p.randomVals = randomVals
	return t
}

func (p *Prover) GetProofData(challenge *big.Int) []*big.Int {
	// z_i = r_i + challenge * secrets[i]
	var proofData = make([]*big.Int, len(p.bases))
	for i, _ := range proofData {
		z_i := p.Group.Mul(challenge, p.secrets[i])
		z_i = p.Group.Add(z_i, p.randomVals[i])
		proofData[i] = z_i
	}
	return proofData
}

// Proof presents all three messages in sigma protocol - useful when challenge
// is generated by prover via Fiat-Shamir.
type Proof struct {
	ProofRandomData *big.Int
	Challenge       *big.Int
	ProofData       []*big.Int
}

func NewProof(proofRandomData, challenge *big.Int,
	proofData []*big.Int) *Proof {
	return &Proof{
		ProofRandomData: proofRandomData,
		Challenge:       challenge,
		ProofData:       proofData,
	}
}

type Verifier struct {
	Group           *Group
	bases           []*big.Int
	proofRandomData *big.Int
	y               *big.Int
	challenge       *big.Int
}

func NewVerifier(group *Group) *Verifier {
	return &Verifier{
		Group: group,
	}
}

// TODO: SetProofRandomData name is not ok - it is not only setting
// proofRandomData, but also bases and y.
// It might be split (a, b for example set in Verifier constructor).
func (v *Verifier) SetProofRandomData(proofRandomData *big.Int, bases []*big.Int,
	y *big.Int) {
	v.proofRandomData = proofRandomData
	v.bases = bases
	v.y = y
}

func (v *Verifier) GetChallenge() *big.Int {
	challenge := common.GetRandomInt(v.Group.Q)
	v.challenge = challenge
	return challenge
}

// SetChallenge is used when Fiat-Shamir is used - when challenge is generated using hash by the prover.
func (v *Verifier) SetChallenge(challenge *big.Int) {
	v.challenge = challenge
}

func (v *Verifier) Verify(proofData []*big.Int) bool {
	// check:
	// g_1^z_1 * ... * g_k^z_k = (g_1^x_1 * ... * g_k^x_k)^challenge * (g_1^r_1 * ... * g_k^r_k)
	left := big.NewInt(1)
	for i := 0; i < len(v.bases); i++ {
		t := v.Group.Exp(v.bases[i], proofData[i])
		left = v.Group.Mul(left, t)
	}

	right := v.Group.Exp(v.y, v.challenge)
	right = v.Group.Mul(right, v.proofRandomData)

	return left.Cmp(right) == 0
}
