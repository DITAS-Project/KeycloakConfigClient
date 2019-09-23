/*
 * Copyright 2018 Information Systems Engineering, TU Berlin, Germany
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *                       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * This is being developed for the DITAS Project: https://www.ditas-project.eu/
 */
package kcc

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/log"
)

type ConfigClient struct {
	endpoint string
	key      *rsa.PublicKey
}

func NewKCC(endpoint string) (client *ConfigClient, err error) {
	client = &ConfigClient{}

	key, err := getKey(endpoint)

	if err != nil {
		return nil, fmt.Errorf("failed to get key %+v\n", err)
	}

	client.endpoint = endpoint
	client.key = key

	return client, err
}

func getKey(address string) (*rsa.PublicKey, error) {
	resp, err := http.Get(address + "/v1/keys")

	if err != nil {
		log.Debug("failed to get key %+v\n", err)
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Debug("failed to get key %+v\n", err)
		return nil, err
	}

	msg := &KeyMessage{}

	err = json.Unmarshal(data, msg)
	if err != nil {
		log.Debug("failed to read key %+v\n", err)
		return nil, err
	}

	decoded, err := base64.StdEncoding.DecodeString(msg.Key)

	crc32q := crc32.MakeTable(0xedb88320)
	keyCheck := crc32.Checksum(decoded, crc32q)
	if msg.CRC != keyCheck {
		return nil, fmt.Errorf("key checksum is false %d == %d\n", msg.CRC, keyCheck)
	} else {
		log.Debug("crc is %d == %d\n", msg.CRC, keyCheck)
	}

	if err != nil {
		log.Debug("failed to get decode %+v\n", err)
		return nil, err
	}

	return bytesToPublicKey(decoded)
}

// bytesToPublicKey bytes to public key
func bytesToPublicKey(pub []byte) (*rsa.PublicKey, error) {
	ifc, err := x509.ParsePKIXPublicKey(pub)
	if err != nil {
		return nil, fmt.Errorf("%+v", err)
	}
	key, ok := ifc.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not ok")
	}
	return key, nil
}

func encryptWithPublicKey(msg []byte, pub *rsa.PublicKey) ([]byte, error) {
	hash := sha256.New()

	text, err := rsa.EncryptOAEP(hash, rand.Reader, pub, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("&+v", err)
	}
	return text, nil
}

func (client *ConfigClient) SendConfig(config Config) error {

	data, err := json.Marshal(config)
	if err != nil {
		log.Error("failed to marshal config")
		return err
	}
	if len(data) > 256 {

		//TODO: messge to big, need to split up!
		configs := config.splitByUser()
		for _, cnf := range configs {
			data, err := json.Marshal(cnf)
			if err != nil {
				log.Error("failed to marshal config")
				return err
			}
			err = client.sendConfig(data, cnf)
			if err != nil {
				return err
			}

		}
	} else {
		return client.sendConfig(data, config)
	}

	return nil
}

func (client *ConfigClient) sendConfig(data []byte, config Config) error {
	msg, err := encryptWithPublicKey(data, client.key)
	if err != nil {
		log.Error("failed to encrypt config")
		return err
	}

	resp, _ := http.Post(client.endpoint+"/v1/"+config.BlueprintID, "plain/text", strings.NewReader(base64.StdEncoding.EncodeToString(msg)))
	data, _ = ioutil.ReadAll(resp.Body)

	if resp.StatusCode > 200 {
		return fmt.Errorf("failed to upadte config %s\n", string(data))
	}
	log.Debug("service response is %s", string(data))

	return nil
}

func (client *ConfigClient) SendBlueprint(blueprint BluePrint) error {
	data, err := json.Marshal(blueprint)
	if err != nil {
		log.Error("failed to marshal blueprint config")
		return err
	}
	resp, err := http.Post(client.endpoint+"/v1/init", "plain/text", bytes.NewReader(data))
	if err != nil {
		log.Error("failed to send blueprint request")
		return err
	}

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode > 300 {
		return fmt.Errorf("failed to upload blueprint %+v", string(data))
	}
	log.Debug("service response is %s", string(data))
	return nil
}
