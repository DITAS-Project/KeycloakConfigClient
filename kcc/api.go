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

type KeyMessage struct {
	Algo string `json:"algorithm"`
	Key  string `json:"key"`
	CRC  uint32 `json:"crc"`
}

type BluePrint struct {
	BlueprintID string `json:"blueprintID"`
	ClientId    string `json:"clientId"`
	RedirectURI string `json:"defaultRedirectUri"`
}

type Config struct {
	BlueprintID string       `json:"blueprintID"`
	Roles       []string     `json:"roles"`
	Users       []UserConfig `json:"users"`
}

type UserConfig struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Roles    []string `json:"realmRoles"`
}

func (config *Config) splitByUser() []Config {
	configs := make([]Config, len(config.Users))
	for i, user := range config.Users {
		configs[i] = Config{
			BlueprintID: config.BlueprintID,
			Roles:       user.Roles,
			Users:       []UserConfig{user},
		}
	}
	return configs
}
